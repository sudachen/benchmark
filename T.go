
package benchmark

import (
	"time"
	"fmt"
	"container/list"
	"runtime/pprof"
	"bytes"
	"runtime"

	"github.com/sudachen/misc"
	"flag"
	"io/ioutil"
	"context"
)

var flagNoGC = flag.Bool("nogc", false, "disable GC on benchmark")
var flagPprof = flag.Bool("pprof", false, "profile benchmarks")
var flagCpuProf = flag.String("cpuprof", misc.NulStr, "where to store cpuprofile")

type messageKind byte

const (
	MsgError messageKind = iota
	MsgInfo
	MsgDebug
	MsgOpt
)

func (mk messageKind) String() string {
	switch mk {
	case MsgError:  return "MsgError"
	case MsgInfo:   return "MsgInfo"
	case MsgDebug:  return "MsgDebug"
	case MsgOpt: 	return "MsgOpt"
	}
	return ""
}

type Message struct {
	Kind messageKind
	Text string
}

type T struct {
	enableGC, isStarted, stopProfiler bool
	startedAt, runOn time.Time
	processor  func(t *T,finished *T)*T
	chActive, chPaused time.Duration

	Err   error
	Label string

	Count int
	Children, Messages, Pprof *list.List
	Active, Total time.Duration
}

func New(label string) *T {
	t := &T{
		Label:label,
		Children: list.New(),
		Messages: list.New(),
		}
	return t
}

func (t *T) run(f func(*T)error) (err error) {
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

const PprofBufferReserve = 1024*1024

func (t *T) pprofRun(f func(*T)error) {
	var buf bytes.Buffer
	if *flagPprof || *flagCpuProf != misc.NulStr {
		buf.Grow(PprofBufferReserve)
		runtime.SetCPUProfileRate(10000)
		pprof.StartCPUProfile(&buf)
	}
	t.run(f)
	if *flagPprof || *flagCpuProf != misc.NulStr {
		pprof.StopCPUProfile()
	}
	if *flagPprof {
		count := 25
		t.Pprof = list.New()
		t.Pprof.PushBack(Top(buf.Bytes(), count, Tagged, Msec, "top"))
		t.Pprof.PushBack(Top(buf.Bytes(), count, Tagged|SortByCum, Msec, "top-cum"))
		t.Pprof.PushBack(Top(buf.Bytes(), count, DefaultReport, Msec, "top-all"))
		t.Pprof.PushBack(Top(buf.Bytes(), count, SortByCum, Msec, "top-all-cum"))
		t.Pprof.PushBack(Top(buf.Bytes(), count, Tagged|RuntimeOnly, Msec, "top-rt"))
		t.Pprof.PushBack(Top(buf.Bytes(), count, Tagged|ExcludeRuntime, Msec, "top-nort"))
	}
	if *flagCpuProf != misc.NulStr {
		ioutil.WriteFile(*flagCpuProf, buf.Bytes(),0644)
	}
}

func Run(label string, f func(*T)error) *T {
	if !flag.Parsed() { flag.Parse() }
	runtime.LockOSThread()
	t := New(label)
	t.pprofRun(f)
	return t
}

func RunWithProcessor(label string, processor func(*T,*T)*T, f func(*T)error) *T {
	if !flag.Parsed() { flag.Parse() }
	runtime.LockOSThread()
	t := New(label)
	t.processor = processor
	t.pprofRun(f)
	if t.processor != nil { t.processor(nil,t) }
	return t
}

func (t *T) Run(label string, f func(*T)error) (err error) {
	t0 := New(label)
	t0.processor = t.processor
	if *flagPprof || *flagCpuProf != misc.NulStr {
		defer pprof.SetGoroutineLabels(context.Background())
	}
	t0.run(f)
	if t.processor != nil { t0 = t.processor(t,t0)}
	if t0 != nil {
		t.Children.PushBack(t0)
		t.chActive += t0.Active
	}
	return
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
		pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(),pprof.Labels("t","active")))
	}

	t.Count++
}

func (t *T) Errorf(ft string, a ...interface{}) {
	m := &Message{MsgError,fmt.Sprintf(ft,a...)}
	t.Messages.PushBack(m)
}

func (t *T) Error(a ...interface{}) {
	m := &Message{MsgError,fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Debugf(ft string, a ...interface{}) {
	m := &Message{MsgDebug,fmt.Sprintf(ft,a...)}
	t.Messages.PushBack(m)
}

func (t *T) Debug(a ...interface{}) {
	m := &Message{MsgDebug,fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Infof(ft string, a ...interface{}) {
	m := &Message{MsgInfo,fmt.Sprintf(ft,a...)}
	t.Messages.PushBack(m)
}

func (t *T) Info(a ...interface{}) {
	m := &Message{MsgInfo,fmt.Sprint(a...)}
	t.Messages.PushBack(m)
}

func (t *T) Opt(a string) {
	m := &Message{MsgOpt,a}
	t.Messages.PushBack(m)
}
