package discord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	cowsay "github.com/Code-Hex/Neo-cowsay"
	dgo "github.com/bwmarrin/discordgo"
)

const (
	botRepo        = "https://github.com/you-win/fpbot"
	dadJokeBaseURL = "https://icanhazdadjoke.com/"
)

var commands = []*dgo.ApplicationCommand{
	{
		Name:        "ping",
		Description: "Ping the bot",
	},
	{
		Name:        "high-five",
		Description: "High five the bot",
	},
	{
		Name:        "whoami",
		Description: "Find out who you are",
	},
	{
		Name:        "cowsay",
		Description: "Make a cow say something",
		Options: []*dgo.ApplicationCommandOption{
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "cowsay-message",
				Description: "Message to make a cow say",
				Required:    true,
			},
		},
	},
	{
		Name:        "joke",
		Description: "Have the bot tell a joke",
	},
	{
		Name:        "repo",
		Description: "Gets the repo this bot's code is stored in",
	},
	{
		Name:        "uptime",
		Description: "Get the current uptime",
	},

	{
		Name:        "admin-command",
		Description: "Why can't subcommands take params :< Also no peeking!",
		Options: []*dgo.ApplicationCommandOption{
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "command-name",
				Description: "Command name",
				Required:    true,
			},
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "command-option-1",
				Description: "Command option 1",
				Required:    false,
			},
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "command-option-2",
				Description: "Command option 2",
				Required:    false,
			},
			{
				Type:        dgo.ApplicationCommandOptionString,
				Name:        "command-option-3",
				Description: "Command option 3",
				Required:    false,
			},
		},
	},
}

var commandHandlers = map[string]func(s *dgo.Session, i *dgo.InteractionCreate){
	"ping": func(s *dgo.Session, i *dgo.InteractionCreate) {
		interactionRespond(s, i, "pong!")
	},
	"high-five": func(s *dgo.Session, i *dgo.InteractionCreate) {
		interactionRespond(s, i, ":clap:")
	},
	"whoami": func(s *dgo.Session, i *dgo.InteractionCreate) {
		interactionRespond(s, i, s.State.User.Username)
	},
	"cowsay": func(s *dgo.Session, i *dgo.InteractionCreate) {
		message, err := cowsay.Say(
			cowsay.Phrase(i.ApplicationCommandData().Options[0].StringValue()),
			cowsay.Random(),
			cowsay.BallonWidth(40),
		)
		if err != nil {
			interactionRespond(s, i, fmt.Sprintf("Unable to cowsay: %s", err.Error()))
		}

		interactionRespond(s, i, fmt.Sprintf("\n```%s```", message))
	},
	"joke": func(s *dgo.Session, i *dgo.InteractionCreate) {
		if time.Since(bd.LastRateLimitedCommandTime).Seconds() < 5 {
			interactionRespond(s, i, "Slow down on the jokes!")
			return
		}
		client := &http.Client{}

		req, err := http.NewRequest("GET", dadJokeBaseURL, nil)
		if err != nil {
			interactionRespond(s, i, fmt.Sprintf("Cannot create GET request: %s", err.Error()))
			return
		}
		req.Header.Set("User-Agent", "Friendly Potato Discord Bot (https://github.com/you-win/fpbot)")
		req.Header.Set("Accept", "application/json")

		res, err := client.Do(req)
		if err != nil {
			interactionRespond(s, i, fmt.Sprintf("Error when sending GET request: %s", err.Error()))
			return
		}

		defer res.Body.Close()
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			interactionRespond(s, i, fmt.Sprintf("Error when reading response body: %s", err.Error()))
			return
		}

		var jsonBody map[string]interface{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			interactionRespond(s, i, fmt.Sprintf("Error when unmarshalling response: %s", err.Error()))
			return
		}

		if jsonBody["status"].(float64) != 200 {
			interactionRespond(s, i, fmt.Sprintf("Non-200 response code: %s", jsonBody))
			return
		}

		interactionRespond(s, i, fmt.Sprintf("```%s```", jsonBody["joke"].(string)))

		bd.LastRateLimitedCommandTime = time.Now()
	},
	"repo": func(s *dgo.Session, i *dgo.InteractionCreate) {
		interactionRespond(s, i, botRepo)
	},
	"uptime": func(s *dgo.Session, i *dgo.InteractionCreate) {
		interactionRespond(s, i, time.Since(bd.StartTime).String())
	},

	"admin-command": func(s *dgo.Session, i *dgo.InteractionCreate) {
		var g *dgo.Guild
		for _, guild := range s.State.Guilds {
			if strings.EqualFold(guild.ID, guildID) {
				g = guild
			}
		}

		roleID := ""
		for _, role := range g.Roles {
			if strings.EqualFold(role.Name, "Admin") {
				roleID = role.ID
			}
		}

		if len(roleID) < 1 {
			interactionRespond(s, i, "Unable to find role")
			return
		}

		if i.Member == nil {
			interactionRespond(s, i, "Unable to respond to dms")
			return
		}

		hasRole := false
		for _, role := range i.Member.Roles {
			if strings.EqualFold(role, roleID) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			interactionRespond(s, i, "Only admins can use this command")
			return
		}

		arg := i.ApplicationCommandData().Options[0].StringValue()

		switch len(i.ApplicationCommandData().Options) {
		case 1:
			switch arg {
			default:
				interactionRespond(s, i, fmt.Sprintf("Unhandled command parameter for: %s", arg))
			}
		case 2:
			arg1 := i.ApplicationCommandData().Options[1].StringValue()
			switch arg {
			case "delete-command":
				appCommands, err := s.ApplicationCommands(s.State.User.ID, guildID)
				if err != nil {
					interactionRespond(s, i, "Unable to find application commands")
				}

				appCmdID := ""
				for _, cmd := range appCommands {
					if strings.EqualFold(cmd.Name, arg1) {
						appCmdID = cmd.ID
					}
				}

				if len(appCmdID) < 1 {
					interactionRespond(s, i, fmt.Sprintf("Unable to find application command: %s", arg1))
					return
				}

				s.ApplicationCommandDelete(s.State.User.ID, guildID, appCmdID)

				interactionRespond(s, i, fmt.Sprintf("Deleted command: %s", arg1))
			default:
				interactionRespond(s, i, fmt.Sprintf("Unhandled command parameters for: %s %s, aborting", arg, arg1))
			}
		case 3:
			arg1 := i.ApplicationCommandData().Options[1].StringValue()
			arg2 := i.ApplicationCommandData().Options[2].StringValue()
			switch arg {
			default:
				interactionRespond(s, i, fmt.Sprintf("Unhandled command parameters for %s %s %s, aborting", arg, arg1, arg2))
			}
		case 4:
			arg1 := i.ApplicationCommandData().Options[1].StringValue()
			arg2 := i.ApplicationCommandData().Options[2].StringValue()
			arg3 := i.ApplicationCommandData().Options[3].StringValue()
			switch arg {
			default:
				interactionRespond(s, i, fmt.Sprintf("Unhandled command parameters for %s %s %s %s, aborting", arg, arg1, arg2, arg3))
			}
		default:
			interactionRespond(s, i, fmt.Sprintf("Unhandled command parameters for %s, aborting", arg))
		}
	},
}

func interactionRespond(s *dgo.Session, i *dgo.InteractionCreate, message string) {
	s.InteractionRespond(i.Interaction, &dgo.InteractionResponse{
		Type: dgo.InteractionResponseChannelMessageWithSource,
		Data: &dgo.InteractionResponseData{
			Content: message,
		},
	})
}

func sendFollowupMessage(s *dgo.Session, i *dgo.InteractionCreate, message string) {
	s.FollowupMessageCreate(s.State.User.ID, i.Interaction, true, &dgo.WebhookParams{
		Content: message,
	})
}
