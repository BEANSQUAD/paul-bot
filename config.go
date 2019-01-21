package main

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

// SetupConfig uses viper to configure the go environment, using the parameters set within the config file.
// This includes setting the API keys for both Discord and Google.
// Throws an error should the config file be unable to be read.
func SetupConfig() {
	config = viper.New()
	config.SetConfigType("toml")
	config.SetConfigName("config")
	config.AddConfigPath("/etc/paul-bot")

	config.SetDefault("DiscordKey", "")
	config.SetDefault("GoogleAPIKey", "")

	err := config.ReadInConfig()
	if err != nil {
		log.Panicf("error reading config file: %v", err)
	}
}
