package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
var (
	TokenArgument string
	TokenFile     string
	Token         string
	err           error
)

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.StringVar(&TokenFile, "f", "", "Bot Token File")
	flag.Parse()
}

func main() {
	// Create a new Discord session using the provided bot token.
	if TokenArgument != "" {
		Token = TokenArgument
	} else if TokenFile != "" {
		var b []byte
		b, err = os.ReadFile(TokenFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		Token = string(b)
	} else {
		flag.Usage()
		return
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageUpdate)
	dg.AddHandler(messageReactionAdd) // Not working??

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	fmt.Println("Shutting down...")
	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content != "h" || len(m.Attachments) > 0 || len(m.Mentions) > 0 || m.Type != 0 {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		printDiscordMessage(m.Message)
	}

}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	s.ChannelMessageDelete(m.ChannelID, m.ID)
	printDiscordMessage(m.Message)
}

// TODO: FIX
func messageReactionAdd(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	fmt.Printf("%s used an emoji, deleting it..\n", e.Member.User.Username)
	s.MessageReactionRemove(e.ChannelID, e.MessageID, e.Emoji.ID, e.UserID)
}

func printDiscordMessage(m *discordgo.Message) {
	res, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(res))
}