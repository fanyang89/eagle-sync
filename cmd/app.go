package cmd

import (
	"os"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"

	"github.com/fanyang89/eagle-sync/eaglesync"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:  "eagle-sync",
		Usage: "Export/sync your eagle library to NAS",
		Commands: []*cli.Command{
			cmdExport,
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

var reSmbConnStr = regexp.MustCompile(`(?m)smb://(?P<address>[^/]+)/(?P<share>[^/]+)/(?P<path>.+)`)

func parseSmbConnectionString(s string) (address string, share string, path string, valid bool) {
	match := reSmbConnStr.FindStringSubmatch(s)
	for i, name := range reSmbConnStr.SubexpNames() {
		if name == "address" {
			address = match[i]
		} else if name == "share" {
			share = match[i]
		} else if name == "path" {
			path = match[i]
		}
	}
	valid = !(len(match) == 0 || address == "" || share == "")
	return
}

var cmdExport = &cli.Command{
	Name: "export",
	Flags: []cli.Flag{
		flagLibraryDir, flagDestDir, flagGroupBySmartFolder, flagOverwrite, flagForce,
		flagSmbUser, flagSmbPassword,
	},
	Action: func(c *cli.Context) error {
		var err error

		dst := c.String("dst")
		var fs afero.Fs
		if strings.HasPrefix(dst, "smb://") {
			address, share, root, ok := parseSmbConnectionString(dst)
			if !ok {
				return errors.New("invalid smb connection string")
			}
			fs, err = eaglesync.NewSmbFs(address, share, eaglesync.SmbFsOption{
				User:     c.String("smb-user"),
				Password: c.String("smb-password"),
			})
			if err != nil {
				return errors.Wrap(err, "create smbfs failed")
			}
			defer func() {
				err := fs.(*eaglesync.SmbFs).Close()
				if err != nil {
					log.Error().Err(err).Msg("close smbfs failed")
				}
			}()
			dst = root
		} else {
			fs = afero.NewOsFs()
		}

		lib := eaglesync.NewLibrary(c.String("library"), fs)
		return lib.Export(dst, eaglesync.ExportOption{
			Bar:                progressbar.Default(-1),
			Overwrite:          c.Bool("overwrite"),
			Force:              c.Bool("force"),
			GroupBySmartFolder: c.Bool("group-by-smart-folder"),
		})
	},
}
