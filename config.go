package main

import (
	"fmt"
	"strings"

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
	"discord-key":    "",
	"google-key":     "",
	"log-channel":    "",
	"status-message": "paul-bot",
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

func configSet(key string, value string) error {
	log.Infof("setting %v => %v", key, value)
	viper.Set(key, value)
	err := viper.MergeInConfig()
	if err != nil {
		return fmt.Errorf("couldn't merge config: %v", err)
	}
	err = viper.WriteConfig()
	if err != nil {
		return fmt.Errorf("couldn't write config: %v", err)
	}
	return nil
}

func checkDefault(defaultMap map[string]string, key string) error {
	keyTrans := strings.Split(strings.ToLower(key), ".")
	key = keyTrans[len(keyTrans)-1]
	validKeys := make([]string, 0, len(defaultMap))

	for k := range defaultMap {
		if !strings.Contains("key", k) {
			validKeys = append(validKeys, k)
		}
	}
	err := fmt.Errorf("key %v must be one of %v", key, validKeys)

	if _, ok := defaultMap[key]; !ok {
		log.Infof("%#v is not in %#v", key, defaultMap)
		return err
	}
	return nil
}

func GuildConfigSet(ctx *exrouter.Context) {
	key := ctx.Args.Get(1)
	value := ctx.Args.Get(2)
	guildKey := fmt.Sprintf("guild.%v.%v", ctx.Msg.GuildID, key)

	if err := checkDefault(DefaultGuildCfg, guildKey); err != nil {
		ctx.Reply(err)
		return
	}

	if value == "" {
		ctx.Reply(fmt.Sprintf("%v = %v", key, viper.Get(guildKey)))
		return
	}

	err := configSet(guildKey, value)
	if err != nil {
		log.Errorf("couldn't set config %v => %v: %v", guildKey, value, err)
	}

	ctx.Reply(fmt.Sprintf("set %v to %v", key, value))
}

func GlobalConfigSet(ctx *exrouter.Context) {
	key := ctx.Args.Get(1)
	value := ctx.Args.Get(2)

	if err := checkDefault(DefaultGlobalCfg, key); err != nil {
		ctx.Reply(err)
		return
	}

	if value == "" {
		ctx.Reply(fmt.Sprintf("%v = %v", key, viper.Get(key)))
		return
	}

	err := configSet(key, value)
	if err != nil {
		log.Errorf("couldn't set config %v => %v: %v", key, value, err)
	}

	ctx.Reply(fmt.Sprintf("globally set %v to %v", key, value))
}
