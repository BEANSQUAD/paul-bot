package main

import (
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

// SetupConfig registers a viper instance to load configuration values from a
// file (e.g. API keys, persistent guild settings)
func SetupConfig() error {
	viper.SetConfigType("toml")
	viper.SetConfigName("vars")
	viper.AddConfigPath("./config/")

	err := viper.ReadInConfig()
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
	guildCfg := viper.GetStringMapString("guild." + e.Guild.ID)
	log.Infof("guild.%v is %v", e.Guild.ID, guildCfg)
	if len(guildCfg) == 0 {
		log.Infof("setting guild %v to default", e.Guild.ID)
		viper.SetDefault("guild."+e.Guild.ID, DefaultGuildCfg)
		err := viper.WriteConfig()
		if err != nil {
			log.Errorf("error writing config while setting %v: %v", e.Guild.ID, err)
		}
	}
	guildCfg = viper.GetStringMapString("guild." + e.Guild.ID)
	log.Infof("guild.%v is %v", e.Guild.ID, guildCfg)
}

func GuildConfigSet(ctx *exrouter.Context) {
}
func GlobalConfigSet(ctx *exrouter.Context) {
}
