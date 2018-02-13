package benchmark

import (
	"bytes"
	"container/list"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"runtime"
	"runtime/pprof"
	"time"
	"encoding/base64"

	"github.com/sudachen/misc"
	ppf "github.com/sudachen/pprof/util"
)

var flagNoGC = flag.Bool("nogc", false, "disable GC on benchmark")
var flagPprof = flag.Bool("pprof", false, "profile benchmarks")
var flagCpuProf = flag.String("cpuprof", misc.NulStr, "where to store cpuprofile")
var flagCallgrapth = flag.Int("callgraph", -1, "count of nodes to write PNG callgraph")

type messageKind byte

const (
	MsgError messageKind = iota
	MsgInfo
	MsgDebug
	MsgOpt
)

func (mk messageKind) String() string {
	switch mk {
	case MsgError:
		return "MsgError"
	case MsgInfo:
		return "MsgInfo"
	case MsgDebug:
		return "MsgDebug"
	case MsgOpt:
		return "MsgOpt"
	}
	return ""
}

type Message struct {
	Kind messageKind
	Text string
}

type T struct {
	enableGC, isStarted, stopProfiler bool

	processor           func(t *T, finished *T) *T
	startedAt, runOn    time.Time
	chActive, chPaused  time.Duration

	Err   error
	Label string
	Count int

	Active, Total time.Duration

	Children, Messages *list.List
}

type Benchmark struct {
	*T
	Pprof *list.List
}

func New(label string) *T {
	t := &T{
		Label:    label,
		Children: list.New(),
		Messages: list.New(),
	}
	return t
}

func (t *T) run(f func(*T) error) (err error) {
	if t.startedAt != (time.Time{}) {
		panic("start is allowed only in leaf tasks")
	}

	t.runOn = time.Now()
	t.Err = f(t)
	t.Total = time.Since(t.runOn)

	if t.Children.Len() != 0 {
		t.Active = t.chActive
	} else {
		if t.isStarted {
			t.Active = time.Since(t.startedAt)
		}
	}

	if t.enableGC {
		enableGC()
	}

	return
}

const PprofBufferReserve = 1024 * 1024

func (t *Benchmark) pprofRun(f func(*T) error) {
	var buf bytes.Buffer
	if *flagPprof || *flagCpuProf != misc.NulStr {
		buf.Grow(PprofBufferReserve)
		//runtime.SetCPUProfileRate(10000)
		pprof.StartCPUProfile(&buf)
	}
	t.run(f)
	if *flagPprof || *flagCpuProf != misc.NulStr {
		pprof.StopCPUProfile()
	}
	if *flagPprof {
		count := 25
		pngcount := *flagCallgrapth
		if pngcount == 0 { pngcount = count }
		t.Pprof = list.New()
		opt := &ppf.Options{ TagFocus: []string{"t:"}, Unit: ppf.Second }
		rpt := ppf.Top(buf.Bytes(), count, opt, "top")
		if pngcount > 0 {
			rpt.Image = base64.StdEncoding.EncodeToString(ppf.Png(buf.Bytes(), pngcount, opt))
		}
		t.Pprof.PushBack(rpt)
		opt = &ppf.Options{ Unit: ppf.Second }
		rpt = ppf.Top(buf.Bytes(), count, opt, "top-all")
		if pngcount > 0 {
			rpt.Image = base64.StdEncoding.EncodeToString(ppf.Png(buf.Bytes(), pngcount, opt))
		}
		t.Pprof.PushBack(rpt)
	}
	if *flagCpuProf != misc.NulStr {
		ioutil.WriteFile(*flagCpuProf, buf.Bytes(), 0644)
	}
}

func Run(label string, f func(*T) error) *Benchmark {
	if !flag.Parsed() {
		flag.Parse()
	}
	runtime.LockOSThread()
	b := &Benchmark{T:New(label)}
	b.pprofRun(f)
	return b
}

func RunWithProcessor(label string, processor func(*T, *T) *T, f func(*T) error) *Benchmark {
	if !flag.Parsed() {
		flag.Parse()
	}
	runtime.LockOSThread()
	b := &Benchmark{T:New(label)}
	b.processor = processor
	b.pprofRun(f)
	if b.processor != nil {
		b.processor(nil, b.T)
	}
	return b
}

func (t *T) Run(label string, f func(*T) error) (err error) {
	t0 := New(label)
	t0.processor = t.processor
	if *flagPprof || *flagCpuProf != misc.NulStr {
		defer pprof.SetGoroutineLabels(context.Background())
	}
	t0.run(f)
	if t.processor != nil {
		t0 = t.processor(t, t0)
	}
	if t0 != nil {
		t.Children.PushBack(t0)
		t.chActive += t0.Active
	}
	return t0.Err
}

func (t *T) Start() {

	if t.Children.Len() != 0 {
		panic("start is allowed only in leaf tasks")
	}

	if !t.isStarted {
		runtime.GC()
		if *flagNoGC {
			t.enableGC = true
			disableGC()
		}
		t.isStarted = true
		t.startedAt = time.Now()
		pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("t", "active")))
	}

	t.Count++
}

func (t *T) Errorf(ft string, a ...interface{}) {
	m := &Message{MsgError, fmt.Sprintf(ft, a...)}
	t.Messages.PushBack(m)
}

func (t *T) Error(a ...interface{}) {
	m := &Message{MsgError, fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Debugf(ft string, a ...interface{}) {
	m := &Message{MsgDebug, fmt.Sprintf(ft, a...)}
	t.Messages.PushBack(m)
}

func (t *T) Debug(a ...interface{}) {
	m := &Message{MsgDebug, fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Infof(ft string, a ...interface{}) {
	m := &Message{MsgInfo, fmt.Sprintf(ft, a...)}
	t.Messages.PushBack(m)
}

func (t *T) Info(a ...interface{}) {
	m := &Message{MsgInfo, fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Opt(a string) {
	m := &Message{MsgOpt, a}
	t.Messages.PushBack(m)
}
