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

var player Player

type Player struct {
	eSession *dca.EncodeSession
	sSession *dca.StreamingSession
	vConn    *discordgo.VoiceConnection
}

func handleErr(err error, output string) {
	log.Printf(output+", Error: %v", err)
}

func Stop(ctx *exrouter.Context) {
	if player.sSession != nil {
		if player.eSession.Running() {
			ctx.Reply("Stopping")
			player.sSession.SetPaused(true)
			Disconnect(ctx)
		}
	} else {
		ctx.Reply("No Sound to Stop")
	}
}

func Pause(ctx *exrouter.Context) {
	if player.sSession != nil {
		if player.sSession.Paused() {
			ctx.Reply("Resuming")
			player.sSession.SetPaused(false)
		} else {
			ctx.Reply("Pausing")
			player.sSession.SetPaused(true)
		}
	} else {
		ctx.Reply("No Sound to Pause")
	}
}

func Play(ctx *exrouter.Context) {
	g, err := ctx.Ses.State.Guild(ctx.Msg.GuildID)
	handleErr(err, "Error Getting Guild Information")

	videos, err := ytSearch(ctx.Args.After(1), 1)
	if err != nil {
		ctx.Reply(fmt.Errorf("error in ytSearch: %v", err))
	}

	var vids []string
	for id := range videos {
		vids = append(vids, id)
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == ctx.Msg.Author.ID {
			ctx.Reply(fmt.Sprintf("https://www.youtube.com/watch?v=%v", vids[0]))
			playSound(ctx.Ses, g.ID, vs.ChannelID, vids[0])
			return
		}
	}
}

func Disconnect(ctx *exrouter.Context) {
	if player.vConn == nil {
		log.Print("Tried to Disconnect when no VoiceConnections existed")
		ctx.Reply("No VoiceConnections to disconnect")
		return
	}

	err := player.vConn.Speaking(false)
	if err != nil {
		log.Printf("error setting vConn.Speaking(): %v", err)
	}
	err = player.vConn.Disconnect()
	if err != nil {
		log.Printf("error calling vConn.Disconnect(): %v", err)
		ctx.Reply("couldn't Disconnect VoiceConnection")
		return
	}
	ctx.Reply("Disconnected")
}

func ytSearch(query string, maxResults int64) (videos map[string]string, err error) {
	if !config.IsSet("GoogleAPIKey") {
		err := fmt.Errorf("GoogleAPIKey is not set in config: %v", config.ConfigFileUsed())
		log.Print(err)
		return nil, err
	}
	client := &http.Client{
		Transport: &transport.APIKey{Key: config.GetString("GoogleAPIKey")},
	}

	service, err := youtube.New(client)
	handleErr(err, "Error Creating New Youtube Client")

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(query).
		MaxResults(maxResults)
	response, err := call.Do()
	handleErr(err, "Error Listing Youtube Videos With Query")

	videos = make(map[string]string)
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
	return videos, nil
}

func playSound(s *discordgo.Session, guildID, channelID string, videoID string) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	handleErr(err, "Error Joining Specified Voice Channel")
	player.vConn = vc

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

	videoInfo, err := ytdl.GetVideoInfo(videoID)
	handleErr(err, "Error Getting Specified Youtube Video Info")

	format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	downloadURL, err := videoInfo.GetDownloadURL(format)
	handleErr(err, "Error Downloading Youtube Video")

	player.eSession, err = dca.EncodeFile(downloadURL.String(), options)
	handleErr(err, "Error Encoding Audio File")
	defer player.eSession.Cleanup()

	done := make(chan error)
	player.sSession = dca.NewStream(player.eSession, vc, done)
	err = <-done
	handleErr(err, "Error Streaming Audio File")
	time.Sleep(250 * time.Millisecond)

	vc.Speaking(false)

	vc.Disconnect()
}
