package ppftool

import (
	"fmt"
	"strings"

	"github.com/google/pprof/driver"
)

type Options struct {
	Unit // Second, Millisecond, Microsecond

	Count    int      // -nodecount=
	CumSort  bool     // -cum=
	TagFocus []string // -tagfocus=
	Focus    []string // -focus=
	Show     []string // -show=
	Hide     []string // -hide=

	Callgraph      // type of callgrapth image NoImage, PNG, SVG
	Callcount int  // -nodecount for callgraph, if <=0 is used Count
	NoLegend  bool // remove legend from callgraph image
}

func (o *Options) flagset(c ...string) driver.FlagSet {
	show := make([]string, 0, 3)
	hide := make([]string, 0, 3)

	c = append(c, "-unit="+o.Unit.String())

	if o.Count > 0 {
		c = append(c, fmt.Sprintf("-nodecount=%d", o.Count))
	}

	if o.CumSort {
		c = append(c, "-cum=true")
	} else {
		c = append(c, "-flat=true")
	}

	if len(o.Hide) != 0 {
		hide = append(hide, o.Hide...)
	}
	c = append(c, "-hide="+strings.Join(hide, "|"))

	if len(o.Show) != 0 {
		hide = append(hide, o.Show...)
	}
	c = append(c, "-show="+strings.Join(show, "|"))

	if len(o.TagFocus) != 0 {
		s := "-tagfocus=" + o.TagFocus[0]
		for _, t := range o.TagFocus[1:] {
			s = s + "|" + t
		}
		c = append(c, s)
	} else {
		c = append(c, "-tagfocus=")
	}

	if len(o.Focus) != 0 {
		s := "-focus=" + o.Focus[0]
		for _, t := range o.Focus[1:] {
			s = s + "|" + t
		}
		c = append(c, s)
	} else {
		c = append(c, "-focus=")
	}

	return Flagset(c...)
}
