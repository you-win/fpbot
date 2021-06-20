package discord

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"fpbot/pkg/chat_games"
    "fpbot/pkg/utils"

	cowsay "github.com/Code-Hex/Neo-cowsay"
	dgo "github.com/bwmarrin/discordgo"
)

const (
	botRepo          = "https://github.com/you-win/fpbot"
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

// TODO this is gross
func help() string {
	return fmt.Sprintf(
		"```\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s```",
		"help               - show this command",
		"ping               - pong",
		"high-five          - high five the bot",
		"whoami             - get your Discord account name",
		"uptime             - get the time since this bot started running",
		"repo               - get the repo where this bot's code is stored",
		"cowsay <text>      - format <text> as spoken by a cow",
		"play <text>        - starts a given game",
		"update-stream-info - Admin only, updates the stream-info message",
	)
}

func (b *botData) handleRegularText(s *dgo.Session, m *dgo.MessageCreate) {
	message, failed := utils.CheckCommand(s, m.Message)
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
	case "repo":
		s.ChannelMessageSend(m.ChannelID, botRepo)
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
		validRole := utils.CheckForRole("Admin", s, m.Message)

		if !validRole {
			s.ChannelMessageSend(m.ChannelID, "Invalid role, aborting.")
			return
		}

		if len(splitMessage) < 2 {
			s.ChannelMessageSend(m.ChannelID, "No update specified, aborting.")
			return
		}

		messageGuild, err := utils.GetMessageGuild(s, m.Message)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		streamInfoChannel, err := utils.GetChannelFromGuild("stream-info", messageGuild)
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

			newSchedule, err = utils.ReplaceStringAt(lastMessage.Content, fmt.Sprintf("%s: ", streamInfoCommand[0]), "\n", streamInfoCommand[1])
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
