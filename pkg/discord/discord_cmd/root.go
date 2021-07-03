package discord_cmd

import (
    "fmt"
    "io/ioutil"
    "bytes"

    fputils "fpbot/pkg/utils"

    "github.com/spf13/cobra"
    dgo "github.com/bwmarrin/discordgo"
)

type DiscordCommand struct {
    Session *dgo.Session
    Message *dgo.Message
    BotData fputils.BotDataAccesser
    buffer bytes.Buffer
}

func (dc *DiscordCommand) Write(p []byte) (n int, err error) {
    dc.buffer.Reset()
    return dc.buffer.Write(p)
}

func NewCommand(s *dgo.Session, m *dgo.Message, b fputils.BotDataAccesser, args []string) *cobra.Command {
    c := &cobra.Command{
        Use: fputils.DiscordBotPrefix,
    }

    c.AddCommand(
        NewPingCommand(s, m, b),
        NewWhoAmICommand(s, m, b),
        NewUptimeCommand(s, m, b),
        NewRepoCommand(s, m, b),
        NewCowsayCommand(s, m, b),
        NewHighFiveCommand(s, m, b),
        NewUserDataCommand(s, m, b),
        NewAdminCommand(s, m, b),
    )

    c.SetArgs(args)
    modifyUsageFunc(c, s, m)
    
    output := ioutil.Discard
    c.SetOut(output)
    c.SetErr(output)

    return c
}

func modifyUsageFunc(c *cobra.Command, s *dgo.Session, m *dgo.Message) {
    usageString := c.UsageString()
    c.SetUsageFunc(func(*cobra.Command) error {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s```", usageString))

        return nil
    })
}
