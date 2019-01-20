package main

import (
	//"io/ioutil"
	"log"
	"os"
	"os/signal"
	//"strings"
	"syscall"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

func main() {
	/*dat, err := ioutil.ReadFile("/etc/paul-bot.key")
	if err != nil {
		log.Panicf("couldn't read /etc/paul-bot.key: %v", err)
	}
	Token := strings.TrimSuffix(string(dat), "\n")*/
	Token := "NTM1NzY2MzQ3NTIzNjg2NDIw.DyZiFw.k0aYTAUOgY8WyN52aZIYaWLV9Rw"

	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Printf("Error creating Discord session: %v", err)
		return
	}

	router := exrouter.New()

	router.On("add", Add).Desc("adds numbers together")

	router.On("play", Play).Desc("plays youtube videos' audio")
	router.On("info", Info).Desc("gets info on current session")
	router.On("stop", Stop).Desc("gets info on current session")

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
