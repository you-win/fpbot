package discord_cmd

import (
    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

type adminCommand struct {
    DiscordCommand
}

func NewAdminCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    // dc := &adminCommand{
    //     DiscordCommand: DiscordCommand{
    //         Session: s,
    //         Message: m,
    //         BotData: b,
    //     },
    // }

    c := &cobra.Command{
        Use: "admin",
        Short: "Run admin commands",
    }

    c.AddCommand(NewSayCommand(s, m, b))

    return c
}
