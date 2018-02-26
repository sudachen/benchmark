package benchmark

import (
	"bytes"
	"container/list"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	ppf "github.com/sudachen/benchmark/ppftool"
	"github.com/sudachen/misc"
)

var flagNoGC = flag.Bool("nogc", false, "disable GC on benchmark")
var flagPprof = flag.Bool("pprof", false, "profile benchmarks")
var flagMprof = flag.Bool("mprof", false, "profile memory allocations")
var flagCpuProf = flag.String("cpuprof", misc.NulStr, "where to store cpuprofile")
var flagMemProf = flag.String("memprof", misc.NulStr, "where to store memprofile")
var flagCallGrapth = flag.Int("callgraph", 0, "count of nodes to write PNG callgraph")
var flagNodeCount = flag.Int("nodecount", 20, "count of nodes to gather top calls")
var flagResult = flag.String("result", misc.NulStr, "file name to write the benchmarking result")

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

	processor          func(t *T, finished *T) *T
	startedAt, runOn   time.Time
	chActive, chPaused time.Duration

	Err   error
	Label string
	Count int

	Active, Total time.Duration

	Children, Messages *list.List
	Heap               *ppf.Report
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

var heapbf bytes.Buffer

/*func (t *T) profileHeap() {
	if *flagMprof {
		runtime.GC()
		pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("_", "heap")))
		heapbf.Reset()
		pprof.WriteHeapProfile(&heapbf)

		opt := &ppf.Options{
			Unit:     ppf.Second,
			Count:    *flagNodeCount,
			TagHide:  []string{"_:"},
			Hide:     []string{"google/pprof\\.:"},
		}

		if rpt, err := ppf.Top(heapbf.Bytes(), opt); err == nil {
			rpt.Label = "heap"
			t.Heap = rpt
		}
		pprof.SetGoroutineLabels(context.Background())
	}
}*/

func (t *T) run(f func(*T) error) (err error) {
	if t.startedAt != (time.Time{}) {
		panic("start is allowed only in leaf tasks")
	} else {
		pprof.SetGoroutineLabels(pprof.WithLabels(context.Background(), pprof.Labels("_", "prepare")))
	}

	t.runOn = time.Now()
	t.Err = f(t)
	t.Total = time.Since(t.runOn)

	if t.Children.Len() != 0 {
		t.Active = t.chActive
	} else {
		if t.isStarted {
			t.Active = time.Since(t.startedAt)
			if t.enableGC {
				enableGC()
			}
			//t.profileHeap()
		}
	}

	return
}

const PprofBufferReserve = 1024 * 1024

func (t *Benchmark) pprofRun(f func(*T) error) {
	var cpubuf bytes.Buffer
	var membuf bytes.Buffer

	if *flagPprof || *flagCpuProf != misc.NulStr {
		cpubuf.Grow(PprofBufferReserve)
		//runtime.SetCPUProfileRate(10000)
		pprof.StartCPUProfile(&cpubuf)
	}

	t.run(f)

	if *flagPprof || *flagCpuProf != misc.NulStr {
		pprof.StopCPUProfile()
	}

	if *flagMemProf != misc.NulStr || *flagMprof {
		runtime.GC()
		pprof.WriteHeapProfile(&membuf)
		if *flagMemProf != misc.NulStr {
			if f, err := os.Create(*flagMemProf); err == nil {
				f.Write(membuf.Bytes())
			} else {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}

	if *flagCpuProf != misc.NulStr {
		if err := ioutil.WriteFile(*flagCpuProf, membuf.Bytes(), 0644); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	if *flagPprof || *flagMprof {
		t.Pprof = list.New()

		count := *flagNodeCount
		pngcount := *flagCallGrapth

		opt := &ppf.Options{
			Count:    count,
			Graph:    ppf.DOT,
			Gcount:   pngcount,
			NoLegend: true,
			TagHide:  []string{"_:"},
			Hide:     []string{"google/pprof\\."},
		}

		if *flagPprof {
			opt.Unit = ppf.Second
			opt.Index = ppf.CpuProfIndex
			rpt, _ := ppf.Top(cpubuf.Bytes(), opt)
			rpt.Label = "top"
			t.Pprof.PushBack(rpt)
		}

		if *flagMprof {
			opt.Unit = ppf.None
			opt.Index = ppf.AllocObjectsIndex
			rpt, _ := ppf.Top(membuf.Bytes(), opt)
			rpt.Label = "alloc"
			t.Pprof.PushBack(rpt)
		}
	}
}

func Run(label string, f func(*T) error) *Benchmark {
	if !flag.Parsed() {
		flag.Parse()
	}
	runtime.LockOSThread()
	b := &Benchmark{T: New(label)}
	b.pprofRun(f)
	return b
}

func RunWithProcessor(label string, processor func(*T, *T) *T, f func(*T) error) *Benchmark {
	if !flag.Parsed() {
		flag.Parse()
	}
	runtime.LockOSThread()
	b := &Benchmark{T: New(label)}
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

func (b *Benchmark) WriteJsonResult() (int, error) {
	if *flagResult != misc.NulStr {
		if f, err := os.Create(*flagResult); err != nil {
			return 0, err
		} else {
			return b.WriteJson(f)
		}
	} else {
		return b.WriteJson(os.Stderr)
	}
}
