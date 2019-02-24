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

	var body []byte
	_, body, err = client.Get(&httd.Request{
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
func NewUpdateVoiceChannelParams() *UpdateChannelParams {
	return &UpdateChannelParams{
		data:    map[string]interface{}{},
		isVoice: true,
	}
}

// NewModifyTextChannelParams create a ModifyChannelParams for a text channel. Prevents changing attributes that
// only exists for voice channels.
func NewModifyTextChannelParams() *UpdateChannelParams {
	return &UpdateChannelParams{
		data:   map[string]interface{}{},
		isText: true,
	}
}

// ModifyChannelParams https://discordapp.com/developers/docs/resources/channel#modify-channel-json-params
type UpdateChannelParams struct {
	data    map[string]interface{}
	isText  bool
	isVoice bool
}

func (m *UpdateChannelParams) init() {
	if m.data != nil {
		return
	}

	m.data = map[string]interface{}{}
}

func (m *UpdateChannelParams) SetName(name string) error {
	if err := validateChannelName(name); err != nil {
		return err
	}

	m.init()
	m.data["name"] = name
	return nil
}
func (m *UpdateChannelParams) SetPosition(pos uint) {
	m.init()
	m.data["position"] = pos
}
func (m *UpdateChannelParams) SetTopic(topic string) error {
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
func (m *UpdateChannelParams) SetNSFW(yes bool) error {
	if m.isVoice {
		return errors.New("cannot set NSFW status for voice channel. Text channels only")
	}
	m.init()
	m.data["nsfw"] = yes
	m.isText = true
	return nil
}
func (m *UpdateChannelParams) SetRateLimitPerUser(seconds uint) error {
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
func (m *UpdateChannelParams) SetBitrate(bitrate uint) error {
	if m.isText {
		return errors.New("cannot set bitrate for text channel. Voice channels only")
	}
	m.init()
	m.data["bitrate"] = bitrate
	m.isVoice = true
	return nil
}
func (m *UpdateChannelParams) SetUserLimit(limit uint) error {
	if m.isText {
		return errors.New("cannot set user limit for text channel. Voice channels only")
	}
	m.init()
	m.data["user_limit"] = limit
	m.isVoice = true
	return nil
}
func (m *UpdateChannelParams) SetPermissionOverwrites(permissions []PermissionOverwrite) {
	m.init()
	m.data["permission_overwrites"] = permissions
}
func (m *UpdateChannelParams) AddPermissionOverwrite(permission PermissionOverwrite) {
	m.init()
	if _, exists := m.data["permission_overwrites"]; !exists {
		m.data["permission_overwrites"] = []PermissionOverwrite{permission}
	} else {
		s := m.data["permission_overwrites"].([]PermissionOverwrite)
		s = append(s, permission)
	}
}
func (m *UpdateChannelParams) AddPermissionOverwrites(permissions []PermissionOverwrite) {
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
func (m *UpdateChannelParams) SetParentID(id Snowflake) error {
	if !m.isVoice && !m.isText {
		return errors.New("can only set parent id for voice and text channels")
	}
	m.init()
	m.data["parent_id"] = id
	return nil
}
func (m *UpdateChannelParams) RemoveParentID() error {
	if !m.isVoice && !m.isText {
		return errors.New("can only set parent id for voice and text channels")
	}
	m.init()
	m.data["parent_id"] = nil
	return nil
}

func (m *UpdateChannelParams) MarshalJSON() ([]byte, error) {
	if len(m.data) == 0 {
		return []byte(`{}`), nil
	}

	return httd.Marshal(m.data)
}

var _ json.Marshaler = (*UpdateChannelParams)(nil)

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
func ModifyChannel(client httd.Patcher, id Snowflake, changes *UpdateChannelParams) (ret *Channel, err error) {
	if id.Empty() {
		err = errors.New("not a valid snowflake")
		return
	}

	var body []byte
	_, body, err = client.Patch(&httd.Request{
		Ratelimiter: ratelimitChannel(id),
		Endpoint:    endpoint.Channel(id),
		Body:        changes,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return nil, err
	}

	ret = &Channel{}
	if err = unmarshal(body, ret); err != nil {
		return nil, err
	}

	return ret, nil
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

	var resp *http.Response
	var body []byte
	resp, body, err = client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannel(id),
		Endpoint:    endpoint.Channel(id),
	})
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusOK)
		err = errors.New(msg)
		return nil, err
	}

	channel = &Channel{}
	if err = unmarshal(body, channel); err != nil {
		return nil, err
	}

	return channel, nil
}

// UpdateChannelPermissionsParams https://discordapp.com/developers/docs/resources/channel#edit-channel-permissions-json-params
type UpdateChannelPermissionsParams struct {
	Allow int    `json:"allow"` // the bitwise value of all allowed permissions
	Deny  int    `json:"deny"`  // the bitwise value of all disallowed permissions
	Type  string `json:"type"`  // "member" for a user or "role" for a role
}

// EditChannelPermissions [REST] Edit the channel permission overwrites for a user or role in a channel. Only usable
// for guild channels. Requires the 'MANAGE_ROLES' permission. Returns a 204 empty response on success.
// For more information about permissions, see permissions.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/permissions/{overwrite.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/permissions
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#edit-channel-permissions
//  Reviewed                2018-06-07
//  Comment                 -
func EditChannelPermissions(client httd.Puter, chanID, overwriteID Snowflake, params *UpdateChannelPermissionsParams) (err error) {
	if chanID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if overwriteID.Empty() {
		return errors.New("overwriteID must be set to target the specific channel permissions")
	}

	var resp *http.Response
	resp, _, err = client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelPermissions(chanID),
		Endpoint:    endpoint.ChannelPermission(chanID, overwriteID),
		ContentType: httd.ContentTypeJSON,
		Body:        params,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}
	return err
}

// UpdateChannelPermissions ...
func (c *client) UpdateChannelPermissions(chanID, overwriteID Snowflake, params *UpdateChannelPermissionsParams, flags ...Flag) (err error) {
	err = EditChannelPermissions(c.req, chanID, overwriteID, params)
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

	var body []byte
	_, body, err = client.Get(&httd.Request{
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
		return nil, err
	}
	if params == nil {
		params = &CreateChannelInvitesParams{} // have to send an empty JSON object ({}). maybe just struct{}?
	}

	var body []byte
	_, body, err = client.Post(&httd.Request{
		Ratelimiter: ratelimitChannelInvites(id),
		Endpoint:    endpoint.ChannelInvites(id),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return nil, err
	}

	ret = &Invite{}
	err = unmarshal(body, ret)
	return nil, err
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

	var resp *http.Response
	resp, _, err = client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelPermissions(channelID),
		Endpoint:    endpoint.ChannelPermission(channelID, overwriteID),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
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
	var resp *http.Response
	resp, _, err = client.Post(&httd.Request{
		Ratelimiter: ratelimitChannelTyping(channelID),
		Endpoint:    endpoint.ChannelTyping(channelID),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
}

// GetPinnedMessages [REST] Returns all pinned messages in the channel as an array of message objects.
//  Method                  GET
//  Endpoint                /channels/{channel.id}/pins
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#get-pinned-messages
//  Reviewed                2018-06-10
//  Comment                 -
func GetPinnedMessages(client httd.Getter, channelID Snowflake) (ret []*Message, err error) {
	var body []byte
	_, body, err = client.Get(&httd.Request{
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
	var resp *http.Response
	resp, _, err = client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelPins(channelID),
		Endpoint:    endpoint.ChannelPin(channelID, msgID),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
}

// GetPinnedMessages .
func (c *client) GetPinnedMessages(channelID Snowflake, flags ...Flag) (ret []*Message, err error) {
	ret, err = GetPinnedMessages(c.req, channelID)
	return
}

func (c *client) PinMessage(message *Message, flags ...Flag) error {
	return c.PinMessageID(message.ChannelID, message.ID, flags...)
}

func (c *client) PinMessageID(channelID, messageID Snowflake, flags ...Flag) error {
	return AddPinnedChannelMessage(c.req, channelID, messageID)
}

func (c *client) UnpinMessage(message *Message, flags ...Flag) error {
	return c.UnpinMessageID(message.ChannelID, message.ID, flags...)
}

func (c *client) UnpinMessageID(channelID, messageID Snowflake, flags ...Flag) error {
	return deletePinnedChannelMessage(c.req, channelID, messageID)
}

type UpdatePinnedMessagesParams struct {
	PinMessages []Snowflake
	PinMessage  Snowflake

	UnpinMessages []Snowflake
	UnpinMessage  Snowflake
}

// UpdatePinnedMessages [REST] Pins one or more messages in a channel. Requires the 'MANAGE_MESSAGES' permission.
// By default, this method returns an error if one of the message ID's you wish to pin is zero/invalid. To
// instead skip those invalid snowflakes/id pass in the disgord flag IgnoreEmptyParams.
//  params := &disgord.UpdatePinnedMessagesParams{
//      AddMessages: []disgord.Snowflake{
//          0,          // empty/invalid
//          1254334682, // ok
//          6582332434, // ok
//          0,          // empty/invalid
//          0,          // empty/invalid
//          683234823,  // ok
//      },
//  }
//
//  // because there exists N empty/invalid message id's, this method will return an error saying the message id
//  // can not be empty/invalid.
//  disgord.UpdatePinnedMessages(channelID, &params) // error
//
//  // The disgord.IgnoreEmptyParams, will in this situation skip all the message id's that are empty/invalid.
//  disgord.UpdatePinnedMessages(channelID, &params, disgord.IgnoreEmptyParams) // ok
//
// This method sends N request, where N equals the number of messages to be pinned.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/pins/{message.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#add-pinned-channel-message
//  Reviewed                2018-06-10
//  Comment                 Each request returns a 204 empty response on success.
func (c *client) UpdatePinnedMessages(channelID Snowflake, params *UpdatePinnedMessagesParams, flags ...Flag) (pinned, unpinned []Snowflake, err error) {
	if channelID.Empty() {
		return nil, nil, errors.New("channelID can not be " + channelID.String())
	}
	if params == nil {
		return nil, nil, errors.New("no message ID was given")
	}

	if len(params.PinMessages) == 0 || !params.PinMessage.Empty() {
		params.PinMessages = append(params.PinMessages, params.PinMessage)
	}
	ctrl := mergeFlags(flags)

	messages := make([]Snowflake, 0, len(params.PinMessages)+1)
	for i := range params.PinMessages {
		id := params.PinMessages[i]
		if id.Empty() {
			err = errors.New("specified messages ID's can not be " + id.String())
			if !ctrl.IgnoreEmptyParams() {
				return nil, nil, err
			}
		} else {
			messages = append(messages, id)
		}
	}
	if len(messages) == 0 {
		return nil, nil, err
	}

	for _, messageID := range messages {
		if err = AddPinnedChannelMessage(c.req, channelID, messageID); err != nil {
			return pinned, nil, err
		}

		pinned = append(pinned, messageID)
	}

	// TODO: unpin messages

	return pinned, nil, nil
}

// DeletePinnedChannelMessage [REST] Delete a pinned message in a channel. Requires the 'MANAGE_MESSAGES' permission.
// Returns a 204 empty response on success. Returns a 204 empty response on success.
//  Method                  DELETE
//  Endpoint                /channels/{channel.id}/pins/{message.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/pins
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#delete-pinned-channel-message
//  Reviewed                2018-06-10
//  Comment                 -
func deletePinnedChannelMessage(client httd.Deleter, channelID, msgID Snowflake) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if msgID.Empty() {
		return errors.New("messageID must be set to target the specific channel message")
	}

	var resp *http.Response
	resp, _, err = client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelPins(channelID),
		Endpoint:    endpoint.ChannelPin(channelID, msgID),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
}

// GroupDMParticipant Information needed to add a recipient to a group chat
type GroupDMParticipant struct {
	AccessToken string    `json:"access_token"`   // access token of a user that has granted your app the gdm.join scope
	Nickname    string    `json:"nick,omitempty"` // nickname of the user being added
	UserID      Snowflake `json:"-"`
}

func (g *GroupDMParticipant) FindErrors() error {
	if g.UserID.Empty() {
		return errors.New("missing userID")
	}
	if g.AccessToken == "" {
		return errors.New("missing access token")
	}
	if err := ValidateUsername(g.Nickname); err != nil && g.Nickname != "" {
		return err
	}

	return nil
}

func (c *client) AddDMParticipant(channelID Snowflake, participant *GroupDMParticipant, flags ...Flag) error {
	return addGroupDMRecipient(c.req, channelID, participant)
}

// addGroupDMRecipient [REST] Adds a recipient to a Group DM using their access token. Returns a 204 empty response
// on success.
//  Method                  PUT
//  Endpoint                /channels/{channel.id}/recipients/{user.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/recipients
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#group-dm-add-recipient
//  Reviewed                2018-06-10
//  Comment                 -
func addGroupDMRecipient(client httd.Puter, channelID Snowflake, params *GroupDMParticipant) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if params == nil {
		return errors.New("params can not be nil")
	}
	if params.UserID.Empty() {
		return errors.New("userID must be set to target the specific recipient")
	}

	var resp *http.Response
	resp, _, err = client.Put(&httd.Request{
		Ratelimiter: ratelimitChannelRecipients(channelID),
		Endpoint:    endpoint.ChannelRecipient(channelID, params.UserID),
		Body:        params,
		ContentType: httd.ContentTypeJSON,
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
}

// removeGroupDMRecipient [REST] Removes a recipient from a Group DM. Returns a 204 empty response on success.
//  Method                  DELETE
//  Endpoint                /channels/{channel.id}/recipients/{user.id}
//  Rate limiter [MAJOR]    /channels/{channel.id}/recipients
//  Discord documentation   https://discordapp.com/developers/docs/resources/channel#group-dm-remove-recipient
//  Reviewed                2018-06-10
//  Comment                 -
func removeGroupDMRecipient(client httd.Deleter, channelID, userID Snowflake) (err error) {
	if channelID.Empty() {
		return errors.New("channelID must be set to target the correct channel")
	}
	if userID.Empty() {
		return errors.New("userID must be set to target the specific recipient")
	}

	var resp *http.Response
	resp, _, err = client.Delete(&httd.Request{
		Ratelimiter: ratelimitChannelRecipients(channelID),
		Endpoint:    endpoint.ChannelRecipient(channelID, userID),
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		msg := "unexpected http response code. Got " + resp.Status + ", wants " + http.StatusText(http.StatusNoContent)
		err = errors.New(msg)
	}

	return err
}
