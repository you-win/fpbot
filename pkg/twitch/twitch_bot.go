package twitch

import (
	"encoding/json"
	"fmt"
	"fpbot/pkg/common"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	tgo "github.com/gempir/go-twitch-irc/v2"
	"github.com/nicklaw5/helix"
)

const (
	refreshURL = "https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s"

	commandChar = "?"
	botUserId   = "699075836" // TODO hardcoded

	isOfflineSwitchCounterMax = 10
)

var bd *common.BotData

type TwitchBot struct {
	Client   *tgo.Client
	SendData chan common.CrossServiceData // Pass things to other services

	TwitchUser string
}

func NewTwitchBot() *TwitchBot {
	return &TwitchBot{
		SendData: make(chan common.CrossServiceData),

		TwitchUser: os.Getenv("TWITCH_USER"),
	}
}

func (tb *TwitchBot) Run(quit chan os.Signal) {
	bd = &common.BotData{
		StartTime:                  time.Now(),
		LastRateLimitedCommandTime: time.Now(),
	}

	client := &http.Client{}

	// twitchUser := os.Getenv("TWITCH_USER")

	twitchUserDiscord := os.Getenv("TWITCH_USER_DISCORD")
	twitchUserGithub := os.Getenv("TWITCH_USER_GITHUB")

	refreshToken := os.Getenv("TWITCH_REFRESH_TOKEN")
	clientID := os.Getenv("TWITCH_CLIENT_ID")
	clientSecret := os.Getenv("TWITCH_CLIENT_SECRET")

	req, err := http.NewRequest("POST", fmt.Sprintf(refreshURL, refreshToken, clientID, clientSecret), nil)
	if err != nil {
		log.Printf("Unable to construct refresh request: %s", err.Error())
		return
	}

	res, err := client.Do(req)
	if err != nil {
		log.Printf("Unable to send refresh request: %s", err.Error())
		return
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Unable to read response body: %s", err.Error())
		return
	}

	var jsonBody map[string]interface{}
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		log.Printf("Unable to unmarshal json response: %s", err.Error())
		return
	}

	if res.StatusCode != 200 {
		log.Printf("Bad response from refresh endpoint: %s", jsonBody)
		return
	}

	twitchClient := tgo.NewClient(tb.TwitchUser, fmt.Sprintf("oauth:%s", jsonBody["access_token"].(string)))

	tb.Client = twitchClient

	capabilities := []string{"TAGS", "COMMANDS", "MEMBERSHIP"}
	twitchClient.Capabilities = capabilities
	// twitchClient := tgo.NewClient(twitchUser, fmt.Sprintf("oauth:%s", twitchToken))

	twitchClient.Join(tb.TwitchUser)

	twitchClient.OnConnect(func() {
		log.Println("Connected to chat")
	})

	twitchClient.OnPrivateMessage(func(message tgo.PrivateMessage) {
		// 100% do not respond to ourselves
		if strings.EqualFold(message.User.ID, botUserId) {
			return
		}

		if !strings.EqualFold(string(message.Message[0]), commandChar) {
			return
		}

		// Rate limit ourselves
		if time.Since(bd.LastRateLimitedCommandTime).Seconds() < 1 {
			return
		}
		bd.LastRateLimitedCommandTime = time.Now()

		commands := strings.Fields(message.Message[1:])
		if len(commands) < 1 {
			return
		}

		switch commands[0] {
		case "help":
			twitchClient.Say(tb.TwitchUser, "Possible commands: ping, lurk, discord, donate, repo, overlay, roll <param>, uptime, tcount")
		case "ping":
			twitchClient.Say(tb.TwitchUser, "pong")
		case "lurk":
			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("Thanks for lurking, %s", message.User.DisplayName))
		case "discord":
			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("%s's Discord can be found here: %s", tb.TwitchUser, twitchUserDiscord))
		case "donate":
			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("Please don't. Dropping a star on one of my github repos is enough: %s. If you really want to, then you can subscribe to the channel.", twitchUserGithub))
		case "repo":
			twitchClient.Say(tb.TwitchUser, "The bot's code can be found here: https://github.com/you-win/fpbot")
		case "overlay":
			twitchClient.Say(tb.TwitchUser, "I wrote the overlay! You can find the code on my Github: https://github.com/you-win/friendly-potato-stream")
		case "roll":
			result := 0
			rand.Seed(time.Now().UnixNano())

			if len(commands) == 1 {
				result = rand.Intn(6) + 1
			} else {
				if val, err := strconv.Atoi(commands[1]); err == nil {
					if val <= 0 {
						val = 6
					}
					result = rand.Intn(val) + 1
				} else {
					result = rand.Intn(6) + 1
				}
			}

			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("Rolled a: %d", result))
		case "uptime":
			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("Bot uptime: %s", time.Since(bd.StartTime).String()))
		case "tcount":
			twitchClient.Say(tb.TwitchUser, "Lower than team_youwin's")
		default:
			twitchClient.Say(tb.TwitchUser, fmt.Sprintf("Unrecognized command: %s Type ~help for a list of commands", commands[0]))
		}
	})

	err = twitchClient.Connect()
	if err != nil {
		log.Printf("Unable to connect to Twitch IRC: %s", err.Error())
		return
	}
	defer twitchClient.Disconnect()

	helixClient, err := helix.NewClient(&helix.Options{
		ClientID: clientID,
	})
	if err != nil {
		tb.SendData <- common.NewCrossServiceData(
			fmt.Sprintf("Unable to create helix client: %s", err.Error()),
			common.Error,
		)
	}

	// TODO add scope
	appAccessTokenResp, err := helixClient.RequestAppAccessToken([]string{""})
	if err != nil {
		tb.SendData <- common.NewCrossServiceData(
			fmt.Sprintf("Unable to get app access token for helix: %s", err.Error()),
			common.Error,
		)
	}

	helixClient.SetAppAccessToken(appAccessTokenResp.Data.AccessToken)

	isLive := false
	isSwitchable := false
	isOfflineSwitchCounter := 0

	for {
		select {
		case <-quit:
			return
		default:
			time.Sleep(time.Second * 60)

			if isOfflineSwitchCounter > isOfflineSwitchCounterMax {
				isSwitchable = true
				isLive = false
				isOfflineSwitchCounter = 0
			}

			resp, err := helixClient.GetStreams(&helix.StreamsParams{
				UserLogins: []string{"team_youwin"},
			})
			if err != nil {
				tb.SendData <- common.NewCrossServiceData(
					fmt.Sprintf("Unable to call get_streams endpoint: %s", err.Error()),
					common.Error,
				)
				isOfflineSwitchCounter += 1
				// continue
			}

			if len(resp.Data.Streams) == 0 {
				isOfflineSwitchCounter += 1
			} else {
				if !isLive && isSwitchable {
					tb.SendData <- common.NewCrossServiceData(
						"<@&901528644382519317> team_youwin is live at https://www.twitch.tv/team_youwin",
						common.DiscordStreamNotifications,
					)
					isLive = true
					isOfflineSwitchCounter = 0
					isSwitchable = false
				}
			}
		}
	}
	// <-quit
}

func (tb *TwitchBot) ReceiveData(data common.CrossServiceData) {
	switch data.Channel {
	case common.Twitch:
		tb.Client.Say(tb.TwitchUser, data.Message)
	default:
		log.Printf("Invalid message channel: %s", data.Message)
	}
}
