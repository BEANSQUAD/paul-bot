FROM golang:alpine as dep_builder
RUN apk add --no-cache git gcc libc-dev
WORKDIR /go/paul-bot
COPY go.sum .
COPY go.mod .
RUN go mod download

FROM dep_builder as proj_builder
ADD . /go/paul-bot
RUN go build -a -o paul-bot main.go

FROM alpine
COPY --from=proj_builder /go/paul-bot /bin/paul-bot
CMD ["/bin/paul-bot"]
