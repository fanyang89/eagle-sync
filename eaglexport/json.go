package eaglexport

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
)

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
