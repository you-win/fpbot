package discord

import (
	"encoding/json"
	"fmt"
	"fpbot/pkg/common"
	"fpbot/pkg/utils"
	"io/ioutil"
	"math/rand"
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

const (
	streamInfoChannelName      = "stream-info"
	streamInfoScheduleTemplate = `Streaming schedule for https://www.twitch.tv/team_youwin
All times in US Eastern time.
Streams may start/end later than listed.
--------------------------------------------------------
Sun: %s
Mon: %s
Tue: %s
Wed: %s
Thu: %s
Fri: %s
Sat: %s
--------------------------------------------------------`
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
		Name:        "roll",
		Description: "Roll a die (dice is plural)",
		Options: []*dgo.ApplicationCommandOption{
			{
				Type:        dgo.ApplicationCommandOptionInteger,
				Name:        "sides",
				Description: "Number of sides on the die",
				Required:    false,
			},
		},
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
	"roll": func(s *dgo.Session, i *dgo.InteractionCreate) {
		result := 0
		sides := 6
		rand.Seed(time.Now().UnixNano())

		if len(i.ApplicationCommandData().Options) == 0 {
			result = rand.Intn(sides) + 1
		} else {
			sides = int(i.ApplicationCommandData().Options[0].IntValue())
			if sides <= 0 {
				sides = 6
			}
			result = rand.Intn(sides) + 1
		}

		interactionRespond(s, i, fmt.Sprintf("Rolling a %d-sided die", sides))
		sendFollowupMessage(s, i, fmt.Sprintf("You rolled a: %d", result))
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
			case "help":
				interactionRespond(s, i, "```help - show this message\nupdate-stream-info <param 1> <param 2> - update stream info\ndelete-command <param 1> - delete a slash command```")
			case "update-stream-info":
				// Assume default
				newSchedule := fmt.Sprintf(streamInfoScheduleTemplate,
					"7-10pm (Programming)",
					"n/a",
					"7-10pm (Programming)",
					"9-11pm (Programming)",
					"7-10pm (Programming)",
					"n/a",
					"n/a",
				)

				streamInfoChannel, streamInfoChannelMessages, err := getStreamInfoData(s, i)
				if err != nil {
					interactionRespond(s, i, fmt.Sprintf("Unable to get stream info data: %s", err.Error()))
					return
				}

				if len(streamInfoChannelMessages) > 0 {
					s.ChannelMessageDelete(streamInfoChannel.ID, streamInfoChannelMessages[0].ID)
				}

				s.ChannelMessageSend(streamInfoChannel.ID, fmt.Sprintf("```%s```", newSchedule))
				interactionRespond(s, i, fmt.Sprintf("Successfully updated %s", streamInfoChannelName))
			case "ping-twitch":
				db.SendData <- common.NewCrossServiceData("discord ping!", common.Twitch)
				interactionRespond(s, i, "Pinged twitch from Discord!")
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
			case "update-stream-info":
				streamInfoChannel, streamInfoChannelMessages, err := getStreamInfoData(s, i)
				if err != nil {
					interactionRespond(s, i, fmt.Sprintf("Unable to get stream info data: %s", err.Error()))
					return
				}

				newSchedule := ""
				if len(streamInfoChannelMessages) > 0 {
					lastMessageID := streamInfoChannelMessages[0].ID

					lastMessage, err := s.ChannelMessage(streamInfoChannel.ID, lastMessageID)
					if err != nil {
						interactionRespond(s, i, fmt.Sprintf("Unable to get last stream info channel message: %s", err.Error()))
						return
					}

					newSchedule = lastMessage.Content
				} else {
					newSchedule = fmt.Sprintf(streamInfoScheduleTemplate,
						"7-10pm (Programming)",
						"n/a",
						"7-10pm (Programming)",
						"9-11pm (Programming)",
						"7-10pm (Programming)",
						"n/a",
						"n/a",
					)
				}

				newSchedule, err = utils.ReplaceStringAt(newSchedule, fmt.Sprintf("%s ", arg1), "\n", arg2)
				if err != nil {
					interactionRespond(s, i, fmt.Sprintf("Unable to create new schedule: %s", err.Error()))
					return
				}

				if len(streamInfoChannelMessages) > 0 {
					s.ChannelMessageDelete(streamInfoChannel.ID, streamInfoChannelMessages[0].ID)
				}

				s.ChannelMessageSend(streamInfoChannel.ID, newSchedule)
				interactionRespond(s, i, fmt.Sprintf("Successfully updated %s", streamInfoChannelName))
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

func getStreamInfoData(s *dgo.Session, i *dgo.InteractionCreate) (*dgo.Channel, []*dgo.Message, error) {
	guild, err := getInteractionGuild(s, i)
	if err != nil {
		return nil, nil, err
	}

	streamInfoChannel, err := getChannelFromGuild(guild, streamInfoChannelName)
	if err != nil {
		return nil, nil, err
	}

	streamInfoChannelMessages, err := s.ChannelMessages(streamInfoChannel.ID, 1, "", "", "")
	if err != nil {
		return nil, nil, err
	}

	return streamInfoChannel, streamInfoChannelMessages, nil
}
