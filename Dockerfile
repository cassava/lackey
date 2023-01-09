FROM golang:1.19

RUN apt-get update && apt-get install -y flac lame mp3info ffmpeg imagemagick
RUN go install github.com/sorenisanerd/gotty@latest

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY audio ./audio
COPY cmd ./cmd
COPY filetype ./filetype
COPY vendor ./vendor
COPY *.go ./
RUN go build -v -o /usr/local/bin ./...

COPY entrypoint.sh /

CMD ["gotty", "--max-connection=1", "-w", "/entrypoint.sh"]
