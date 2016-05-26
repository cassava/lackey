// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"
	"time"

	"github.com/cassava/lackey"
	"github.com/goulash/audio"
	"github.com/goulash/units"
	"github.com/spf13/cobra"
)

var (
	showStandard bool
)

func init() {
	MainCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVarP(&showStandard, "standard", "s", false, "show standard statistics")
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "show library statistics",
	Long: `Read the library and show statistics:

  First of all, simple statistics are shown, such as how large
  the library is, how many artists, albums, songs, and so on.

  Other things of interest can be seen by passing the options.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := lackey.ReadLibrary(Conf.LibraryPath)
		if err != nil {
			return err
		}

		if showStandard {
			standardStats(db)
		}

		return nil
	},
}

func standardStats(db *lackey.Database) {
	var (
		totalBytes   int64
		musicBytes   int64
		length       time.Duration
		artists      = make(map[string]bool)
		albums       = make(map[string]bool)
		albumartists = make(map[string]bool)
		genres       = make(map[string]bool)
		composers    = make(map[string]bool)
		songs        int
	)

	addto := func(m map[string]bool, s string) {
		for _, f := range strings.Split(s, "/") {
			m[f] = true
		}
	}

	totalBytes = db.Size()
	db.Walk(func(e *lackey.Entry) error {
		if !e.IsMusic() {
			return nil
		}

		musicBytes += e.Size()
		md, ok := e.Data().(audio.Metadata)
		if !ok {
			col.Fprintf(os.Stderr, "@yWarning:@| cannot read audio metadata from %s\n  Data: %s\n", e.AbsPath(), e.Data())
			return nil
		}

		songs++
		length += md.Length()
		addto(artists, md.Artist())
		addto(albums, md.Album())
		addto(albumartists, md.AlbumArtist())
		addto(genres, md.Genre())
		addto(composers, md.Composer())
		return nil
	})

	col.Printf("Standard stats:\n")
	col.Printf("  @!Total size@|    %s\n", units.Bytes10(totalBytes))
	col.Printf("  @!Music size@|    %s\n", units.Bytes10(musicBytes))
	col.Printf("  @!Play time@|     %s\n", length)
	col.Printf("  @!Songs@|         %d\n", songs)
	col.Printf("  @!Artists@|       %d\n", len(artists))
	col.Printf("  @!Albums@|        %d\n", len(albums))
	col.Printf("  @!Album artists@| %d\n", len(albumartists))
	col.Printf("  @!Genres@|        %d\n", len(genres))
	col.Printf("  @!Composers@|     %d\n", len(composers))
}
