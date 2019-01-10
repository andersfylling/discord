package disgord

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/andersfylling/disgord/endpoint"
	"github.com/andersfylling/disgord/httd"
)

func ratelimitChannel(id Snowflake) string {
	return "c:" + id.String()
}
func ratelimitChannelPermissions(id Snowflake) string {
	return ratelimitChannel(id) + ":perm"
}
func ratelimitChannelInvites(id Snowflake) string {
	return ratelimitChannel(id) + ":i"
}
func ratelimitChannelTyping(id Snowflake) string {
	return ratelimitChannel(id) + ":t"
}
func ratelimitChannelPins(id Snowflake) string {
	return ratelimitChannel(id) + ":pins"
}
func ratelimitChannelRecipients(id Snowflake) string {
	return ratelimitChannel(id) + ":r"
}
func ratelimitChannelMessages(id Snowflake) string {
	return ratelimitChannel(id) + ":m"
}
func ratelimitChannelMessagesDelete(id Snowflake) string {
	return ratelimitChannelMessages(id) + "_"
}
func ratelimitChannelWebhooks(id Snowflake) string {
	return ratelimitChannel(id) + ":w"
}

// GetChannel [REST] Get a channel by Snowflake. Returns a channel object.
//  Method                  GET
//  Endpoint                /channels/{channel.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#get-channel
//  Reviewed                2018-06-07
//  Comment                 -
func GetChannel(client httd.Getter, id Snowflake) (ret *Channel, err error) {
	if id.Empty() {
		return nil, errors.New("not a valid snowflake")
	}

	_, body, err := client.Get(&httd.Request{
		Ratelimiter: ratelimitChannel(id),
		Endpoint:    endpoint.Channel(id),
	})
	if err != nil {
		return
	}

	ret = &Channel{}
	err = unmarshal(body, ret)
	return
}

// NewModifyVoiceChannelParams create a ModifyChannelParams for a voice channel. Prevents changing attributes that
// only exists for text channels.
func NewModifyVoiceChannelParams() *ModifyChannelParams {
	return &ModifyChannelParams{
		data:    map[string]interface{}{},
		isVoice: true,
	}
}

// NewModifyTextChannelParams create a ModifyChannelParams for a text channel. Prevents changing attributes that
// only exists for voice channels.
func NewModifyTextChannelParams() *ModifyChannelParams {
	return &ModifyChannelParams{
		data:   map[string]interface{}{},
		isText: true,
	}
}

// ModifyChannelParams https://discordapp.com/developers/docs/resources/channel#modify-channel-json-params
type ModifyChannelParams struct {
	data    map[string]interface{}
	isText  bool
	isVoice bool
}

func (m *ModifyChannelParams) init() {
	if m.data != nil {
		return
	}

	m.data = map[string]interface{}{}
}

func (m *ModifyChannelParams) SetName(name string) error {
	if err := validateChannelName(name); err != nil {
		return err
	}

	m.init()
	m.data["name"] = name
	return nil
}
func (m *ModifyChannelParams) SetPosition(pos uint) {
	m.init()
	m.data["position"] = pos
}
func (m *ModifyChannelParams) SetTopic(topic string) error {
	if m.isVoice {
		return errors.New("cannot set topic for a voice channel. Text channels only")
	}
	if len(topic) > 1024 {
		return errors.New("topic is too long. max is 1024 character")
	}

	m.init()
	m.data["topic"] = topic
	m.isText = true
	return nil
}
func (m *ModifyChannelParams) SetNSFW(yes bool) error {
	if m.isVoice {
		return errors.New("cannot set NSFW status for voice channel. Text channels only")
	}
	m.init()
	m.data["nsfw"] = yes
	m.isText = true
	return nil
}
func (m *ModifyChannelParams) SetRateLimitPerUser(seconds uint) error {
	if m.isVoice {
		return errors.New("cannot set rate limit for a voice channel. Text channels only")
	}
	if seconds > 120 {
		return errors.New("limit can be maximum 120 seconds")
	}

	m.init()
	m.data["rate_limit_per_user"] = seconds
	m.isText = true
	return nil
}
func (m *ModifyChannelParams) SetBitrate(bitrate uint) error {
	if m.isText {
		return errors.New("cannot set bitrate for text channel. Voice channels only")
	}
	m.init()
	m.data["bitrate"] = bitrate
	m.isVoice = true
	return nil
}
func (m *ModifyChannelParams) SetUserLimit(limit uint) error {
	if m.isText {
		return errors.New("cannot set user limit for text channel. Voice channels only")
	}
	m.init()
	m.data["user_limit"] = limit
	m.isVoice = true
	return nil
}
func (m *ModifyChannelParams) SetPermissionOverwrites(permissions []PermissionOverwrite) {
	m.init()
	m.data["permission_overwrites"] = permissions
}
func (m *ModifyChannelParams) AddPermissionOverwrite(permission PermissionOverwrite) {
	m.init()
	if _, exists := m.data["permission_overwrites"]; !exists {
		m.data["permission_overwrites"] = []PermissionOverwrite{permission}
	} else {
		s := m.data["permission_overwrites"].([]PermissionOverwrite)
		s = append(s, permission)
	}
}
func (m *ModifyChannelParams) AddPermissionOverwrites(permissions []PermissionOverwrite) {
	m.init()
	if _, exists := m.data["permission_overwrites"]; !exists {
		m.data["permission_overwrites"] = permissions
	} else {
		s := m.data["permission_overwrites"].([]PermissionOverwrite)
		for i := range permissions {
			s = append(s, permissions[i])
		}
	}
}
func (m *ModifyChannelParams) SetParentID(id Snowflake) error {
	if !m.isVoice && !m.isText {
		return errors.New("can only set parent id for voice and text channels")
	}
	m.init()
	m.data["parent_id"] = id
	return nil
}
func (m *ModifyChannelParams) RemoveParentID() error {
	if !m.isVoice && !m.isText {
		return errors.New("can only set parent id for voice and text channels")
	}
	m.init()
	m.data["parent_id"] = nil
	return nil
}

func (m *ModifyChannelParams) MarshalJSON() ([]byte, error) {
	if len(m.data) == 0 {
		return []byte(`{}`), nil
	}

	return httd.Marshal(m.data)
}

var _ json.Marshaler = (*ModifyChannelParams)(nil)

// ModifyChannel [REST] Update a channels settings. Requires the 'MANAGE_CHANNELS' permission for the guild. Returns
// a channel on success, and a 400 BAD REQUEST on invalid parameters. Fires a Channel Update Gateway event. If
// modifying a category, individual Channel Update events will fire for each child channel that also changes.
// For the PATCH method, all the JSON Params are optional.
//  Method                  PUT/PATCH
//  Endpoint                /channels/{channel.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#modify-channel
//  Reviewed                2018-06-07
//  Comment                 andersfylling: only implemented the patch method, as its parameters are optional.
func ModifyChannel(client httd.Patcher, id Snowflake, changes *ModifyChannelParams) (ret *Channel, err error) {
	if id.Empty() {
		err = errors.New("not a valid snowflake")
		return
	}

	_, body, err := client.Patch(&httd.Request{
		Ratelimiter: ratelimitChannel(id),
		Endpoint:    endpoint.Channel(id),
		Body:        changes,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return
	}

	ret = &Channel{}
	err = unmarshal(body, ret)
	return
}

// DeleteChannel [REST] Delete a channel, or close a private message. Requires the 'MANAGE_CHANNELS' permission for
// the guild. Deleting a category does not delete its child channels; they will have their parent_id removed and a
// Channel Update Gateway event will fire for each of them. Returns a channel object on success.
// Fires a Channel Delete Gateway event.
//  Method                  Delete
//  Endpoint                /channels/{channel.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#deleteclose-channel
//  Reviewed                2018-10-09
//  Comment                 Deleting a guild channel cannot be undone. Use this with caution, as it
//                          is impossible to undo this action when performed on a guild channel. In
//                          contrast, when used with a private message, it is possible to undo the
//                          action by opening a private message with the recipient again.
func DeleteChannel(client httd.Deleter, id Snowflake) (channel *Channel, err error) {
	if id.Empty() {
		err = errors.New("not a valid snowflake")
		return
	}

	resp, body, err := client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannel(id),
		Endpoint:    endpoint.Channel(id),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusOK)
		err = errors.New(msg)
	}

	channel = &Channel{}
	err = unmarshal(body, channel)
	return
}

// EditChannelPermissionsParams https://discordapp.com/developers/docs/resources/channel#edit-channel-permissions-json-params
type EditChannelPermissionsParams struct {
	Allow int    `json:"allow"` // the bitwise value of all allowed permissions
	Deny  int    `json:"deny"`  // the bitwise value of all disallowed permissions
	Type  string `json:"type"`  // "member" for a user or "role" for a role
}

// SetChannelPermissionsParams is an alias for EditChannelPermissionsParams because Discord uses a single endpoint
// for both editing and adding permission overwrites.
type SetChannelPermissionsParams = EditChannelPermissionsParams

// EditChannelPermissions [REST] Edit the channel permission overwrites for a user or role in a channel. Only usable
// for guild channels. Requires the 'MANAGE_ROLES' permission. Returns a 204 empty response on success.
// For more information about permissions, see permissions.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/permissions/{overwrite.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/permissions
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#edit-channel-permissions
//  Reviewed                2018-06-07
//  Comment                 -
func EditChannelPermissions(client httd.Puter, chanID, overwriteID Snowflake, params *EditChannelPermissionsParams) (err error) {
	if chanID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if overwriteID.Empty() {
		return errors.New("overwriteID must be set to target the specific channel permissions")
	}

	resp, _, err := client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelPermissions(chanID),
		Endpoint:    endpoint.ChannelPermission(chanID, overwriteID),
		ContentType: httd.ContentTypeJSON,
		Body:        params,
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// GetChannelInvites [REST] Returns a list of invite objects (with invite metadata) for the channel. Only usable for
// guild channels. Requires the 'MANAGE_CHANNELS' permission.
//  Method                  GET
//  Endpoint                /channels/{channel.id}/invites
//  Rate limiter [MAJOR]    /channels/{channel.id}/invites
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#get-channel-invites
//  Reviewed                2018-06-07
//  Comment                 -
func GetChannelInvites(client httd.Getter, id Snowflake) (ret []*Invite, err error) {
	if id.Empty() {
		err = errors.New("channelID must be set to target the correct channel")
		return
	}

	_, body, err := client.Get(&httd.Request{
		Ratelimiter: ratelimitChannelInvites(id),
		Endpoint:    endpoint.ChannelInvites(id),
	})
	if err != nil {
		return
	}

	ret = []*Invite{}
	err = unmarshal(body, ret)
	return
}

// CreateChannelInvitesParams https://discordapp.com/developers/docs/resources/channel#create-channel-invite-json-params
type CreateChannelInvitesParams struct {
	MaxAge    int  `json:"max_age,omitempty"`   // duration of invite in seconds before expiry, or 0 for never. default 86400 (24 hours)
	MaxUses   int  `json:"max_uses,omitempty"`  // max number of uses or 0 for unlimited. default 0
	Temporary bool `json:"temporary,omitempty"` // whether this invite only grants temporary membership. default false
	Unique    bool `json:"unique,omitempty"`    // if true, don't try to reuse a similar invite (useful for creating many unique one time use invites). default false
}

// CreateChannelInvites [REST] Create a new invite object for the channel. Only usable for guild channels. Requires
// the CREATE_INSTANT_INVITE permission. All JSON parameters for this route are optional, however the request body is
// not. If you are not sending any fields, you still have to send an empty JSON object ({}). Returns an invite object.
//  Method                  POST
//  Endpoint                /channels/{channel.id}/invites
//  Rate limiter [MAJOR]    /channels/{channel.id}/invites
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#create-channel-invite
//  Reviewed                2018-06-07
//  Comment                 -
func CreateChannelInvites(client httd.Poster, id Snowflake, params *CreateChannelInvitesParams) (ret *Invite, err error) {
	if id.Empty() {
		err = errors.New("channelID must be set to target the correct channel")
		return
	}
	if params == nil {
		params = &CreateChannelInvitesParams{} // have to send an empty JSON object ({}). maybe just struct{}?
	}

	_, body, err := client.Post(&httd.Request{
		Ratelimiter: ratelimitChannelInvites(id),
		Endpoint:    endpoint.ChannelInvites(id),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return
	}

	ret = &Invite{}
	err = unmarshal(body, ret)
	return
}

// DeleteChannelPermission [REST] Delete a channel permission overwrite for a user or role in a channel. Only usable
// for guild channels. Requires the 'MANAGE_ROLES' permission. Returns a 204 empty response on success. For more
// information about permissions, see permissions: https://discordapp.com/developers/docs/topics/permissions#permissions
//  Method                  DELETE
//  Endpoint                /channels/{channel.id}/permissions/{overwrite.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/permissions
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#delete-channel-permission
//  Reviewed                2018-06-07
//  Comment                 -
func DeleteChannelPermission(client httd.Deleter, channelID, overwriteID Snowflake) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if overwriteID.Empty() {
		return errors.New("overwriteID must be set to target the specific channel permissions")
	}

	resp, _, err := client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelPermissions(channelID),
		Endpoint:    endpoint.ChannelPermission(channelID, overwriteID),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// TriggerTypingIndicator [REST] Post a typing indicator for the specified channel. Generally bots should not implement
// this route. However, if a bot is responding to a command and expects the computation to take a few seconds, this
// endpoint may be called to let the user know that the bot is processing their message. Returns a 204 empty response
// on success. Fires a Typing Start Gateway event.
//  Method                  POST
//  Endpoint                /channels/{channel.id}/typing
//  Rate limiter [MAJOR]    /channels/{channel.id}/typing
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#trigger-typing-indicator
//  Reviewed                2018-06-10
//  Comment                 -
func TriggerTypingIndicator(client httd.Poster, channelID Snowflake) (err error) {
	resp, _, err := client.Post(&httd.Request{
		Ratelimiter: ratelimitChannelTyping(channelID),
		Endpoint:    endpoint.ChannelTyping(channelID),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// GetPinnedMessages [REST] Returns all pinned messages in the channel as an array of message objects.
//  Method                  GET
//  Endpoint                /channels/{channel.id}/pins
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#get-pinned-messages
//  Reviewed                2018-06-10
//  Comment                 -
func GetPinnedMessages(client httd.Getter, channelID Snowflake) (ret []*Message, err error) {
	_, body, err := client.Get(&httd.Request{
		Ratelimiter: ratelimitChannelPins(channelID),
		Endpoint:    endpoint.ChannelPins(channelID),
	})
	if err != nil {
		return
	}

	ret = []*Message{}
	err = unmarshal(body, ret)
	return
}

// AddPinnedChannelMessage [REST] Pin a message in a channel. Requires the 'MANAGE_MESSAGES' permission.
// Returns a 204 empty response on success.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/pins/{message.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#add-pinned-channel-message
//  Reviewed                2018-06-10
//  Comment                 -
func AddPinnedChannelMessage(client httd.Puter, channelID, msgID Snowflake) (err error) {
	resp, _, err := client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelPins(channelID),
		Endpoint:    endpoint.ChannelPin(channelID, msgID),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// DeletePinnedChannelMessage [REST] Delete a pinned message in a channel. Requires the 'MANAGE_MESSAGES' permission.
// Returns a 204 empty response on success. Returns a 204 empty response on success.
//  Method                  DELETE
//  Endpoint                /channels/{channel.id}/pins/{message.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#delete-pinned-channel-message
//  Reviewed                2018-06-10
//  Comment                 -
func DeletePinnedChannelMessage(client httd.Deleter, channelID, msgID Snowflake) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if msgID.Empty() {
		return errors.New("messageID must be set to target the specific channel message")
	}

	resp, _, err := client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelPins(channelID),
		Endpoint:    endpoint.ChannelPin(channelID, msgID),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// GroupDMAddRecipientParams JSON params for GroupDMAddRecipient
type GroupDMAddRecipientParams struct {
	AccessToken string `json:"access_token"` // access botToken of a user that has granted your app the gdm.join scope
	Nickname    string `json:"nick"`         // nickname of the user being added
}

// GroupDMAddRecipient [REST] Adds a recipient to a Group DM using their access botToken. Returns a 204 empty response
// on success.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/recipients/{user.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/recipients
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#group-dm-add-recipient
//  Reviewed                2018-06-10
//  Comment                 -
func GroupDMAddRecipient(client httd.Puter, channelID, userID Snowflake, params *GroupDMAddRecipientParams) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if userID.Empty() {
		return errors.New("userID must be set to target the specific recipient")
	}

	resp, _, err := client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelRecipients(channelID),
		Endpoint:    endpoint.ChannelRecipient(channelID, userID),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}

// GroupDMRemoveRecipient [REST] Removes a recipient from a Group DM. Returns a 204 empty response on success.
//  Method                  DELETE
//  Endpoint                /channels/{channel.id}/recipients/{user.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/recipients
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#group-dm-remove-recipient
//  Reviewed                2018-06-10
//  Comment                 -
func GroupDMRemoveRecipient(client httd.Deleter, channelID, userID Snowflake) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if userID.Empty() {
		return errors.New("userID must be set to target the specific recipient")
	}

	resp, _, err := client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelRecipients(channelID),
		Endpoint:    endpoint.ChannelRecipient(channelID, userID),
	})
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return
}
