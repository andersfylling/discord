package disgord

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/andersfylling/disgord/logger"

	"github.com/andersfylling/snowflake/v3"

	"github.com/andersfylling/disgord/constant"
	"github.com/andersfylling/disgord/websocket"
	"golang.org/x/net/proxy"

	"github.com/andersfylling/disgord/event"
	"github.com/andersfylling/disgord/httd"
)

// NewRESTClient creates a client for sending and handling Discord protocols such as rate limiting
func NewRESTClient(conf *Config) (*httd.Client, error) {
	return httd.NewClient(&httd.Config{
		APIVersion:                   constant.DiscordVersion,
		BotToken:                     conf.BotToken,
		UserAgentSourceURL:           constant.GitHubURL,
		UserAgentVersion:             constant.Version,
		HTTPClient:                   conf.HTTPClient,
		CancelRequestWhenRateLimited: conf.CancelRequestWhenRateLimited,
	})
}

// New create a client. But panics on configuration/setup errors.
func New(conf *Config) (c *client) {
	var err error
	if c, err = NewClient(conf); err != nil {
		panic(err)
	}

	return c

}

// NewClient creates a new DisGord client and returns an error on configuration issues
func NewClient(conf *Config) (c *client, err error) {
	if conf.HTTPClient == nil {
		conf.HTTPClient = &http.Client{
			Timeout: time.Second * 10,
		}
	}
	if conf.Proxy != nil {
		conf.HTTPClient.Transport = &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (conn net.Conn, e error) {
				return conf.Proxy.Dial(network, addr)
			},
		}
	}

	if conf.ProjectName == "" {
		conf.ProjectName = LibraryInfo()
	}

	conf.shutdownChan = make(chan interface{})

	if conf.Logger == nil {
		conf.Logger = logger.Empty{}
	}

	// request client for REST requests
	reqClient, err := NewRESTClient(conf)
	if err != nil {
		return nil, err
	}

	if conf.WSShardManagerConfig == nil {
		conf.WSShardManagerConfig = &WSShardManagerConfig{}
	}
	if conf.WSShardManagerConfig.ShardRateLimit == 0 {
		conf.WSShardManagerConfig.ShardRateLimit = DefaultShardRateLimit
	}
	sharding := NewShardManager(conf)

	// caching
	var cacher *Cache
	if !conf.DisableCache {
		if conf.CacheConfig == nil {
			conf.CacheConfig = DefaultCacheConfig()
		} else {
			ensureBasicCacheConfig(conf.CacheConfig)
		}
		cacher, err = newCache(conf.CacheConfig)
		if err != nil {
			return nil, err
		}

		// register for events for activate caches
		if !conf.CacheConfig.DisableUserCaching {
			sharding.TrackEvent.Add(event.Ready)
			sharding.TrackEvent.Add(event.UserUpdate)
		}
		if !conf.CacheConfig.DisableChannelCaching {
			sharding.TrackEvent.Add(event.ChannelCreate)
			sharding.TrackEvent.Add(event.ChannelUpdate)
			sharding.TrackEvent.Add(event.ChannelPinsUpdate)
			sharding.TrackEvent.Add(event.ChannelDelete)
		}
		if !conf.CacheConfig.DisableGuildCaching {
			sharding.TrackEvent.Add(event.GuildCreate)
			sharding.TrackEvent.Add(event.GuildDelete)
			sharding.TrackEvent.Add(event.GuildUpdate)
			sharding.TrackEvent.Add(event.GuildEmojisUpdate)
			sharding.TrackEvent.Add(event.GuildMemberAdd)
			sharding.TrackEvent.Add(event.GuildMemberRemove)
			sharding.TrackEvent.Add(event.GuildMembersChunk)
			sharding.TrackEvent.Add(event.GuildMemberUpdate)
			sharding.TrackEvent.Add(event.GuildRoleCreate)
			sharding.TrackEvent.Add(event.GuildRoleDelete)
			sharding.TrackEvent.Add(event.GuildRoleUpdate)
			sharding.TrackEvent.Add(event.GuildIntegrationsUpdate)
		}
	} else {
		// create an empty cache to avoid nil panics
		cacher, err = newCache(&CacheConfig{
			DisableUserCaching:       true,
			DisableChannelCaching:    true,
			DisableGuildCaching:      true,
			DisableVoiceStateCaching: true,
		})
		if err != nil {
			return nil, err
		}
	}

	// Required for voice operation
	sharding.TrackEvent.Add(event.VoiceStateUpdate)
	sharding.TrackEvent.Add(event.VoiceServerUpdate)

	// event dispatcher
	eventChanSize := 20
	evtDemultiplexer := newEvtDemultiplexer(conf.ActivateEventChannels, eventChanSize)

	// create a disgord client/instance/session
	c = &client{
		shutdownChan:     conf.shutdownChan,
		config:           conf,
		shardManager:     sharding,
		httpClient:       conf.HTTPClient,
		Proxy:            conf.Proxy,
		botToken:         conf.BotToken,
		evtDemultiplexer: evtDemultiplexer,
		req:              reqClient,
		cache:            cacher,
		log:              conf.Logger,
		pool:             newPools(),
	}
	c.evtDemultiplexer.session = c
	c.voiceRepository = newVoiceRepository(c)
	sharding.client = c

	return c, err
}

// Config Configuration for the DisGord client
type Config struct {
	BotToken   string
	HTTPClient *http.Client
	Proxy      proxy.Dialer

	CancelRequestWhenRateLimited bool

	DisableCache         bool
	CacheConfig          *CacheConfig
	WSShardManagerConfig *WSShardManagerConfig
	Presence             *UpdateStatusCommand

	//ImmutableCache bool

	//LoadAllMembers   bool
	//LoadAllChannels  bool
	//LoadAllRoles     bool
	//LoadAllPresences bool

	// for cancellation
	shutdownChan chan interface{}

	// your project name, name of bot, or application
	ProjectName string

	// ActivateEventChannels signifies that the developer will use channels to handle incoming events. May it be
	// in addition to handlers or not. This forces the use of a scheduler to empty the buffered channels when they
	// reach their capacity. Since it requires extra resources, others who have no interest in utilizing channels
	// should not experience any performance penalty (even though it might be unnoticeable).
	ActivateEventChannels bool

	// Logger is a dependency that must be injected to support logging.
	// disgord.DefaultLogger() can be used
	Logger Logger
}

// client is the main disgord client to hold your state and data. You must always initiate it using the constructor
// methods (eg. New(..) or NewClient(..)).
//
// Note that this client holds all the REST methods, and is split across files, into whatever category
// the REST methods regards.
type client struct {
	sync.RWMutex

	shutdownChan chan interface{}
	config       *Config
	botToken     string

	myID        Snowflake
	permissions int

	// reactor demultiplexer for events
	evtDemultiplexer *evtDemultiplexer

	// cancelRequestWhenRateLimited by default the client waits until either the HTTPClient.timeout or
	// the rate limit ends before closing a request channel. If activated, in stead, requests will
	// instantly be denied, and the process ended with a rate limited error.
	cancelRequestWhenRateLimited bool

	// req holds the rate limiting logic and error parsing unique for Discord
	req *httd.Client

	// http client used for connections
	httpClient *http.Client
	Proxy      proxy.Dialer

	shardManager *WSShardManager

	cache *Cache

	log Logger

	// voice
	*voiceRepository

	// pools
	pool *pools
}

//////////////////////////////////////////////////////
//
// COMPLIANCE'S / IMPLEMENTATIONS
//
//////////////////////////////////////////////////////
var _ fmt.Stringer = (*client)(nil)
var _ Session = (*client)(nil)
var _ Link = (*client)(nil)

//////////////////////////////////////////////////////
//
// METHODS
//
//////////////////////////////////////////////////////

func (c *client) Pool() *pools {
	return c.pool
}

// AddPermission adds a minimum required permission to the bot. If the permission is negative, it is overwritten to 0.
// This is useful for creating the bot URL.
//
// At the moment, this holds no other effect than aesthetics.
func (c *client) AddPermission(permission int) (updatedPermissions int) {
	if permission < 0 {
		permission = 0
	}

	c.permissions |= permission
	return c.GetPermissions()
}

// GetPermissions returns the minimum bot requirements.
func (c *client) GetPermissions() (permissions int) {
	return c.permissions
}

// CreateBotURL creates a URL that can be used to invite this bot to a guild/server.
// Note that it depends on the bot ID to be after the Discord update where the client ID
// is the same as the Bot ID.
//
// By default the permissions will be 0, as in none. If you want to add/set the minimum required permissions
// for your bot to run successfully, you should utilise
//  client.
func (c *client) CreateBotURL() (u string, err error) {
	_, _ = c.GetCurrentUser() // update c.myID

	if c.myID.Empty() {
		err = errors.New("unable to get bot id")
		return "", err
	}

	// make sure the snowflake is new enough to be used as a client ID
	t, err := time.Parse("2006-01-02 15:04:05", "2016-08-07 05:39:21.906")
	if err != nil {
		return "", err
	}

	loc, _ := time.LoadLocation("America/Los_Angeles")
	t = t.In(loc)

	if !c.myID.Date().After(t) {
		err = errors.New("the bot was not created after " + t.String() + " and can therefore not use the bot ID to generate a invite link")
		return "", err
	}

	format := "https://discordapp.com/oauth2/authorize?scope=bot&client_id=%s&permissions=%d"
	u = fmt.Sprintf(format, c.myID.String(), c.permissions)
	return u, nil
}

// HeartbeatLatency checks the duration of waiting before receiving a response from Discord when a
// heartbeat packet was sent. Note that heartbeats are usually sent around once a minute and is not a accurate
// way to measure delay between the client and Discord server
func (c *client) HeartbeatLatency() (duration time.Duration, err error) {
	return c.shardManager.GetAvgHeartbeatLatency()
}

// Myself get the current user / connected user
// Deprecated: use GetCurrentUser instead
func (c *client) Myself() (user *User, err error) {
	return c.GetCurrentUser()
}

// GetConnectedGuilds get a list over guild IDs that this client is "connected to"; or have joined through the ws connection. This will always hold the different Guild IDs, while the GetGuilds or GetCurrentUserGuilds might be affected by cache configuration.
func (c *client) GetConnectedGuilds() []snowflake.ID {
	c.shardManager.RLock()
	defer c.shardManager.RUnlock()

	var guilds []snowflake.ID
	for i := range c.shardManager.shards {
		guilds = append(guilds, c.shardManager.shards[i].guilds...)
	}

	return guilds
}

// Logger returns the log instance of DisGord.
// Note that this instance is never nil. When the conf.Logger is not assigned
// an empty struct is used instead. Such that all calls are simply discarded at compile time
// removing the need for nil checks.
func (c *client) Logger() logger.Logger {
	return c.log
}

func (c *client) String() string {
	return LibraryInfo()
}

// RateLimiter return the rate limiter object
func (c *client) RateLimiter() httd.RateLimiter {
	return c.req.RateLimiter()
}

// Req return the request object. Used in REST requests to handle rate limits,
// wrong http responses, etc.
func (c *client) Req() httd.Requester {
	return c.req
}

// Cache returns the cacheLink manager for the session
func (c *client) Cache() Cacher {
	return c.cache
}

//////////////////////////////////////////////////////
//
// Socket connection
//
//////////////////////////////////////////////////////

func (c *client) setupConnectEnv() {
	// set the user ID upon connection
	// only works with socket logic
	c.On(event.UserUpdate, c.handlerUpdateSelfBot)
	c.On(event.GuildCreate, c.handlerAddToConnectedGuilds)
	c.On(event.GuildDelete, c.handlerRemoveFromConnectedGuilds)

	// setup event observer
	go c.eventHandler()
	//c.evtDispatch.start()
}

// Connect establishes a websocket connection to the discord API
func (c *client) Connect() (err error) {
	// set the user ID upon connection
	// only works for socketing
	//
	// also verifies that the correct credentials were supplied
	var me *User
	if me, err = c.GetCurrentUser(); err != nil {
		return err
	}
	c.myID = me.ID

	url, shardCount, err := c.shardManager.GetConnectionDetails(c.req)
	if err != nil {
		return err
	}

	if c.config.WSShardManagerConfig.URL == "" {
		c.config.WSShardManagerConfig.URL = url
	}
	if c.config.WSShardManagerConfig.ShardLimit == 0 {
		c.config.WSShardManagerConfig.ShardLimit = shardCount
	}

	_ = c.shardManager.Prepare(c.config)
	c.setupConnectEnv() // calling this before the c.ShardManager.Prepare will cause a evtChan deadlock

	c.log.Info("Connecting to discord Gateway")
	if err = c.shardManager.Connect(); err != nil {
		c.log.Info(err)
		return err
	}

	c.log.Info("Connected")
	return nil
}

// Disconnect closes the discord websocket connection
func (c *client) Disconnect() (err error) {
	fmt.Println() // to keep ^C on it's own line
	c.log.Info("Closing Discord gateway connection")
	close(c.evtDemultiplexer.shutdown)
	if err = c.shardManager.Disconnect(); err != nil {
		c.log.Error(err)
		return err
	}
	close(c.shutdownChan)
	c.log.Info("Disconnected")

	return nil
}

// Suspend in case you want to temporary disconnect from the Gateway. But plan on
// connecting again without restarting your software/application, this should be used.
func (c *client) Suspend() (err error) {
	c.log.Info("Closing Discord gateway connection")
	if err = c.shardManager.Disconnect(); err != nil {
		return err
	}
	c.log.Info("Suspended")

	return nil
}

// DisconnectOnInterrupt wait until a termination signal is detected
func (c *client) DisconnectOnInterrupt() (err error) {
	// create a channel to listen for termination signals (graceful shutdown)
	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-termSignal

	return c.Disconnect()
}

// StayConnectedUntilInterrupted is a simple wrapper for connect, and disconnect that listens for system interrupts.
// When a error happens you can terminate the application without worries.
func (c *client) StayConnectedUntilInterrupted() (err error) {
	if err = c.Connect(); err != nil {
		c.log.Error(err)
		return err
	}

	if err = c.DisconnectOnInterrupt(); err != nil {
		c.log.Error(err)
		return err
	}

	return nil
}

//////////////////////////////////////////////////////
//
// Internal event handlers
//
//////////////////////////////////////////////////////

// handlerGuildDelete update internal state when joining or creating a guild
func (c *client) handlerAddToConnectedGuilds(_ Session, evt *GuildCreate) {
	// NOTE: during unit tests, you must remember that shards are usually added dynamically at runtime
	//  meaning, you might have to add your own shards if you get a panic here
	shard, _ := c.shardManager.GetShard(evt.Guild.ID)
	shard.Lock()
	defer shard.Unlock()

	// don't add an entry if there already is one
	for i := range shard.guilds {
		if shard.guilds[i] == evt.Guild.ID {
			return
		}
	}
	shard.guilds = append(shard.guilds, evt.Guild.ID)
}

// handlerGuildDelete update internal state when deleting or leaving a guild
func (c *client) handlerRemoveFromConnectedGuilds(_ Session, evt *GuildDelete) {
	// NOTE: during unit tests, you must remember that shards are usually added dynamically at runtime
	//  meaning, you might have to add your own shards if you get a panic here
	shard, _ := c.shardManager.GetShard(evt.UnavailableGuild.ID)
	shard.Lock()
	defer shard.Unlock()

	for i := range shard.guilds {
		if shard.guilds[i] != evt.UnavailableGuild.ID {
			continue
		}
		shard.guilds[i] = shard.guilds[len(shard.guilds)-1]
		shard.guilds = shard.guilds[:len(shard.guilds)-1]
	}
}

func (c *client) handlerUpdateSelfBot(_ Session, update *UserUpdate) {
	_ = c.cache.Update(UserCache, update.User)
}

//////////////////////////////////////////////////////
//
// Socket utilities
//
//////////////////////////////////////////////////////

// Ready triggers a given callback when all shards has gotten their first Ready event
// Warning: if you run client.Disconnect and want to run Connect again later, this will
//  not work. The callback will be triggered instantly, as all the shards have already
//  successfully connected once.
// TODO: allow this method to be reused (see Warning paragraph)
func (c *client) Ready(cb func()) {
	// TODO: optimize..
	go func() {
		for {
			<-time.After(1 * time.Second)
			if c.shardManager.InitialReadyReceived() {
				break
			}
		}
		cb()
	}()
}

// On binds a singular or multiple event handlers to the stated event, with the same middlewares.
//  On("MESSAGE_CREATE", mdlwHasMentions, handleMsgsWithMentions)
// the mdlwHasMentions is optional, as are any middleware and controller. But a handler must be specified.
//  On("MESSAGE_CREATE", mdlwHasMentions, handleMsgsWithMentions, &ctrl{remaining:3})
// here ctrl is a custom struct that implements disgord.HandlerCtrl, and after 3 calls it is no longer callable.
//
//  On("MESSAGE_CREATE", mdlwHasMentions, handleMsgsWithMentions, saveToDB, &ctrl{remaining:3})
// the On statement takes all it's parameters as a unique definition. These are not two separate
// handlers. But rather, two handler run in sequence after the middleware (3 times).
//
// Another example is to create a voting system where you specify a deadline instead of a remaining counter:
//  On("MESSAGE_CREATE", mdlwHasMentions, handleMsgsWithMentions, saveToDB, &ctrl{deadline:time.Now().Add(time.Hour)})
// again, you specify the IsDead() method to comply with the disgord.HandlerCtrl interface, so you can do whatever
// you want.
//
// If the HandlerCtrl.OnInsert returns an error, the related handlers are still added to the demultiplexer reactor.
// But the error is logged to the injected logger instance (.Error).
//
// This ctrl feature was inspired by https://github.com/discordjs/discord.js
func (c *client) On(event string, inputs ...interface{}) {
	if err := ValidateHandlerInputs(inputs...); err != nil {
		panic(err)
	}
	c.shardManager.TrackEvent.Add(event)

	// detect middleware then handlers. Ordering is important.
	spec := &handlerSpec{}
	if err := spec.populate(inputs...); err != nil { // TODO: improve redundant checking
		panic(err) // if the pattern is wrong: (event,[ ...middlewares,] ...handlers[, controller])
		// if you want to error check before you use the .On, you can use disgord.ValidateHandlerInputs(...)
	}

	c.evtDemultiplexer.Lock()
	c.evtDemultiplexer.handlers[event] = append(c.evtDemultiplexer.handlers[event], spec)
	if err := spec.ctrl.OnInsert(c); err != nil {
		c.log.Error(err)
	}
	c.evtDemultiplexer.Unlock()
}

// Emit sends a socket command directly to Discord.
func (c *client) Emit(command SocketCommand, data interface{}) error {
	switch command {
	case CommandUpdateStatus, CommandUpdateVoiceState, CommandRequestGuildMembers:
	default:
		return errors.New("command is not supported")
	}
	return c.shardManager.Emit(command, data)
}

// EventChan get a event channel using the event name
func (c *client) EventChan(event string) (channel interface{}, err error) {
	return c.evtDemultiplexer.EventChan(event)
}

// EventChannels get access to all the event channels
func (c *client) EventChannels() (channels EventChannels) {
	return c.evtDemultiplexer
}

// AcceptEvent only events registered using this method is accepted from the Discord socket API. The rest is discarded
// to reduce unnecessary marshalling and controls.
func (c *client) AcceptEvent(events ...string) {
	for _, evt := range events {
		c.shardManager.TrackEvent.Add(evt)
	}
}

//////////////////////////////////////////////////////
//
// Abstract CRUD operations
//
//////////////////////////////////////////////////////

// DeleteFromDiscord if the given object has implemented the private interface discordDeleter this method can
// be used to delete said object.
func (c *client) DeleteFromDiscord(obj discordDeleter) (err error) {
	err = obj.deleteFromDiscord(c)
	return
}

// SaveToDiscord saves an object to the Discord servers. This supports creation of new objects. Given two arguments,
// the original object before changes (that reflects the most recent known discord state) and the new object that
// represents what the state should look like, the changes can be determined by a basic diff check and let DisGord
// try to udpating/modify the object at Discord.
//
// client.SaveToDiscord(object) -> saves an entirely new object (must not have a discord id)
// client.SaveToDiscord(original, new) -> updates an existing object (must have a discord id)
// TODO: flag support
func (c *client) SaveToDiscord(original discordSaver, changes ...discordSaver) (err error) {
	if original == nil {
		return errors.New("you must specify at least one discord object to be saved to Discord")
	}

	var updated discordSaver
	if len(changes) > 1 {
		updated = changes[1]
	}

	err = original.saveToDiscord(c, updated)
	return
}

//////////////////////////////////////////////////////
//
// REST Methods
// customs
//
//////////////////////////////////////////////////////

func (c *client) GetGuilds(params *GetCurrentUserGuildsParams, flags ...Flag) ([]*Guild, error) {
	// TODO: populate these partial guild objects
	return c.GetCurrentUserGuilds(params)
}

// SendMsg Input anything and it will be converted to a message and sent. If you
// supply it with multiple data's, it will simply merge them. Even if they are multiple Message objects.
// However, if you supply multiple CreateMessageParams objects, you will face issues. But at this point
// you really need to reconsider your own code.
//
// Note that sending a &Message will simply refer to it, and not copy over the contents into
// the reply. example output: message{6434732342356}
//
// If you want to affect the actual message data besides .Content; provide a
// MessageCreateParams. The reply message will be updated by the last one provided.
func (c *client) SendMsg(channelID Snowflake, data ...interface{}) (msg *Message, err error) {

	var flags []Flag
	params := &CreateMessageParams{}
	for i := range data {
		if data[i] == nil {
			continue
		}

		var s string
		switch t := data[i].(type) {
		case *CreateMessageParams:
			*params = *t
		case CreateMessageParams:
			*params = t
		case string:
			s = t
		case *Flag:
			flags = append(flags, *t)
		case Flag:
			flags = append(flags, t)
		default:
			if str, ok := t.(fmt.Stringer); ok {
				s = str.String()
			} else {
				s = fmt.Sprint(t)
			}
		}

		if s != "" {
			params.Content += " " + s
		}
	}
	if data == nil {
		if mergeFlags(flags).IgnoreEmptyParams() {
			params.Content = ""
		} else {
			return nil, errors.New("params were nil")
		}
	}

	return c.CreateMessage(channelID, params, flags...)
}

/* status updates */

// UpdateStatus updates the client's game status
// note: for simple games, check out UpdateStatusString
func (c *client) UpdateStatus(s *UpdateStatusCommand) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return c.shardManager.Emit(CommandUpdateStatus, s)
}

// UpdateStatusString sets the client's game activity to the provided string, status to online
// and type to Playing
func (c *client) UpdateStatusString(s string) error {
	updateData := &UpdateStatusCommand{
		Since: nil,
		Game: &Activity{
			Name: s,
			Type: 0,
		},
		Status: StatusOnline,
		AFK:    false,
	}
	return c.UpdateStatus(updateData)
}

// eventHandler Takes a incoming event from the websocket package, parses it, and sends
// trigger requests to the event dispatcher and state cacher.
func (c *client) eventHandler() {
	for {
		var err error
		var evt *websocket.Event
		var alive bool

		select {
		case evt, alive = <-c.shardManager.evtChan:
			if !alive {
				return
			}
		case <-c.shutdownChan:
			return
		}

		var box eventBox

		switch evt.Name {
		case EventReady:
			box = &Ready{}
		case EventResumed:
			box = &Resumed{}
		case EventChannelCreate:
			box = &ChannelCreate{}
		case EventChannelUpdate:
			box = &ChannelUpdate{}
		case EventChannelDelete:
			box = &ChannelDelete{}
		case EventChannelPinsUpdate:
			box = &ChannelPinsUpdate{}
		case EventGuildCreate:
			box = &GuildCreate{}
		case EventGuildUpdate:
			box = &GuildUpdate{}
		case EventGuildDelete:
			box = &GuildDelete{}
		case EventGuildBanAdd:
			box = &GuildBanAdd{}
		case EventGuildBanRemove:
			box = &GuildBanRemove{}
		case EventGuildEmojisUpdate:
			box = &GuildEmojisUpdate{}
		case EventGuildIntegrationsUpdate:
			box = &GuildIntegrationsUpdate{}
		case EventGuildMemberAdd:
			box = &GuildMemberAdd{}
		case EventGuildMemberRemove:
			box = &GuildMemberRemove{}
		case EventGuildMemberUpdate:
			box = &GuildMemberUpdate{}
		case EventGuildMembersChunk:
			box = &GuildMembersChunk{}
		case EventGuildRoleCreate:
			box = &GuildRoleCreate{}
		case EventGuildRoleUpdate:
			box = &GuildRoleUpdate{}
		case EventGuildRoleDelete:
			box = &GuildRoleDelete{}
		case EventMessageCreate:
			//box = c.pool.msgCreate.Get().(*MessageCreate)
			box = &MessageCreate{}
		case EventMessageUpdate:
			box = &MessageUpdate{}
		case EventMessageDelete:
			box = &MessageDelete{}
		case EventMessageDeleteBulk:
			box = &MessageDeleteBulk{}
		case EventMessageReactionAdd:
			box = &MessageReactionAdd{}
		case EventMessageReactionRemove:
			box = &MessageReactionRemove{}
		case EventMessageReactionRemoveAll:
			box = &MessageReactionRemoveAll{}
		case EventPresenceUpdate:
			box = &PresenceUpdate{}
		case EventTypingStart:
			box = &TypingStart{}
		case EventUserUpdate:
			box = &UserUpdate{}
		case EventVoiceStateUpdate:
			box = &VoiceStateUpdate{}
		case EventVoiceServerUpdate:
			box = &VoiceServerUpdate{}
		case EventWebhooksUpdate:
			box = &WebhooksUpdate{}
		default:
			fmt.Printf("------\nTODO\nImplement event handler for `%s`, data: \n%+v\n------\n\n", evt.Name, string(evt.Data))
			continue // move on to next event
		}

		// populate box
		ctx := context.Background()
		box.registerContext(ctx)

		// first unmarshal to get identifiers
		//tmp := *box

		// unmarshal into cacheLink
		//err := c.cacheEvent2(evtName, box)

		if err = httd.Unmarshal(evt.Data, box); err != nil {
			c.log.Error(err)
			continue // ignore event
			// TODO: if an event is ignored, should it not at least send a signal for listeners with no parameters?
		}
		executeInternalUpdater(evt)
		executeInternalClientUpdater(c, evt)

		// cache
		if !c.config.DisableCache {
			cacheEvent(c.cache, evt.Name, box, evt.Data)
		}

		// trigger listeners
		c.evtDemultiplexer.triggerChan(ctx, evt.Name, c, box)
		go c.evtDemultiplexer.triggerHandlers(ctx, evt.Name, c, box)
	}
}

func (c *client) newRESTRequest(conf *httd.Request, flags []Flag) *rest {
	r := &rest{
		c:    c,
		conf: conf,
	}
	r.init()
	r.flags = mergeFlags(flags)

	return r
}

//////////////////////////////////////////////////////
//
// Deprecated / Legacy supported REST methods
//
// As I want to keep the method names simple, I
// do not want to make it difficult to use
// DisGord side by side with the documentation.
//
// Below I've added all the REST methods with
// their names, as in, the discord documentation.
//
//////////////////////////////////////////////////////

// Deprecated: use UpdateChannel
func (c *client) ModifyChannel(id Snowflake, flags ...Flag) *updateChannelBuilder {
	return c.UpdateChannel(id, flags...)
}

// Deprecated: use DeleteChannel
func (c *client) CloseChannel(id Snowflake, flags ...Flag) (*Channel, error) {
	return c.DeleteChannel(id, flags...)
}

// Deprecated: use DeleteMessages
func (c *client) BulkDeleteMessages(id Snowflake, params *DeleteMessagesParams, flags ...Flag) error {
	return c.DeleteMessages(id, params, flags...)
}

// Deprecated: use UpdateMessage
func (c *client) EditMessage(chanID, msgID Snowflake, flags ...Flag) *updateMessageBuilder {
	return c.UpdateMessage(chanID, msgID, flags...)
}

// Deprecated: use UpdateChannelPermissions
func (c *client) EditChannelPermissions(channelID, overwriteID Snowflake, params *UpdateChannelPermissionsParams, flags ...Flag) error {
	return c.UpdateChannelPermissions(channelID, overwriteID, params, flags...)
}

// Deprecated: use PinMessage or PinMessageID
func (c *client) AddPinnedChannelMessage(channelID, messageID Snowflake, flags ...Flag) (err error) {
	return c.PinMessageID(channelID, messageID, flags...)
}

// Deprecated: use UnpinMessage or UnpinMessageID
func (c *client) DeletePinnedChannelMessage(channelID, messageID Snowflake, flags ...Flag) (err error) {
	return c.UnpinMessageID(channelID, messageID, flags...)
}

// Deprecated: use AddDMParticipant
func (c *client) GroupDMAddRecipient(channelID Snowflake, recipient *GroupDMParticipant, flags ...Flag) (err error) {
	return c.AddDMParticipant(channelID, recipient, flags...)
}

// Deprecated: use KickParticipant
func (c *client) GroupDMRemoveRecipient(channelID, userID Snowflake, flags ...Flag) error {
	return c.KickParticipant(channelID, userID, flags...)
}

// Deprecated: use GetGuildEmojis
func (c *client) ListGuildEmojis(guildID Snowflake, flags ...Flag) ([]*Emoji, error) {
	return c.GetGuildEmojis(guildID, flags...)
}

// Deprecated: use UpdateGuildEmoji
func (c *client) ModifyGuildEmoji(guildID, emojiID Snowflake, flags ...Flag) *updateGuildEmojiBuilder {
	return c.UpdateGuildEmoji(guildID, emojiID, flags...)
}

// Deprecated: use UpdateGuild
func (c *client) ModifyGuild(id Snowflake, flags ...Flag) *updateGuildBuilder {
	return c.UpdateGuild(id, flags...)
}

// Deprecated: use UpdateGuildChannelPositions
func (c *client) ModifyGuildChannelPositions(id Snowflake, params []UpdateGuildChannelPositionsParams, flags ...Flag) error {
	return c.UpdateGuildChannelPositions(id, params, flags...)
}

// Deprecated: use GetGuildMembers
func (c *client) ListGuildMembers(id, after Snowflake, limit int, flags ...Flag) ([]*Member, error) {
	return c.GetGuildMembers(id, &GetGuildMembersParams{
		After: after,
		Limit: limit,
	}, flags...)
}

// TODO: AddGuildMember => CreateGuildMember

// Deprecated: use UpdateGuildMember
func (c *client) ModifyGuildMember(guildID, userID Snowflake, flags ...Flag) *updateGuildMemberBuilder {
	return c.UpdateGuildMember(guildID, userID, flags...)
}

// Deprecated: use SetCurrentUserNick
func (c *client) ModifyCurrentUserNick(guildID Snowflake, nick string, flags ...Flag) (string, error) {
	return c.SetCurrentUserNick(guildID, nick, flags...)
}

// TODO: AddGuildMemberRole => UpdateGuildMember
// TODO: RemoveGuildMemberRole => UpdateGuildMember

// Deprecated: use KickMember
func (c *client) RemoveGuildMember(guildID, userID Snowflake, flags ...Flag) error {
	return c.KickMember(guildID, userID, flags...)
}

// Deprecated: use UnbanMember
func (c *client) RemoveGuildBan(guildID, userID Snowflake, flags ...Flag) error {
	return c.UnbanMember(guildID, userID, flags...)
}

// Deprecated: use UpdateGuildRolePositions
func (c *client) ModifyGuildRolePositions(guildID Snowflake, params []UpdateGuildRolePositionsParams, flags ...Flag) (ret []*Role, err error) {
	return c.UpdateGuildRolePositions(guildID, params, flags...)
}

// Deprecated: use DeleteGuildRole
func (c *client) RemoveGuildRole(guildID, roleID Snowflake, flags ...Flag) error {
	return c.DeleteGuildRole(guildID, roleID, flags...)
}

// Deprecated: use PruneMembers
func (c *client) BeginGuildPrune(guildID Snowflake, nrOfDays int, flags ...Flag) error {
	return c.PruneMembers(guildID, nrOfDays, flags...)
}

// Deprecated: use EstimatePruneMembersCount
func (c *client) GetGuildPruneCount(guildID Snowflake, nrOfDays int, flags ...Flag) (int, error) {
	return c.EstimatePruneMembersCount(guildID, nrOfDays, flags...)
}

// Deprecated: use UpdateGuildIntegration
func (c *client) ModifyGuildIntegration(guildID, integrationID Snowflake, params *UpdateGuildIntegrationParams, flags ...Flag) error {
	return c.UpdateGuildIntegration(guildID, integrationID, params, flags...)
}

// Deprecated: use UpdateGuildEmbed
func (c *client) ModifyGuildEmbed(guildID Snowflake, flags ...Flag) *updateGuildEmbedBuilder {
	return c.UpdateGuildEmbed(guildID, flags...)
}

// Deprecated: use UpdateCurrentUser
func (c *client) ModifyCurrentUser(flags ...Flag) *updateCurrentUserBuilder {
	return c.UpdateCurrentUser(flags...)
}

// Deprecated: use LeaveGuild
func (c *client) ListVoiceRegions(flags ...Flag) ([]*VoiceRegion, error) {
	return c.GetVoiceRegions(flags...)
}

// Deprecated: use UpdateWebhook
func (c *client) ModifyWebhook(id Snowflake, flags ...Flag) *updateWebhookBuilder {
	return c.UpdateWebhook(id, flags...)
}

// Deprecated: use UpdateWebhookWithToken
func (c *client) ModifyWebhookWithToken(newWebhook *Webhook, flags ...Flag) *updateWebhookBuilder {
	return c.UpdateWebhookWithToken(newWebhook.ID, newWebhook.Token, flags...)
}
