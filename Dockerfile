FROM golang:alpine
RUN apk add --no-cache git gcc libc-dev ca-certificates
WORKDIR /app
COPY go.mod go.sum /app
RUN go mod download
COPY . /app
RUN go build -v -a -o paul-bot main.go
CMD ["/app/paul-bot"]
