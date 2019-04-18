package disgord

// Code generated - This file has been automatically generated by generate/events/main.go - DO NOT EDIT.
// Warning: This file is overwritten at "go generate", instead adapt events.go and event/events.go and run go generate

import (
	"context"

	"github.com/andersfylling/disgord/event"
)

// ---------------------------

// EvtChannelCreate Sent when a new channel is created, relevant to the current user. The inner payload is a DM channel or
// guild channel object.
//
const EvtChannelCreate = event.ChannelCreate

func (h *ChannelCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelCreate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtChannelDelete Sent when a channel relevant to the current user is deleted. The inner payload is a DM or Guild channel object.
//
const EvtChannelDelete = event.ChannelDelete

func (h *ChannelDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelDelete) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtChannelUpdate Sent when a channel is updated. The inner payload is a guild channel object.
//
const EvtChannelUpdate = event.ChannelUpdate

func (h *ChannelUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *ChannelUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildBanAdd Sent when a user is banned from a guild. The inner payload is a user object, with an extra guild_id key.
//
const EvtGuildBanAdd = event.GuildBanAdd

func (h *GuildBanAdd) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildBanAdd) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildBanRemove Sent when a user is unbanned from a guild. The inner payload is a user object, with an extra guild_id key.
//
const EvtGuildBanRemove = event.GuildBanRemove

func (h *GuildBanRemove) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildBanRemove) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtGuildDelete Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
// The inner payload is an unavailable guild object. If the unavailable field is not set, the user was removed
// from the guild.
//
const EvtGuildDelete = event.GuildDelete

func (h *GuildDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildDelete) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildEmojisUpdate Sent when a guild's emojis have been updated.
//  Fields:
//  - GuildID Snowflake
//  - Emojis []*Emoji
//
const EvtGuildEmojisUpdate = event.GuildEmojisUpdate

func (h *GuildEmojisUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildEmojisUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildIntegrationsUpdate Sent when a guild integration is updated.
//  Fields:
//  - GuildID Snowflake
//
const EvtGuildIntegrationsUpdate = event.GuildIntegrationsUpdate

func (h *GuildIntegrationsUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildIntegrationsUpdate) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtGuildMemberRemove Sent when a user is removed from a guild (leave/kick/ban).
//  Fields:
//  - GuildID   Snowflake
//  - User      *User
//
const EvtGuildMemberRemove = event.GuildMemberRemove

func (h *GuildMemberRemove) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMemberRemove) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtGuildMembersChunk Sent in response to Gateway Request Guild Members.
//  Fields:
//  - GuildID Snowflake
//  - Members []*Member
//
const EvtGuildMembersChunk = event.GuildMembersChunk

func (h *GuildMembersChunk) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildMembersChunk) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildRoleCreate Sent when a guild role is created.
//  Fields:
//  - GuildID   Snowflake
//  - Role      *Role
//
const EvtGuildRoleCreate = event.GuildRoleCreate

func (h *GuildRoleCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleCreate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildRoleDelete Sent when a guild role is created.
//  Fields:
//  - GuildID Snowflake
//  - RoleID  Snowflake
//
const EvtGuildRoleDelete = event.GuildRoleDelete

func (h *GuildRoleDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleDelete) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildRoleUpdate Sent when a guild role is created.
//  Fields:
//  - GuildID Snowflake
//  - Role    *Role
//
const EvtGuildRoleUpdate = event.GuildRoleUpdate

func (h *GuildRoleUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildRoleUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtGuildUpdate Sent when a guild is updated. The inner payload is a guild object.
//
const EvtGuildUpdate = event.GuildUpdate

func (h *GuildUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *GuildUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtMessageCreate Sent when a message is created. The inner payload is a message object.
//
const EvtMessageCreate = event.MessageCreate

func (h *MessageCreate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageCreate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtMessageDelete Sent when a message is deleted.
//  Fields:
//  - ID        Snowflake
//  - ChannelID Snowflake
//
const EvtMessageDelete = event.MessageDelete

func (h *MessageDelete) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageDelete) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtMessageDeleteBulk Sent when multiple messages are deleted at once.
//  Fields:
//  - IDs       []Snowflake
//  - ChannelID Snowflake
//
const EvtMessageDeleteBulk = event.MessageDeleteBulk

func (h *MessageDeleteBulk) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageDeleteBulk) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtMessageReactionRemoveAll Sent when a user explicitly removes all reactions from a message.
//  Fields:
//  - ChannelID Snowflake
//  - MessageID Snowflake
//
const EvtMessageReactionRemoveAll = event.MessageReactionRemoveAll

func (h *MessageReactionRemoveAll) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageReactionRemoveAll) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtMessageUpdate Sent when a message is updated. The inner payload is a message object.
//
// NOTE! Has _at_least_ the GuildID and ChannelID fields.
//
const EvtMessageUpdate = event.MessageUpdate

func (h *MessageUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *MessageUpdate) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtPresencesReplace Holds and array of presence update objects
//
const EvtPresencesReplace = event.PresencesReplace

func (h *PresencesReplace) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *PresencesReplace) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtResumed The resumed event is dispatched when a client has sent a resume payload to the gateway
// (for resuming existing sessions).
//  Fields:
//  - Trace []string
//
const EvtResumed = event.Resumed

func (h *Resumed) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *Resumed) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtUserUpdate Sent when properties about the user change. Inner payload is a user object.
//
const EvtUserUpdate = event.UserUpdate

func (h *UserUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *UserUpdate) setShardID(id uint)                  { h.ShardID = id }

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

// ---------------------------

// EvtVoiceStateUpdate Sent when someone joins/leaves/moves voice channels. Inner payload is a voice state object.
//
const EvtVoiceStateUpdate = event.VoiceStateUpdate

func (h *VoiceStateUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *VoiceStateUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------

// EvtWebhooksUpdate Sent when a guild channel's webhook is created, updated, or deleted.
//  Fields:
//  - GuildID   Snowflake
//  - ChannelID Snowflake
//
const EvtWebhooksUpdate = event.WebhooksUpdate

func (h *WebhooksUpdate) registerContext(ctx context.Context) { h.Ctx = ctx }
func (h *WebhooksUpdate) setShardID(id uint)                  { h.ShardID = id }

// ---------------------------
