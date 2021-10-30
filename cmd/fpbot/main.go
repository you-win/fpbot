package main

import (
	"fmt"
	"fpbot/internal/runner"

	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Starting bot")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	botRunner := runner.NewRunner(sc)
	go botRunner.Run()

	<-sc
}
