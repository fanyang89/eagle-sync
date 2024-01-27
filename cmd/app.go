package cmd

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"

	"github.com/fanyang89/eagle-sync/eaglesync"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:  "eagle-sync",
		Usage: "Export/sync your eagle library to NAS",
		Commands: []*cli.Command{
			cmdSync, cmdExport,
		},
	}
}

var flagLibraryDir = &cli.StringFlag{
	Name:    "library",
	Aliases: []string{"d"},
	Action: func(c *cli.Context, s string) error {
		_, err := os.Stat(s)
		if err != nil {
			if os.IsNotExist(err) {
				return errors.Wrap(err, "library not exists")
			}
			return errors.Wrap(err, "unexpected error")
		}
		return nil
	},
}

var flagDestDir = &cli.StringFlag{
	Name:    "dst",
	Aliases: []string{"o"},
	Action: func(c *cli.Context, s string) error {
		if s == "" {
			return errors.New("dst is empty")
		}

		_, err := os.Stat(s)
		if err != nil {
			if os.IsNotExist(err) {
				return os.MkdirAll(s, 0755)
			}
		}
		return nil
	},
}

var flagBySmartFolder = &cli.BoolFlag{
	Name:    "by-smart-folder",
	Aliases: []string{"s"},
}

var cmdSync = &cli.Command{
	Name:  "sync",
	Flags: []cli.Flag{flagLibraryDir, flagDestDir},
	Action: func(c *cli.Context) error {
		return nil
	},
}

var cmdExport = &cli.Command{
	Name:  "export",
	Flags: []cli.Flag{flagLibraryDir, flagDestDir, flagBySmartFolder},
	Action: func(c *cli.Context) error {
		lib := eaglesync.NewLibrary(c.String("library"))
		return lib.Export(c.String("dst"), progressbar.Default(100))
	},
}
