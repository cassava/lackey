FROM golang:1.19

RUN apt-get update && apt-get install -y flac lame mp3info ffmpeg imagemagick

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin ./...

CMD ["lackey"]
