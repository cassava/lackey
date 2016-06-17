// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"runtime"

	"github.com/goulash/color"
	"github.com/goulash/errs"
	"github.com/goulash/osutil"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
)

// col lets us print in colors.
var col = color.New()

type Configuration struct {
	SourceDirectory  string
	SourceMaxBitrate int
	TargetDirectory  string
	TargetQuality    int

	DryRun      bool
	FailOnError bool
	NoUpdate    bool
	LinkFiles   bool
	OnlyMusic   bool

	Parallel int
}

var MainCmd = &cobra.Command{
	Use:   "lackey",
	Short: "Convert your high-quality music collection",
	Long: `Lackey converts your high-quality music collection to a lower-quality
one that can be used on-the-go, for example on your MP3-player.

It does modify your high-quality music collection in any way to this end.
`,
}

var (
	SourceDirectory  = ""
	SourceMaxBitrate = 256

	TargetDirectory = ""
	TargetQuality   = 2

	DryRun      = false
	FailOnError = false
	NoUpdate    = false
	LinkFiles   = false
	OnlyMusic   = false

	Parallel = runtime.NumCPU()
)

func init() {
	flag.StringVar(&SourceDirectory, "source", SourceDirectory, "directory containing library of high-quality music")
	flag.IntVarP(&SourceMaxBitrate, "t", "threshold", SourceMaxBitrate, "after what bitrate music is converted")
	flag.StringVar(&TargetDirectory, "target", TargetDirectory, "directory where converted files will be copied to")
	flag.IntVar(&TargetQuality, "quality", TargetQuality, "target mp3 quality, from 0 to 9")
	flag.BoolVarP(&DryRun, "n", "dry-run", DryRun, "do not modify filesystem, just print what would happen")
	flag.BoolVarP(&FailOnError, "f", "fail", FailOnError, "do not continue when an error occurs")
	flag.BoolVar(&NoUpdate, "no-update", NoUpdate, "always copy/link/convert all files over")
	flag.BoolVarP(&LinkFiles, "l", "link", LinkFiles, "make hard links where possible")
	flag.BoolVarP(&OnlyMusic, "m", "only-music", OnlyMusic, "only copy music")
	flag.IntVarP(&Parallel, "p", "parallel", Parallel, "run this many jobs at the same time")
}

func main() {
	flag.Parse()

	// We need to be able to run at least one job at a time.
	if Parallel <= 0 {
		log.Fatal("Fatal: parallel value must be at least 1")
	}

	// Source directory should exist.
	ex, err := osutil.DirExists(SourceDirectory)
	if err != nil {
		log.Fatal("Fatal:", err)
	}
	if !ex {
		log.Fatal("Fatal: source directory does not exist:", SourceDirectory)
	}

	if SourceMaxBitrate < 32 || 500 < SourceMaxBitrate {
		log.Fatal("Fatal: source bitrate threshold should be between 32 and 500")
	}

	// Target quality should be between 0 and 9
	if TargetQuality < 0 || 9 < TargetQuality {
		log.Fatal("Fatal: target quality should be between 0 (highest) and 9 (lowest)")
	}

	// Target directory should exist, create it otherwise.
	ex, err = osutil.DirExists(TargetDirectory)
	if err != nil {
		log.Fatal("Fatal:", err)
	}
	if !ex && !DryRun {
		os.MkdirAll(TargetDirectory, 0777)
	}

	w := Walker{
		SourceMaxBitrate: SourceMaxBitrate,
		TargetCodec:      MP3,
		TargetQuality:    TargetQuality,

		DryRun:    DryRun,
		NoUpdate:  NoUpdate,
		LinkFiles: LinkFiles,
		OnlyMusic: OnlyMusic,
		Parallel:  Parallel,
	}

	eh := func(err error) error {
		log.Println("Error:", err)
		return nil
	}
	if FailOnError {
		eh = errs.Quit
	}

	err = w.Walk(SourceDirectory, TargetDirectory, eh)
	if err != nil {
		log.Fatal(err)
	}
}
