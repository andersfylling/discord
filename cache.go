package disgord

import (
	"sync"

	"github.com/andersfylling/disgord/internal/crs"
)

type idHolder struct {
	ID      Snowflake `json:"id"`
	Channel struct {
		ID Snowflake `json:"id"`
	} `json:"channel"`
	Guild struct {
		ID Snowflake `json:"id"`
	} `json:"guild"`
	User struct {
		ID Snowflake `json:"id"`
	} `json:"user"`
	UserID    Snowflake `json:"user_id"`
	GuildID   Snowflake `json:"guild_id"`
	ChannelID Snowflake `json:"channel_id"`
}

func newSimpleLFUCache(limitUsers, limitVoiceStates, limitChannels, limitGuilds uint) Cache {
	lfus := &CacheLFU{
		CurrentUser: &User{},
	}
	crs.SetLimit(&lfus.Users, limitUsers)
	crs.SetLimit(&lfus.VoiceStates, limitVoiceStates)
	crs.SetLimit(&lfus.Channels, limitChannels)
	crs.SetLimit(&lfus.Guilds, limitGuilds)

	return lfus
}

// CacheLFU cache with CRS support for Users and voice states
type CacheLFU struct {
	CacheNop

	CurrentUserMu sync.Mutex
	CurrentUser   *User

	Users       crs.LFU
	VoiceStates crs.LFU
	Channels    crs.LFU
	Guilds      crs.LFU
}

var _ Cache = (*CacheLFU)(nil)

func (c *CacheLFU) Ready(data []byte) (*Ready, error) {
	c.CurrentUserMu.Lock()
	defer c.CurrentUserMu.Unlock()

	rdy := &Ready{
		User: c.CurrentUser,
	}

	err := c.unmarshalUpdate(data, rdy)
	return rdy, err
}

func (c *CacheLFU) ChannelCreate(data []byte) (*ChannelCreate, error) {
	// assumption#1: Create may take place after an update to the channel
	// assumption#2: The set of fields in both ChannelCreate and ChannelUpdate are the same
	// assumption#3: a channel can not change from one type to another (text => news, text => voice)

	wrap := func(c *Channel) *ChannelCreate {
		return &ChannelCreate{Channel: c}
	}

	var channel *Channel
	if err := c.unmarshalUpdate(data, &channel); err != nil {
		return nil, err
	}

	c.Channels.Lock()
	defer c.Channels.Unlock()
	if wrapper, exists := c.Channels.Get(channel.ID); exists {
		// don't update it. It might overwrite a update event(!)
		// TODO: timestamps would be helpful here(?)
		//  or some queue of updates
		err := c.unmarshalUpdate(data, wrapper.Val)
		return wrap(channel), err
	}

	item := c.Channels.CreateCacheableItem(channel)
	c.Channels.Set(channel.ID, item)

	return wrap(channel), nil
}

func (c *CacheLFU) ChannelUpdate(data []byte) (*ChannelUpdate, error) {
	// assumption#1: Create may not take place before an update event
	// assumption#2: The set of fields in both ChannelCreate and ChannelUpdate are the same
	// assumption#3: a channel can not change from one type to another (text => news, text => voice)

	var metadata *idHolder
	if err := c.unmarshalUpdate(data, &metadata); err != nil {
		return nil, err
	}
	channelID := metadata.ID

	c.Channels.Lock()
	defer c.Channels.Unlock()

	var channel *Channel
	if item, exists := c.Channels.Get(channelID); exists {
		channel = item.Val.(*Channel)
		if err := c.unmarshalUpdate(data, channel); err != nil {
			return nil, err
		}
	} else {
		if err := c.unmarshalUpdate(data, &channel); err != nil {
			return nil, err
		}
		item := c.Channels.CreateCacheableItem(channel)
		c.Channels.Set(channelID, item)
	}

	return &ChannelUpdate{Channel: channel}, nil
}

func (c *CacheLFU) ChannelDelete(data []byte) (*ChannelDelete, error) {
	var cd *ChannelDelete
	if err := c.unmarshalUpdate(data, &cd); err != nil {
		return nil, err
	}

	c.Channels.Lock()
	defer c.Channels.Unlock()
	c.Channels.Delete(cd.Channel.ID)

	return cd, nil
}

func (c *CacheLFU) ChannelPinsUpdate(data []byte) (*ChannelPinsUpdate, error) {
	// assumption#1: not sent on deleted pins

	var cpu *ChannelPinsUpdate
	if err := c.unmarshalUpdate(data, &cpu); err != nil {
		return nil, err
	}

	if cpu.LastPinTimestamp.IsZero() {
		return cpu, nil
	}

	c.Channels.Lock()
	defer c.Channels.Unlock()
	if item, exists := c.Channels.Get(cpu.ChannelID); exists {
		channel := item.Val.(*Channel)
		channel.LastPinTimestamp = cpu.LastPinTimestamp
	}

	return cpu, nil
}

//func (c *CacheLFU) VoiceStateUpdate(data []byte) (*VoiceStateUpdate, error) {
//	// assumption#1: not sent on deleted pins
//
//	type voiceStateUpdateHolder struct {
//
//	}
//
//	var vsu *VoiceStateUpdate
//	if err := util.UnmarshalUpdate(data, &vsu); err != nil {
//		return nil, err
//	}
//
//	c.Channels.Lock()
//	defer c.Channels.Unlock()
//	if item, exists := c.Channels.Get(cpu.ChannelID); exists {
//		channel := item.Val.(*Channel)
//		channel.LastPinTimestamp = cpu.LastPinTimestamp
//	}
//
//	return cpu, nil
//}

func (c *CacheLFU) UserUpdate(data []byte) (*UserUpdate, error) {
	update := &UserUpdate{User: c.CurrentUser}

	c.CurrentUserMu.Lock()
	defer c.CurrentUserMu.Unlock()
	if err := c.unmarshalUpdate(data, update); err != nil {
		return nil, err
	}

	return update, nil
}

func (c *CacheLFU) VoiceServerUpdate(data []byte) (*VoiceServerUpdate, error) {
	var vsu *VoiceServerUpdate
	if err := c.unmarshalUpdate(data, &vsu); err != nil {
		return nil, err
	}

	return vsu, nil
}

func (c *CacheLFU) GuildMemberRemove(data []byte) (*GuildMemberRemove, error) {
	var gmr *GuildMemberRemove
	if err := c.unmarshalUpdate(data, &gmr); err != nil {
		return nil, err
	}

	c.Guilds.Lock()
	defer c.Guilds.Unlock()

	if item, exists := c.Guilds.Get(gmr.GuildID); exists {
		guild := item.Val.(*Guild)

		for i := range guild.Members {
			if guild.Members[i].UserID == gmr.User.ID {
				guild.MemberCount--
				guild.Members[i] = guild.Members[len(guild.Members)-1]
				guild.Members = guild.Members[:len(guild.Members)-1]
			}
		}
	}

	return gmr, nil
}

func (c *CacheLFU) GuildMemberAdd(data []byte) (*GuildMemberAdd, error) {
	var gmr *GuildMemberAdd
	if err := c.unmarshalUpdate(data, &gmr); err != nil {
		return nil, err
	}

	userID := gmr.Member.User.ID
	c.Users.Lock()
	if _, exists := c.Users.Get(userID); !exists {
		c.Users.Set(userID, c.Users.CreateCacheableItem(gmr.Member.User))
	}
	c.Users.Unlock()

	c.Guilds.Lock()
	defer c.Guilds.Unlock()

	if item, exists := c.Guilds.Get(gmr.Member.GuildID); exists {
		guild := item.Val.(*Guild)

		var member *Member
		for i := range guild.Members { // slow... map instead?
			if guild.Members[i].UserID == gmr.Member.User.ID {
				member = guild.Members[i]
				if err := c.unmarshalUpdate(data, member); err != nil {
					return nil, err
				}
				break
			}
		}
		if member == nil {
			member = &Member{}
			*member = *gmr.Member

			guild.Members = append(guild.Members, member)
			guild.MemberCount++
		}
		member.User = nil
	}

	return gmr, nil
}

func (c *CacheLFU) GuildCreate(data []byte) (*GuildCreate, error) {
	var guildEvt *GuildCreate
	if err := c.unmarshalUpdate(data, &guildEvt); err != nil {
		return nil, err
	}

	c.Guilds.Lock()
	defer c.Guilds.Unlock()

	if item, exists := c.Guilds.Get(guildEvt.Guild.ID); exists {
		guild := item.Val.(*Guild)
		if !guild.Unavailable {
			if len(guild.Members) > 0 {
				// seems like an update event came before create
				// this kinda... isn't good
				_ = c.unmarshalUpdate(data, item.Val)
			} else {
				// duplicate event
				return guildEvt, nil
			}
		} else {
			item.Val = guildEvt.Guild
		}
	} else {
		e := c.Guilds.CreateCacheableItem(guildEvt.Guild)
		c.Guilds.Set(guildEvt.Guild.ID, e)
	}

	return guildEvt, nil
}

func (c *CacheLFU) GuildUpdate(data []byte) (*GuildUpdate, error) {
	var guildEvt *GuildUpdate
	if err := c.unmarshalUpdate(data, &guildEvt); err != nil {
		return nil, err
	}

	c.Guilds.Lock()
	defer c.Guilds.Unlock()

	if item, exists := c.Guilds.Get(guildEvt.Guild.ID); exists {
		guild := item.Val.(*Guild)
		if guild.Unavailable {
			item.Val = guildEvt.Guild
		} else if err := c.unmarshalUpdate(data, item.Val); err != nil {
			return nil, err
		}
	} else {
		e := c.Guilds.CreateCacheableItem(guildEvt.Guild)
		c.Guilds.Set(guildEvt.Guild.ID, e)
	}

	return guildEvt, nil
}

func (c *CacheLFU) GuildDelete(data []byte) (*GuildDelete, error) {
	var guildEvt *GuildDelete
	if err := c.unmarshalUpdate(data, &guildEvt); err != nil {
		return nil, err
	}

	c.Guilds.Lock()
	defer c.Guilds.Unlock()
	c.Guilds.Delete(guildEvt.UnavailableGuild.ID)

	return guildEvt, nil
}
