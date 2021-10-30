package runner

import (
	"fpbot/pkg/discord"
	"fpbot/pkg/twitch"
	"os"
)

type Runner struct {
	DB         *discord.DiscordBot
	TB         *twitch.TwitchBot
	ExitSignal chan os.Signal
}

func NewRunner(exitSignal chan os.Signal) *Runner {
	return &Runner{
		DB:         discord.NewDiscordBot(),
		TB:         twitch.NewTwitchBot(),
		ExitSignal: exitSignal,
	}
}

func (r *Runner) Run() {
	go r.DB.Run(r.ExitSignal)
	go r.TB.Run(r.ExitSignal)

	for {
		select {
		case data := <-r.DB.SendData:
			r.TB.ReceiveData(data)
		case data := <-r.TB.SendData:
			r.DB.ReceiveData(data)
		case <-r.ExitSignal:
			return
		}
	}
}
