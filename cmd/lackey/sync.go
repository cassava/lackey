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
	syncDeleteBefore bool
	syncDryRun       bool
	syncOnlyMusic    bool
)

func init() {
	MainCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVarP(&syncDryRun, "dryrun", "n", "just show what will be done, without doing it")
	syncCmd.Flags().BoolVarP(&syncDeleteBefore, "delete-before", "d", "delete extra files in destination")
	syncCmd.Flags().BoolVarP(&syncOnlyMusic, "only-music", "m", "only synchronize music")
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

		var op lackey.Operator
		if syncDryRun {
			op = &lackey.DryRunner{
				Color:     col,
				Strip:     true,
				SrcPrefix: sdb.Path(),
				DstPrefix: ddb.Path(),
			}
		} else {
			panic("not implemented")
		}
		p := lackey.NewPlanner(sdb, ddb, op)
		p.IgnoreData = syncOnlyMusic
		p.DeleteBefore = syncDeleteBefore
		return p.Plan()
	},
}
