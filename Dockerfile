FROM golang:alpine as builder
RUN apk add --no-cache git gcc libc-dev ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go build -v github.com/bwmarrin/discordgo \
    github.com/Necroforger/dgrouter \
    github.com/Necroforger/dgrouter/exrouter \
    github.com/jonas747/dca \
    github.com/rylio/ytdl \
    github.com/spf13/viper \
    google.golang.org/api/youtube/v3 \
    google.golang.org/api/googleapi/transport
COPY . /app
RUN go build -v -o paul-bot .

FROM alpine
RUN apk add --no-cache ca-certificates youtube-dl ffmpeg
WORKDIR /app
COPY --from=builder /app/paul-bot /app/paul-bot
CMD ["/app/paul-bot"]
