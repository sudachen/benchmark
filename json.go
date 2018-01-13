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

	if t.Pprof != nil && t.Pprof.Len() != 0 {
		ppf := make([]interface{}, 0, t.Pprof.Len())
		for e := t.Pprof.Front(); e != nil; e = e.Next() {
			p := e.Value.(*Report)
			v := make(map[string]interface{})
			v["label"] = p.Label
			v["unit"] = p.Unit.String()
			v["options"] = p.Options.String()
			r := make([]interface{}, len(p.Rows))
			for i, x := range p.Rows {
				r0 := make(map[string]string)
				r0["flat"] = strconv.FormatFloat(x.Flat, 'f', -1, 64)
				r0["flat%"] = strconv.FormatFloat(x.FlatPercent, 'f', -1, 64)
				r0["cum"] = strconv.FormatFloat(x.Cum, 'f', -1, 64)
				r0["cum%"] = strconv.FormatFloat(x.CumPercent, 'f', -1, 64)
				r0["sum%"] = strconv.FormatFloat(x.SumPercent, 'f', -1, 64)
				r0["function"] = x.Function
				r[i] = r0
			}
			v["rows"] = r
			r = make([]interface{}, len(p.Errors))
			for i, x := range p.Errors {
				r[i] = x
			}
			v["errors"] = r
			ppf = append(ppf, v)
		}
		m["pprof"] = ppf
	}

	return m
}

func (t *T) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.toMap())
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

	if v, ok := m["pprof"]; ok {
		t.Pprof = list.New()
		for _, x := range v.([]interface{}) {
			y := x.(map[string]interface{})
			p0 := &Report{}
			if err := p0.Unit.fromString(y["unit"].(string)); err != nil {
				return err
			}
			if err := p0.Options.fromString(y["options"].(string)); err != nil {
				return err
			}
			p0.Label = y["label"].(string)
			if u, ok := y["rows"]; ok {
				a := u.([]interface{})
				p0.Rows = make(Rows, 0, len(a))
				for _, z := range a {
					var err error
					w := z.(map[string]interface{})
					r := &Row{}
					if r.Flat, err = strconv.ParseFloat(w["flat"].(string), 64); err != nil {
						return err
					}
					if r.FlatPercent, err = strconv.ParseFloat(w["flat%"].(string), 64); err != nil {
						return err
					}
					if r.SumPercent, err = strconv.ParseFloat(w["sum%"].(string), 64); err != nil {
						return err
					}
					if r.Cum, err = strconv.ParseFloat(w["cum"].(string), 64); err != nil {
						return err
					}
					if r.CumPercent, err = strconv.ParseFloat(w["cum%"].(string), 64); err != nil {
						return err
					}
					r.Function = w["function"].(string)
					p0.Rows = append(p0.Rows, r)
				}
			}
			if u, ok := y["errors"]; ok {
				a := u.([]interface{})
				p0.Errors = make([]string, 0, len(a))
				for _, w := range a {
					p0.Errors = append(p0.Errors, w.(string))
				}
			}
			t.Pprof.PushBack(p0)
		}
	}

	return nil
}

func (t *T) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	return t.fromMap(m)
}

func (t *T) WriteJson(wr io.Writer) (int, error) {
	if b, err := json.MarshalIndent(t.toMap(), "", "\t"); err != nil {
		return 0, err
	} else {
		return wr.Write(b)
	}
}

func (t *T) ReadJson(rd io.Reader) error {
	if b, err := ioutil.ReadAll(rd); err != nil {
		return err
	} else {
		return json.Unmarshal(b, t)
	}
}
