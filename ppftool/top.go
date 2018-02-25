package ppftool

import (
	"io/ioutil"
	"os"
	"encoding/base64"

	"github.com/google/pprof/driver"
)

func Top(b []byte, o *Options) (*Report, error) {

	tempfile := TempFileName() // if driver.Options.Writer is skipped by pprof
	                           // output will be recorded to this file

	rpt := &Report{Unit: o.unit()}

	err := driver.PProf(&driver.Options{
		Fetch:   &fetcher{b},
		Flagset: o.flagset("-top", "-output="+tempfile),
		UI:      &ui{rpt},
		Writer:  &writer{rpt},
		Obj:	 &objtool{},
	})

	if err != nil {
		return nil, err
	}

	if rpt.Rows == nil { // old pprof does not use writer in report generator
		if _, err := os.Stat(tempfile); err == nil {
			defer os.Remove(tempfile)
			if b, err := ioutil.ReadFile(tempfile); err != nil {
				return nil, err
			} else {
				rpt.Write(b)
			}
		}
	}

	if o.Graph != NoImage {
		if img, err := Image(b, o, &ui{rpt}); err != nil {
			return nil, err
		} else {
			rpt.Image = base64.StdEncoding.EncodeToString(img)
		}
	}

	return rpt, nil
}
