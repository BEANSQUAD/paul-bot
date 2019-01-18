FROM golang:alpine as dep_builder
RUN apk add --no-cache git gcc libc-dev
WORKDIR /go/paul-bot
COPY go.mod go.sum .
RUN go mod download

FROM dep_builder as proj_builder
COPY . /go/paul-bot
RUN go build -a -o paul-bot main.go

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=proj_builder /go/paul-bot/paul-bot /paul-bot
CMD ["/paul-bot"]
