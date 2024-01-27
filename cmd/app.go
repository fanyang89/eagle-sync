package cmd

import (
	"github.com/urfave/cli/v2"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:  "eagle-sync",
		Usage: "Export/sync your eagle library to NAS",
		Commands: []*cli.Command{
			cmdExport, cmdSync,
		},
	}
}

var flagLibraryDir = &cli.StringFlag{
	Name:    "library",
	Aliases: []string{"d"},
	Action: func(context *cli.Context, s string) error {
		return nil
	},
}

var flagDestDir = &cli.StringFlag{
	Name:    "dst",
	Aliases: []string{"o"},
	Action: func(context *cli.Context, s string) error {
		return nil
	},
}

var cmdExport = &cli.Command{
	Name:  "sync",
	Flags: []cli.Flag{flagLibraryDir},
	Action: func(context *cli.Context) error {
		return nil
	},
}

var cmdSync = &cli.Command{
	Name:  "export",
	Flags: []cli.Flag{flagLibraryDir},
	Action: func(context *cli.Context) error {
		return nil
	},
}
