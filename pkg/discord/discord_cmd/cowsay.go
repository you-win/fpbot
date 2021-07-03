package discord_cmd

import (
    "fmt"

    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
    cowsay "github.com/Code-Hex/Neo-cowsay"
)

type cowsayCommand struct {
    DiscordCommand
}

func (c *cowsayCommand) run(args []string) {
    cowMessage := "moo"
    if len(args) > 0 {
        cowMessage = args[0]
    }
    say, err := cowsay.Say(
        cowsay.Phrase(cowMessage),
        cowsay.Random(),
        cowsay.BallonWidth(40),
    )
    if err != nil {
        c.Session.ChannelMessageSend(c.Message.ChannelID, err.Error())
        return
    }

    c.Session.ChannelMessageSend(c.Message.ChannelID, fmt.Sprintf("\n```%s```", say))
}

func NewCowsayCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    dc := &cowsayCommand{
        DiscordCommand: DiscordCommand{
            Session: s,
            Message: m,
            BotData: b,
        },
    }

    c := &cobra.Command{
        Use: "cowsay",
        Short: "Have a cow say something",
        Run: func(cmd *cobra.Command, args []string){
            dc.run(args)
        },
    }
    c.SetOut(dc)
    c.SetErr(dc)

    return c
}