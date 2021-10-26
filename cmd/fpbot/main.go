package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"fpbot/pkg/discord"
	"fpbot/pkg/twitch"
)

const (
	refreshURL = "https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s"
)

func main() {
	fmt.Println("Starting bot")

	// Get new twitch token

	client := &http.Client{}

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

	twitchToken := jsonBody["access_token"].(string)

	go discord.Run(twitchToken)
	go twitch.Run(twitchToken)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}
