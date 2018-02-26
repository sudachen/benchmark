package ppftool

import (
	"fmt"
	"strings"

	"github.com/google/pprof/driver"
)

type Options struct {
	Unit // Second, Millisecond, Microsecond
	Index

	Count     int      // -nodecount=
	CumSort   bool     // -cum=
	TagShow   []string // -tagshow=
	TagHide   []string // -taghide=
	TagIgnore []string // -tagignore=
	TagFocus  []string // -tagfocus=
	Ignore    []string // -ignore=
	Focus     []string // -focus=
	Show      []string // -show=
	Hide      []string // -hide=

	Graph         // type of callgrapth image NoImage, PNG, SVG, DOT
	Gcount   int  // -nodecount for callgraph, if <=0 is used Count
	NoLegend bool // remove legend from callgraph image
}

type Index byte

const (
	CpuProfIndex Index = iota
	AllocObjectsIndex
	AllocSpaceIndex
	InuseObjectsIndex
	InuseSpaceIndex
)

func (o *Options) unit() Unit {
	if o.Unit == DefaultUnit {
		switch o.Index {
		case AllocObjectsIndex, InuseObjectsIndex:
			return None
		case AllocSpaceIndex, InuseSpaceIndex:
			return Megabyte
		}
	}
	return o.Unit
}

func (o *Options) flagset(c ...string) driver.FlagSet {
	if o.Count > 0 {
		c = append(c, fmt.Sprintf("-nodecount=%d", o.Count))
	}

	if o.CumSort {
		c = append(c, "-cum=true")
	} else {
		c = append(c, "-flat=true")
	}

	if len(o.Hide) > 0 {
		c = append(c, "-hide="+strings.Join(o.Hide, "|"))
	}
	if len(o.Show) > 0 {
		c = append(c, "-show="+strings.Join(o.Show, "|"))
	}
	if len(o.Ignore) > 0 {
		c = append(c, "-ignore="+strings.Join(o.Ignore, "|"))
	}
	if len(o.Focus) > 0 {
		c = append(c, "-focus="+strings.Join(o.Focus, "|"))
	}
	if len(o.TagHide) > 0 {
		c = append(c, "-taghide="+strings.Join(o.TagHide, "|"))
	}
	if len(o.TagShow) > 0 {
		c = append(c, "-tagshow="+strings.Join(o.TagShow, "|"))
	}
	if len(o.TagIgnore) > 0 {
		c = append(c, "-tagignore="+strings.Join(o.TagIgnore, "|"))
	}
	if len(o.TagFocus) != 0 {
		c = append(c, "-tagfocus="+strings.Join(o.TagFocus, "|"))
	}

	c = append(c, "-unit="+o.unit().String())

	switch o.Index {
	case AllocObjectsIndex:
		c = append(c, "-sample_index=alloc_objects")
	case AllocSpaceIndex:
		c = append(c, "-sample_index=alloc_space")
	case InuseObjectsIndex:
		c = append(c, "-sample_index=inuse_objects")
	case InuseSpaceIndex:
		c = append(c, "-sample_index=inuse_space")
	default:
	}

	return Flagset(c...)
}
