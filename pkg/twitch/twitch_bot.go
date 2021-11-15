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
)

const (
	refreshURL = "https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s"

	commandChar = "?"
	botUserId   = "699075836" // TODO hardcoded

	isOfflineSwitchCounterMax = 60
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
		log.Println("Connected to Twitch IRC chat")
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

	go twitchClient.Connect()
	// err = twitchClient.Connect()
	// if err != nil {
	// 	log.Printf("Unable to connect to Twitch IRC: %s", err.Error())
	// 	return
	// }
	defer twitchClient.Disconnect()

	log.Println("Twitch bot sucessfully started, probably")

	// helixClient, err := helix.NewClient(&helix.Options{
	// 	ClientID: clientID,
	// })
	// if err != nil {
	// 	tb.SendData <- common.NewCrossServiceData(
	// 		fmt.Sprintf("Unable to create helix client: %s", err.Error()),
	// 		common.Error,
	// 	)
	// }

	// appAccessTokenResp, err := helixClient.RequestAppAccessToken([]string{""})
	// if err != nil {
	// 	log.Printf("Error getting app access token: %s", err.Error())
	// 	tb.SendData <- common.NewCrossServiceData(
	// 		fmt.Sprintf("Unable to get app access token for helix: %s", err.Error()),
	// 		common.Error,
	// 	)
	// }

	// helixClient.SetAppAccessToken(appAccessTokenResp.Data.AccessToken)

	res, err = client.Post(
		fmt.Sprintf("https://id.twitch.tv/oauth2/token?client_id=%s&client_secret=%s&grant_type=client_credentials", clientID, clientSecret),
		"application/x-www-form-urlencoded",
		nil,
	)
	if err != nil {
		log.Printf("Unable to get app access token: %s", err.Error())
		return
	}

	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("Unable to read response body: %s", err.Error())
		res.Body.Close()
		return
	}

	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		log.Printf("Error when unmarshalling data: %s", err.Error())
		res.Body.Close()
		return
	}

	res.Body.Close()

	if res.StatusCode != 200 {
		log.Printf("Bad response from app access token endpoint: %s", err.Error())
		return
	}

	appAccessToken := jsonBody["access_token"]

	// req, err = http.NewRequest(
	// 	"GET",
	// 	fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", "44149998"),
	// 	nil,
	// )
	req, err = http.NewRequest(
		"GET",
		fmt.Sprintf("https://api.twitch.tv/helix/streams?user_login=%s", tb.TwitchUser),
		nil,
	)
	if err != nil {
		log.Printf("Unable to create streams request: %s", err.Error())
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", appAccessToken))
	req.Header.Add("Client-Id", clientID)

	isLive := false
	isSwitchable := true
	isOfflineSwitchCounter := 0

	for {
		select {
		case <-quit:
			return
		default:
			if isSwitchable {
				isSwitchable = false
				go func() {
					time.Sleep(time.Second * 60)

					if isOfflineSwitchCounter > isOfflineSwitchCounterMax {
						isLive = false
						isOfflineSwitchCounter = 0
						isSwitchable = true
						return
					}

					res, err = client.Do(req)
					if err != nil {
						log.Printf("Error when polling Twitch stream: %s", err.Error())
						isOfflineSwitchCounter += 1
						isSwitchable = true
						return
					}

					body, err = ioutil.ReadAll(res.Body)
					if err != nil {
						log.Printf("Unable to read response body: %s", err.Error())
						isOfflineSwitchCounter += 1
						isSwitchable = true
						return
					}

					defer res.Body.Close()

					var streamJson map[string]interface{}
					err = json.Unmarshal(body, &streamJson)
					if err != nil {
						log.Printf("Unable to unmarshal json response: %s", err.Error())
						isOfflineSwitchCounter += 1
						isSwitchable = true
						return
					}

					if res.StatusCode != 200 {
						log.Printf("Bad response from streams endpoint: %s", err.Error())
						isOfflineSwitchCounter += 1
						isSwitchable = true
						return
					}

					data := streamJson["data"].([]interface{})

					if len(data) == 0 {
						isOfflineSwitchCounter += 1
					} else {
						if !isLive {
							tb.SendData <- common.NewCrossServiceData(
								"<@&901528644382519317> team_youwin is live at https://www.twitch.tv/team_youwin",
								common.DiscordStreamNotifications,
							)
						}
						isLive = true
						isOfflineSwitchCounter = 0
					}

					isSwitchable = true
				}()
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
