package main

import (
	"log"

	"github.com/spf13/viper"
)

var config *viper.Viper

func SetupConfig() {
	config = viper.New()
	config.SetConfigType("toml")
	config.SetConfigName("config")
	config.AddConfigPath("/etc/paul-bot")

	err := config.ReadInConfig()
	if err != nil {
		log.Panicf("error reading config file: %v", err)
	}
}
