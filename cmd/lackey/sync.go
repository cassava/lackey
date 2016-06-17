// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"errors"

	"github.com/cassava/lackey"
	"github.com/spf13/cobra"
)

var (
	syncDeleteBefore     bool
	syncDryRun           bool
	syncOnlyMusic        bool
	syncForceTranscode   bool
	syncBitrateThreshold int
	syncTargetQuality    int
)

func init() {
	MainCmd.AddCommand(syncCmd)
	syncCmd.Flags().IntVarP(&syncBitrateThreshold, "threshold", "t", 256, "bitrate threshold")
	syncCmd.Flags().BoolVarP(&syncForceTranscode, "force", "f", false, "force transcode for all audio")
	syncCmd.Flags().IntVarP(&syncTargetQuality, "quality", "q", 4, "target mp3 quality")
	syncCmd.Flags().BoolVarP(&syncDryRun, "dryrun", "n", false, "just show what will be done, without doing it")
	syncCmd.Flags().BoolVarP(&syncDeleteBefore, "delete-before", "d", false, "delete extra files in destination")
	syncCmd.Flags().BoolVarP(&syncOnlyMusic, "only-music", "m", false, "only synchronize music")
}

var syncCmd = &cobra.Command{
	Use:   "sync <destination>",
	Short: "synchronize libraries",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("missing destination library destination argument")
		}
		col.Println("@.Reading source library (this might take a while)...")
		sdb, err := lackey.ReadLibrary(Conf.LibraryPath)
		if err != nil {
			return err
		}

		col.Println("@.Reading destination library (this might take a while)...")
		ddb, err := lackey.ReadLibrary(args[0])
		if err != nil {
			return err
		}

		op := &lackey.Runner{
			Color:            col,
			BitrateThreshold: syncBitrateThreshold,
			TargetQuality:    syncTargetQuality,
			ForceTranscode:   syncForceTranscode,
			DryRun:           syncDryRun,
			Verbose:          Conf.Verbose,
			Strip:            true,
			SrcPrefix:        sdb.Path() + "/",
			DstPrefix:        ddb.Path() + "/",
		}
		p := lackey.NewPlanner(sdb, ddb, op)
		p.IgnoreData = syncOnlyMusic
		p.DeleteBefore = syncDeleteBefore
		return p.Plan()
	},
}
