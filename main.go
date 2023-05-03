package main

import (
	"encoding/json"
	//"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Variables used for command line parameters
/*
var (
	Token string
)

func init() {
	flag.Parse()
	Token = flag.Arg(1)
}
*/

func main() {
	// Create a new Discord session using the provided bot token.

	Token := os.Args[len(os.Args)-1]
	if len(Token) != 70 {
		fmt.Print("invalid authkey")
		return
	}

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Each type of event to respond to must add a handler.
	dg.AddHandler(messageCreate)
	dg.AddHandler(messageUpdate)
	dg.AddHandler(MessageReactionAdd)
	dg.AddHandler(guildMemberadd)
	dg.AddHandler(threadCreate)

	// Intents for permission to do things.
	// The `|=` operator is the bitwise OR operator
	dg.Identify.Intents |= discordgo.IntentGuilds                // Delete threads, not working for some reason?
	dg.Identify.Intents |= discordgo.IntentGuildMembers          // Manage nicknames
	dg.Identify.Intents |= discordgo.IntentGuildMessages         // Manage messages
	dg.Identify.Intents |= discordgo.IntentGuildMessageReactions // Manage reactions
	dg.Identify.Intents |= discordgo.IntentDirectMessages        // Just incase for private commands
	dg.Identify.Intents |= discordgo.IntentsMessageContent       // read message content

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
	if isViolatingMessage(s, m.Message) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func messageUpdate(s *discordgo.Session, m *discordgo.MessageUpdate) {
	if isViolatingMessage(s, m.Message) {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func MessageReactionAdd(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
	if err := s.MessageReactionsRemoveAll(r.ChannelID, r.MessageID); err != nil {
		fmt.Printf("Error, could not remove all reactions from %s\n%s", r.MessageID, err)
	}
}

// TODO: currently broken due to permissions error.
func threadCreate(s *discordgo.Session, t *discordgo.ThreadCreate) {
	parent, err := s.Channel(t.ID)
	switch {
	case err != nil:
		fmt.Println(err)
		return
	case parent.Name != "h":
		return
	}
	if _, err := s.ChannelDelete(t.ID); err != nil {
		fmt.Printf("Failed to delete thread:\nID: %s\nName: %s\nReason: %s\n", t.ID, t.Name, err)
	}
}

func guildMemberadd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	if err := s.GuildMemberNickname(m.GuildID, m.User.ID, "h"); err != nil {
		fmt.Println(err)
	}
}

func guildMemberUpdate(s *discordgo.Session, m *discordgo.GuildMemberUpdate) {
	if m.Nick != "h" {
		if err := s.GuildMemberNickname(m.GuildID, m.User.ID, "h"); err != nil {
			fmt.Println(err)
		}
	}
}

func printJson(b []byte) error {
	if j, err := json.MarshalIndent(b, "", "  "); err != nil {
		return err
	} else {
		fmt.Println(string(j))
		return nil
	}
}

func printDiscordMessage(m *discordgo.Message) {
	var b []byte
	if err := m.UnmarshalJSON(b); err != nil {
		fmt.Println("UnmarshalJSON error:", err)
	}
	printJson(b)
}

func isViolatingMessage(s *discordgo.Session, m *discordgo.Message) bool {
	channel, err := s.Channel(m.ChannelID)
	switch {
	case err != nil:
		fmt.Println(err)
		return false
	// Common objects that get passed via `*discordgo.MessageUpdate` events cause crashes
	// due to attributes being null. An exception is that those objects always have a
	// zeroed timestamp. This check fixed crashes related to thread creation and deletion.
	// Timestamp is deprecated and may disappear in the future, so potentially come up
	// with a different method to avoid this.
	case m.Timestamp.IsZero():
		return false
	case m.Author.ID == s.State.User.ID:
		return false
	case channel.Name != "h":
		return false
	case m.Type != discordgo.MessageTypeDefault:
		return true
	case m.Member.Nick != "h":
		return true
	case m.Content != "h":
		return true
	case len(m.Attachments) > 0:
		return true
	case len(m.Mentions) > 0:
		return true
	case len(m.Reactions) > 0:
		return true
	default:
		return false
	}
}
