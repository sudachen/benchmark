package ppftool

import (
	"fmt"
	"time"

	"github.com/google/pprof/driver"
	"github.com/google/pprof/profile"
)

type fetcher struct {
	b []byte
}

func (f *fetcher) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	if src == "" {
		p, err := profile.ParseData(f.b)
		return p, "", err
	}
	return nil, "", fmt.Errorf("unknown source %s", src)
}

func Fetcher(b []byte) driver.Fetcher {
	return &fetcher{b}
}
