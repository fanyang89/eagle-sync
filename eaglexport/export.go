package eaglexport

import (
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/errors"
	"github.com/djherbis/times"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc/pool"
	"github.com/spf13/afero"
	"github.com/vbauerster/mpb/v8"
)

type Library struct {
	baseDir string
	history *History
	fs      afero.Fs
}

func NewLibrary(baseDir string, fs afero.Fs, history *History) *Library {
	return &Library{
		baseDir: baseDir,
		history: history,
		fs:      fs,
	}
}

type ExportOption struct {
	// Overwrite the existing file
	Overwrite bool

	// Force clean up the destination directory before export
	Force bool

	// GroupBySmartFolder export group by smart folder
	GroupBySmartFolder bool

	SpeedBar *mpb.Bar
	ItemBar  *mpb.Bar
}

func (e *Library) Close() {
	e.history.Close()
}

func (e *Library) Export(outputDir string, option ExportOption) error {
	if option.Force {
		err := e.fs.RemoveAll(outputDir)
		if err != nil {
			return errors.Wrapf(err, "delete directory '%v' failed", outputDir)
		}
	}

	var mtimeMap Mtime
	err := parseJsonFile(filepath.Join(e.baseDir, "mtime.json"), &mtimeMap)
	if err != nil {
		return err
	}

	var libraryMetadata LibraryInfo
	err = parseJsonFile(filepath.Join(e.baseDir, "metadata.json"), &libraryMetadata)
	if err != nil {
		return err
	}

	filter := NewFolderFilter(&libraryMetadata)

	count, ok := mtimeMap["all"]
	if !ok {
		return errors.New("field 'all' not exists")
	}

	if option.ItemBar != nil {
		option.ItemBar.SetTotal(count, false)
		defer func() { option.ItemBar.SetCurrent(count) }()
	}

	p := pool.New().WithErrors().WithMaxGoroutines(runtime.NumCPU())
	for fileInfoName, mtime := range mtimeMap {
		fileInfoName := fileInfoName
		mtime := mtime

		if fileInfoName == "all" {
			continue
		}

		p.Go(func() error {
			var fileInfo FileInfo
			fileMetadataPath := filepath.Join(e.baseDir, "images", fileInfoName+".info", "metadata.json")
			err = parseJsonFile(fileMetadataPath, &fileInfo)
			if err != nil {
				return err
			}

			if fileInfo.IsDeleted {
				return nil
			}

			infoDir := filepath.Join(e.baseDir, "images", fileInfoName+".info")
			fileName := fileInfo.Name + "." + fileInfo.Ext
			src := filepath.Join(infoDir, fileName)

			var dst string
			if option.GroupBySmartFolder {
				var category string
				category, err = filter.Evaluate(&fileInfo)
				if err != nil {
					return err
				}

				if category == "" {
					dst = filepath.Join(outputDir, "uncategorized", fileName)
				} else {
					dst = filepath.Join(outputDir, category, fileName)
				}
			} else {
				dst = filepath.Join(outputDir, fileName)
			}

			return e.copyFile(src, dst, mtime, &option)
		})
	}
	return p.Wait()
}

func (e *Library) copyFile(src string, dst string, fileMtime int64, option *ExportOption) error {
	// TODO: src file is always in the OS fs or not?
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open src file failed")
	}
	defer func() { _ = srcFile.Close() }()

	srcStat, err := times.StatFile(srcFile)
	if err != nil {
		return errors.Wrap(err, "stat src file failed")
	}

	srcStat2, err := srcFile.Stat()
	if err != nil {
		return errors.Wrap(err, "stat2 src file failed")
	}

	_ = e.fs.MkdirAll(filepath.Dir(dst), 0755)
	dstFile, err := e.fs.OpenFile(dst, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0655)
	if err != nil {
		return errors.Wrap(err, "open dst file failed")
	}
	defer func() { _ = dstFile.Close() }()

	dstStat, err := dstFile.Stat()
	if err != nil {
		return errors.Wrap(err, "stat dst file failed")
	}

	copied := false
	if e.history != nil {
		m, ok := e.history.Get(src)
		if ok {
			copied = m == srcStat.ModTime()
		}
	}

	if !copied {
		if srcStat.ModTime() != dstStat.ModTime() || fileMtime != dstStat.ModTime().UnixMilli() || option.Overwrite {
			var n int64
			n, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return errors.Wrap(err, "copy file failed")
			}
			if option.SpeedBar != nil {
				option.SpeedBar.IncrInt64(n)
			}

			err = e.fs.Chtimes(dst, srcStat.AccessTime(), srcStat.ModTime())
			if err != nil {
				return errors.Wrapf(err, "chtimes failed, path: %v", dst)
			}

			if e.history != nil {
				err = e.history.Append(HistoryEntry{
					Path:  src,
					MTime: srcStat.ModTime(),
				})
				if err != nil {
					log.Warn().Err(err).Msg("append history failed")
				}
			}
		} else {
			return errors.Wrap(err, "stat dst file failed")
		}
	} else {
		//log.Info().Str("file", src).Msg("skip copy")
		if option.SpeedBar != nil {
			option.SpeedBar.IncrInt64(srcStat2.Size())
		}
	}

	if option.ItemBar != nil {
		option.ItemBar.IncrBy(1)
	}

	return nil
}
