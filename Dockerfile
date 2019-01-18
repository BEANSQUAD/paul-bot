FROM golang:alpine as builder
RUN apk add --no-cache git gcc
ADD . /go/paul-bot
WORKDIR /go/paul-bot
RUN go mod download
RUN go build -a -o paul-bot main.go

FROM alpine:latest
COPY --from=builder /go/paul-bot /paul-bot
CMD ["/paul-bot"]
