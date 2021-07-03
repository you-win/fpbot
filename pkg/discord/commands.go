package discord

import (
	"strings"
	"time"

    "fpbot/pkg/utils"
    "fpbot/pkg/discord/discord_cmd"

	dgo "github.com/bwmarrin/discordgo"
    // db "github.com/replit/database-go"
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

type BotData struct {
	startTime    time.Time
}

func NewBotData() BotData {
    return BotData {
        startTime:    time.Now(),
    }
}

func (bd BotData) GetStartTime() time.Time {
    return bd.startTime
}

func (bd *BotData) handleRegularText(s *dgo.Session, m *dgo.MessageCreate) {
	message, failed := utils.CheckCommand(s, m.Message)
	if failed || utils.CheckForSelf(s, m.Message) {
        return
	}

    splitMessage := strings.Split(message, " ")

    cmd := discord_cmd.NewCommand(s, m.Message, bd, splitMessage)

    err := cmd.Execute()
    if err != nil {
        s.ChannelMessageSend(m.ChannelID, err.Error())
    }
}
