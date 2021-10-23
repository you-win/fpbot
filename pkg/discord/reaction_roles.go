package discord

import (
	"log"

	dgo "github.com/bwmarrin/discordgo"
)

// TODO hardcoded channel id for now
var getRolesChannelID = "901529922974130306"

type ReactionRoles struct {
	roles   map[string]string
	adminID string
}

func newReactionRoles() *ReactionRoles {
	return &ReactionRoles{
		roles:   make(map[string]string),
		adminID: "76825724970864640", // TODO hardcoded
	}
}

func (rr *ReactionRoles) handleReady(s *dgo.Session, r *dgo.Ready) {
	roles, err := s.GuildRoles(guildID)
	if err != nil {
		log.Fatalf("Unable to get guild roles: %s", err.Error())
		return
	}

	for _, role := range roles {
		rr.roles[role.Name] = role.ID
	}
}

func (rr *ReactionRoles) handleReactionAdd(s *dgo.Session, r *dgo.MessageReactionAdd) {
	if r.ChannelID != getRolesChannelID || r.GuildID != guildID {
		return
	}

	if r.UserID == rr.adminID {
		return
	}

	switch r.Emoji.Name {
	case "👀":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Announcements"])
	case "🖥️":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Stream Notifications"])

	case "🍥":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Weeb"])
	case "💋":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["ur mum"])
	case "📼":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Content Creator"])
	case "🔢":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Developer"])
	case "🎶":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Musician"])
	case "🎨":
		s.GuildMemberRoleAdd(guildID, r.UserID, rr.roles["Artist"])

	default:
		log.Printf("unhandled emoji: %s", r.Emoji.Name)
	}
}

func (rr *ReactionRoles) handleReactionRemove(s *dgo.Session, r *dgo.MessageReactionRemove) {
	if r.ChannelID != getRolesChannelID || r.GuildID != guildID {
		return
	}

	if r.UserID == rr.adminID {
		return
	}

	switch r.Emoji.Name {
	case "👀":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Announcements"])
	case "🖥️":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Stream Notifications"])

	case "🍥":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Weeb"])
	case "💋":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["ur mum"])
	case "📼":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Content Creator"])
	case "🔢":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Developer"])
	case "🎨":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Artist"])
	case "🎶":
		s.GuildMemberRoleRemove(guildID, r.UserID, rr.roles["Musician"])

	default:
		log.Printf("unhandled emoji: %s", r.Emoji.Name)
	}
}
