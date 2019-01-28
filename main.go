package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func main() {
	err := SetupConfig()
	if err != nil {
		log.Fatalf("error setting up config: %v", err)
	}

	for viper.GetString("discord-key") == "" {
		log.Print("couldn't read discord-key from config file")
		time.Sleep(time.Duration(5) * time.Second)
	}

	dg, err := discordgo.New("Bot " + viper.GetString("discord-key"))
	if err != nil {
		log.Printf("error creating Discord session: %v", err)
		return
	}

	SetupLogger(dg)

	router := exrouter.New()

	// math
	router.On("add", Add).Desc("adds numbers together")

	// media
	router.On("play", Play).Desc("plays audio from source/query audio")
	router.On("stop", Stop).Desc("stops current audio playing")
	router.On("pause", Pause).Desc("pause and unpause current audio playing")
	router.On("disconnect", Disconnect).Desc("disconnect from the current guilds voice channel")
	router.On("skip", Skip).Desc("skips the currently playing video")
	router.On("queue", Queue).Desc("shows the video queue")

	// logging
	router.On("log", GenerateLogEvent).Desc("makes a call to log.Print with a message")

	// config
	router.On("set", GuildConfigSet).Desc("changes settings for the current server")
	router.On("gset", GlobalConfigSet).Desc("changes global settings")

	// misc
	router.On("exit", Exit).Desc("exits the bot")

	// TEMP POO, ALEX NO HURT ME
	router.On("poo", tempPoo).Desc("returns the url of cytube/memebase")

	dg.AddHandler(ready)
	dg.AddHandler(initGuildCfg)

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.ID == s.State.User.ID {
			return
		}
		prefix := viper.GetString(fmt.Sprintf("guild.%v.prefix", m.GuildID))
		router.FindAndExecute(dg, prefix, dg.State.User.ID, m.Message)
	})

	// Open the websocket and begin listening for events.
	err = dg.Open()
	if err != nil {
		log.Printf("error opening Discord session: %v", err)
	}
	defer dg.Close()

	log.Println("running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}

func ready(s *discordgo.Session, _ *discordgo.Ready) {
	s.UpdateStatus(0, viper.GetString("status-message"))
}

func tempPoo(ctx *exrouter.Context) {
	ctx.Reply("https://cytu.be/r/meme_base password: botnetqq")
}

// Exit disconnects the bot from any voice channels, and calls os.Exit.
func Exit(ctx *exrouter.Context) {
	Disconnect(ctx)
	log.Printf("Exiting")
	defer os.Exit(0)
}
