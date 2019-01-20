package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

func main() {
	config := SetupConfig()
	if !config.IsSet("DiscordKey") {
		log.Panicf("couldn't read DiscordKey from config file: %v", config.ConfigFileUsed())
	}

	dg, err := discordgo.New("Bot " + config.GetString("DiscordKey"))
	if err != nil {
		log.Printf("Error creating Discord session: %v", err)
		return
	}

	router := exrouter.New()

	router.On("add", Add).Desc("adds numbers together")

	router.On("play", Media).Desc("plays youtube videos' audio")

	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		router.FindAndExecute(dg, "!", dg.State.User.ID, m.Message)
	})

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)

	// Open the websocket and begin listening for events.
	err = dg.Open()
	if err != nil {
		log.Printf("Error opening Discord session: %v", err)
	}

	log.Println("running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStatus(0, "Botting It Up")
}

func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, err := s.ChannelMessageSend(channel.ID, "bot")
			if err != nil {
				log.Printf("couldn't send guild startup message %v", err)
			}
		}
	}
}
