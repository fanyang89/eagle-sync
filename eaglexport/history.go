package eaglexport

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/goccy/go-json"
	"github.com/rs/zerolog/log"
)

type HistoryEntry struct {
	Path  string    `json:"path"`
	MTime time.Time `json:"mtime"`
}

type History struct {
	path    string
	writer  io.WriteCloser
	reader  io.Reader
	encoder *json.Encoder
	data    map[string]time.Time
	m       sync.RWMutex
}

func NewHistory(path string) (*History, error) {
	fh, err := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "open history file failed")
	}

	return &History{
		path:    path,
		writer:  fh,
		reader:  fh,
		encoder: json.NewEncoder(fh),
		data:    make(map[string]time.Time),
	}, nil
}

func (r *History) Load() {
	r.m.Lock()
	defer r.m.Unlock()

	decoder := json.NewDecoder(r.reader)
	for {
		var h HistoryEntry
		err := decoder.Decode(&h)
		if err != nil {
			if err != io.EOF {
				log.Warn().Err(err).Msg("Load history failed")
			}
			break
		}
		r.data[h.Path] = h.MTime
	}
	log.Info().Int("history", len(r.data)).Msg("History loaded")
}

func (r *History) Close() {
	r.m.Lock()
	defer r.m.Unlock()

	err := r.writer.Close()
	if err != nil {
		log.Error().Err(err).Msg("Close file failed")
	}
}

func (r *History) Append(h HistoryEntry) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.data[h.Path] = h.MTime
	return r.encoder.Encode(h)
}

func (r *History) Get(path string) (time.Time, bool) {
	r.m.RLock()
	defer r.m.RUnlock()
	t, ok := r.data[path]
	return t, ok
}
