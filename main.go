// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"
	"runtime"

	"github.com/goulash/errs"
	"github.com/goulash/osutil"
)

var (
	SourceDirectory  = ""
	SourceMaxBitrate = 256

	TargetDirectory = ""
	TargetQuality   = 2

	FailOnError = false
	LinkFiles   = false
	OnlyMusic   = false

	Parallel = runtime.NumCPU()
)

func init() {
	flag.StringVar(&SourceDirectory, "source", SourceDirectory, "directory containing library of high-quality music")
	flag.IntVar(&SourceMaxBitrate, "threshold", SourceMaxBitrate, "after what bitrate music is converted")
	flag.StringVar(&TargetDirectory, "target", TargetDirectory, "directory where converted files will be copied to")
	flag.IntVar(&TargetQuality, "quality", TargetQuality, "target mp3 quality, from 0 to 9")
	flag.BoolVar(&FailOnError, "fail", FailOnError, "do not continue when an error occurs")
	flag.BoolVar(&LinkFiles, "link", LinkFiles, "make hard links where possible")
	flag.BoolVar(&OnlyMusic, "only-music", OnlyMusic, "only copy music")
	flag.IntVar(&Parallel, "parallel", Parallel, "run this many jobs at the same time")
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
	if !ex {
		os.MkdirAll(TargetDirectory, 0777)
	}

	w := Walker{
		SourceMaxBitrate: SourceMaxBitrate,
		TargetCodec:      MP3,
		TargetQuality:    TargetQuality,

		OnlyMusic: OnlyMusic,
		LinkFiles: LinkFiles,
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
