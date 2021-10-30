package common

import "time"

// In the case of Discord, these are channel selectors
// In the case of Twitch, this is a nonce
const (
	None                       = iota
	Twitch                     = iota
	DiscordGeneral             = iota
	DiscordAnnouncements       = iota
	DiscordStreamNotifications = iota
	DiscordBotController       = iota
	Error                      = iota
)

type BotData struct {
	StartTime                  time.Time
	LastRateLimitedCommandTime time.Time
}

type CrossServiceData struct {
	Message string
	Channel int
}

func NewCrossServiceData(message string, channelParam ...int) CrossServiceData {
	channel := None
	if len(channelParam) > 0 {
		channel = channelParam[0]
	}
	return CrossServiceData{
		Message: message,
		Channel: channel,
	}
}
