package discord

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func Run() {
	discordToken := os.Getenv("DISCORD_TOKEN")

	dg, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session, ", err)
		return
	}

	defer dg.Close()

	// bd := botData{
	// 	startTime:    time.Now(),
	// 	requireIdent: true,
	// }
    bd := NewBotData()

	dg.AddHandler(bd.handleRegularText)

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
