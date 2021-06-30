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
    db "github.com/replit/database-go"
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

type botData struct {
	startTime    time.Time
    activeChatGames map[string]*chat_games.CountUp // TODO interfaces are difficult, abstract this out somehow
}

func NewBotData() botData {
    return botData {
        startTime:    time.Now(),
        activeChatGames: make(map[string]*chat_games.CountUp),
    }
}

func (b *botData) handleRegularText(s *dgo.Session, m *dgo.MessageCreate) {
	message, failed := utils.CheckCommand(s, m.Message)
    activeChatGamesLen := len(b.activeChatGames)
	if failed {
        if activeChatGamesLen < 1 {
            return
        }
        if utils.CheckForSelf(s, m.Message) {
            return
        }
        // We're playing a game!
        for k, v := range b.activeChatGames {
            switch k {
            case chat_games.CountUpName:
                if v.GetChannelID() != m.ChannelID {
                    return
                }
                if !v.Play(m.Message.Content) {
                    err := v.WriteScore()
                    if err != nil {
                        s.ChannelMessageSend(m.ChannelID, err.Error())
                    }
                    highScore, err := v.ReadScore()
                    if err != nil {
                        highScore = "0"
                    }
                    s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s score: %d\nHigh score: %s```", chat_games.CountUpName, v.Score, highScore))
                    delete(b.activeChatGames, chat_games.CountUpName)
                }
            default:
                // Shouldn't be possible?
            }
        }
        // Don't run the rest of the code
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
        case "help":
            s.ChannelMessageSend(m.ChannelID, "The only game is `CountUp` lmao")
		case chat_games.CountUpName:
            countUp := chat_games.NewCountUp(m.ChannelID)
			b.activeChatGames[chat_games.CountUpName] = &countUp
            s.ChannelMessageSend(m.ChannelID, "Starting game `CountUp`")
		default:
			s.ChannelMessageSend(m.ChannelID, gameToPlay)
		}
    // NOTE Admin only commands
    case "say": // Undocumented intentionally
        validRole := utils.CheckForRole("Admin", s, m.Message)

		if !validRole {
			s.ChannelMessageSend(m.ChannelID, "Invalid role, aborting.")
			return
		}

        if len(splitMessage) < 2 {
            s.ChannelMessageSend(m.ChannelID, "No params specified.")
            return
        }

        commands := strings.SplitN(splitMessage[1], " ", 2)
        if len(commands) < 2 {
            s.ChannelMessageSend(m.ChannelID, "Not enough params specified. Need 2 params: Channel and Message")
        }

        messageGuild, err := utils.GetMessageGuild(s, m.Message)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		channelToSendTo, err := utils.GetChannelFromGuild(commands[0], messageGuild)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

        s.ChannelMessageSend(channelToSendTo.ID, commands[1])
    case "db-delete":
        validRole := utils.CheckForRole("Admin", s, m.Message)

		if !validRole {
			s.ChannelMessageSend(m.ChannelID, "Invalid role, aborting.")
			return
		}

        if len(splitMessage) < 2 {
            s.ChannelMessageSend(m.ChannelID, "No params specified.")
            return
        }
        
        _, err := db.Get(splitMessage[1])
        if err != nil {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s does not exist in db: %s", splitMessage[1], err))
            return
        }
        err = db.Delete(splitMessage[1])
        if err != nil {
            s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Unable to delete key %s: %s", splitMessage[1], err.Error()))
            return
        }

        s.ChannelMessageSend(m.ChannelID, "Probably a success.")
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

    // Cleanup chat games only when the bot is invoked
    var chatGamesToBeDeleted []string
    for k, v := range b.activeChatGames {
        if time.Now().After(v.GetCleanupTime()) {
            chatGamesToBeDeleted = append(chatGamesToBeDeleted, k)
        }
    }

    if len(chatGamesToBeDeleted) > 0 {
        for _, v := range chatGamesToBeDeleted {
            delete(b.activeChatGames, v)
        }
    }
}
