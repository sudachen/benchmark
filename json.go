package benchmark

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	ppf "github.com/sudachen/benchmark/ppftool"
)

func (k messageKind) MarshalJSON() ([]byte, error) {
	return []byte("\"" + k.String() + "\""), nil
}

func (k *messageKind) fromString(s string) error {
	switch s {
	case "MsgError":
		*k = MsgError
	case "MsgInfo":
		*k = MsgInfo
	case "MsgDebug":
		*k = MsgDebug
	case "MsgOpt":
		*k = MsgOpt
	default:
		return fmt.Errorf("invalid message kind %s", s)
	}
	return nil
}

func (b *Benchmark) toMap() map[string]interface{} {
	m := b.T.toMap()

	if b.Pprof != nil && b.Pprof.Len() != 0 {
		f := make([]interface{}, 0, b.Pprof.Len())
		for e := b.Pprof.Front(); e != nil; e = e.Next() {
			p := e.Value.(*ppf.Report)
			v := p.ToMap()
			f = append(f, v)
		}
		m["pprof"] = f
	}

	return m
}

func (t *T) toMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["label"] = t.Label
	m["count"] = fmt.Sprintf("%v", t.Count)
	m["active"] = fmt.Sprintf("%v", uint64(t.Active))
	m["total"] = fmt.Sprintf("%v", uint64(t.Total))

	if t.Err != nil {
		m["error"] = t.Err.Error()
	}

	if t.Children != nil && t.Children.Len() != 0 {
		children := make([]interface{}, 0, t.Children.Len())
		for e := t.Children.Front(); e != nil; e = e.Next() {
			children = append(children, e.Value)
		}
		m["children"] = children
	}

	if t.Messages != nil && t.Messages.Len() != 0 {
		messages := make([]interface{}, 0, t.Messages.Len())
		for e := t.Messages.Front(); e != nil; e = e.Next() {
			msg := e.Value.(*Message)
			v := make(map[string]string)
			v["kind"] = msg.Kind.String()
			v["text"] = msg.Text
			messages = append(messages, v)
		}
		m["messages"] = messages
	}

	if t.Heap != nil {
		m["heap"] = t.Heap.ToMap()
	}

	return m
}

func (t *T) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toMap())
}

func (b *Benchmark) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.toMap())
}

func (b *Benchmark) fromMap(m map[string]interface{}) error {
	b.T = &T{}
	if err := b.T.fromMap(m); err != nil {
		return err
	}
	if v, ok := m["pprof"]; ok {
		b.Pprof = list.New()
		for _, x := range v.([]interface{}) {
			y := x.(map[string]interface{})
			p0 := &ppf.Report{}
			p0.FromMap(y)
			b.Pprof.PushBack(p0)
		}
	}
	return nil
}

func (t *T) fromMap(m map[string]interface{}) error {
	t.Label = m["label"].(string)

	if e, ok := m["error"]; ok {
		t.Err = errors.New(e.(string))
	}

	if v, err := strconv.ParseInt(m["count"].(string), 10, 64); err != nil {
		return err
	} else {
		t.Count = int(v)
	}
	if v, err := strconv.ParseInt(m["active"].(string), 10, 64); err != nil {
		return err
	} else {
		t.Active = time.Duration(v)
	}
	if v, err := strconv.ParseInt(m["total"].(string), 10, 64); err != nil {
		return err
	} else {
		t.Total = time.Duration(v)
	}

	t.Children = list.New()
	if v, ok := m["children"]; ok {
		for _, x := range v.([]interface{}) {
			t0 := &T{}
			t0.fromMap(x.(map[string]interface{}))
			t.Children.PushBack(t0)
		}
	}

	t.Messages = list.New()
	if v, ok := m["messages"]; ok {
		for _, x := range v.([]interface{}) {
			y := x.(map[string]interface{})
			m0 := &Message{}
			m0.Text = y["text"].(string)
			if err := m0.Kind.fromString(y["kind"].(string)); err != nil {
				return err
			}
			t.Messages.PushBack(m0)
		}
	}

	if v, ok := m["heap"]; ok {
		y := v.(map[string]interface{})
		p0 := &ppf.Report{}
		p0.FromMap(y)
		t.Heap = p0
	}

	return nil
}

func (b *Benchmark) UnmarshalJSON(bs []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(bs, &m); err != nil {
		return err
	}
	return b.fromMap(m)
}

func (b *Benchmark) WriteJson(wr io.Writer) (int, error) {
	if bs, err := json.MarshalIndent(b.toMap(), "", "\t"); err != nil {
		return 0, err
	} else {
		return wr.Write(bs)
	}
}

func (b *Benchmark) ReadJson(rd io.Reader) error {
	if bs, err := ioutil.ReadAll(rd); err != nil {
		return err
	} else {
		return json.Unmarshal(bs, b)
	}
}
