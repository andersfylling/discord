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

	"github.com/andersfylling/disgord/endpoint"

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

// Req return the request object. Used in REST requests to handle rate limits,
// wrong http responses, etc.
func (c *client) Req() httd.Requester {
	return c.req
}

// Cache returns the cacheLink manager for the session
func (c *client) Cache() Cacher {
	return c.cache
}

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

func ValidateHandlerInputs(inputs ...interface{}) (err error) {
	var i int
	var ok bool

	// make sure that middlewares are only at beginning
	for j := i; j < len(inputs); j++ {
		if _, ok = inputs[j].(Middleware); ok {
			if j != i {
				return errors.New("middlewares can only be in the beginning. Grouped together")
			}
			i++
		}
	}

	// there should now be N handlers, 0 < N.
	if len(inputs) <= i {
		return errors.New("missing handler(s)")
	}

	// check for extra controllers
	for j := len(inputs) - 2; j >= i; j-- {
		if _, ok = inputs[j].(HandlerCtrl); ok {
			return errors.New("a handlerCtrl's can only be at the end of the definition. Expected a handler")
		}
	}

	// TODO: test for all handler types?

	return nil
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

// Generic CRUDS

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

// REST

// Guild

// CreateGuild .
func (c *client) CreateGuild(params *CreateGuildParams, flags ...Flag) (ret *Guild, err error) {
	ret, err = CreateGuild(c.req, params)
	return
}

// GetGuild .
func (c *client) GetGuild(id Snowflake, flags ...Flag) (ret *Guild, err error) {
	ret, err = c.cache.GetGuild(id)
	if err != nil {
		ret, err = GetGuild(c.req, id)
		if err != nil {
			return nil, err
		}
		c.cache.SetGuild(ret)
	}
	return ret, err
}

// ModifyGuild .
func (c *client) UpdateGuild(id Snowflake, params *UpdateGuildParams, flags ...Flag) (ret *Guild, err error) {
	ret, err = ModifyGuild(c.req, id, params)
	return
}

// DeleteGuild .
func (c *client) DeleteGuild(id Snowflake, flags ...Flag) (err error) {
	err = DeleteGuild(c.req, id)
	return
}

// GetGuildChannels .
func (c *client) GetGuildChannels(id Snowflake, flags ...Flag) (ret []*Channel, err error) {
	var guild *Guild
	guild, err = c.cache.GetGuild(id)
	if err != nil {
		ret, err = GetGuildChannels(c.req, id)
		if err != nil {
			return nil, err
		}
		c.cache.SetGuild(&Guild{
			ID:       id,
			Channels: ret,
		})
	} else {
		ret = guild.Channels
	}
	return ret, nil
}

// CreateGuildChannel .
func (c *client) CreateGuildChannel(id Snowflake, params *CreateGuildChannelParams, flags ...Flag) (ret *Channel, err error) {
	ret, err = CreateGuildChannel(c.req, id, params)
	return
}

// ModifyGuildChannelPositions .
func (c *client) UpdateGuildChannelPositions(id Snowflake, params []UpdateGuildChannelPositionsParams, flags ...Flag) (ret *Guild, err error) {
	ret, err = ModifyGuildChannelPositions(c.req, id, params)
	return
}

// GetGuildMember .
func (c *client) GetGuildMember(guildID, userID Snowflake, flags ...Flag) (ret *Member, err error) {
	ret, err = c.cache.GetGuildMember(guildID, userID)
	if err != nil {
		ret, err = GetGuildMember(c.req, guildID, userID)
		if err != nil {
			return
		}
		c.cache.SetGuildMember(guildID, ret)
		_ = c.cache.Update(UserCache, ret.User)
	}
	return
}

// GetGuildMembers .
func (c *client) GetGuildMembers(guildID, after Snowflake, limit int, flags ...Flag) (ret []*Member, err error) {
	ret, err = c.cache.GetGuildMembersAfter(guildID, after, limit)
	if err != nil {
		ret, err = GetGuildMembers(c.req, guildID, after, limit)
		if err != nil {
			return nil, err
		}
		c.cache.SetGuildMembers(guildID, ret)
		//c.cacheLink.Update(UserCache, ret.User)
		// TODO: update users
	}
	return ret, nil
}

// AddGuildMember .
func (c *client) AddGuildMember(guildID, userID Snowflake, params *AddGuildMemberParams, flags ...Flag) (ret *Member, err error) {
	ret, err = AddGuildMember(c.req, guildID, userID, params)
	return
}

// ModifyGuildMember .
func (c *client) UpdateGuildMember(guildID, userID Snowflake, params *UpdateGuildMemberParams, flags ...Flag) (err error) {
	err = ModifyGuildMember(c.req, guildID, userID, params)
	return
}

// updateCurrentUserNickParams ...
// https://discordapp.com/developers/docs/resources/guild#modify-guild-member-json-params
type updateCurrentUserNickParams struct {
	Nick string `json:"nick"` // :CHANGE_NICKNAME
}

type nickNameResponse struct {
	Nickname string `json:"nickname"`
}

// SetCurrentUserNick [REST] Modifies the nickname of the current user in a guild. Returns a 200
// with the nickname on success. Fires a Guild Member Update Gateway event.
//  Method                  PATCH
//  Endpoint                /guilds/{guild.id}/members/@me/nick
//  Rate limiter            /guilds/{guild.id}/members/@me/nick
//  Discord documentation   https://discordapp.com/developers/docs/resources/guild#modify-current-user-nick
//  Reviewed                2018-08-18
//  Comment                 -
func (c *client) SetCurrentUserNick(id Snowflake, nick string, flags ...Flag) (newNick string, err error) {
	params := &updateCurrentUserNickParams{
		Nick: nick,
	}

	r := c.newRESTRequest(&httd.Request{
		Method:      http.MethodPatch,
		Ratelimiter: ratelimitGuildMembers(id),
		Endpoint:    endpoint.GuildMembersMeNick(id),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
	}, flags)
	r.expectsStatusCode = http.StatusOK
	r.factory = func() interface{} {
		return &nickNameResponse{}
	}

	return getNickName(r.Execute)
}

// AddGuildMemberRole .
func (c *client) AddGuildMemberRole(guildID, userID, roleID Snowflake, flags ...Flag) (err error) {
	err = AddGuildMemberRole(c.req, guildID, userID, roleID)
	return
}

// RemoveGuildMemberRole .
// TODO: merge with UpdateGuildMember
func (c *client) RemoveGuildMemberRole(guildID, userID, roleID Snowflake, flags ...Flag) (err error) {
	err = RemoveGuildMemberRole(c.req, guildID, userID, roleID)
	return
}

// KickMember .
func (c *client) KickMember(guildID, userID Snowflake, flags ...Flag) (err error) {
	err = RemoveGuildMember(c.req, guildID, userID)
	return
}

// GetGuildRoles .
func (c *client) GetGuildRoles(guildID Snowflake, flags ...Flag) (ret []*Role, err error) {
	ret, err = c.cache.GetGuildRoles(guildID)
	if err != nil {
		ret, err = GetGuildRoles(c.req, guildID)
		if err != nil {
			return nil, err
		}
		c.cache.SetGuildRoles(guildID, ret)
	}

	return ret, nil
}

// CreateGuildRole .
func (c *client) CreateGuildRole(id Snowflake, params *CreateGuildRoleParams, flags ...Flag) (ret *Role, err error) {
	ret, err = CreateGuildRole(c.req, id, params)
	return
}

// UpdateGuildRolePositions .
func (c *client) UpdateGuildRolePositions(guildID Snowflake, params []UpdateGuildRolePositionsParams, flags ...Flag) (ret []*Role, err error) {
	ret, err = UpdateGuildRolePositions(c.req, guildID, params)
	return
}

// DeleteGuildRole .
func (c *client) DeleteGuildRole(guildID, roleID Snowflake, flags ...Flag) (err error) {
	err = DeleteGuildRole(c.req, guildID, roleID)
	return
}

// GetGuildVoiceRegions .
func (c *client) GetGuildVoiceRegions(id Snowflake, flags ...Flag) (ret []*VoiceRegion, err error) {
	ret, err = GetGuildVoiceRegions(c.req, id)
	return
}

// GetGuildInvites .
func (c *client) GetGuildInvites(id Snowflake, flags ...Flag) (ret []*Invite, err error) {
	ret, err = GetGuildInvites(c.req, id)
	return
}

// GetGuildIntegrations .
func (c *client) GetGuildIntegrations(id Snowflake, flags ...Flag) (ret []*Integration, err error) {
	ret, err = GetGuildIntegrations(c.req, id)
	return
}

// CreateGuildIntegration .
func (c *client) CreateGuildIntegration(guildID Snowflake, params *CreateGuildIntegrationParams, flags ...Flag) (err error) {
	err = CreateGuildIntegration(c.req, guildID, params)
	return
}

// UpdateGuildIntegration .
func (c *client) UpdateGuildIntegration(guildID, integrationID Snowflake, params *UpdateGuildIntegrationParams, flags ...Flag) (err error) {
	err = UpdateGuildIntegration(c.req, guildID, integrationID, params)
	return
}

// DeleteGuildIntegration .
func (c *client) DeleteGuildIntegration(guildID, integrationID Snowflake, flags ...Flag) (err error) {
	err = DeleteGuildIntegration(c.req, guildID, integrationID)
	return
}

// SyncGuildIntegration .
func (c *client) SyncGuildIntegration(guildID, integrationID Snowflake, flags ...Flag) (err error) {
	err = SyncGuildIntegration(c.req, guildID, integrationID)
	return
}

// GetGuildEmbed .
func (c *client) GetGuildEmbed(guildID Snowflake, flags ...Flag) (ret *GuildEmbed, err error) {
	ret, err = GetGuildEmbed(c.req, guildID)
	return
}

// ModifyGuildEmbed .
func (c *client) UpdateGuildEmbed(guildID Snowflake, params *GuildEmbed, flags ...Flag) (ret *GuildEmbed, err error) {
	ret, err = ModifyGuildEmbed(c.req, guildID, params)
	return
}

// GetGuildVanityURL .
func (c *client) GetGuildVanityURL(guildID Snowflake, flags ...Flag) (ret *PartialInvite, err error) {
	ret, err = GetGuildVanityURL(c.req, guildID)
	return
}

// Webhook

// CreateWebhook .
func (c *client) CreateWebhook(channelID Snowflake, params *CreateWebhookParams, flags ...Flag) (ret *Webhook, err error) {
	ret, err = CreateWebhook(c.req, channelID, params)
	return
}

// GetChannelWebhooks .
func (c *client) GetChannelWebhooks(channelID Snowflake, flags ...Flag) (ret []*Webhook, err error) {
	ret, err = GetChannelWebhooks(c.req, channelID)
	return
}

// GetGuildWebhooks .
func (c *client) GetGuildWebhooks(guildID Snowflake, flags ...Flag) (ret []*Webhook, err error) {
	ret, err = GetGuildWebhooks(c.req, guildID)
	return
}

// GetWebhook .
func (c *client) GetWebhook(id Snowflake, flags ...Flag) (ret *Webhook, err error) {
	ret, err = GetWebhook(c.req, id)
	return
}

// GetWebhookWithToken .
func (c *client) GetWebhookWithToken(id Snowflake, token string, flags ...Flag) (ret *Webhook, err error) {
	ret, err = GetWebhookWithToken(c.req, id, token)
	return
}

// ModifyWebhook .
func (c *client) UpdateWebhook(id Snowflake, params *UpdateWebhookParams, flags ...Flag) (ret *Webhook, err error) {
	if id.Empty() {
		err = errors.New("given webhook ID was not set, there is nothing to modify")
		return
	}
	if params == nil {
		err = errors.New("given param object was nil, there is nothing to modify")
		return
	}
	if params.Empty() {
		err = errors.New("given param object was empty, there is nothing to modify")
		return
	}

	// verify avatar string prefix
	if params.avatarIsSet && params.avatar != "" && !validAvatarPrefix(params.avatar) {
		err = errors.New("given avatar string is invalid. Must specify data encoding. Eg. 'data:image/jpeg;base64,'")
		return
	}

	// TODO: check if user has permission to modify webhook
	ret, err = UpdateWebhook(c.req, id, params)
	return
}

// ModifyWebhookWithToken .
func (c *client) UpdateWebhookWithToken(newWebhook *Webhook, flags ...Flag) (ret *Webhook, err error) {
	ret, err = UpdateWebhookWithToken(c.req, newWebhook)
	return
}

// DeleteWebhook .
func (c *client) DeleteWebhook(webhookID Snowflake, flags ...Flag) (err error) {
	err = DeleteWebhook(c.req, webhookID)
	return
}

// DeleteWebhookWithToken .
func (c *client) DeleteWebhookWithToken(id Snowflake, token string, flags ...Flag) (err error) {
	err = DeleteWebhookWithToken(c.req, id, token)
	return
}

// ExecuteWebhook .
func (c *client) ExecuteWebhook(params *ExecuteWebhookParams, wait bool, URLSuffix string, flags ...Flag) (err error) {
	err = ExecuteWebhook(c.req, params, wait, URLSuffix)
	return
}

// ExecuteSlackWebhook .
func (c *client) ExecuteSlackWebhook(params *ExecuteWebhookParams, wait bool, flags ...Flag) (err error) {
	err = ExecuteSlackWebhook(c.req, params, wait)
	return
}

// ExecuteGitHubWebhook .
func (c *client) ExecuteGitHubWebhook(params *ExecuteWebhookParams, wait bool, flags ...Flag) (err error) {
	err = ExecuteGitHubWebhook(c.req, params, wait)
	return
}

//
// #########################################################################
//
//
// Custom methods are usually reused by the resource package for readability
// -----

func (c *client) GetGuilds(params *GetCurrentUserGuildsParams, flags ...Flag) ([]*Guild, error) {
	// TODO: populate these partial guild objects
	return c.GetCurrentUserGuilds(params)
}

// SendMsg .
func (c *client) SendMsg(channelID Snowflake, message *Message, flags ...Flag) (msg *Message, err error) {
	if constant.LockedMethods {
		message.RLock()
	}
	params := &CreateMessageParams{
		Content:                  message.Content,
		Tts:                      message.Tts,
		SpoilerTagContent:        message.SpoilerTagContent,
		SpoilerTagAllAttachments: message.SpoilerTagAllAttachments,
		// File: ...
		// Embed: ...
	}
	if !message.Nonce.Empty() {
		params.Nonce = message.Nonce
	}
	if len(message.Embeds) > 0 {
		params.Embed = message.Embeds[0]
	}

	if constant.LockedMethods {
		message.RUnlock()
	}

	return c.CreateMessage(channelID, params)
}

// SendMsgString .
func (c *client) SendMsgString(channelID Snowflake, content string, flags ...Flag) (msg *Message, err error) {
	params := &CreateMessageParams{
		Content: content,
	}

	msg, err = c.CreateMessage(channelID, params)
	return
}

func waitForEvent(eventEmitter <-chan *websocket.Event) (event *websocket.Event, err error) {
	var alive bool
	event, alive = <-eventEmitter
	if !alive {
		err = errors.New("event emitter (channel) is dead")
	}

	return
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
// I don't agree with the way Discord have named their
// REST methods, and I've restricted DisGord to only use
// CRUD prefixes (READ => GET). Here I list every REST
// method that does not match the one in the Discord docs.
//
// This makes it intuitive to use DisGord while using
// the Discord docs for reference.
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
func (c *client) EditMessage(chanID, msgID Snowflake, params *UpdateMessageParams, flags ...Flag) (ret *Message, err error) {
	return c.UpdateMessage(chanID, msgID, params, flags...)
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
func (c *client) ModifyGuild(id Snowflake, params *UpdateGuildParams, flags ...Flag) (*Guild, error) {
	return c.UpdateGuild(id, params, flags...)
}

// Deprecated: use UpdateGuildChannelPositions
func (c *client) ModifyGuildChannelPositions(id Snowflake, params []UpdateGuildChannelPositionsParams, flags ...Flag) (*Guild, error) {
	return c.UpdateGuildChannelPositions(id, params, flags...)
}

// Deprecated: use GetGuildMembers
func (c *client) ListGuildMembers(id, after Snowflake, limit int, flags ...Flag) ([]*Member, error) {
	return c.GetGuildMembers(id, after, limit, flags...)
}

// TODO: AddGuildMember => CreateGuildMember

// Deprecated: use UpdateGuildMember
func (c *client) ModifyGuildMember(guildID, userID Snowflake, params *UpdateGuildMemberParams, flags ...Flag) error {
	return c.UpdateGuildMember(guildID, userID, params, flags...)
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
func (c *client) ModifyGuildEmbed(guildID Snowflake, params *GuildEmbed, flags ...Flag) (*GuildEmbed, error) {
	return c.UpdateGuildEmbed(guildID, params, flags...)
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
func (c *client) ModifyWebhook(id Snowflake, params *UpdateWebhookParams, flags ...Flag) (ret *Webhook, err error) {
	return c.UpdateWebhook(id, params, flags...)
}

// Deprecated: use UpdateWebhookWithToken
func (c *client) ModifyWebhookWithToken(newWebhook *Webhook, flags ...Flag) (ret *Webhook, err error) {
	return c.UpdateWebhookWithToken(newWebhook, flags...)
}
