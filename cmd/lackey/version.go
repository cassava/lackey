// Copyright (c) 2016, Ben Morgan. All rights reserved.
// Use of this source code is governed by an MIT license
// that can be found in the LICENSE file.

package main

import (
	"os"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	MainCmd.AddCommand(versionCmd)
}

type programInfo struct {
	Name      string
	Author    string
	Email     string
	Version   string
	Date      string
	Homepage  string
	Copyright string
	License   string
}

const versionTmpl = `{{.Name}} version {{.Version}} ({{.Date}})
Copyright {{.Copyright}}, {{.Author}} <{{.Email}}>

You may find {{.Name}} on the Internet at
    {{.Homepage}}
Please report any bugs you may encounter.

The source code of {{.Name}} is licensed under the {{.License}} license.
`

var progInfo = programInfo{
	Name:      "lackey",
	Author:    "Ben Morgan",
	Email:     "neembi@gmail.com",
	Version:   "0.4",
	Date:      "",
	Copyright: "2016–2017",
	Homepage:  "https://github.com/cassava/lackey",
	License:   "MIT",
}

var versionCmd = &cobra.Command{
	Use:               "version",
	Short:             "show version and date information",
	Long:              "Show the official version number of lackey, as well as the release date.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error { return nil },
	Run: func(cmd *cobra.Command, args []string) {
		if progInfo.Date == "" {
			progInfo.Date = time.Now().Format("2 January 2006")
		}
		template.Must(template.New("version").Parse(versionTmpl)).Execute(os.Stdout, progInfo)
	},
}
