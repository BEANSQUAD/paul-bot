FROM golang:alpine as builder
RUN apk add --no-cache git gcc libc-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go build -v github.com/bwmarrin/discordgo
COPY . /app
RUN go build -v -o paul-bot main.go

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/paul-bot /paul-bot
CMD ["/paul-bot"]
