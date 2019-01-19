ARG PROJPATH=/go/src/github.com/BEANSQUAD/paul-bot

FROM golang:alpine as dep_builder
RUN apk add --no-cache git gcc libc-dev
WORKDIR $PROJPATH
COPY go.mod go.sum ./
RUN go mod download

FROM dep_builder as proj_builder
COPY . $PROJPATH
RUN go build -v -a -o paul-bot main.go

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=proj_builder $PROJPATH/paul-bot /paul-bot
CMD ["/paul-bot"]
