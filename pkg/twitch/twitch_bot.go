package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	tgo "github.com/gempir/go-twitch-irc/v2"
)

const (
	refreshURL = "https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s"

	commandChar = "?"
)

func Run() {
	client := &http.Client{}

	twitchUser := os.Getenv("TWITCH_USER")

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

	twitchClient := tgo.NewClient(twitchUser, fmt.Sprintf("oauth:%s", jsonBody["access_token"].(string)))

	twitchClient.Join(twitchUser)

	twitchClient.OnConnect(func() {
		log.Println("Connected to chat")
	})

	twitchClient.OnPrivateMessage(func(message tgo.PrivateMessage) {
		if !strings.EqualFold(string(message.Message[0]), commandChar) {
			return
		}

		commands := strings.Fields(message.Message[1:])
		if len(commands) < 1 {
			return
		}

		switch commands[0] {
		case "help":
			twitchClient.Say(twitchUser, "Possible commands: ping, lurk, discord, donate, repo")
		case "ping":
			twitchClient.Say(twitchUser, "pong")
		case "lurk":
			twitchClient.Say(twitchUser, fmt.Sprintf("Thanks for lurking, %s", message.User.DisplayName))
		case "discord":
			twitchClient.Say(twitchUser, fmt.Sprintf("%s's Discord can be found here: %s", twitchUser, twitchUserDiscord))
		case "donate":
			twitchClient.Say(twitchUser, fmt.Sprintf("Please don't. Dropping a star on one of my github repos is enough: %s. If you really want to, then you can subscribe to the channel.", twitchUserGithub))
		case "repo":
			twitchClient.Say(twitchUser, "The bot's code can be found here: https://github.com/you-win/fpbot")
		default:
			twitchClient.Say(twitchUser, fmt.Sprintf("Unrecognized command: %s Type ~help for a list of commands", commands[0]))
		}
	})

	err = twitchClient.Connect()
	if err != nil {
		log.Printf("Unable to connect to Twitch IRC: %s", err.Error())
		return
	}
	defer twitchClient.Disconnect()

	log.Println("Connected to Twitch IRC")
}
