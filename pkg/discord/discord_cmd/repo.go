package discord_cmd

import (
    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

const botRepo = "https://github.com/you-win/fpbot"

type repoCommand struct {
    DiscordCommand
}

func (c *repoCommand) run() {
    c.Session.ChannelMessageSend(c.Message.ChannelID, botRepo)
}

func NewRepoCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    dc := &repoCommand{
        DiscordCommand: DiscordCommand{
            Session: s,
            Message: m,
            BotData: b,
        },
    }

    c := &cobra.Command{
        Use: "repo",
        Short: "Gets the repo this bot's code is stored in",
        Args: cobra.ExactArgs(0),
        Run: func(cmd *cobra.Command, args []string){
            dc.run()
        },
    }
    c.SetOut(dc)
    c.SetErr(dc)

    return c
}