package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
	"github.com/Necroforger/dgrouter/exrouter"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

func Play(ctx *exrouter.Context){

	m := ctx.Msg
	s := ctx.Ses

	// Find the channel that the message came from.
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		// Could not find channel.
		return
	}

	// Find the guild for that channel.
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		// Could not find guild.
		return
	}

	// Look for the message sender in that guild's current voice states.
	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			go playSound(s, g.ID, vs.ChannelID, strings.TrimLeft(m.Content, "!play "))
			if err != nil {
				fmt.Println("Error playing sound:", err)
			}

			return
		}
	}
}

const developerKey = "AIzaSyDxE51o2JqlECAQYCMJ9ytjYzgLH_uON-Y" //this is temp and a bit of a bodge to get the youtube API working for now

func ytSearch(query string, maxResults int64) map[string]string {
	client := &http.Client{
		Transport: &transport.APIKey{Key: developerKey},
	}

	service, err := youtube.New(client)
	if err != nil {
		log.Fatalf("Error creating new YouTube client: %v", err)
	}

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(query).
		MaxResults(maxResults)
	response, err := call.Do()
	if err != nil {
	}

	// Group video, channel, and playlist results in separate lists.
	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

	// Iterate through each item and add it to the correct list.
	for _, item := range response.Items {
		switch item.Id.Kind {
		case "youtube#video":
			videos[item.Id.VideoId] = item.Snippet.Title
		case "youtube#channel":
			channels[item.Id.ChannelId] = item.Snippet.Title
		case "youtube#playlist":
			playlists[item.Id.PlaylistId] = item.Snippet.Title
		}
	}

	return videos
}

// PlaySound plays the current buffer to the provided channel.
func playSound(s *discordgo.Session, guildID, channelID string, search string) {
	videos := ytSearch(search, 1)
	var vids []string

	for id := range videos {
		vids = append(vids, id)
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Change these accordingly
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = "audio"
	options.Volume = 256
	options.CompressionLevel = 10
	options.PacketLoss = 1
	options.BufferedFrames = 100

	videoInfo, err := ytdl.GetVideoInfo(vids[0])
	if err != nil {
	}

	format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	downloadURL, err := videoInfo.GetDownloadURL(format)
	if err != nil {
	}

	fmt.Println("1")

	encodingSession, err := dca.EncodeFile(downloadURL.String(), options)
	if err != nil {
	}
	defer encodingSession.Cleanup()

	fmt.Println("2")

	done := make(chan error)
	dca.NewStream(encodingSession, vc, done)
	err = <-done
	if err != nil {
	}

	fmt.Println("3")

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	vc.Disconnect()
}
