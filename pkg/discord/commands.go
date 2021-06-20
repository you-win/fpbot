package discord

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"fpbot/pkg/chat_games"

	cowsay "github.com/Code-Hex/Neo-cowsay"
	dgo "github.com/bwmarrin/discordgo"
)

const (
	twitchURL        = "https://www.twitch.tv/team_youwin"
	scheduleTemplate = `Streaming schedule for https://www.twitch.tv/team_youwin
All times in US Eastern time.
Streams may start/end later than listed.
--------------------------------------------------------
Sun: %s
Mon: %s
Tue: %s
Wed: %s
Thu: %s
Fri: %s
Sat: %s
--------------------------------------------------------`
)

func checkCommand(s *dgo.Session, m *dgo.Message) (string, bool) {
	if m.Author.ID == s.State.User.ID {
		return "", true
	}

	if !strings.HasPrefix(m.Content, "$") {
		return "", true
	}

	newString := strings.Replace(m.Content, "$", "", 1)

	return newString, false
}

func getMessageGuild(s *dgo.Session, m *dgo.Message) (*dgo.Guild, error) {
	for _, guild := range s.State.Guilds {
		if strings.EqualFold(guild.ID, m.GuildID) {
			return guild, nil
		}
	}
	return nil, errors.New("Unable to find guild for message")
}

func getChannelFromGuild(channelName string, guild *dgo.Guild) (*dgo.Channel, error) {
	for _, channel := range guild.Channels {
		if strings.EqualFold(channel.Name, channelName) {
			return channel, nil
		}
	}
	return nil, errors.New("Unable to find channel")
}

func checkForRole(roleName string, s *dgo.Session, m *dgo.Message) bool {
	messageGuild, err := getMessageGuild(s, m)
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

// TODO this is gross
func help() string {
	return fmt.Sprintf(
		"```\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s```",
		"help               - show this command",
		"ping               - pong",
		"high-five          - high five the bot",
		"whoami             - get your Discord account name",
		"uptime             - get the time since this bot started running",
		"cowsay <text>      - format <text> as spoken by a cow",
		"play <text>        - starts a given game",
		"update-stream-info - Admin only, updates the stream-info message",
	)
}

func replaceStringAt(input string, startString string, endString string, newString string) (string, error) {
	startIndex := strings.Index(input, startString)
	if startIndex == -1 {
		return "", errors.New(fmt.Sprintf("Could not find: %s", startString))
	}
	startIndex += len(startString)

	endIndex := strings.Index(input[startIndex:], endString)
	if endIndex == -1 {
		return "", errors.New(fmt.Sprintf("Could not find: %s", endString))
	}
	endIndex = startIndex + endIndex

	var sb strings.Builder

	sb.WriteString(input[:startIndex])
	sb.WriteString(newString)
	sb.WriteString(input[endIndex:])

	return sb.String(), nil
}

func (b *botData) handleRegularText(s *dgo.Session, m *dgo.MessageCreate) {
	message, failed := checkCommand(s, m.Message)
	if failed && b.requireIdent {
		return
	}

	splitMessage := strings.SplitN(message, " ", 2)
	if len(splitMessage) < 1 {
		return
	}

	switch splitMessage[0] {
	case "help":
		s.ChannelMessageSend(m.ChannelID, help())
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	case "whoami":
		whoamiName := m.Message.Author.Username
		if len(m.Member.Nick) > 0 {
			whoamiName = m.Member.Nick
		}
		s.ChannelMessageSend(m.ChannelID, whoamiName)
	case "uptime":
		s.ChannelMessageSend(m.ChannelID, time.Since(b.startTime).String())
	case "high-five":
		emoji, err := strconv.ParseInt(strings.TrimPrefix("\\U0001f44f", "\\U"), 16, 32)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, string(emoji))
	case "cowsay":
		cowMessage := "moo"
		if len(splitMessage) > 1 {
			cowMessage = splitMessage[1]
		}
		say, err := cowsay.Say(
			cowsay.Phrase(cowMessage),
			cowsay.Random(),
			cowsay.BallonWidth(40),
		)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\n```%s```", say))
	case "play":
		gameToPlay := "Play what?"
		if len(splitMessage) > 1 {
			gameToPlay = splitMessage[1]
		}

		switch gameToPlay {
		case chat_games.CountUpName:
			// TODO
			s.ChannelMessageSend(m.ChannelID, "Not yet implemented")
		default:
			s.ChannelMessageSend(m.ChannelID, gameToPlay)
		}
	case "update-stream-info":
		validRole := checkForRole("Admin", s, m.Message)

		if !validRole {
			s.ChannelMessageSend(m.ChannelID, "Invalid role, aborting.")
			return
		}

		if len(splitMessage) < 2 {
			s.ChannelMessageSend(m.ChannelID, "No update specified, aborting.")
			return
		}

		messageGuild, err := getMessageGuild(s, m.Message)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		streamInfoChannel, err := getChannelFromGuild("stream-info", messageGuild)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		streamInfoChannelMessages, err := s.ChannelMessages(streamInfoChannel.ID, 1, "", "", "")
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		newSchedule := ""
		streamInfoCommand := strings.SplitN(splitMessage[1], " ", 2)
		switch streamInfoCommand[0] {
		case "default":
			newSchedule = fmt.Sprintf(scheduleTemplate,
				"7-10pm (Programming)",
				"n/a",
				"7-10pm (Programming)",
				"9-11pm (Programming)",
				"7-10pm (Programming)",
				"n/a",
				"n/a",
			)
			if len(streamInfoChannelMessages) < 1 {
				s.ChannelMessageSend(m.ChannelID, "stream-info channel doesn't have a message to edit")
				return
			} else {
				s.ChannelMessageDelete(streamInfoChannel.ID, streamInfoChannelMessages[0].ID)
			}

			s.ChannelMessageSend(streamInfoChannel.ID, fmt.Sprintf("```%s```", newSchedule))
		default:
			if len(streamInfoChannelMessages) < 1 {
				s.ChannelMessageSend(m.ChannelID, "stream-info channel doesn't have a message to edit")
				return
			}
			lastMessageID := streamInfoChannelMessages[0].ID

			if len(streamInfoCommand) < 2 {
				s.ChannelMessageSend(m.ChannelID, "Not enough commands to update stream-info")
				return
			}
			lastMessage, err := s.ChannelMessage(streamInfoChannel.ID, lastMessageID)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			newSchedule, err = replaceStringAt(lastMessage.Content, fmt.Sprintf("%s: ", streamInfoCommand[0]), "\n", streamInfoCommand[1])
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			s.ChannelMessageDelete(streamInfoChannel.ID, lastMessageID)
			s.ChannelMessageSend(streamInfoChannel.ID, newSchedule)
		}

		s.ChannelMessageSend(m.ChannelID, "Updated stream-info")
	default:
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unrecognized command %s", splitMessage[0]))
	}
}
