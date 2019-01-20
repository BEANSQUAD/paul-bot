package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

func Play(ctx *exrouter.Context) {
	g, err := ctx.Ses.State.Guild(ctx.Msg.GuildID)
	if err != nil {
		return
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == ctx.Msg.Author.ID {
			go playSound(ctx.Ses, g.ID, vs.ChannelID, ctx.Args.After(1))
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

	videos := make(map[string]string)
	channels := make(map[string]string)
	playlists := make(map[string]string)

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

func playSound(s *discordgo.Session, guildID, channelID string, search string) {
	videos := ytSearch(search, 1)
	var vids []string

	for id := range videos {
		vids = append(vids, id)
	}

	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
	}

	vc.Speaking(true)
	time.Sleep(250 * time.Millisecond)

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
	time.Sleep(250 * time.Millisecond)

	fmt.Println("3")

	vc.Speaking(false)

	vc.Disconnect()
}
