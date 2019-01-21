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
	vQueue   []videoQuery
	playing bool
	ingesting bool
}

type videoQuery struct{
	videoInfo *ytdl.VideoInfo
	query     string
	requester *discordgo.User
}

func handleErr(err error, output string) {
	if err != nil {
		log.Printf(output+", Error: %v", err)

	}
}

func Stop(ctx *exrouter.Context) {
	if player.sSession != nil {
		player.vQueue = player.vQueue[:0]
		if player.eSession.Running() {
			ctx.Reply("Stopping")
			err := player.eSession.Stop()
			handleErr(err, "Error Stopping Encoding Session")
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
	for !player.ingesting {
		time.Sleep(10 * time.Millisecond)
	}
	player.ingesting = true
	g, err := ctx.Ses.State.Guild(ctx.Msg.GuildID)
	handleErr(err, "Error Getting Guild Information")
	var vSes string
	for _, vs := range g.VoiceStates {
		if vs.UserID == ctx.Msg.Author.ID {
			vSes = vs.ChannelID
		}
	}

	player.vConn, err = ctx.Ses.ChannelVoiceJoin(g.ID, vSes, false, true)
	handleErr(err, "Error Joining Specified Voice Channel")

	videos, err := ytSearch(ctx.Args.After(1), 1)
	handleErr(err, "Error Searching Using Query")

	var vids []string
	for id := range videos {
		vids = append(vids, id)
	}

	videoStruct, err := ytdl.GetVideoInfo(vids[0])
	handleErr(err, "Error Getting Video Info")

	player.vQueue = append(player.vQueue, videoQuery{videoStruct, ctx.Args.After(1), ctx.Msg.Author})

	ctx.Reply(fmt.Sprintf("Added "+ vids[0] + " to queue"))
	
	if player.eSession == nil || !player.eSession.Running() {
		ctx.Reply(fmt.Sprintf("Playing: https://www.youtube.com/watch?v=%v", vids[0]))
		playSound(*player.vQueue[0].videoInfo)
	}
	player.ingesting = false
}

func Skip(ctx *exrouter.Context) {
	if len(player.vQueue) > 1 {
		player.vQueue = player.vQueue[1:]
		err := player.eSession.Stop()
		handleErr(err, "Error Stopping Encoding Session")
		for !player.playing{
			time.Sleep(10 * time.Millisecond)
		}
		ctx.Reply(fmt.Sprintf("Playing: https://www.youtube.com/watch?v=%v", player.vQueue[0].videoInfo.ID))
		playSound(*player.vQueue[0].videoInfo)
	}
}

func Queue(ctx *exrouter.Context) {
	for i := range player.vQueue {
		ctx.Reply(player.vQueue[i].query)
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
	if config.GetString("GoogleAPIKey") == "" {
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
	log.Printf("Calling to Youtube")
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

func playSound(videoInfo ytdl.VideoInfo) {
	player.playing = true

	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = "audio"
	options.Volume = 256
	options.CompressionLevel = 10
	options.PacketLoss = 1
	options.BufferedFrames = 1

	format := videoInfo.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	downloadURL, err := videoInfo.GetDownloadURL(format)
	handleErr(err, "Error Downloading Youtube Video")

	player.eSession, err = dca.EncodeFile(downloadURL.String(), options)
	handleErr(err, "Error Encoding Audio File")

	player.vConn.Speaking(true)

	done := make(chan error)
	player.sSession = dca.NewStream(player.eSession, player.vConn, done)
	err = <-done
	handleErr(err, "Error Streaming Audio File")

	player.vConn.Speaking(false)
	player.eSession.Cleanup()

	if len(player.vQueue) > 1{
		player.vQueue = player.vQueue[1:]
		playSound(*player.vQueue[0].videoInfo)
	}else if len(player.vQueue) == 0 {
		log.Printf("Disconnecting")
		player.vConn.Disconnect()
	}
	player.playing = false
}
