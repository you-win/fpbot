package discord_cmd

import (
    "time"

    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

type uptimeCommand struct {
    DiscordCommand
}

func (c *uptimeCommand) run() {
    c.Session.ChannelMessageSend(c.Message.ChannelID, time.Since(c.BotData.GetStartTime()).String())

    if c.buffer.Len() > 0 {
        c.Session.ChannelMessageSend(c.Message.ChannelID, c.buffer.String())
    }
}

func NewUptimeCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    dc := &uptimeCommand{
        DiscordCommand: DiscordCommand{
            Session: s,
            Message: m,
            BotData: b,
        },
    }

    c := &cobra.Command{
        Use: "uptime",
        Short: "Get the current uptime",
        Long: "Finds information about the current user",
        Args: cobra.ExactArgs(0),
        Run: func(cmd *cobra.Command, args []string){
            dc.run()
        },
    }
    
    modifyUsageFunc(c, s, m)

    return c
}