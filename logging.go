package main

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lalamove/konfig"
	"github.com/sirupsen/logrus"
)

type DiscordHook struct {
	sess *discordgo.Session
}

func (h *DiscordHook) Fire(e *logrus.Entry) error {

	color := 0x0
	switch e.Level {
	case logrus.PanicLevel, logrus.FatalLevel:
		color = 0xe00000 // red
	case logrus.ErrorLevel:
		color = 0xed6262 // soft red
	case logrus.WarnLevel:
		color = 0xf7ba13 // orange
	case logrus.InfoLevel:
		color = 0x458fff // blue
	case logrus.DebugLevel, logrus.TraceLevel:
		color = 0xffffff // white
	}
	_, err := h.sess.ChannelMessageSendEmbed(konfig.String("LogChannel"), &discordgo.MessageEmbed{
		Title:       "Log event",
		Description: e.Message,
		Timestamp:   e.Time.Format(time.RFC3339),
		Color:       color,
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
