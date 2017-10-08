// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"
	"runtime"

	"github.com/cassava/lackey"
	"github.com/spf13/cobra"
)

var (
	syncDeleteBefore   bool
	syncDryRun         bool
	syncOnlyMusic      bool
	syncForceTranscode bool
	syncConcurrent     int

	// MP3:
	syncBitrateThreshold int
	syncTargetQuality    int

	// OPUS:
	syncOPUS          bool
	syncTargetBitrate string
	syncUseOGG        bool
)

func init() {
	MainCmd.AddCommand(syncCmd)
	syncCmd.Flags().IntVarP(&syncConcurrent, "concurrent", "w", runtime.NumCPU(), "number of concurrent workers")
	syncCmd.Flags().BoolVarP(&syncForceTranscode, "force", "f", false, "force transcode for all audio")
	syncCmd.Flags().BoolVarP(&syncDryRun, "dryrun", "n", false, "just show what will be done, without doing it")
	syncCmd.Flags().BoolVarP(&syncDeleteBefore, "delete-before", "d", false, "delete extra files in destination")
	syncCmd.Flags().BoolVarP(&syncOnlyMusic, "only-music", "m", false, "only synchronize music")

	// MP3:
	syncCmd.Flags().IntVarP(&syncBitrateThreshold, "threshold", "t", 256, "bitrate threshold at which we copy instead of transcoding")
	syncCmd.Flags().IntVarP(&syncTargetQuality, "quality", "q", 4, "target MP3 quality (0=highest, largest; 9=lowest, smallest)")

	// OPUS:
	syncCmd.Flags().StringVarP(&syncTargetBitrate, "bitrate", "r", "96k", "target OPUS bitrate, in bps")
	syncCmd.Flags().BoolVarP(&syncOPUS, "opus", "u", false, "output codec is OPUS not MP3")
	syncCmd.Flags().BoolVar(&syncUseOGG, "use-ogg-extension", false, "use the .ogg extension instead of .opus")
}

var syncCmd = &cobra.Command{
	Use:   "sync <destination>",
	Short: "synchronize libraries",
	Long: `Synchronize from a high-quality library to a lower-quality mirror.

  The high-quality library is specified with the -L flag.
  Making it an explicit parameter leads to making it a configuration option
  later, and prevents costly mistakes as to which is source and which is
  destination.

  The relevant default options are as follows:
  
    - it will follow symlinks (--follow-symlinks=true)
    - it will unconditionally convert non-MP3 music with the LAME encoder
      with a quality setting of 4 (--quality=4). See the LAME encoder on
      what the quality setting means. Lower is better.
    - it will convert existing MP3s if they have a bitrate higher than 256kbps,
      and copy them otherwise (--threshold=256)
    - it will copy all data files that are not music
    - it will delete all unexpected files in the destination (like rsync)
    - it will use the number of cores as the number of workers to use
      (e.g. --concurrent=4)
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("missing destination library destination argument")
		}
		col.Println("@.Reading source library (this might take a while)...")
		sdb, err := Conf.ReadLibrary(Conf.LibraryPath)
		if err != nil {
			return err
		}

		col.Println("@.Reading destination library (this might take a while)...")
		ddb, err := Conf.ReadLibrary(args[0])
		if err != nil {
			return err
		}

		var e lackey.Encoder
		if syncOPUS {
			ext := ".opus"
			if syncUseOGG {
				ext = ".ogg"
			}
			e = &lackey.OPUSEncoder{
				Extension:     ext,
				TargetBitrate: syncTargetBitrate,
			}
		} else {
			e = &lackey.MP3Encoder{
				TargetQuality:    syncTargetQuality,
				BitrateThreshold: syncBitrateThreshold,
			}
		}

		op := &lackey.Runner{
			Color:          col,
			Encoder:        e,
			ForceTranscode: syncForceTranscode,
			DryRun:         syncDryRun,
			Verbose:        Conf.Verbose,
			Strip:          true,
			SrcPrefix:      sdb.Path() + "/",
			DstPrefix:      ddb.Path() + "/",
		}
		p := lackey.NewPlanner(sdb, ddb, op)
		p.IgnoreData = syncOnlyMusic
		p.DeleteBefore = syncDeleteBefore
		p.Concurrent = syncConcurrent
		return p.Plan()
	},
}
