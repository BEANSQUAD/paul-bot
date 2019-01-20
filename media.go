package main

import (
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

type Player struct {
	eSession *dca.EncodeSession
	sSession *dca.StreamingSession
}

func handleErr(err error, output string){
	log.Printf(output + ", Error: %v", err)
}

func Media(ctx *exrouter.Context) {
	g, err := ctx.Ses.State.Guild(ctx.Msg.GuildID)
	handleErr(err, "Error Getting Guild Information")
	
	videos := ytSearch(ctx.Args.After(1), 1)
	var vids []string

	for id := range videos {
		vids = append(vids, id)
	}

	var player Player

	for _, vs := range g.VoiceStates {
		if vs.UserID == ctx.Msg.Author.ID {
			go playSound(ctx.Ses, g.ID, vs.ChannelID, vids[0], player)
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
	handleErr(err, "Error Creating New Youtube Client")

	// Make the API call to YouTube.
	call := service.Search.List("id,snippet").
		Q(query).
		MaxResults(maxResults)
	response, err := call.Do()
	handleErr(err, "Error Listing Youtube Videos With Query")

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

func playSound(s *discordgo.Session, guildID, channelID string, videoID string, player Player) {
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	handleErr(err, "Error Joining Specified Voice Channel")

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
