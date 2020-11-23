// Code generated by generate/interfaces; DO NOT EDIT.

package disgord

func (v *VoiceState) deepCopy() interface{} {
	cp := &VoiceState{}
	_ = DeepCopyOver(cp, v)
	return cp
}

func (v *VoiceRegion) deepCopy() interface{} {
	cp := &VoiceRegion{}
	_ = DeepCopyOver(cp, v)
	return cp
}

func (m *MentionChannel) deepCopy() interface{} {
	cp := &MentionChannel{}
	_ = DeepCopyOver(cp, m)
	return cp
}

func (u *User) deepCopy() interface{} {
	cp := &User{}
	_ = DeepCopyOver(cp, u)
	return cp
}

func (m *Message) deepCopy() interface{} {
	cp := &Message{}
	_ = DeepCopyOver(cp, m)
	return cp
}

func (m *messageSticker) deepCopy() interface{} {
	cp := &messageSticker{}
	_ = DeepCopyOver(cp, m)
	return cp
}

func (a *AuditLogOption) deepCopy() interface{} {
	cp := &AuditLogOption{}
	_ = DeepCopyOver(cp, a)
	return cp
}

func (e *Emoji) deepCopy() interface{} {
	cp := &Emoji{}
	_ = DeepCopyOver(cp, e)
	return cp
}

func (i *Invite) deepCopy() interface{} {
	cp := &Invite{}
	_ = DeepCopyOver(cp, i)
	return cp
}

func (i *InviteMetadata) deepCopy() interface{} {
	cp := &InviteMetadata{}
	_ = DeepCopyOver(cp, i)
	return cp
}

func (r *Role) deepCopy() interface{} {
	cp := &Role{}
	_ = DeepCopyOver(cp, r)
	return cp
}

func (g *Guild) deepCopy() interface{} {
	cp := &Guild{}
	_ = DeepCopyOver(cp, g)
	return cp
}

func (c *Channel) deepCopy() interface{} {
	cp := &Channel{}
	_ = DeepCopyOver(cp, c)
	return cp
}

func (a *AuditLogEntry) deepCopy() interface{} {
	cp := &AuditLogEntry{}
	_ = DeepCopyOver(cp, a)
	return cp
}

func (a *AuditLogChanges) deepCopy() interface{} {
	cp := &AuditLogChanges{}
	_ = DeepCopyOver(cp, a)
	return cp
}

func (a *AuditLog) deepCopy() interface{} {
	cp := &AuditLog{}
	_ = DeepCopyOver(cp, a)
	return cp
}
