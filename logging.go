package main

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/lalamove/konfig"
	"github.com/sirupsen/logrus"
)

type DiscordHook struct {
	sess *discordgo.Session
}

func (h *DiscordHook) Fire(e *logrus.Entry) error {
	msg, err := e.String()
	if err != nil {
		return fmt.Errorf("error in coaxing Entry %v to string: %v", e, err)
	}
	_, err = h.sess.ChannelMessageSendEmbed(konfig.String("LogChannel"), &discordgo.MessageEmbed{
		Title:       "Log event",
		Description: msg,
	})
	if err != nil {
		return fmt.Errorf("error sending log embed: %v", err)
	}
	return nil
}

func (h *DiscordHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func SetupLogger(s *discordgo.Session) {
	logrus.AddHook(&DiscordHook{s})
}
