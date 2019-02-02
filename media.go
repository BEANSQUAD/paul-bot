package main

import (
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/Necroforger/dgrouter/exrouter"
	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
	"github.com/rylio/ytdl"
	"github.com/spf13/viper"

	"google.golang.org/api/googleapi/transport"
	"google.golang.org/api/youtube/v3"
)

var player Player

// Player is a struct grouping together relevant information about the bot's playing capabilities.
// This ensures consistency across play calls.
type Player struct {
	eSession *dca.EncodeSession
	sSession *dca.StreamingSession
	vConn    *discordgo.VoiceConnection
	vQueue   []videoQuery
}

type videoQuery struct {
	videoInfo *ytdl.VideoInfo
	query     string
	requester *discordgo.User
}

func (plr *Player) skipAudio() bool {
	if len(plr.vQueue) > 1 {
		plr.stopAudio()
		return true
	} else {
		plr.disconnect()
		return false
	}
}

func (plr *Player) stopAudio() {
	err := plr.eSession.Stop()
	handleErr(err, "error stopping streaming session")
	plr.eSession.Cleanup()
}

func (plr *Player) stop() bool {
	if plr.sSession != nil {
		plr.vQueue = plr.vQueue[:1]
		if plr.eSession.Running() {
			plr.stopAudio()
			plr.disconnect()
		}
		return true
	} else {
		return false
	}
}

func (plr *Player) disconnect() bool {
	if plr.vConn == nil {
		log.Print("Tried to Disconnect when no VoiceConnections existed")
		return false
	} else {
		err := plr.vConn.Speaking(false)
		handleErr(err, "error setting vConn.Speaking()")

		err = plr.vConn.Disconnect()
		handleErr(err, "error calling vConn.Disconnect()")
		return true
	}
}

func (plr *Player) pause() bool {
	if plr.sSession != nil {
		if plr.sSession.Paused() {
			plr.sSession.SetPaused(false)
		} else {
			plr.sSession.SetPaused(true)
		}
		return true
	} else {
		return false
	}
}

func (plr *Player) startQueue() {
	for len(plr.vQueue) > 0 {
		playSound(*plr.vQueue[0].videoInfo)
		player.vQueue = plr.vQueue[1:]
	}
	player.stopAudio()
	player.disconnect()
}

// handleErr handles an error, checking if the error returned from a function isn't nil.
// If it isn't, logs the error.
func handleErr(err error, output string) {
	if err != nil {
		log.Printf(output+", Error: %v", err)
	}
}

// Stop stops the currently playing media, resets the queue, and disconnects.
// If there is nothing playing, will tell the channel as much.
// Throws an error if it cannot stop the media properly.
func Stop(ctx *exrouter.Context) {
	if player.stop() {
		ctx.Reply("Stopping")
	} else {
		ctx.Reply("No Sound to Stop")
	}
}

// Pause toggles the currently playing media between paused and playing.
// Prints to the channel when it does either, or if there is nothing to pause.
func Pause(ctx *exrouter.Context) {
	if player.pause() {
		ctx.Reply("Toggleing Pause")
	} else {
		ctx.Reply("No Sound to Pause")
	}
}

func getVidString(input string) []string {
	var vids []string
	if strings.HasPrefix(input, "https://www.youtube.com/watch?v=") {
		vids = append(vids, strings.TrimLeft(input, "https://www.youtube.com/watch?v="))
	} else {
		videos, err := ytSearch(input, 1)
		handleErr(err, "Error Searching Using Query")
		if videos != nil {
			for id := range videos {
				vids = append(vids, id)
			}
		}
	}
	return vids
}

// Play searches for the given string on youtube, and adds the first result to the queue.
// Throws a variety of errors, should the bot have issues getting the discord guild info,
// joining the channel, or searching for/retrieving the media requested.
// Prints to the current channel the retrieved media.
func Play(ctx *exrouter.Context) {

	ctx.Reply("test")

	g, err := ctx.Ses.State.Guild(ctx.Msg.GuildID)
	handleErr(err, "Error Getting Guild Information")

	var vSes string
	for _, vs := range g.VoiceStates {
		if vs.UserID == ctx.Msg.Author.ID {
			vSes = vs.ChannelID
			break
		}
	}

	player.vConn, err = ctx.Ses.ChannelVoiceJoin(g.ID, vSes, false, true)
	handleErr(err, "Error Joining Specified Voice Channel")

	vids := getVidString(ctx.Args.After(1))

	if vids != nil {

		videoStruct, err := ytdl.GetVideoInfo(vids[0])
		handleErr(err, "Error Getting Video Info")

		player.vQueue = append(player.vQueue, videoQuery{videoStruct, ctx.Args.After(1), ctx.Msg.Author})

		defer ctx.Reply(fmt.Sprintf("Added " + vids[0] + " to queue"))

		if player.eSession == nil || !player.eSession.Running() {
			defer ctx.Reply(fmt.Sprintf("Playing: https://www.youtube.com/watch?v=%v", vids[0]))
			go player.startQueue()
		}
	} else {
		defer ctx.Reply("YoutubeAPI Quota Exceeded")
		Disconnect(ctx)
	}
}

// Skip skips the currently playing media, moving to the next one.
// Prints to the channel the new media that is being played.
func Skip(ctx *exrouter.Context) {
	if player.skipAudio() {
		ctx.Reply(fmt.Sprintf("Playing: https://www.youtube.com/watch?v=%v", player.vQueue[1].videoInfo.ID))
	} else {
		ctx.Reply("Current Song is Last In Queue, Stopping")
	}
}

// Queue prints the currently queued media to the channel.
func Queue(ctx *exrouter.Context) {
	ctx.Reply("Videos Currently in Queue: ")
	for i := range player.vQueue {
		ctx.Reply(player.vQueue[i].query)
	}
}

// Disconnect disconnects the bot from it's current voice channel, and prints to the channel as such.
// Throws an error and prints to the channel if it tries to disconnect when not in a channel.
// Throws an error if it cannot disconnect from the current voice channel properly.
func Disconnect(ctx *exrouter.Context) {
	if !player.disconnect() {
		ctx.Reply("No VoiceConnections to disconnect")
	} else {
		ctx.Reply("Disconnected")
	}
}

// ytSearch searches youtube for the specified string using the google API.
// Will return a specified amount of results in a map, along with any errors.
// Errors occur should the bot not have an API key, or if it cannot search youtube.
func ytSearch(query string, maxResults int64) (videos map[string]string, err error) {
	if viper.GetString("google-key") == "" {
		err := fmt.Errorf("google-key is not set in config file")
		return nil, err
	}
	client := &http.Client{
		Transport: &transport.APIKey{Key: viper.GetString("google-key")},
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

	if response != nil {
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
	} else {
		return nil, nil
	}

	return videos, nil
}

// playSound configures audio settings to a default, and plays the audio of a specified
// youtube video from the queue.
// Throws errors should the bot be unable to download, encode, or stream the audio.
func playSound(info ytdl.VideoInfo) {
	options := dca.StdEncodeOptions
	options.RawOutput = true
	options.Bitrate = 64
	options.Application = "audio"
	options.Volume = 256
	options.CompressionLevel = 10
	options.PacketLoss = 1
	options.BufferedFrames = 100

	format := info.Formats.Extremes(ytdl.FormatAudioBitrateKey, true)[0]
	downloadURL, err := info.GetDownloadURL(format)
	handleErr(err, "Error Downloading Youtube Video")

	player.eSession, err = dca.EncodeFile(downloadURL.String(), options)
	handleErr(err, "Error Encoding Audio File")
	defer player.eSession.Cleanup()

	player.vConn.Speaking(true)
	defer player.vConn.Speaking(false)

	done := make(chan error)
	player.sSession = dca.NewStream(player.eSession, player.vConn, done)
	err = <-done
	handleErr(err, "Error Streaming Audio File")
}
