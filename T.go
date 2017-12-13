
package benchmark

import (
	"time"
	"fmt"
	"container/list"
)

type messageKind byte;

const (
	MsgError messageKind = iota
	MsgInfo
	MsgDebug
)

func (mk messageKind) String() string {
	switch mk {
	case MsgError: return "MsgError"
	case MsgInfo:  return "MsgInfo"
	case MsgDebug: return "MsgDebug"
	}
	return ""
}

type Message struct {
	Kind messageKind `json:"kind"`
	Text string		 `json:"text"`
}

type T struct {
	pauseCount int
	startedAt, pausedAt time.Time
	processor  func(t *T,finished *T)*T

	Err   error
	Label string

	Children, Messages     *list.List
	Active, Paused, Total  time.Duration
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
	t.startedAt = time.Now()
	t.Err = f(t)
	t.Total = time.Since(t.startedAt)
	t.Active = t.Total - t.Paused
	return
}

func Run(label string, f func(*T)error) *T {
	t := New(label)
	t.run(f)
	return t
}

func RunWithProcessor(label string, processor func(*T,*T)*T, f func(*T)error) *T {
	t := New(label)
	t.processor = processor
	t.run(f)
	if t.processor != nil { t.processor(nil,t) }
	return t
}

func (t *T) Run(label string, f func(*T)error) (err error) {
	t0 := New(label)
	t0.processor = t.processor
	t0.run(f)
	if t.processor != nil { t0 = t.processor(t,t0)}
	if t0 != nil {
		t.Children.PushBack(t0)
	}
	return
}

func (t *T) Pause() {
	t.pauseCount++
	if t.pauseCount == 1 {
		t.pausedAt = time.Now()
	}
}

func (t *T) Resume() {
	if t.pauseCount > 0 {
		t.pauseCount--
		if t.pauseCount == 0 {
			t.Paused = t.Paused + time.Since(t.pausedAt)
		}
	}
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

