// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"

	"github.com/cassava/lackey"
	"github.com/goulash/color"
	"github.com/spf13/cobra"
)

// Conf loads and stores the configuration (apart from command line
// configuration) of this program, including where the repository is.
var Conf struct {
	LibraryPath string
	Verbose     bool
	Color       string

	lackey.LibraryReader

	BitrateThreshold int
	TargetQuality    int
}

// col lets us print in colors.
var col = color.New()

type UsageError struct {
	Cmd   string
	Msg   string
	Usage func() error
}

func (e *UsageError) Error() string {
	return fmt.Sprintf("%s", e.Msg)
}

var MainCmd = &cobra.Command{
	Use:   "lackey",
	Short: "manage your high-quality music library",
	Long: `Lackey primarily helps you keep a lower-quality mirror of your
primary high-quality music library. For example, you may have FLACs of your
music, but want an up-to-date MP3 mirror of this.
`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This function can be overriden if it's not necessary for a command.
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		return nil
	},
}

// main loads the configuration and executes the primary command.
func main() {
	/*
		err := Conf.MergeAll()
		if err != nil {
			// We didn't manage to load any configuration, which means that repoctl
			// is unconfigured. There are some commands that work nonetheless, so
			// we can't stop now -- which is why we don't os.Exit(1).
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		}
	*/

	// Arguments from the command line override the configuration file,
	// so we have to add the flags after loading the configuration.
	//
	// TODO: Maybe in the future we will make it possible to specify the
	// configuration file via the command line; right now it is not a priority.
	col.Set(Conf.Color) // set default, which will be auto if Conf.Color is empty or invalid
	MainCmd.PersistentFlags().Var(col, "color", "when to use color (auto|never|always)")
	MainCmd.PersistentFlags().BoolVarP(&Conf.Verbose, "verbose", "v", Conf.Verbose, "show minimal amount of information")
	MainCmd.PersistentFlags().StringVarP(&Conf.LibraryPath, "library", "L", "", "path to primary library")
	MainCmd.PersistentFlags().BoolVar(&Conf.LibraryReader.IgnoreHidden, "ignore-hidden", true, "ignore hidden files")
	MainCmd.PersistentFlags().BoolVar(&Conf.LibraryReader.FollowSymlinks, "follow-symlinks", true, "do not follow symlinks")

	err := MainCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		if e, ok := err.(*UsageError); ok {
			e.Usage()
		}
		os.Exit(1)
	}
}
