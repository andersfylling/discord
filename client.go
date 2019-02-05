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

	"github.com/andersfylling/snowflake/v3"

	"github.com/andersfylling/disgord/constant"
	"github.com/andersfylling/disgord/websocket"
	"golang.org/x/net/proxy"

	"github.com/andersfylling/disgord/event"
	"github.com/andersfylling/disgord/httd"
)

// NewRESTClient creates a client for sending and handling Discord protocols such as rate limiting
func NewRESTClient(conf *Config) (client *httd.Client) {
	// request client
	reqConf := &httd.Config{
		APIVersion:                   constant.DiscordVersion,
		BotToken:                     conf.BotToken,
		UserAgentSourceURL:           constant.GitHubURL,
		UserAgentVersion:             constant.Version,
		HTTPClient:                   conf.HTTPClient,
		CancelRequestWhenRateLimited: conf.CancelRequestWhenRateLimited,
	}
	client = httd.NewClient(reqConf)
	return
}

// NewSession create a client and return the Session interface
func NewClient(conf *Config) (*Client, error) {
	var err error
	if conf.HTTPClient == nil {
		// http client configuration
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

	// request client
	reqClient := NewRESTClient(conf)

	if conf.WSShardManagerConfig == nil {
		conf.WSShardManagerConfig = &WSShardManagerConfig{}
	}
	if conf.WSShardManagerConfig.ShardRateLimit == 0 {
		conf.WSShardManagerConfig.ShardRateLimit = DefaultShardRateLimit
	}
	sharding := NewShardManager(conf)

	// caching
	// TODO: should not pre-set the cache sizes as some guilds might be small while others huge.
	//  might spend more memory than needed
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

		// Required for voice operation
		sharding.TrackEvent.Add(event.VoiceStateUpdate)
		sharding.TrackEvent.Add(event.VoiceServerUpdate)
	}

	sharding.TrackEvent.Add(event.VoiceStateUpdate)
	sharding.TrackEvent.Add(event.VoiceServerUpdate)

	// event dispatcher
	eventChanSize := 20
	evtDispatcher := NewDispatch(conf.ActivateEventChannels, eventChanSize)

	// create a disgord client/instance/session
	c := &Client{
		shutdownChan: conf.shutdownChan,
		config:       conf,
		shardManager: sharding,
		httpClient:   conf.HTTPClient,
		Proxy:        conf.Proxy,
		botToken:     conf.BotToken,
		evtDispatch:  evtDispatcher,
		req:          reqClient,
		cache:        cacher,
		log:          conf.Logger,
	}
	c.voiceRepository = newVoiceRepository(c)

	return c, nil
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

// conClient is the main disgord client to hold your state and data
type Client struct {
	sync.RWMutex

	shutdownChan chan interface{}
	config       *Config
	botToken     string

	myID Snowflake

	// register listeners for events
	evtDispatch *Dispatch

	// cancelRequestWhenRateLimited by default the client waits until either the HTTPClient.timeout or
	// the rate limit ends before closing a request channel. If activated, in stead, requests will
	// instantly be denied, and the channel closed.
	cancelRequestWhenRateLimited bool

	// discord http api
	req *httd.Client

	// http client used for connections
	httpClient *http.Client
	Proxy      proxy.Dialer

	shardManager *WSShardManager

	cache *Cache

	log Logger

	// voice
	*voiceRepository
}

var _ Session = (*Client)(nil)

// HeartbeatLatency checks the duration of waiting before receiving a response from Discord when a
// heartbeat packet was sent. Note that heartbeats are usually sent around once a minute and is not a accurate
// way to measure delay between the client and Discord server
func (c *Client) HeartbeatLatency() (duration time.Duration, err error) {
	return c.shardManager.GetAvgHeartbeatLatency()
}

// Myself get the current user / connected user
func (c *Client) Myself() (user *User, err error) {
	if c.myID.Empty() {
		user, err = c.GetCurrentUser().IgnoreCache().Execute()
		if err == nil {
			c.myID = user.ID
		}
		return
	}

	var usr interface{}
	usr, err = c.cache.Get(UserCache, c.myID)
	if err == nil {
		user = usr.(*User)
	}
	return
}

// GetConnectedGuilds get a list over guild IDs that this client is "connected to"; or have joined through the ws connection. This will always hold the different Guild IDs, while the GetGuilds or GetCurrentUserGuilds might be affected by cache configuration.
func (c *Client) GetConnectedGuilds() []snowflake.ID {
	c.shardManager.RLock()
	defer c.shardManager.RUnlock()

	var guilds []snowflake.ID
	for i := range c.shardManager.shards {
		guilds = append(guilds, c.shardManager.shards[i].guilds...)
	}

	return guilds
}

// Deprecated
func (c *Client) logInfo(msg string) {
	c.Info(msg)
}

// Deprecated
func (c *Client) logErr(msg string) {
	c.Error(msg)
}

func (c *Client) Info(v ...interface{}) {
	if c.log != nil {
		c.log.Info(v)
	}
}
func (c *Client) Debug(v ...interface{}) {
	if c.log != nil {
		c.log.Debug(v)
	}
}
func (c *Client) Error(v ...interface{}) {
	if c.log != nil {
		c.log.Error(v)
	}
}

var _ Logger = (*Client)(nil)

func (c *Client) String() string {
	return LibraryInfo()
}

var _ fmt.Stringer = (*Client)(nil)

// RateLimiter return the rate limiter object
func (c *Client) RateLimiter() httd.RateLimiter {
	return c.req.RateLimiter()
}

func (c *Client) setupConnectEnv() {
	fmt.Println("Setting up connect env")
	// set the user ID upon connection
	// only works with socket logic
	c.Once(event.Ready, c.handlerSetSelfBotID)
	c.On(event.UserUpdate, c.handlerUpdateSelfBot)
	c.On(event.GuildCreate, c.handlerAddToConnectedGuilds)
	c.On(event.GuildDelete, c.handlerRemoveFromConnectedGuilds)

	// setup event observer
	go c.eventHandler()
	//c.evtDispatch.start()
}

// Connect establishes a websocket connection to the discord API
func (c *Client) Connect() (err error) {
	// set the user ID upon connection
	// only works for socketing
	//
	// also verifies that the correct credentials were supplied
	var me *User
	me, err = c.GetCurrentUser().Execute()
	if err != nil {
		return
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

	c.Info("Connecting to discord Gateway")
	err = c.shardManager.Connect()
	if err != nil {
		c.Info(err)
		return
	}
	c.Info("Connected")

	return nil
}

// Disconnect closes the discord websocket connection
func (c *Client) Disconnect() (err error) {
	fmt.Println() // to keep ^C on it's own line
	c.Info("Closing Discord gateway connection")
	close(c.evtDispatch.shutdown)
	err = c.shardManager.Disconnect()
	if err != nil {
		c.Error(err)
		return
	}
	close(c.shutdownChan)
	c.Info("Disconnected")

	return nil
}

var _ Link = (*Client)(nil)

// DisconnectOnInterrupt wait until a termination signal is detected
func (c *Client) DisconnectOnInterrupt() (err error) {
	// create a channel to listen for termination signals (graceful shutdown)
	termSignal := make(chan os.Signal, 1)
	signal.Notify(termSignal, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-termSignal

	return c.Disconnect()
}

// Req return the request object. Used in REST requests to handle rate limits,
// wrong http responses, etc.
func (c *Client) Req() httd.Requester {
	return c.req
}

// Cache returns the cacheLink manager for the session
func (c *Client) Cache() Cacher {
	return c.cache
}

// Ready triggers a given callback when all shards has gotten their first Ready event
// Warning: if you run client.Disconnect and want to run Connect again later, this will
//  not work. The callback will be triggered instantly, as all the shards have already
//  successfully connected once.
// TODO: allow this method to be reused (see Warning paragraph)
func (c *Client) Ready(cb func()) {
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

// On adds a event handler on the given event.
// On => event => handle the content like this
func (c *Client) On(event string, handlers ...interface{}) {
	c.shardManager.TrackEvent.Add(event)

	c.evtDispatch.listenersLock.Lock()
	defer c.evtDispatch.listenersLock.Unlock()

	for _, handler := range handlers {
		c.evtDispatch.listeners[event] = append(c.evtDispatch.listeners[event], handler)
	}
}

// Once same as `On`, however, once the handler is triggered, it is removed. In other words, it is only triggered once.
func (c *Client) Once(event string, handlers ...interface{}) {
	c.shardManager.TrackEvent.Add(event) // TODO: remove event after firing. unless there are more handlers

	c.evtDispatch.listenersLock.Lock()
	defer c.evtDispatch.listenersLock.Unlock()
	for _, handler := range handlers {
		index := len(c.evtDispatch.listeners[event])
		c.evtDispatch.listeners[event] = append(c.evtDispatch.listeners[event], handler)
		c.evtDispatch.listenOnceOnly[event] = append(c.evtDispatch.listenOnceOnly[event], index)
	}
}

// Emit sends a socket command directly to Discord.
func (c *Client) Emit(command SocketCommand, data interface{}) error {
	switch command {
	case CommandUpdateStatus, CommandUpdateVoiceState, CommandRequestGuildMembers:
	default:
		return errors.New("command is not supported")
	}
	return c.shardManager.Emit(command, data)
}

// EventChan get a event channel using the event name
func (c *Client) EventChan(event string) (channel interface{}, err error) {
	return c.evtDispatch.EventChan(event)
}

// EventChannels get access to all the event channels
func (c *Client) EventChannels() (channels EventChannels) {
	return c.evtDispatch
}

// AcceptEvent only events registered using this method is accepted from the Discord socket API. The rest is discarded
// to reduce unnecessary marshalling and controls.
func (c *Client) AcceptEvent(events ...string) {
	for _, evt := range events {
		c.shardManager.TrackEvent.Add(evt)
	}
}

// Generic CRUDS

// DeleteFromDiscord if the given object has implemented the private interface discordDeleter this method can
// be used to delete said object.
func (c *Client) DeleteFromDiscord(obj discordDeleter) (err error) {
	err = obj.deleteFromDiscord(c)
	return
}

// SaveToDiscord saves an object to the Discord servers. This supports creation of new objects or udpating/modifying
// existing objects. It really depends on how the object has implemented the private interface discordSaver.
func (c *Client) SaveToDiscord(obj discordSaver) (err error) {
	err = obj.saveToDiscord(c)
	return
}

// REST

// Channel

// GetChannel ...
func (c *Client) GetChannel(id Snowflake) (ret *Channel, err error) {
	if ret, err = c.cache.GetChannel(id); err != nil {
		ret, err = GetChannel(c.req, id)
		if err != nil {
			return
		}
		_ = c.cache.Update(ChannelCache, ret)
	}
	return
}

// ModifyChannel ...
func (c *Client) ModifyChannel(id Snowflake, changes *ModifyChannelParams) (ret *Channel, err error) {
	ret, err = ModifyChannel(c.req, id, changes) // should trigger a socket event, no need to update cacheLink
	return
}

// DeleteChannel ...
func (c *Client) DeleteChannel(id Snowflake) (channel *Channel, err error) {
	channel, err = DeleteChannel(c.req, id) // should trigger a socket event, no need to update cacheLink
	return
}

// SetChannelPermissions ...
func (c *Client) SetChannelPermissions(chanID, overwriteID Snowflake, params *SetChannelPermissionsParams) (err error) {
	err = EditChannelPermissions(c.req, chanID, overwriteID, params)
	return
}

// GetChannelInvites ...
func (c *Client) GetChannelInvites(id Snowflake) (ret []*Invite, err error) {
	ret, err = GetChannelInvites(c.req, id)
	return
}

// CreateChannelInvites ...
func (c *Client) CreateChannelInvites(id Snowflake, params *CreateChannelInvitesParams) (ret *Invite, err error) {
	ret, err = CreateChannelInvites(c.req, id, params)
	return
}

// DeleteChannelPermission .
func (c *Client) DeleteChannelPermission(channelID, overwriteID Snowflake) (err error) {
	err = DeleteChannelPermission(c.req, channelID, overwriteID)
	return
}

// TriggerTypingIndicator .
func (c *Client) TriggerTypingIndicator(channelID Snowflake) (err error) {
	err = TriggerTypingIndicator(c.req, channelID)
	return
}

// GetPinnedMessages .
func (c *Client) GetPinnedMessages(channelID Snowflake) (ret []*Message, err error) {
	ret, err = GetPinnedMessages(c.req, channelID)
	return
}

// AddPinnedChannelMessage .
func (c *Client) AddPinnedChannelMessage(channelID, msgID Snowflake) (err error) {
	err = AddPinnedChannelMessage(c.req, channelID, msgID)
	return
}

// DeletePinnedChannelMessage .
func (c *Client) DeletePinnedChannelMessage(channelID, msgID Snowflake) (err error) {
	err = DeletePinnedChannelMessage(c.req, channelID, msgID)
	return
}

// GroupDMAddRecipient .
func (c *Client) GroupDMAddRecipient(channelID, userID Snowflake, params *GroupDMAddRecipientParams) (err error) {
	err = GroupDMAddRecipient(c.req, channelID, userID, params)
	return
}

// GroupDMRemoveRecipient .
func (c *Client) GroupDMRemoveRecipient(channelID, userID Snowflake) (err error) {
	err = GroupDMRemoveRecipient(c.req, channelID, userID)
	return
}

// GetChannelMessages .
func (c *Client) GetChannelMessages(channelID Snowflake, params URLParameters) (ret []*Message, err error) {
	ret, err = GetChannelMessages(c.req, channelID, params)
	return
}

// GetChannelMessage .
func (c *Client) GetChannelMessage(channelID, messageID Snowflake) (ret *Message, err error) {
	ret, err = GetChannelMessage(c.req, channelID, messageID)
	return
}

// CreateChannelMessage .
func (c *Client) CreateChannelMessage(channelID Snowflake, params *CreateChannelMessageParams) (ret *Message, err error) {
	ret, err = CreateChannelMessage(c.req, channelID, params)
	return
}

// EditMessage .
func (c *Client) EditMessage(chanID, msgID Snowflake, params *EditMessageParams) (ret *Message, err error) {
	ret, err = EditMessage(c.req, chanID, msgID, params)
	return
}

// DeleteMessage .
func (c *Client) DeleteMessage(channelID, msgID Snowflake) (err error) {
	err = DeleteMessage(c.req, channelID, msgID)
	return
}

// BulkDeleteMessages .
func (c *Client) BulkDeleteMessages(chanID Snowflake, params *BulkDeleteMessagesParams) (err error) {
	err = BulkDeleteMessages(c.req, chanID, params)
	return
}

// CreateReaction .
func (c *Client) CreateReaction(channelID, messageID Snowflake, emoji interface{}) (err error) {
	err = CreateReaction(c.req, channelID, messageID, emoji)
	return
}

// DeleteOwnReaction .
func (c *Client) DeleteOwnReaction(channelID, messageID Snowflake, emoji interface{}) (err error) {
	err = DeleteOwnReaction(c.req, channelID, messageID, emoji)
	return
}

// DeleteUserReaction .
func (c *Client) DeleteUserReaction(channelID, messageID, userID Snowflake, emoji interface{}) (err error) {
	err = DeleteUserReaction(c.req, channelID, messageID, userID, emoji)
	return
}

// GetReaction .
func (c *Client) GetReaction(channelID, messageID Snowflake, emoji interface{}, params URLParameters) (ret []*User, err error) {
	ret, err = GetReaction(c.req, channelID, messageID, emoji, params)
	return
}

// DeleteAllReactions .
func (c *Client) DeleteAllReactions(channelID, messageID Snowflake) (err error) {
	err = DeleteAllReactions(c.req, channelID, messageID)
	return
}

// Guild

// CreateGuild .
func (c *Client) CreateGuild(params *CreateGuildParams) (ret *Guild, err error) {
	ret, err = CreateGuild(c.req, params)
	return
}

// GetGuild .
func (c *Client) GetGuild(id Snowflake) (ret *Guild, err error) {
	ret, err = c.cache.GetGuild(id)
	if err != nil {
		ret, err = GetGuild(c.req, id)
		if err != nil {
			return
		}
		c.cache.SetGuild(ret)
	}
	return
}

// ModifyGuild .
func (c *Client) ModifyGuild(id Snowflake, params *ModifyGuildParams) (ret *Guild, err error) {
	ret, err = ModifyGuild(c.req, id, params)
	return
}

// DeleteGuild .
func (c *Client) DeleteGuild(id Snowflake) (err error) {
	err = DeleteGuild(c.req, id)
	return
}

// GetGuildChannels .
func (c *Client) GetGuildChannels(id Snowflake) (ret []*Channel, err error) {
	var guild *Guild
	guild, err = c.cache.GetGuild(id)
	if err != nil {
		ret, err = GetGuildChannels(c.req, id)
		if err != nil {
			return
		}
		c.cache.SetGuild(&Guild{
			ID:       id,
			Channels: ret,
		})
	} else {
		ret = guild.Channels
	}
	return
}

// CreateGuildChannel .
func (c *Client) CreateGuildChannel(id Snowflake, params *CreateGuildChannelParams) (ret *Channel, err error) {
	ret, err = CreateGuildChannel(c.req, id, params)
	return
}

// ModifyGuildChannelPositions .
func (c *Client) ModifyGuildChannelPositions(id Snowflake, params []ModifyGuildChannelPositionsParams) (ret *Guild, err error) {
	ret, err = ModifyGuildChannelPositions(c.req, id, params)
	return
}

// GetGuildMember .
func (c *Client) GetGuildMember(guildID, userID Snowflake) (ret *Member, err error) {
	ret, err = c.cache.GetGuildMember(guildID, userID)
	if err != nil {
		ret, err = GetGuildMember(c.req, guildID, userID)
		if err != nil {
			return
		}
		c.cache.SetGuildMember(guildID, ret)
		c.cache.Update(UserCache, ret.User)
	}
	return
}

// GetGuildMembers .
func (c *Client) GetGuildMembers(guildID, after Snowflake, limit int) (ret []*Member, err error) {
	ret, err = c.cache.GetGuildMembersAfter(guildID, after, limit)
	if err != nil {
		ret, err = GetGuildMembers(c.req, guildID, after, limit)
		if err != nil {
			return
		}
		c.cache.SetGuildMembers(guildID, ret)
		//c.cacheLink.Update(UserCache, ret.User)
		// TODO: update users
	}
	return
}

// AddGuildMember .
func (c *Client) AddGuildMember(guildID, userID Snowflake, params *AddGuildMemberParams) (ret *Member, err error) {
	ret, err = AddGuildMember(c.req, guildID, userID, params)
	return
}

// ModifyGuildMember .
func (c *Client) ModifyGuildMember(guildID, userID Snowflake, params *ModifyGuildMemberParams) (err error) {
	err = ModifyGuildMember(c.req, guildID, userID, params)
	return
}

// ModifyCurrentUserNick .
func (c *Client) ModifyCurrentUserNick(id Snowflake, params *ModifyCurrentUserNickParams) (nick string, err error) {
	nick, err = ModifyCurrentUserNick(c.req, id, params)
	return
}

// AddGuildMemberRole .
func (c *Client) AddGuildMemberRole(guildID, userID, roleID Snowflake) (err error) {
	err = AddGuildMemberRole(c.req, guildID, userID, roleID)
	return
}

// RemoveGuildMemberRole .
func (c *Client) RemoveGuildMemberRole(guildID, userID, roleID Snowflake) (err error) {
	err = RemoveGuildMemberRole(c.req, guildID, userID, roleID)
	return
}

// RemoveGuildMember .
func (c *Client) RemoveGuildMember(guildID, userID Snowflake) (err error) {
	err = RemoveGuildMember(c.req, guildID, userID)
	return
}

// GetGuildBans .
func (c *Client) GetGuildBans(id Snowflake) (ret []*Ban, err error) {
	ret, err = GetGuildBans(c.req, id)
	return
}

// GetGuildBan .
func (c *Client) GetGuildBan(guildID, userID Snowflake) (ret *Ban, err error) {
	ret, err = GetGuildBan(c.req, guildID, userID)
	return
}

// CreateGuildBan .
func (c *Client) CreateGuildBan(guildID, userID Snowflake, params *CreateGuildBanParams) (err error) {
	err = CreateGuildBan(c.req, guildID, userID, params)
	return
}

// RemoveGuildBan .
func (c *Client) RemoveGuildBan(guildID, userID Snowflake) (err error) {
	err = RemoveGuildBan(c.req, guildID, userID)
	return
}

// GetGuildRoles .
func (c *Client) GetGuildRoles(guildID Snowflake) (ret []*Role, err error) {
	ret, err = c.cache.GetGuildRoles(guildID)
	if err != nil {
		ret, err = GetGuildRoles(c.req, guildID)
		if err != nil {
			return
		}
		c.cache.SetGuildRoles(guildID, ret)
	}

	return
}

// CreateGuildRole .
func (c *Client) CreateGuildRole(id Snowflake, params *CreateGuildRoleParams) (ret *Role, err error) {
	ret, err = CreateGuildRole(c.req, id, params)
	return
}

// ModifyGuildRolePositions .
func (c *Client) ModifyGuildRolePositions(guildID Snowflake, params []ModifyGuildRolePositionsParams) (ret []*Role, err error) {
	ret, err = ModifyGuildRolePositions(c.req, guildID, params)
	return
}

// ModifyGuildRole .
func (c *Client) ModifyGuildRole(guildID, roleID Snowflake, params *ModifyGuildRoleParams) (ret *Role, err error) {
	ret, err = ModifyGuildRole(c.req, guildID, roleID, params)
	return
}

// DeleteGuildRole .
func (c *Client) DeleteGuildRole(guildID, roleID Snowflake) (err error) {
	err = DeleteGuildRole(c.req, guildID, roleID)
	return
}

// GetGuildPruneCount .
func (c *Client) GetGuildPruneCount(id Snowflake, params *GuildPruneParams) (ret *GuildPruneCount, err error) {
	ret, err = GetGuildPruneCount(c.req, id, params)
	return
}

// BeginGuildPrune .
func (c *Client) BeginGuildPrune(id Snowflake, params *GuildPruneParams) (ret *GuildPruneCount, err error) {
	ret, err = BeginGuildPrune(c.req, id, params)
	return
}

// GetGuildVoiceRegions .
func (c *Client) GetGuildVoiceRegions(id Snowflake) (ret []*VoiceRegion, err error) {
	ret, err = GetGuildVoiceRegions(c.req, id)
	return
}

// GetGuildInvites .
func (c *Client) GetGuildInvites(id Snowflake) (ret []*Invite, err error) {
	ret, err = GetGuildInvites(c.req, id)
	return
}

// GetGuildIntegrations .
func (c *Client) GetGuildIntegrations(id Snowflake) (ret []*Integration, err error) {
	ret, err = GetGuildIntegrations(c.req, id)
	return
}

// CreateGuildIntegration .
func (c *Client) CreateGuildIntegration(guildID Snowflake, params *CreateGuildIntegrationParams) (err error) {
	err = CreateGuildIntegration(c.req, guildID, params)
	return
}

// ModifyGuildIntegration .
func (c *Client) ModifyGuildIntegration(guildID, integrationID Snowflake, params *ModifyGuildIntegrationParams) (err error) {
	err = ModifyGuildIntegration(c.req, guildID, integrationID, params)
	return
}

// DeleteGuildIntegration .
func (c *Client) DeleteGuildIntegration(guildID, integrationID Snowflake) (err error) {
	err = DeleteGuildIntegration(c.req, guildID, integrationID)
	return
}

// SyncGuildIntegration .
func (c *Client) SyncGuildIntegration(guildID, integrationID Snowflake) (err error) {
	err = SyncGuildIntegration(c.req, guildID, integrationID)
	return
}

// GetGuildEmbed .
func (c *Client) GetGuildEmbed(guildID Snowflake) (ret *GuildEmbed, err error) {
	ret, err = GetGuildEmbed(c.req, guildID)
	return
}

// ModifyGuildEmbed .
func (c *Client) ModifyGuildEmbed(guildID Snowflake, params *GuildEmbed) (ret *GuildEmbed, err error) {
	ret, err = ModifyGuildEmbed(c.req, guildID, params)
	return
}

// GetGuildVanityURL .
func (c *Client) GetGuildVanityURL(guildID Snowflake) (ret *PartialInvite, err error) {
	ret, err = GetGuildVanityURL(c.req, guildID)
	return
}

// Webhook

// CreateWebhook .
func (c *Client) CreateWebhook(channelID Snowflake, params *CreateWebhookParams) (ret *Webhook, err error) {
	ret, err = CreateWebhook(c.req, channelID, params)
	return
}

// GetChannelWebhooks .
func (c *Client) GetChannelWebhooks(channelID Snowflake) (ret []*Webhook, err error) {
	ret, err = GetChannelWebhooks(c.req, channelID)
	return
}

// GetGuildWebhooks .
func (c *Client) GetGuildWebhooks(guildID Snowflake) (ret []*Webhook, err error) {
	ret, err = GetGuildWebhooks(c.req, guildID)
	return
}

// GetWebhook .
func (c *Client) GetWebhook(id Snowflake) (ret *Webhook, err error) {
	ret, err = GetWebhook(c.req, id)
	return
}

// GetWebhookWithToken .
func (c *Client) GetWebhookWithToken(id Snowflake, token string) (ret *Webhook, err error) {
	ret, err = GetWebhookWithToken(c.req, id, token)
	return
}

// ModifyWebhook .
func (c *Client) ModifyWebhook(id Snowflake, params *ModifyWebhookParams) (ret *Webhook, err error) {
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
		err = errors.New("given avatar string is invalid. Must specify data encoding. Eg. data:image/jpeg;base64,")
		return
	}

	// TODO: check if user has permission to modify webhook
	ret, err = ModifyWebhook(c.req, id, params)
	return
}

// ModifyWebhookWithToken .
func (c *Client) ModifyWebhookWithToken(newWebhook *Webhook) (ret *Webhook, err error) {
	ret, err = ModifyWebhookWithToken(c.req, newWebhook)
	return
}

// DeleteWebhook .
func (c *Client) DeleteWebhook(webhookID Snowflake) (err error) {
	err = DeleteWebhook(c.req, webhookID)
	return
}

// DeleteWebhookWithToken .
func (c *Client) DeleteWebhookWithToken(id Snowflake, token string) (err error) {
	err = DeleteWebhookWithToken(c.req, id, token)
	return
}

// ExecuteWebhook .
func (c *Client) ExecuteWebhook(params *ExecuteWebhookParams, wait bool, URLSuffix string) (err error) {
	err = ExecuteWebhook(c.req, params, wait, URLSuffix)
	return
}

// ExecuteSlackWebhook .
func (c *Client) ExecuteSlackWebhook(params *ExecuteWebhookParams, wait bool) (err error) {
	err = ExecuteSlackWebhook(c.req, params, wait)
	return
}

// ExecuteGitHubWebhook .
func (c *Client) ExecuteGitHubWebhook(params *ExecuteWebhookParams, wait bool) (err error) {
	err = ExecuteGitHubWebhook(c.req, params, wait)
	return
}

//
// #########################################################################
//
//
// Custom methods are usually reused by the resource package for readability
// -----

func (c *Client) GetGuilds(params *GetCurrentUserGuildsParams) ([]*Guild, error) {
	return c.GetCurrentUserGuilds(params)
}

// SendMsg .
func (c *Client) SendMsg(channelID Snowflake, message *Message) (msg *Message, err error) {
	if constant.LockedMethods {
		message.RLock()
	}
	params := &CreateChannelMessageParams{
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

	return c.CreateChannelMessage(channelID, params)
}

// SendMsgString .
func (c *Client) SendMsgString(channelID Snowflake, content string) (msg *Message, err error) {
	params := &CreateChannelMessageParams{
		Content: content,
	}

	msg, err = c.CreateChannelMessage(channelID, params)
	return
}

// UpdateMessage .
func (c *Client) UpdateMessage(message *Message) (msg *Message, err error) {
	if constant.LockedMethods {
		message.RLock()
		defer message.RUnlock()
	}

	params := &EditMessageParams{
		Content: message.Content,
	}
	if len(message.Embeds) > 0 {
		params.Embed = message.Embeds[0]
	}

	msg, err = c.EditMessage(message.ChannelID, message.ID, params)
	return
}

// UpdateChannel Not implemented yet
func (c *Client) UpdateChannel(channel *Channel) (err error) {
	// there are several different REST calls that needs to be made in order
	// to update the channel. But how exactly do we know what has changed?
	return errors.New("not implemented")
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
func (c *Client) UpdateStatus(s *UpdateStatusCommand) error {
	return c.Emit(CommandUpdateStatus, s)
}

// UpdateStatusString sets the client's game activity to the provided string, status to online
// and type to Playing
func (c *Client) UpdateStatusString(s string) error {
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
func (c *Client) eventHandler() {
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

		err = unmarshal(evt.Data, box)
		if err != nil {
			c.Error(err.Error())
			continue // ignore event
			// TODO: if an event is ignored, should it not at least send a signal for listeners with no parameters?
		}

		fmt.Println("Got event", evt.Name)
		if updater, implements := box.(internalUpdater); implements {
			updater.updateInternals()
		}
		if updater, implements := box.(internalClientUpdater); implements {
			updater.updateInternalsWithClient(c)
		}

		// cacheLink
		if !c.config.DisableCache {
			cacheEvent(c.cache, evt.Name, box)
		}

		// voice
		switch evt.Name {
		case EventVoiceStateUpdate:
			c.voiceRepository.onVoiceStateUpdate(box.(*VoiceStateUpdate))
		case EventVoiceServerUpdate:
			c.voiceRepository.onVoiceServerUpdate(box.(*VoiceServerUpdate))
		}

		// trigger listeners
		prepareBox(evt.Name, box)
		c.evtDispatch.triggerChan(ctx, evt.Name, c, box)
		go c.evtDispatch.triggerHandlers(ctx, evt.Name, c, box)
	}
}
