# syntax=docker/dockerfile:1

# Stage 1: Build
FROM golang:1.19 AS build

# Build latest gotty for providing lackey as a web service together
# with the entrypoint script.
RUN go install github.com/sorenisanerd/gotty@latest

# Build lackey from source.
WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY vendor ./vendor
COPY audio ./audio
COPY filetype ./filetype
COPY cmd ./cmd
COPY *.go ./
RUN go build -v -o /go/bin ./...

# Stage 2: Deploy
FROM ubuntu AS deploy
LABEL org.opencontainers.image.source https://github.com/cassava/lackey

# Install system dependencies of lackey.
RUN apt-get update && \
    apt-get install --no-install-recommends -y flac lame mp3info ffmpeg imagemagick

COPY --from=build /go/bin/gotty /go/bin/lackey /usr/local/bin/
COPY entrypoint.sh /

CMD ["gotty", "--max-connection=1", "-w", "/entrypoint.sh"]
