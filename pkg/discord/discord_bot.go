package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
    "time"

	"github.com/bwmarrin/discordgo"
)

type botData struct {
    startTime time.Time
}

func Run() {
	discordToken := os.Getenv("DISCORD_TOKEN")

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session, ", err)
		return
	}

    defer dg.Close()

    bd := botData {
        startTime: time.Now(),
    }

    dg.AddHandler(bd.handleRegularText)

	// dg.AddHandler(bd.ping)

    // dg.AddHandler(bd.whoAmI)

    // dg.AddHandler(bd.uptime)

	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection, ", err)
		return
	}

	fmt.Println("Bot is running...")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
