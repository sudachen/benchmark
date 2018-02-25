package ppftool

import (
	"bytes"
	"fmt"
	"github.com/google/pprof/driver"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Graph byte

const (
	NoImage Graph = iota
	PNG
	SVG
	DOT
)

func Image(b []byte, o *Options, u driver.UI) ([]byte, error) {

	bf := &bytes.Buffer{}

	if o.Gcount > 0 {
		o0 := *o
		o0.Count = o.Gcount
		o = &o0
	}

	cmd := "-png"
	if o.NoLegend || o.Graph == DOT {
		cmd = "-dot"
	} else if o.Graph == SVG {
		cmd = "-svg"
	}

	tempfile := TempFileName()
	if u == nil {
		u = &ui{}
	}
	err := driver.PProf(&driver.Options{
		Fetch:   &fetcher{b},
		Flagset: o.flagset(cmd, "-output="+tempfile),
		UI:      u,
		Writer:  &writer{bf},
	})

	if err != nil {
		return nil, err
	}

	if bf.Len() == 0 { // old pprof does not use writer in report generator
		if _, err := os.Stat(tempfile); err == nil {
			defer os.Remove(tempfile)
			if b, err := ioutil.ReadFile(tempfile); err != nil {
				return nil, err
			} else {
				bf.Write(b)
			}
		}
	}

	if o.NoLegend {
		a := strings.Split(string(bf.Bytes()), "\n")
		bf.Reset()

		for _, s := range a {
			if strings.Index(s, "subgraph cluster_L") != 0 {
				bf.WriteString(s)
				bf.WriteByte('\n')
			}
		}

		if o.Graph != DOT {

			obf := &bytes.Buffer{}
			format := "png"
			if o.Graph == SVG {
				cmd = "-svg"
			}

			cmd := exec.Command("dot", "-T"+format)
			cmd.Stdin, cmd.Stdout, cmd.Stderr = bf, obf, os.Stderr
			if err := cmd.Run(); err != nil {
				return nil, fmt.Errorf("Failed to execute dot. Is Graphviz installed? Error: %v", err)
			}

			bf = obf
		}
	}

	return bf.Bytes(), nil
}
