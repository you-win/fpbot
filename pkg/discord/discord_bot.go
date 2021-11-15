package discord

import (
	"fmt"
	"fpbot/pkg/common"
	"log"

	"os"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	discordToken string
	guildID      string
)

const DiscordEpoch = 1420070400000

type DiscordBot struct {
	Session  *discordgo.Session
	SendData chan common.CrossServiceData // Pass things to other services

	LastStreamNotificationTime *time.Time
}

var bd *common.BotData

var db *DiscordBot

func NewDiscordBot() *DiscordBot {
	oneHourBefore := time.Now().AddDate(0, -1, 0)
	db = &DiscordBot{
		SendData: make(chan common.CrossServiceData),

		LastStreamNotificationTime: &oneHourBefore,
	}
	return db
}

func (db *DiscordBot) Run(quit chan os.Signal) {
	bd = &common.BotData{
		StartTime:                  time.Now(),
		LastRateLimitedCommandTime: time.Now(),
	}

	discordToken = os.Getenv("DISCORD_TOKEN")
	guildID = os.Getenv("GUILD_ID")

	s, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
	}

	db.Session = s

	// Slash commands
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Anti-spam
	as := NewAntiSpam()
	s.AddHandler(as.handleSpam)

	// Reaction roles
	rr := newReactionRoles()
	s.AddHandler(rr.handleReady)
	s.AddHandler(rr.handleReactionAdd)
	s.AddHandler(rr.handleReactionRemove)

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Printf("Added %s", v.Name)
	}

	log.Println("Discord bot is running...")

	<-quit
}

func (db *DiscordBot) ReceiveData(data common.CrossServiceData) {
	switch data.Channel {
	case common.DiscordGeneral: // TODO actually directs to bot playground
		db.Session.ChannelMessageSend("854954868334264351", data.Message)
	case common.DiscordAnnouncements:
		db.Session.ChannelMessageSend("853476898855845900", data.Message)
	case common.DiscordStreamNotifications:
		if time.Since(*db.LastStreamNotificationTime).Hours() < 3.0 {
			db.Session.ChannelMessageSend("856373732813963274", "Bot tried to send a stream notification too quickly")
			return
		}

		currentTime := time.Now()

		db.LastStreamNotificationTime = &currentTime

		db.Session.ChannelMessageSend("901833984542134293", data.Message)
	case common.DiscordBotController:
		db.Session.ChannelMessageSend("856373732813963274", data.Message)
	default:
		db.LogError(fmt.Sprintf("Invalid Discord message channel: %s", data.Message))
	}
}

func (db *DiscordBot) LogError(message string) {
	db.Session.ChannelMessageSend("854954868334264351", fmt.Sprintf("[ERROR] %s", message))
}
