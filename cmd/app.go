package cmd

import (
	"os"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/afero"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"

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
		flagSmbUser, flagSmbPassword, flagHistoryFile,
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

		historyFile := c.String("history-file")
		var history *eaglexport.History
		if historyFile != "" {
			history, err = eaglexport.NewHistory(historyFile)
			if err != nil {
				return err
			}
			history.Load()
		}

		lib := eaglexport.NewLibrary(c.String("library"), fs, history)
		p := mpb.New(mpb.WithRefreshRate(180 * time.Millisecond))
		speedBar := p.New(0,
			mpb.BarStyle(),
			mpb.PrependDecorators(
				decor.Counters(decor.SizeB1024(0), "% .2f / % .2f"),
			),
			mpb.AppendDecorators(
				decor.EwmaETA(decor.ET_STYLE_GO, 30),
				decor.Name(" ] "),
				decor.EwmaSpeed(decor.SizeB1024(0), "% .2f", 60),
			))
		itemBar := p.New(0, mpb.BarStyle().Rbound("|"),
			mpb.PrependDecorators(
				decor.Name("item ", decor.WC{C: decor.DindentRight | decor.DextraSpace}),
				decor.CountersNoUnit("%d / %d ", decor.WCSyncWidth),
				decor.OnComplete(decor.AverageETA(decor.ET_STYLE_GO), "done"),
			),
			mpb.AppendDecorators(decor.Percentage()))

		option := eaglexport.ExportOption{
			SpeedBar:           speedBar,
			ItemBar:            itemBar,
			Overwrite:          c.Bool("overwrite"),
			Force:              c.Bool("force"),
			GroupBySmartFolder: c.Bool("group-by-smart-folder"),
		}

		profileFile, err := os.OpenFile("cpu.profile", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			log.Error().Err(err).Msg("open file failed")
		}
		defer profileFile.Close()

		err = pprof.StartCPUProfile(profileFile)
		if err != nil {
			log.Error().Err(err).Msg("start CPU profile failed")
		}
		defer pprof.StopCPUProfile()

		return lib.Export(dst, option)
	},
}
