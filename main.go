package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/lalamove/konfig"
)

func main() {
	err := SetupConfig()
	if err != nil {
		log.Fatalf("error setting up config: %v", err)
	}

	for konfig.String("DiscordAPIKey") == "" {
		log.Print("couldn't read DiscordAPIKey from config file")
		time.Sleep(time.Duration(5) * time.Second)
	}

	dg, err := discordgo.New("Bot " + konfig.String("DiscordAPIKey"))
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
	router.On("buffer", Buffer).Desc("tweaks the websocket audio buffer frame count")

	// misc
	router.On("exit", Exit).Desc("exits the bot")
	router.On("log", Log).Desc("makes a call to log.Print with a message")

	dg.AddHandler(ready)
	dg.AddHandler(guildCreate)

	dg.AddHandler(func(_ *discordgo.Session, m *discordgo.MessageCreate) {
		guildCfg := konfig.StringMapString("guildCfg-" + m.GuildID)
		if guildCfg == nil { // map zero type is nil
			log.Infof("guildCfg-%v is nil, setting default", m.GuildID)
			konfig.Set("guildCfg-"+m.GuildID, DefaultGuildCfg)
		}
		log.Infof("prefix is: %v", guildCfg["prefix"])
		router.FindAndExecute(dg, guildCfg["prefix"], dg.State.User.ID, m.Message)
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
	s.UpdateStatus(0, "Botting It Up")
}

func guildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	if e.Guild.Unavailable {
		return
	}
	guildCfg := konfig.StringMapString("guildCfg-" + e.Guild.ID)
	if guildCfg == nil { // map zero type is nil
		log.Infof("guildCfg-%v is nil, setting default", m.GuildID)
		konfig.Set("guildCfg-"+e.Guild.ID, DefaultGuildCfg)
	}
	log.Infof("guildCfg-%v contents: %v", e.Guild.ID, guildCfg)
}

// Exit disconnects the bot from any voice channels, and calls os.Exit.
func Exit(ctx *exrouter.Context) {
	err := player.vConn.Speaking(false)
	if err != nil {
		log.Printf("error setting vConn.Speaking(): %v", err)
	}
	err = player.vConn.Disconnect()
	if err != nil {
		log.Printf("error calling vConn.Disconnect(): %v", err)
	}
	ctx.Reply("Exiting")
	os.Exit(0)
}

func Log(ctx *exrouter.Context) {
	msg := ctx.Args.After(1)
	ctx.Reply("logged: " + msg)
	log.Printf("log command: %v", msg)
}
