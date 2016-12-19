// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cassava/lackey"
	"github.com/cassava/lackey/audio/mp3"
	"github.com/goulash/audio"
	"github.com/goulash/audio/flac"
	"github.com/goulash/stat"
	"github.com/goulash/units"
	"github.com/spf13/cobra"
)

var (
	showStandard bool
	showRuntime  bool
	showTree     bool
)

func init() {
	MainCmd.AddCommand(statsCmd)
	statsCmd.Flags().BoolVarP(&showStandard, "standard", "s", false, "show standard statistics")
	statsCmd.Flags().BoolVarP(&showRuntime, "runtime", "r", false, "show runtime statistics")
	statsCmd.Flags().BoolVarP(&showTree, "tree", "t", false, "show library tree")
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
		db, err := Conf.ReadLibrary(Conf.LibraryPath)
		if err != nil {
			return err
		}

		if showTree {
			printTree(db)
		}
		if showRuntime {
			runtimeStats()
		}
		if showStandard {
			standardStats(db)
		}

		return nil
	},
}

func printTree(db *lackey.Database) {
	e := db.Root()
	col.Println("Tree:")
	printEntry(e, 0)
	col.Println()
}

func printEntry(e *lackey.Entry, level int) {
	prefix := strings.Repeat("  ", level)

	switch e.Type() {
	case lackey.DirEntry:
		col.Printf("%s@!%s@|/\n", prefix, e.Filename())
		for _, c := range e.Children() {
			printEntry(c, level+1)
		}
	case lackey.FileEntry:
		col.Printf("%s%s\n", prefix, e.Filename())
	case lackey.MusicEntry:
		col.Printf("%s@b%s@|\n", prefix, e.Filename())
	default:
		col.Printf("%s@r%s@|\n", prefix, e.Filename())
	}
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
	col.Println()
}

func runtimeStats() {
	stats := func(r *stat.Run) string {
		return fmt.Sprintf("μ=%s, σ=%s, n=%d", time.Duration(r.Mean()), time.Duration(r.Std()), r.N())
	}

	col.Printf("Runtime stats:\n")
	col.Printf("  audio.@!Identify@|        %s\n", stats(&audio.Stats.Identify))
	col.Printf("  audio.@!ReadMetadata@|    %s\n", stats(&audio.Stats.ReadMetadata))
	col.Printf("  flac.@!Identify@|         %s\n", stats(&flac.Stats.Identify))
	col.Printf("  flac.@!ReadFileMetadata@| %s\n", stats(&flac.Stats.ReadFileMetadata))
	col.Printf("  flac.@!ReadMetadata@|     %s\n", stats(&flac.Stats.ReadMetadata))
	col.Printf("  mp3.@!Assert@|            %s\n", stats(&mp3.Stats.Assert))
	col.Printf("  mp3.@!ReadMetadata@|      %s\n", stats(&mp3.Stats.ReadMetadata))
	col.Printf("  mp3.@!ReadMetadataMeta@|  %s\n", stats(&mp3.Stats.ReadMetadataMeta))
	col.Printf("  mp3.@!ReadMetadataBrDu@|  %s\n", stats(&mp3.Stats.ReadMetadataBrDu))
	col.Printf("  mp3.@!ToolMP3INFO@|       %s\n", stats(&mp3.Stats.ToolMP3INFO))
	col.Printf("  mp3.@!ToolEXIFTOOL@|      %s\n", stats(&mp3.Stats.ToolEXIFTOOL))
	col.Println()
}
