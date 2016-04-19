// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"

	"github.com/goulash/color"

	"github.com/cassava/lackey/music"
	_ "github.com/cassava/lackey/music/flac"
	_ "github.com/cassava/lackey/music/mp3"
)

var col = color.New()

func printMetadata(file string) {
	col.Printf("@!%s\n", file)
	m, err := music.ReadMetadata(file)
	if err != nil {
		col.Printf("\t@r%s\n", err)
		return
	}

	col.Printf("\tTitle:        %s\n", m.Title())
	col.Printf("\tAlbum:        %s\n", m.Album())
	col.Printf("\tArtist:       %s\n", m.Artist())
	col.Printf("\tAlbum artist: %s\n", m.AlbumArtist())
	col.Printf("\tComposer:     %s\n", m.Composer())
	col.Printf("\tYear:         %d\n", m.Year())
	col.Printf("\tGenre:        %s\n", m.Genre())
	i, n := m.Track()
	col.Printf("\tTrack:        %d/%d\n", i, n)
	i, n = m.Disc()
	col.Printf("\tDisc:         %d/%d\n", i, n)
	col.Printf("\tLength:       %s\n", m.Length())
	col.Printf("\tComment:      %s\n", m.Comment())
	col.Printf("\tCopyright:    %s\n", m.Copyright())
	col.Printf("\tWebsite:      %s\n", m.Website())
	col.Println()
	col.Printf("\tEncoded by:       %s\n", m.EncodedBy())
	col.Printf("\tEncoder settings: %s\n", m.EncoderSettings())
	col.Printf("\tEncoding:         %s\n", m.Encoding())
	col.Printf("\tEncoding bitrate: %d Kbps\n", m.EncodingBitrate())
	col.Println()
	col.Printf("\tOriginal filename: %s\n", m.OriginalFilename())
	col.Printf("\tPrivate data:      %q\n", string(m.PrivateData()))
}

func main() {
	flag.Parse()

	for _, f := range flag.Args() {
		printMetadata(f)
	}
}
