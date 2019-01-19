package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/rylio/ytdl"
	"github.com/jonas747/dca"

	"google.golang.org/api/youtube/v3"
	"google.golang.org/api/googleapi/transport"
)

func main() {
	dat, err := ioutil.ReadFile("/etc/paul-bot.key")
	if err != nil {
		panic(err)
	}
	Token := strings.TrimSuffix(string(dat), "\n")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong6!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	tokens := strings.Split(m.Content, " ")
	if tokens[0] == "!add" {
		args := tokens[1:]
		s.ChannelMessageSend(m.ChannelID, strconv.Itoa(sum(args)))
	}
}

func sum(nums []string) int {
	total := 0
	for _, num := range nums {
		n, err := strconv.Atoi(num)
		if err != nil {
			fmt.Println("error casting string to int,", err)
		}
		total += n
	}
	return total
}

func onGuild(s *discordgo.Session, evt *discordgo.GuildCreate) {
	s.ChannelMessageSend(evt.SystemChannelID, "good2go")
}
