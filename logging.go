package main

import (
	"io"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/lalamove/konfig"
	"github.com/sirupsen/logrus"
)

type discordWriter struct {
	sess *discordgo.Session
}

func (w *discordWriter) Write(b []byte) (int, error) {
	n := len(b)
	w.sess.ChannelMessageSend(konfig.String("LogChannel"), string(b))
	return n, nil
}

func SetupLogger(s *discordgo.Session) {
	dw := discordWriter{s}
	multi := io.MultiWriter(os.Stdout, &dw)
	logrus.SetOutput(multi)
}
