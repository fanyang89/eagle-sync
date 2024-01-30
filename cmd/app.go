package cmd

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"

	"github.com/fanyang89/eaglexport/eaglexport"
)

func NewApp() *cli.App {
	return &cli.App{
		Name:  "eagle-export",
		Usage: "Export your eagle library to NAS",
		Commands: []*cli.Command{
			cmdExport,
		},
	}
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
			fs, err = eaglexport.NewSmbFs(address, share, eaglexport.SmbFsOption{
				User:     c.String("smb-user"),
				Password: c.String("smb-password"),
			})
			if err != nil {
				return errors.Wrap(err, "create smbfs failed")
			}
			defer func() {
				err := fs.(*eaglexport.SmbFs).Close()
				if err != nil {
					log.Error().Err(err).Msg("close smbfs failed")
				}
			}()
			dst = root
		} else {
			fs = afero.NewOsFs()
		}

		lib := eaglexport.NewLibrary(c.String("library"), fs)
		return lib.Export(dst, eaglexport.ExportOption{
			Bar:                progressbar.DefaultBytes(-1, "exporting..."),
			Overwrite:          c.Bool("overwrite"),
			Force:              c.Bool("force"),
			GroupBySmartFolder: c.Bool("group-by-smart-folder"),
		})
	},
}
