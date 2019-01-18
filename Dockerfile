FROM golang:alpine as builder
RUN apk add --no-cache git gcc libc-dev
ADD . /go/paul-bot
WORKDIR /go/paul-bot
RUN go mod download
RUN go build -a -o paul-bot main.go

FROM alpine:latest
COPY --from=builder /go/paul-bot /bin/paul-bot
CMD ["/bin/paul-bot"]
