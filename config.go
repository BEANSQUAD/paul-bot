package main

import (
	"fmt"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// DefaultGuildCfg contains the default options that should exist for a guild
var DefaultGuildCfg = map[string]string{
	"prefix": "!",
}

// DefaultGlobalCfg contains the default options that should exist for the bot
var DefaultGlobalCfg = map[string]string{
	"DiscordAPIKey": "",
	"GoogleAPIKey":  "",
	"LogChannel":    "",
	"StatusMessage": "paul-bot",
}

// SetupConfig registers a viper instance, setting default values
// for the bot and to load configuration values from a file
// (e.g. API keys, persistent guild settings)
func SetupConfig() error {
	viper.SetConfigType("toml")
	viper.SetConfigName("vars")
	viper.AddConfigPath("./config/")

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	for k, v := range DefaultGlobalCfg {
		viper.SetDefault(k, v)
	}

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		log.Debugf("Config file changed: %v", e.Name)
	})
	return nil
}

func configSet(key string, value interface{}) error {
	// shouldn't be needed due to WatchConfig() looking for changes
	err := viper.MergeInConfig()
	if err != nil {
		return err
	}

	viper.Set(key, value)

	err = viper.WriteConfig()
	if err != nil {
		return err
	}
	return nil
}

func initGuildCfg(s *discordgo.Session, e *discordgo.GuildCreate) {
	if e.Guild.Unavailable {
		return
	}
	log.Infof("setting defaults for new guild: %v (%v)", e.Guild.Name, e.Guild.ID)
	viper.SetDefault("guild."+e.Guild.ID, DefaultGuildCfg)
	err := viper.WriteConfig()
	if err != nil {
		log.Errorf("could not write config while setting %v: %v", e.Guild.ID, err)
	}
}

func GuildConfigSet(ctx *exrouter.Context) {
	key := ctx.Args.Get(1)
	value := ctx.Args.Get(2)
	guildKey := fmt.Sprintf("guild.%v.%v", ctx.Msg.GuildID, key)

	err := configSet(guildKey, value)
	if err != nil {
		log.Errorf("could not set guild config %v to %v: %v", guildKey, value, err)
		ctx.Reply("couldn't set %v to %v", key, value)
	}
}
func GlobalConfigSet(ctx *exrouter.Context) {
}
