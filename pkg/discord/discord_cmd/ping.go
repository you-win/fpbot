package discord_cmd

import (
    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

type pingCommand struct {
    DiscordCommand
}

func (c *pingCommand) run() {
    c.Session.ChannelMessageSend(c.Message.ChannelID, "Pong!")
}

func NewPingCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    dc := &pingCommand{
        DiscordCommand: DiscordCommand{
            Session: s,
            Message: m,
            BotData: b,
        },
    }

    c := &cobra.Command{
        Use: "ping",
        Short: "Ping the bot",
        Args: cobra.ExactArgs(0),
        Run: func(cmd *cobra.Command, args []string){
            dc.run()
        },
    }
    
    modifyUsageFunc(c, s, m)

    return c
}
