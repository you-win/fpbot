package twitch

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgo "github.com/gempir/go-twitch-irc/v2"
)

const refreshURL = "https://id.twitch.tv/oauth2/token?grant_type=refresh_token&refresh_token=%s&client_id=%s&client_secret=%s"

func Run() {
	client := &http.Client{}

	twitchUser := os.Getenv("TWITCH_USER")

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
		log.Println("Connected")
	})

	twitchClient.OnPrivateMessage(func(message tgo.PrivateMessage) {
		log.Println(message.Message)
	})

	err = twitchClient.Connect()
	if err != nil {
		log.Printf("Unable to connect to Twitch IRC: %s", err.Error())
		return
	}
	defer twitchClient.Disconnect()

	log.Println("Connected to Twitch IRC")
}
