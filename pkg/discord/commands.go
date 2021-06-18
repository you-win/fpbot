package discord

import (
	"fmt"
	"log"
	"strings"
	"time"

	"fpbot/pkg/chat_games"

	cowsay "github.com/Code-Hex/Neo-cowsay"
	dgo "github.com/bwmarrin/discordgo"
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

func help() string {
	return fmt.Sprintf(
		"```\n%s\n%s\n%s\n%s\n%s\n%s```",
		"help          - show this command",
		"ping          - pong",
		"whoami        - get your Discord account name",
		"uptime        - get the time since this bot started running",
		"cowsay <text> - format <text> as spoken by a cow",
		"play <text>   - starts a given game",
	)
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
		s.ChannelMessageSend(m.ChannelID, m.Message.Author.Username)
	case "uptime":
		s.ChannelMessageSend(m.ChannelID, time.Since(b.startTime).String())
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
			log.Fatal(err)
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
	default:
		log.Println("Unrecognized command " + splitMessage[0])
	}
}
