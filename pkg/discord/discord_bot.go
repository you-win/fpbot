package discord

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	// fpmodel "fpbot/pkg/model"

	"github.com/bwmarrin/discordgo"
)

var (
	discordToken string
	guildID      string
)

type BotData struct {
	StartTime                  time.Time
	LastRateLimitedCommandTime time.Time
}

var bd *BotData

var s *discordgo.Session

func Run() {
	bd = &BotData{
		StartTime:                  time.Now(),
		LastRateLimitedCommandTime: time.Now(),
	}

	discordToken = os.Getenv("DISCORD_TOKEN")
	guildID = os.Getenv("GUILD_ID")

	s, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
	}

	// Slash commands
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Anti-spam
	as := NewAntiSpam()
	s.AddHandler(as.handleSpam)

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

	log.Println("Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
	log.Println("Bot leaving this plane of existence. おやすみ")
}
