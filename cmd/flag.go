package cmd

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
	"github.com/urfave/cli/v2"
)

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

var flagGroupBySmartFolder = &cli.BoolFlag{
	Name:    "group-by-smart-folder",
	Aliases: []string{"s"},
	Value:   true,
}

var flagOverwrite = &cli.BoolFlag{
	Name: "overwrite",
}

var flagForce = &cli.BoolFlag{
	Name: "force",
}

var flagSmbUser = &cli.StringFlag{
	Name: "smb-user",
	Action: func(context *cli.Context, s string) error {
		if s == "" {
			return errors.New("empty smb user")
		}
		return nil
	},
}

var flagSmbPassword = &cli.StringFlag{
	Name: "smb-password",
	Action: func(context *cli.Context, s string) error {
		if s == "" {
			return errors.New("empty smb password")
		}
		return nil
	},
}

var flagHistoryFile = &cli.StringFlag{
	Name:  "history-file",
	Value: "",
	Action: func(context *cli.Context, s string) error {
		if s != "" {
			dir := filepath.Dir(s)
			return os.MkdirAll(dir, 0755)
		}
		return nil
	},
}
