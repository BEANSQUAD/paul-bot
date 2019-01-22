package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
)

// Exit disconnects the bot, and exits the bot.
// The configuration of docker means that the bot is automatically restarted.
func Exit(ctx *exrouter.Context) {
	err := player.vConn.Speaking(false)
	if err != nil {
		log.Printf("error setting vConn.Speaking(): %v", err)
	}
	// I've commented this out because apperently it's an 'ineffectual assignment' and I don't know what it will break
	// The disconnect func returns an error but i forgot to handle the error here
	err = player.vConn.Disconnect()
	if err != nil {
		log.Printf("error calling vConn.Disconnect(): %v", err)
	}
	ctx.Reply("Restarting")
	os.Exit(1)
}

func main() {
	SetupConfig()
	if config.GetString("DiscordKey") == "" {
		log.Panicf("couldn't read DiscordKey from config file: %v", config.ConfigFileUsed())
	}

	dg, err := discordgo.New("Bot " + config.GetString("DiscordKey"))
	if err != nil {
		log.Printf("Error creating Discord session: %v", err)
		return
	}

	router := exrouter.New()

	router.On("add", Add).Desc("adds numbers together")

	router.On("play", Play).Desc("plays audio from source/query audio")
	router.On("stop", Stop).Desc("stops current audio playing")
	router.On("pause", Pause).Desc("(un)pause current audio playing")
	router.On("disconnect", Disconnect).Desc("disconnect from the current guilds voice channel")
	router.On("skip", Skip).Desc("disconnect from the current guilds voice channel")
	router.On("queue", Queue).Desc("disconnect from the current guilds voice channel")
	router.On("fuckoff", Exit).Desc("Calls os.exit with")

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
			//JMJ wasnt a fan of it spamming the chat so i have disabled this for now
			/*_, err := s.ChannelMessageSend(channel.ID, "(╯°□°）╯︵ ┻━┻)")
			if err != nil {
				log.Printf("couldn't send guild startup message %v", err)
			}*/
		}
	}
}
