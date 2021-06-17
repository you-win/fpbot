package discord

import (
    "time"
    "strings"
    "log"
    "fmt"

    dgo "github.com/bwmarrin/discordgo"
    cowsay "github.com/Code-Hex/Neo-cowsay"
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

func (b *botData) handleRegularText(s *dgo.Session, m *dgo.MessageCreate) {
	message, failed := checkCommand(s, m.Message)
	if failed {
		return
	}

    splitMessage := strings.SplitN(message, " ", 2)
    if len(splitMessage) < 1 {
        return
    }

    switch splitMessage[0] {
    case "ping":
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    case "whoami":
        s.ChannelMessageSend(m.ChannelID, m.Message.Author.Username)
    case "uptime":
        s.ChannelMessageSend(m.ChannelID, time.Since(b.startTime).String())
    case "cowsay":
        log.Println(splitMessage[1])
        say, err := cowsay.Say(
            cowsay.Phrase(splitMessage[1]),
            cowsay.Random(),
            cowsay.BallonWidth(40),
        )
        if err != nil {
            log.Fatal(err)
            return
        }

        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("\n```%s```", say))
    default:
        log.Println("Unrecognized command " + splitMessage[0])
    }
}
