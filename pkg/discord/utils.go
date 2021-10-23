package discord

import (
	"errors"
	"strings"

	dgo "github.com/bwmarrin/discordgo"
)

func checkForSelf(s *dgo.Session, m *dgo.Message) bool {
	if m.Author.ID == s.State.User.ID {
		return true
	}
	return false
}

func getMessageGuild(s *dgo.Session, m *dgo.Message) (*dgo.Guild, error) {
	for _, guild := range s.State.Guilds {
		if strings.EqualFold(guild.ID, m.GuildID) {
			return guild, nil
		}
	}
	return nil, errors.New("Unable to find guild for message")
}

func getInteractionGuild(s *dgo.Session, i *dgo.InteractionCreate) (*dgo.Guild, error) {
	for _, guild := range s.State.Guilds {
		if strings.EqualFold(guild.ID, i.GuildID) {
			return guild, nil
		}
	}
	return nil, errors.New("Unable to find guild for interaction")
}

func getChannelFromGuild(guild *dgo.Guild, channelName string) (*dgo.Channel, error) {
	for _, channel := range guild.Channels {
		if strings.EqualFold(channel.Name, channelName) {
			return channel, nil
		}
	}
	return nil, errors.New("Unable to find channel")
}

func checkInteractionForRole(s *dgo.Session, i *dgo.InteractionCreate, roleName string) bool {
	if i.Member == nil {
		return false
	}

	interactionGuild, err := getInteractionGuild(s, i)
	if err != nil {
		return false
	}

	roleID := ""
	for _, role := range interactionGuild.Roles {
		if strings.EqualFold(role.Name, roleName) {
			roleID = role.ID
		}
	}

	if len(roleID) < 1 {
		return false
	}

	for _, memberRoleID := range i.Member.Roles {
		if strings.EqualFold(memberRoleID, roleID) {
			return true
		}
	}

	return false
}
