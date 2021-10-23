package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fpbot/pkg/discord"
	"fpbot/pkg/twitch"
)

func main() {
	fmt.Println("Hello world")
	go discord.Run()
	go twitch.Run()

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
