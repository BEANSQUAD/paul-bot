package main

import (
	"log"

	"github.com/spf13/viper"
)

func SetupConfig() *viper.Viper {
	config := viper.New()
	config.SetConfigType("toml")
	config.SetConfigName("config")
	config.AddConfigPath("/etc/paul-bot")
	config.SetDefault("DiscordKey", "")

	err := config.ReadInConfig()
	if err != nil {
		log.Panicf("error reading config file: %v", err)
	}
	return config
}
