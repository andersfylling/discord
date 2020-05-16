package disgord

// Code generated - This file has been automatically generated by generate/events/main.go - DO NOT EDIT.
// Warning: This file is overwritten at "go generate", instead adapt events.go and event/events.go and run go generate

import (
	"context"

	"github.com/andersfylling/disgord/internal/event"
)

func AllEvents(except ...string) []string {
	evtsMap := map[string]int8{

		EvtChannelCreate: 0,

		EvtChannelDelete: 0,

		EvtChannelPinsUpdate: 0,

		EvtChannelUpdate: 0,

		EvtGuildBanAdd: 0,

		EvtGuildBanRemove: 0,

		EvtGuildCreate: 0,

		EvtGuildDelete: 0,

		EvtGuildEmojisUpdate: 0,

		EvtGuildIntegrationsUpdate: 0,

		EvtGuildMemberAdd: 0,

		EvtGuildMemberRemove: 0,

		EvtGuildMemberUpdate: 0,

		EvtGuildMembersChunk: 0,

		EvtGuildRoleCreate: 0,

		EvtGuildRoleDelete: 0,

		EvtGuildRoleUpdate: 0,

		EvtGuildUpdate: 0,

		EvtInviteCreate: 0,

		EvtInviteDelete: 0,

		EvtMessageCreate: 0,

		EvtMessageDelete: 0,

		EvtMessageDeleteBulk: 0,

		EvtMessageReactionAdd: 0,

		EvtMessageReactionRemove: 0,

		EvtMessageReactionRemoveAll: 0,

		EvtMessageUpdate: 0,

		EvtPresenceUpdate: 0,

		EvtReady: 0,

		EvtResumed: 0,

		EvtTypingStart: 0,

		EvtUserUpdate: 0,

		EvtVoiceServerUpdate: 0,

		EvtVoiceStateUpdate: 0,

		EvtWebhooksUpdate: 0,
	}

	for i := range except {
		delete(evtsMap, except[i])
	}

	evts := make([]string, 0, len(evtsMap))
	for k := range evtsMap {
		evts = append(evts, k)
	}
	return evts
}

// ---------------------------

// EvtChannelCreate Sent when a new channel is created, relevant to the current user. The inner payload is a DM channel or
// guild channel object.
//
const EvtChannelCreate = event.ChannelCreate

func (h *ChannelCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelCreate) setShardID(id uint)                  { h.ShardID = id }

type HandlerChannelCreate = func(Session, *ChannelCreate)

func (c *Client) OnChannelCreate(mdlws []Middleware, handlers []HandlerChannelCreate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtChannelCreate, inputs...)
}

// ---------------------------

// EvtChannelDelete Sent when a channel relevant to the current user is deleted. The inner payload is a DM or Guild channel object.
//
const EvtChannelDelete = event.ChannelDelete

func (h *ChannelDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelDelete) setShardID(id uint)                  { h.ShardID = id }

type HandlerChannelDelete = func(Session, *ChannelDelete)

func (c *Client) OnChannelDelete(mdlws []Middleware, handlers []HandlerChannelDelete, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtChannelDelete, inputs...)
}

// ---------------------------

// EvtChannelPinsUpdate Sent when a message is pinned or unpinned in a text channel. This is not sent when a pinned message is deleted.
//  Fields:
//  - ChannelID int64 or Snowflake
//  - LastPinTimestamp time.Now().UTC().Format(time.RFC3339)
// TODO fix.
//
const EvtChannelPinsUpdate = event.ChannelPinsUpdate

func (h *ChannelPinsUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelPinsUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerChannelPinsUpdate = func(Session, *ChannelPinsUpdate)

func (c *Client) OnChannelPinsUpdate(mdlws []Middleware, handlers []HandlerChannelPinsUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtChannelPinsUpdate, inputs...)
}

// ---------------------------

// EvtChannelUpdate Sent when a channel is updated. The inner payload is a guild channel object.
//
const EvtChannelUpdate = event.ChannelUpdate

func (h *ChannelUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerChannelUpdate = func(Session, *ChannelUpdate)

func (c *Client) OnChannelUpdate(mdlws []Middleware, handlers []HandlerChannelUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtChannelUpdate, inputs...)
}

// ---------------------------

// EvtGuildBanAdd Sent when a user is banned from a guild. The inner payload is a user object, with an extra guild_id key.
//
const EvtGuildBanAdd = event.GuildBanAdd

func (h *GuildBanAdd) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildBanAdd) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildBanAdd = func(Session, *GuildBanAdd)

func (c *Client) OnGuildBanAdd(mdlws []Middleware, handlers []HandlerGuildBanAdd, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildBanAdd, inputs...)
}

// ---------------------------

// EvtGuildBanRemove Sent when a user is unbanned from a guild. The inner payload is a user object, with an extra guild_id key.
//
const EvtGuildBanRemove = event.GuildBanRemove

func (h *GuildBanRemove) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildBanRemove) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildBanRemove = func(Session, *GuildBanRemove)

func (c *Client) OnGuildBanRemove(mdlws []Middleware, handlers []HandlerGuildBanRemove, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildBanRemove, inputs...)
}

// ---------------------------

// EvtGuildCreate This event can be sent in three different scenarios:
//  1. When a user is initially connecting, to lazily load and backfill information for all unavailable guilds
//     sent in the Ready event.
// 	2. When a Guild becomes available again to the client.
// 	3. When the current user joins a new Guild.
//
const EvtGuildCreate = event.GuildCreate

func (h *GuildCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildCreate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildCreate = func(Session, *GuildCreate)

func (c *Client) OnGuildCreate(mdlws []Middleware, handlers []HandlerGuildCreate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildCreate, inputs...)
}

// ---------------------------

// EvtGuildDelete Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
// The inner payload is an unavailable guild object. If the unavailable field is not set, the user was removed
// from the guild.
//
const EvtGuildDelete = event.GuildDelete

func (h *GuildDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildDelete) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildDelete = func(Session, *GuildDelete)

func (c *Client) OnGuildDelete(mdlws []Middleware, handlers []HandlerGuildDelete, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildDelete, inputs...)
}

// ---------------------------

// EvtGuildEmojisUpdate Sent when a guild's emojis have been updated.
//  Fields:
//  - GuildID Snowflake
//  - Emojis []*Emoji
//
const EvtGuildEmojisUpdate = event.GuildEmojisUpdate

func (h *GuildEmojisUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildEmojisUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildEmojisUpdate = func(Session, *GuildEmojisUpdate)

func (c *Client) OnGuildEmojisUpdate(mdlws []Middleware, handlers []HandlerGuildEmojisUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildEmojisUpdate, inputs...)
}

// ---------------------------

// EvtGuildIntegrationsUpdate Sent when a guild integration is updated.
//  Fields:
//  - GuildID Snowflake
//
const EvtGuildIntegrationsUpdate = event.GuildIntegrationsUpdate

func (h *GuildIntegrationsUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildIntegrationsUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildIntegrationsUpdate = func(Session, *GuildIntegrationsUpdate)

func (c *Client) OnGuildIntegrationsUpdate(mdlws []Middleware, handlers []HandlerGuildIntegrationsUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildIntegrationsUpdate, inputs...)
}

// ---------------------------

// EvtGuildMemberAdd Sent when a new user joins a guild. The inner payload is a guild member object with these extra fields:
//  - GuildID Snowflake
//
//  Fields:
//  - Member *Member
//
const EvtGuildMemberAdd = event.GuildMemberAdd

func (h *GuildMemberAdd) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMemberAdd) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildMemberAdd = func(Session, *GuildMemberAdd)

func (c *Client) OnGuildMemberAdd(mdlws []Middleware, handlers []HandlerGuildMemberAdd, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildMemberAdd, inputs...)
}

// ---------------------------

// EvtGuildMemberRemove Sent when a user is removed from a guild (leave/kick/ban).
//  Fields:
//  - GuildID   Snowflake
//  - User      *User
//
const EvtGuildMemberRemove = event.GuildMemberRemove

func (h *GuildMemberRemove) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMemberRemove) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildMemberRemove = func(Session, *GuildMemberRemove)

func (c *Client) OnGuildMemberRemove(mdlws []Middleware, handlers []HandlerGuildMemberRemove, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildMemberRemove, inputs...)
}

// ---------------------------

// EvtGuildMemberUpdate Sent when a guild member is updated.
//  Fields:
//  - GuildID   Snowflake
//  - Roles     []Snowflake
//  - User      *User
//  - Nick      string
//
const EvtGuildMemberUpdate = event.GuildMemberUpdate

func (h *GuildMemberUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMemberUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildMemberUpdate = func(Session, *GuildMemberUpdate)

func (c *Client) OnGuildMemberUpdate(mdlws []Middleware, handlers []HandlerGuildMemberUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildMemberUpdate, inputs...)
}

// ---------------------------

// EvtGuildMembersChunk Sent in response to Gateway Request Guild Members.
//  Fields:
//  - GuildID Snowflake
//  - Members []*Member
//
const EvtGuildMembersChunk = event.GuildMembersChunk

func (h *GuildMembersChunk) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMembersChunk) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildMembersChunk = func(Session, *GuildMembersChunk)

func (c *Client) OnGuildMembersChunk(mdlws []Middleware, handlers []HandlerGuildMembersChunk, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildMembersChunk, inputs...)
}

// ---------------------------

// EvtGuildRoleCreate Sent when a guild role is created.
//  Fields:
//  - GuildID   Snowflake
//  - Role      *Role
//
const EvtGuildRoleCreate = event.GuildRoleCreate

func (h *GuildRoleCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleCreate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildRoleCreate = func(Session, *GuildRoleCreate)

func (c *Client) OnGuildRoleCreate(mdlws []Middleware, handlers []HandlerGuildRoleCreate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildRoleCreate, inputs...)
}

// ---------------------------

// EvtGuildRoleDelete Sent when a guild role is created.
//  Fields:
//  - GuildID Snowflake
//  - RoleID  Snowflake
//
const EvtGuildRoleDelete = event.GuildRoleDelete

func (h *GuildRoleDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleDelete) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildRoleDelete = func(Session, *GuildRoleDelete)

func (c *Client) OnGuildRoleDelete(mdlws []Middleware, handlers []HandlerGuildRoleDelete, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildRoleDelete, inputs...)
}

// ---------------------------

// EvtGuildRoleUpdate Sent when a guild role is created.
//  Fields:
//  - GuildID Snowflake
//  - Role    *Role
//
const EvtGuildRoleUpdate = event.GuildRoleUpdate

func (h *GuildRoleUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildRoleUpdate = func(Session, *GuildRoleUpdate)

func (c *Client) OnGuildRoleUpdate(mdlws []Middleware, handlers []HandlerGuildRoleUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildRoleUpdate, inputs...)
}

// ---------------------------

// EvtGuildUpdate Sent when a guild is updated. The inner payload is a guild object.
//
const EvtGuildUpdate = event.GuildUpdate

func (h *GuildUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerGuildUpdate = func(Session, *GuildUpdate)

func (c *Client) OnGuildUpdate(mdlws []Middleware, handlers []HandlerGuildUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtGuildUpdate, inputs...)
}

// ---------------------------

// EvtInviteCreate Sent when a guild's invite is created.
//  Fields:
//  - Code String
//  - Guild   *Guild
//  - Channel *Channel
//  - Inviter *User
//  - CreatedAt Time
//  - MaxAge int
//  - MaxUses int
//  - Temporary bool
//  - Uses int
//  - Revoked bool
//  - Unique bool
//  - ApproximatePresenceCount int
//  - ApproximateMemberCount int
//
const EvtInviteCreate = event.InviteCreate

func (h *InviteCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *InviteCreate) setShardID(id uint)                  { h.ShardID = id }

type HandlerInviteCreate = func(Session, *InviteCreate)

func (c *Client) OnInviteCreate(mdlws []Middleware, handlers []HandlerInviteCreate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtInviteCreate, inputs...)
}

// ---------------------------

// EvtInviteDelete Sent when an invite is deleted.
//
const EvtInviteDelete = event.InviteDelete

func (h *InviteDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *InviteDelete) setShardID(id uint)                  { h.ShardID = id }

type HandlerInviteDelete = func(Session, *InviteDelete)

func (c *Client) OnInviteDelete(mdlws []Middleware, handlers []HandlerInviteDelete, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtInviteDelete, inputs...)
}

// ---------------------------

// EvtMessageCreate Sent when a message is created. The inner payload is a message object.
//
const EvtMessageCreate = event.MessageCreate

func (h *MessageCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageCreate) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageCreate = func(Session, *MessageCreate)

func (c *Client) OnMessageCreate(mdlws []Middleware, handlers []HandlerMessageCreate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageCreate, inputs...)
}

// ---------------------------

// EvtMessageDelete Sent when a message is deleted.
//  Fields:
//  - ID        Snowflake
//  - ChannelID Snowflake
//
const EvtMessageDelete = event.MessageDelete

func (h *MessageDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageDelete) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageDelete = func(Session, *MessageDelete)

func (c *Client) OnMessageDelete(mdlws []Middleware, handlers []HandlerMessageDelete, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageDelete, inputs...)
}

// ---------------------------

// EvtMessageDeleteBulk Sent when multiple messages are deleted at once.
//  Fields:
//  - IDs       []Snowflake
//  - ChannelID Snowflake
//
const EvtMessageDeleteBulk = event.MessageDeleteBulk

func (h *MessageDeleteBulk) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageDeleteBulk) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageDeleteBulk = func(Session, *MessageDeleteBulk)

func (c *Client) OnMessageDeleteBulk(mdlws []Middleware, handlers []HandlerMessageDeleteBulk, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageDeleteBulk, inputs...)
}

// ---------------------------

// EvtMessageReactionAdd Sent when a user adds a reaction to a message.
//  Fields:
//  - UserID     Snowflake
//  - ChannelID  Snowflake
//  - MessageID  Snowflake
//  - Emoji      *Emoji
//
const EvtMessageReactionAdd = event.MessageReactionAdd

func (h *MessageReactionAdd) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageReactionAdd) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageReactionAdd = func(Session, *MessageReactionAdd)

func (c *Client) OnMessageReactionAdd(mdlws []Middleware, handlers []HandlerMessageReactionAdd, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageReactionAdd, inputs...)
}

// ---------------------------

// EvtMessageReactionRemove Sent when a user removes a reaction from a message.
//  Fields:
//  - UserID     Snowflake
//  - ChannelID  Snowflake
//  - MessageID  Snowflake
//  - Emoji      *Emoji
//
const EvtMessageReactionRemove = event.MessageReactionRemove

func (h *MessageReactionRemove) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageReactionRemove) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageReactionRemove = func(Session, *MessageReactionRemove)

func (c *Client) OnMessageReactionRemove(mdlws []Middleware, handlers []HandlerMessageReactionRemove, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageReactionRemove, inputs...)
}

// ---------------------------

// EvtMessageReactionRemoveAll Sent when a user explicitly removes all reactions from a message.
//  Fields:
//  - ChannelID Snowflake
//  - MessageID Snowflake
//
const EvtMessageReactionRemoveAll = event.MessageReactionRemoveAll

func (h *MessageReactionRemoveAll) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageReactionRemoveAll) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageReactionRemoveAll = func(Session, *MessageReactionRemoveAll)

func (c *Client) OnMessageReactionRemoveAll(mdlws []Middleware, handlers []HandlerMessageReactionRemoveAll, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageReactionRemoveAll, inputs...)
}

// ---------------------------

// EvtMessageUpdate Sent when a message is updated. The inner payload is a message object.
//
// NOTE! Has _at_least_ the GuildID and ChannelID fields.
//
const EvtMessageUpdate = event.MessageUpdate

func (h *MessageUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerMessageUpdate = func(Session, *MessageUpdate)

func (c *Client) OnMessageUpdate(mdlws []Middleware, handlers []HandlerMessageUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtMessageUpdate, inputs...)
}

// ---------------------------

// EvtPresenceUpdate A user's presence is their current state on a guild. This event is sent when a user's presence is updated for a guild.
//  Fields:
//  - User    *User
//  - Roles   []Snowflake
//  - Game    *Activity
//  - GuildID Snowflake
//  - Status  string
//
const EvtPresenceUpdate = event.PresenceUpdate

func (h *PresenceUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *PresenceUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerPresenceUpdate = func(Session, *PresenceUpdate)

func (c *Client) OnPresenceUpdate(mdlws []Middleware, handlers []HandlerPresenceUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtPresenceUpdate, inputs...)
}

// ---------------------------

// EvtReady The ready event is dispatched when a client has completed the initial handshake with the gateway (for new sessions).
// // The ready event can be the largest and most complex event the gateway will send, as it contains all the state
// // required for a client to begin interacting with the rest of the platform.
// //  Fields:
// //  - V int
// //  - User *User
// //  - PrivateChannels []*Channel
// //  - Guilds []*GuildUnavailable
// //  - SessionID string
// //  - Trace []string
//
const EvtReady = event.Ready

func (h *Ready) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *Ready) setShardID(id uint)                  { h.ShardID = id }

type HandlerReady = func(Session, *Ready)

func (c *Client) OnReady(mdlws []Middleware, handlers []HandlerReady, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtReady, inputs...)
}

// ---------------------------

// EvtResumed The resumed event is dispatched when a client has sent a resume payload to the gateway
// (for resuming existing sessions).
//  Fields:
//  - Trace []string
//
const EvtResumed = event.Resumed

func (h *Resumed) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *Resumed) setShardID(id uint)                  { h.ShardID = id }

type HandlerResumed = func(Session, *Resumed)

func (c *Client) OnResumed(mdlws []Middleware, handlers []HandlerResumed, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtResumed, inputs...)
}

// ---------------------------

// EvtTypingStart Sent when a user starts typing in a channel.
//  Fields:
//  - ChannelID     Snowflake
//  - UserID        Snowflake
//  - TimestampUnix int
//
const EvtTypingStart = event.TypingStart

func (h *TypingStart) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *TypingStart) setShardID(id uint)                  { h.ShardID = id }

type HandlerTypingStart = func(Session, *TypingStart)

func (c *Client) OnTypingStart(mdlws []Middleware, handlers []HandlerTypingStart, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtTypingStart, inputs...)
}

// ---------------------------

// EvtUserUpdate Sent when properties about the user change. Inner payload is a user object.
//
const EvtUserUpdate = event.UserUpdate

func (h *UserUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *UserUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerUserUpdate = func(Session, *UserUpdate)

func (c *Client) OnUserUpdate(mdlws []Middleware, handlers []HandlerUserUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtUserUpdate, inputs...)
}

// ---------------------------

// EvtVoiceServerUpdate Sent when a guild's voice server is updated. This is sent when initially connecting to voice, and when the current
// voice instance fails over to a new server.
//  Fields:
//  - Token     string
//  - ChannelID Snowflake
//  - Endpoint  string
//
const EvtVoiceServerUpdate = event.VoiceServerUpdate

func (h *VoiceServerUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *VoiceServerUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerVoiceServerUpdate = func(Session, *VoiceServerUpdate)

func (c *Client) OnVoiceServerUpdate(mdlws []Middleware, handlers []HandlerVoiceServerUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtVoiceServerUpdate, inputs...)
}

// ---------------------------

// EvtVoiceStateUpdate Sent when someone joins/leaves/moves voice channels. Inner payload is a voice state object.
//
const EvtVoiceStateUpdate = event.VoiceStateUpdate

func (h *VoiceStateUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *VoiceStateUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerVoiceStateUpdate = func(Session, *VoiceStateUpdate)

func (c *Client) OnVoiceStateUpdate(mdlws []Middleware, handlers []HandlerVoiceStateUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtVoiceStateUpdate, inputs...)
}

// ---------------------------

// EvtWebhooksUpdate Sent when a guild channel's WebHook is created, updated, or deleted.
//  Fields:
//  - GuildID   Snowflake
//  - ChannelID Snowflake
//
const EvtWebhooksUpdate = event.WebhooksUpdate

func (h *WebhooksUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *WebhooksUpdate) setShardID(id uint)                  { h.ShardID = id }

type HandlerWebhooksUpdate = func(Session, *WebhooksUpdate)

func (c *Client) OnWebhooksUpdate(mdlws []Middleware, handlers []HandlerWebhooksUpdate, ctrl ...HandlerCtrl) {
	var inputs []interface{}
	for mdlw := range mdlws {
		inputs = append(inputs, mdlw)
	}
	for handler := range handlers {
		inputs = append(inputs, handler)
	}
	if len(ctrl) > 0 {
		inputs = append(inputs, ctrl[0])
	}

	c.On(EvtWebhooksUpdate, inputs...)
}

// ---------------------------

type SocketHandlerRegistrators interface {
	OnChannelCreate([]Middleware, []HandlerChannelCreate, ...HandlerCtrl)
	OnChannelDelete([]Middleware, []HandlerChannelDelete, ...HandlerCtrl)
	OnChannelPinsUpdate([]Middleware, []HandlerChannelPinsUpdate, ...HandlerCtrl)
	OnChannelUpdate([]Middleware, []HandlerChannelUpdate, ...HandlerCtrl)
	OnGuildBanAdd([]Middleware, []HandlerGuildBanAdd, ...HandlerCtrl)
	OnGuildBanRemove([]Middleware, []HandlerGuildBanRemove, ...HandlerCtrl)
	OnGuildCreate([]Middleware, []HandlerGuildCreate, ...HandlerCtrl)
	OnGuildDelete([]Middleware, []HandlerGuildDelete, ...HandlerCtrl)
	OnGuildEmojisUpdate([]Middleware, []HandlerGuildEmojisUpdate, ...HandlerCtrl)
	OnGuildIntegrationsUpdate([]Middleware, []HandlerGuildIntegrationsUpdate, ...HandlerCtrl)
	OnGuildMemberAdd([]Middleware, []HandlerGuildMemberAdd, ...HandlerCtrl)
	OnGuildMemberRemove([]Middleware, []HandlerGuildMemberRemove, ...HandlerCtrl)
	OnGuildMemberUpdate([]Middleware, []HandlerGuildMemberUpdate, ...HandlerCtrl)
	OnGuildMembersChunk([]Middleware, []HandlerGuildMembersChunk, ...HandlerCtrl)
	OnGuildRoleCreate([]Middleware, []HandlerGuildRoleCreate, ...HandlerCtrl)
	OnGuildRoleDelete([]Middleware, []HandlerGuildRoleDelete, ...HandlerCtrl)
	OnGuildRoleUpdate([]Middleware, []HandlerGuildRoleUpdate, ...HandlerCtrl)
	OnGuildUpdate([]Middleware, []HandlerGuildUpdate, ...HandlerCtrl)
	OnInviteCreate([]Middleware, []HandlerInviteCreate, ...HandlerCtrl)
	OnInviteDelete([]Middleware, []HandlerInviteDelete, ...HandlerCtrl)
	OnMessageCreate([]Middleware, []HandlerMessageCreate, ...HandlerCtrl)
	OnMessageDelete([]Middleware, []HandlerMessageDelete, ...HandlerCtrl)
	OnMessageDeleteBulk([]Middleware, []HandlerMessageDeleteBulk, ...HandlerCtrl)
	OnMessageReactionAdd([]Middleware, []HandlerMessageReactionAdd, ...HandlerCtrl)
	OnMessageReactionRemove([]Middleware, []HandlerMessageReactionRemove, ...HandlerCtrl)
	OnMessageReactionRemoveAll([]Middleware, []HandlerMessageReactionRemoveAll, ...HandlerCtrl)
	OnMessageUpdate([]Middleware, []HandlerMessageUpdate, ...HandlerCtrl)
	OnPresenceUpdate([]Middleware, []HandlerPresenceUpdate, ...HandlerCtrl)
	OnReady([]Middleware, []HandlerReady, ...HandlerCtrl)
	OnResumed([]Middleware, []HandlerResumed, ...HandlerCtrl)
	OnTypingStart([]Middleware, []HandlerTypingStart, ...HandlerCtrl)
	OnUserUpdate([]Middleware, []HandlerUserUpdate, ...HandlerCtrl)
	OnVoiceServerUpdate([]Middleware, []HandlerVoiceServerUpdate, ...HandlerCtrl)
	OnVoiceStateUpdate([]Middleware, []HandlerVoiceStateUpdate, ...HandlerCtrl)
	OnWebhooksUpdate([]Middleware, []HandlerWebhooksUpdate, ...HandlerCtrl)
}
