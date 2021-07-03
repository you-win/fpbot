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
        Use: "<>",
        Short: "uwu",
        Long: "UwU",
    }

    c.AddCommand(
        NewPingCommand(s, m, b),
        NewWhoAmICommand(s, m, b),
        NewUptimeCommand(s, m, b),
        NewRepoCommand(s, m, b),
        NewCowsayCommand(s, m, b),
        NewHighFiveCommand(s, m, b),

        // Admin commands
        NewAdminCommand(s, m, b),
    )

    c.SetArgs(args)
    usageString := c.UsageString()
    c.SetUsageFunc(func(*cobra.Command) error {
        s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("```%s```", usageString))

        return nil
    })
    output := ioutil.Discard
    c.SetOut(output)
    c.SetErr(output)

    return c
}
