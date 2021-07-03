package discord_cmd

import (
    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

type sayCommand struct {
    DiscordCommand
}

func (c *sayCommand) run(args []string) {
    validRole := fputils.CheckForRole("Admin", c.Session, c.Message)

    if !validRole {
        c.Session.ChannelMessageSend(c.Message.ChannelID, "Invalid role, aborting.")
        return
    }

    if len(args) < 2 {
        c.Session.ChannelMessageSend(c.Message.ChannelID, "Not enough params specified. Need 2 params: Channel and Message")
    }

    messageGuild, err := fputils.GetMessageGuild(c.Session, c.Message)
    if err != nil {
        c.Session.ChannelMessageSend(c.Message.ChannelID, err.Error())
        return
    }

    channelToSendTo, err := fputils.GetChannelFromGuild(args[0], messageGuild)
    if err != nil {
        c.Session.ChannelMessageSend(c.Message.ChannelID, err.Error())
        return
    }

    c.Session.ChannelMessageSend(channelToSendTo.ID, args[1])
}

func NewSayCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser) *cobra.Command {
    dc := &sayCommand{
        DiscordCommand: DiscordCommand{
            Session: s,
            Message: m,
            BotData: b,
        },
    }

    c := &cobra.Command{
        Use: "say",
        Short: "Have the bot say something",
        Run: func(cmd *cobra.Command, args []string){
            dc.run(args)
        },
    }

    return c
}
