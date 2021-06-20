package utils

import (
    "strings"
    "errors"
    "fmt"

    dgo "github.com/bwmarrin/discordgo"
)

func CheckCommand(s *dgo.Session, m *dgo.Message) (string, bool) {
	if m.Author.ID == s.State.User.ID {
		return "", true
	}

	if !strings.HasPrefix(m.Content, "$") {
		return "", true
	}

	newString := strings.Replace(m.Content, "$", "", 1)

	return newString, false
}

func GetMessageGuild(s *dgo.Session, m *dgo.Message) (*dgo.Guild, error) {
	for _, guild := range s.State.Guilds {
		if strings.EqualFold(guild.ID, m.GuildID) {
			return guild, nil
		}
	}
	return nil, errors.New("Unable to find guild for message")
}

func GetChannelFromGuild(channelName string, guild *dgo.Guild) (*dgo.Channel, error) {
	for _, channel := range guild.Channels {
		if strings.EqualFold(channel.Name, channelName) {
			return channel, nil
		}
	}
	return nil, errors.New("Unable to find channel")
}

func CheckForRole(roleName string, s *dgo.Session, m *dgo.Message) bool {
	messageGuild, err := GetMessageGuild(s, m)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Couldn't find guild with error: %s", err))
	}

	roleID := ""
	for _, role := range messageGuild.Roles {
		if strings.EqualFold(role.Name, roleName) {
			roleID = role.ID
		}
	}

	if len(roleID) < 1 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Couldn't find role: %s", roleName))
		return false
	}

	for _, memberRoleID := range m.Member.Roles {
		if strings.EqualFold(memberRoleID, roleID) {
			return true
		}
	}

	return false
}
