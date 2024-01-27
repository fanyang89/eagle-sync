package eaglesync

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/errors"
	"github.com/djherbis/times"
	"github.com/schollz/progressbar/v3"
	"github.com/sourcegraph/conc/pool"
)

type Library struct {
	BaseDir string
}

func NewLibrary(baseDir string) *Library {
	return &Library{
		BaseDir: baseDir,
	}
}

func (e *Library) Export(outputDir string, bar *progressbar.ProgressBar) error {
	var mtimeMap Mtime
	err := parseJsonFile(filepath.Join(e.BaseDir, "mtime.json"), &mtimeMap)
	if err != nil {
		return err
	}

	var libraryMetadata LibraryInfo
	err = parseJsonFile(filepath.Join(e.BaseDir, "metadata.json"), &libraryMetadata)
	if err != nil {
		return err
	}

	filter := NewFolderFilter(&libraryMetadata)

	count, ok := mtimeMap["all"]
	if !ok {
		return errors.New("field 'all' not exists")
	}

	if bar != nil {
		bar.ChangeMax64(count)
		defer func() { _ = bar.Finish() }()
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
			fileMetadataPath := filepath.Join(e.BaseDir, "images", fileInfoName+".info", "metadata.json")
			err = parseJsonFile(fileMetadataPath, &fileInfo)
			if err != nil {
				return err
			}

			if fileInfo.IsDeleted {
				return nil
			}

			var category string
			category, err = filter.Evaluate(&fileInfo)
			if err != nil {
				return err
			}

			infoDir := filepath.Join(e.BaseDir, "images", fileInfoName+".info")
			fileName := fileInfo.Name + "." + fileInfo.Ext
			src := filepath.Join(infoDir, fileName)
			var dst string
			if category == "" {
				dst = filepath.Join(outputDir, "uncategorized", fileName)
			} else {
				dst = filepath.Join(outputDir, category, fileName)
			}

			err = copyFile(src, dst, mtime)
			if err != nil {
				return err
			}

			if bar != nil {
				_ = bar.Add(1)
			}
			return nil
		})
	}
	return p.Wait()
}

func parseJsonFile(path string, out interface{}) error {
	fh, err := os.Open(path)
	if err != nil {
		return errors.Wrapf(err, "open file failed, path: %v", path)
	}
	defer func() { _ = fh.Close() }()

	dec := json.NewDecoder(fh)
	err = dec.Decode(out)
	if err != nil {
		return errors.Wrap(err, "decode failed")
	}
	return nil
}

func copyFile(src string, dst string, fileMtime int64) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "open src file failed")
	}
	defer func() { _ = srcFile.Close() }()

	srcStat, err := times.StatFile(srcFile)
	if err != nil {
		return errors.Wrap(err, "stat src file failed")
	}

	_ = os.MkdirAll(filepath.Dir(dst), 0755)
	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrap(err, "open dst file failed")
	}
	defer func() { _ = dstFile.Close() }()

	dstStat, err := dstFile.Stat()
	if err != nil {
		return errors.Wrap(err, "stat dst file failed")
	}

	if srcStat.ModTime() != dstStat.ModTime() || fileMtime != dstStat.ModTime().UnixMilli() {
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return errors.Wrap(err, "copy file failed")
		}
		err = os.Chtimes(dst, srcStat.AccessTime(), srcStat.ModTime())
		if err != nil {
			return errors.Wrapf(err, "chtimes failed, path: %v", dst)
		}
	} else {
		return errors.Wrap(err, "stat dst file failed")
	}

	return nil
}
