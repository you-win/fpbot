package discord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	// fpmodel "fpbot/pkg/model"

	"github.com/bwmarrin/discordgo"
)

var (
	discordToken string
	guildID      string
)

type DiscordRunner struct {
	// twitchMessages chan
}

type BotData struct {
	StartTime                  time.Time
	LastRateLimitedCommandTime time.Time
}

var bd *BotData

var s *discordgo.Session

func Run(twitchToken string) {
	bd = &BotData{
		StartTime:                  time.Now(),
		LastRateLimitedCommandTime: time.Now(),
	}

	discordToken = os.Getenv("DISCORD_TOKEN")
	guildID = os.Getenv("GUILD_ID")

	s, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Fatal("Error creating Discord session: ", err)
	}

	// Slash commands
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	// Anti-spam
	as := NewAntiSpam()
	s.AddHandler(as.handleSpam)

	// Reaction roles
	rr := newReactionRoles()
	s.AddHandler(rr.handleReady)
	s.AddHandler(rr.handleReactionAdd)
	s.AddHandler(rr.handleReactionRemove)

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	defer s.Close()

	for _, v := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, v)
		if err != nil {
			log.Fatalf("Cannot create '%v' command: %v", v.Name, err)
		}
		log.Printf("Added %s", v.Name)
	}

	log.Println("Bot is running...")

	twitchClient := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.twitch.tv/helix/streams?user_id=%s", "44149998"), nil)
	if err != nil {
		log.Printf("Unable to create request for polling twitch stream: %s", err.Error())
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", twitchToken))
	req.Header.Add("Client-Id", os.Getenv("TWITCH_CLIENT_ID"))

	isLive := false
	for {
		time.Sleep(time.Second * 60)

		res, err := twitchClient.Do(req)
		if err != nil {
			log.Printf("Error when polling twitch stream: %s", err.Error())
			// Don't exit here, since I guess this can just fail sometimes
			if res != nil && res.Body != nil {
				res.Body.Close()
			}
			continue
		}
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Printf("Unable to read response body: %s", err.Error())
			res.Body.Close()
			continue
		}

		var jsonBody map[string]interface{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			log.Printf("Unable to unmarshal json response: %s", err.Error())
			res.Body.Close()
			continue
		}

		if res.StatusCode != 200 {
			log.Printf("Bad response from user endpoint: %s", jsonBody)
			res.Body.Close()
			continue
		}

		data := jsonBody["data"].([]interface{})

		if len(data) == 0 {
			isLive = false
		} else {
			if !isLive {
				s.ChannelMessageSend("901833984542134293", "<@&901528644382519317> team_youwin is live at https://www.twitch.tv/team_youwin")
				isLive = true
			}
		}

		res.Body.Close()
	}

	// sc := make(chan os.Signal, 1)
	// signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	// <-sc
	// log.Println("Bot leaving this plane of existence. おやすみ")
}
