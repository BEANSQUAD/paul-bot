package main

import (
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
