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
	col.Set(Conf.Color) // set default, which will be auto if Conf.Color is empty or invalid
	MainCmd.PersistentFlags().Var(col, "color", "when to use color (auto|never|always)")
	MainCmd.PersistentFlags().BoolVarP(&Conf.Verbose, "verbose", "v", Conf.Verbose, "show more information")
	MainCmd.PersistentFlags().StringVarP(&Conf.LibraryPath, "library", "L", "", "path to primary library")
	MainCmd.PersistentFlags().BoolVar(&Conf.LibraryReader.IgnoreHidden, "ignore-hidden", true, "ignore hidden files")
	MainCmd.PersistentFlags().BoolVar(&Conf.LibraryReader.FollowSymlinks, "follow-symlinks", true, "follow symlinks")

	err := MainCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		if e, ok := err.(*UsageError); ok {
			e.Usage()
		}
		os.Exit(1)
	}
}
