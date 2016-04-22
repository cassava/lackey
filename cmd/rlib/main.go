// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"net/http"
	_ "net/http/pprof"

	"github.com/goulash/color"

	"github.com/cassava/lackey"
	"github.com/cassava/lackey/music"
	_ "github.com/cassava/lackey/music/flac"
	_ "github.com/cassava/lackey/music/mp3"
)

var col = color.New()
var dirPrefix = col.Color("@.  | @|")

func PrintEntryWithPrefix(e *lackey.Entry, prefix string) {
	switch e.Type() {
	case lackey.DirEntry:
		col.Printf("%s@!%s/\n", prefix, e.Filename())
		p := prefix + dirPrefix
		for _, v := range e.Children() {
			PrintEntryWithPrefix(v, p)
		}
	case lackey.FileEntry:
		col.Printf("%s@y%s\n", prefix, e.Filename())
	case lackey.MusicEntry:
		if m, ok := e.Data().(music.Metadata); ok {
			n, _ := m.Track()
			col.Printf("%s@g%d-%s@| (%s)\n", prefix, n, m.Title(), m.Encoding())
			break
		}
		fallthrough
	case lackey.ErrorEntry:
		col.Printf("%s@r%s@|: %s\n", prefix, e.Filename(), e.Data())
	default:
		col.Printf("%s@R%s: %s\n", prefix, e.Filename(), e.Data())
	}
}

func PrintDatabase(db *lackey.Database) {
	col.Println("Library path:", db.Path())
	col.Println("Library size:", db.SizeString())
	PrintEntryWithPrefix(db.Root(), "")
}

func main() {
	listen := flag.String("listen", "localhost:6060", "address to listen on")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Println("Error: require single path argument")
		os.Exit(1)
	}

	go func() {
		log.Println(http.ListenAndServe(*listen, nil))
	}()

	path := flag.Arg(0)
	fmt.Println("Reading library...")
	db, err := lackey.ReadLibrary(path)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	PrintDatabase(db)
}
