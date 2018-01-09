package benchmark

import(
	"github.com/google/pprof/profile"
	"github.com/google/pprof/driver"
	"io"
	"flag"
	"fmt"
	"time"
	"strings"
	"errors"
)

type Options byte
const (
	SortByCum Options = 1 << iota
	ExcludeRuntime
	RuntimeOnly
	Tagged
	DefaultReport Options = 0
)

func (o Options) String() string {
	s := make([]string,0,4)
	if (o & SortByCum) != 0 {
		s = append(s,"SortByCum")
	}
	if (o & ExcludeRuntime) != 0 {
		s = append(s,"ExcludeRuntime")
	}
	if (o & RuntimeOnly) != 0 {
		s = append(s,"RuntimeOnly")
	}
	if (o & Tagged) != 0 {
		s = append(s,"Tagged")
	}
	if len(s) > 0 {
		return strings.Join(s,"|")
	}
	return "DefaultReport"
}

func (o *Options) fromString(s string) error {
	tf := func(s string) (r []string) {
		r = make([]string, 0, len(s))
		for _, v := range strings.Split(s,"|") {
			if s := strings.TrimSpace(v); len(s) > 0 {
				r = append(r,s)
			}
		}
		return
	}

	*o = DefaultReport
	for _,x := range tf(s) {
		switch x {
		case "SortByCum": *o |= SortByCum
		case "ExcludeRuntime": *o |= ExcludeRuntime
		case "RuntimeOnly": *o |= RuntimeOnly
		case "Tagged": *o |= Tagged
		}
	}

	return nil
}

type Unit byte
const (
	Usec Unit = iota
	Msec
	Sec
)

func (u Unit) String() string {
	switch u {
	case Usec: return "Usec"
	case Msec: return "Msec"
	case Sec: return "Sec"
	}
	panic("invalid Unit value")
}

func (u *Unit) fromString(s string) error {
	switch s {
	case "Usec": *u = Usec
	case "Msec": *u = Msec
	case "Sec": *u = Sec
	default: return errors.New("unknown Unit "+s)
	}
	return nil
}

type Row struct {
	Flat, FlatPercent, SumPercent, Cum, CumPercent float64
	Function string
}

type Rows []*Row

type Report struct {
	Unit
	Options
	Rows
	Label string
	Errors []string
}

func Top(b []byte, count int, options Options, unit Unit, label string) *Report {
	rpt := &Report{Label: label, Unit: unit, Options: options, Rows: make(Rows,0,count)}
	c := append(tuneBy(options,unit),"output=@",fmt.Sprintf("top%d",count))
	f := &flagset{ flag.NewFlagSet("ppf",flag.ContinueOnError ), []string{defaultProfile} }
	o := &driver.Options{
		Flagset: f,
		Fetch: &fetcher{b},
		Writer: &writer{rpt},
		UI: &ui{rpt,c,0},
	}
	driver.PProf(o)
	return rpt
}

func tuneBy(o Options,u Unit) []string{
	var c []string

	if (o & SortByCum) != 0 {
		c = append(c,"cum=true")
	} else {
		c = append(c,"flat=true")
	}

	if (o & ExcludeRuntime) != 0 {
		c = append(c,"hide=^runtime\\..*$")
		c = append(c,"show=")
	} else if (o & RuntimeOnly) != 0 {
		c = append(c,"show=^runtime\\..*$")
		c = append(c,"hide=")

	}

	if (o & Tagged) != 0 {
		c = append(c,"tagfocus=t:active")
	} else {
		c = append(c,"tagfocus=")
	}

	switch u {
	case Usec: c = append(c,"unit=us")
	case Msec: c = append(c,"unit=ms")
	case Sec: c = append(c,"unit=s")
	}

	return c
}

const defaultProfile = "@="

type fetcher struct {
	b []byte
}

func (f *fetcher) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	if src == defaultProfile {
		p, err := profile.ParseData(f.b)
		return p, "", err
	}
	return nil, "", fmt.Errorf("unknown source %s",src)
}

type ui struct {
	*Report
	command []string
	index int
}

func (u *ui) ReadLine(prompt string) (string, error) {
	if u.index < len(u.command) {
		u.index++
		return u.command[u.index-1], nil
	}
	return "quit", nil
}

func (u *ui) PrintErr(a ...interface{}) {
	u.Report.Errors = append(u.Report.Errors,fmt.Sprint(a...))
}

func (u *ui) Print(a ...interface{}) {}
func (u *ui) IsTerminal() bool { return false }
func (u *ui) SetAutoComplete(complete func(string) string) {}

type writer struct { *Report }

func (w *writer) Open(name string) (io.WriteCloser, error) {
	return w, nil
}

func (w *writer) Write(p []byte) (n int, err error) {
	tf := func(s string) (r []string) {
		r = make([]string, 0, len(s))
		for _, v := range strings.Fields(s) {
			if s := strings.TrimSpace(v); len(s) > 0 {
				r = append(r,s)
			}
		}
		return
	}

	skip := true
	for _, l := range strings.Split(string(p),"\n") {
		a := tf(l)
		if skip && "flat flat% sum% cum cum%" == strings.Join(a," ") {
			skip = false
		}
		if !skip && len(a) > 5 {
			i := &Row{}
			fmt.Sscanf(a[0],"%f",&i.Flat)
			fmt.Sscanf(a[1],"%f",&i.FlatPercent)
			fmt.Sscanf(a[2],"%f",&i.SumPercent)
			fmt.Sscanf(a[3],"%f",&i.Cum)
			fmt.Sscanf(a[4],"%f",&i.CumPercent)
			i.Function = a[5]
			w.Report.Rows = append(w.Report.Rows,i)
		}
	}
	return
}

func (w *writer) Close() error {
	return nil
}

type flagset struct {
	*flag.FlagSet
	args []string
}

func (f *flagset) Bool(name string, def bool, usage string) *bool {
	return f.FlagSet.Bool(name,def,usage)
}

func (f *flagset) Int(name string, def int, usage string) *int {
	return f.FlagSet.Int(name,def,usage)
}

func (f *flagset) Float64(name string, def float64, usage string) *float64 {
	return f.FlagSet.Float64(name,def,usage)
}

func (f *flagset) String(name string, def string, usage string) *string {
	return f.FlagSet.String(name,def,usage)
}

func (f *flagset) BoolVar(pointer *bool, name string, def bool, usage string) {
	f.FlagSet.BoolVar(pointer,name,def,usage)
}

func (f *flagset) IntVar(pointer *int, name string, def int, usage string) {
	f.FlagSet.IntVar(pointer,name,def,usage)
}

func (f *flagset) Float64Var(pointer *float64, name string, def float64, usage string) {
	f.FlagSet.Float64Var(pointer,name,def,usage)
}

func (f *flagset) StringVar(pointer *string, name string, def string, usage string) {
	f.FlagSet.StringVar(pointer,name,def,usage)
}

func (f *flagset) StringList(name string, def string, usage string) *[]*string {
	return &[]*string{f.FlagSet.String(name, def, usage)}

}

func (f *flagset) ExtraUsage() string {
	return ""
}

func (f *flagset) Parse(usage func()) []string {
	//f.FlagSet.Usage = usage
	f.FlagSet.Usage = func(){}
	f.FlagSet.Parse(f.args)
	return f.FlagSet.Args()
}
