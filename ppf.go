package benchmark

import(
	"github.com/google/pprof/profile"
	"github.com/google/pprof/driver"
	"io"
	"flag"
	"fmt"
	"time"
	"strings"
)

const DefaultProfile = "@="

type fetcher struct {
	b []byte
}

func (f *fetcher) Fetch(src string, duration, timeout time.Duration) (*profile.Profile, string, error) {
	if src == DefaultProfile {
		p, err := profile.ParseData(f.b)
		return p, "", err
	}
	return nil, "", fmt.Errorf("unknown source %s",src)
}

type ui struct {
	*T
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

func (u *ui) Print(a ...interface{}) {}
func (u *ui) PrintErr(a ...interface{}) {}
func (u *ui) IsTerminal() bool { return false }
func (u *ui) SetAutoComplete(complete func(string) string) {}

type writer struct { *T }

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
		q := strings.Join(tf(l)," ")
		if skip && q == "flat flat% sum% cum cum%" {
			skip = false
		}
		if !skip && len(q) > 0 {
			w.T.Pprof(q)
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

func WritePprofReport(b []byte, t *T, c ...string) {
	f := &flagset{ flag.NewFlagSet("ppf",flag.ContinueOnError ), []string{DefaultProfile} }
	o := &driver.Options{
		Flagset: f,
		Fetch: &fetcher{b},
		Writer: &writer{t},
		UI: &ui{t,append([]string{"output=@"}, c...), 0},
	}
	driver.PProf(o)
}
