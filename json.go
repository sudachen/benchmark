
package benchmark

import (
	"io"
	"encoding/json"
	"fmt"
	"container/list"
	"time"
	"errors"
	"strconv"
	"io/ioutil"
)

func (k messageKind) MarshalJSON() ([]byte,error) {
	return []byte("\""+k.String()+"\""), nil
}

func (k *messageKind) fromString(s string) error {
	switch s {
	case "MsgError": *k = MsgError
	case "MsgInfo":  *k = MsgInfo
	case "MsgDebug": *k = MsgDebug
	case "MsgOpt":   *k = MsgOpt
	case "MsgPprof": *k = MsgPprof
	default:
		return fmt.Errorf("invalid message kind %s",s)
	}
	return nil
}

func (t *T) toMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["label"]  = t.Label
	m["count"]  = fmt.Sprintf("%v",t.Count)
	m["active"] = fmt.Sprintf("%v",uint64(t.Active))
	m["total"]  = fmt.Sprintf("%v",uint64(t.Total))

	if t.Err != nil { m["error"] = t.Err.Error() }

	if t.Children.Len() != 0 {
		children := make([]interface{}, 0, t.Children.Len())
		for e := t.Children.Front(); e != nil; e = e.Next() {
			children = append(children, e.Value)
		}
		m["children"] = children
	}

	if t.Messages.Len() != 0 {
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

	return m
}

func (t *T) MarshalJSON() ([]byte,error) {
	return json.Marshal(t.toMap())
}

func (t *T) fromMap(m map[string]interface{}) error {
	t.Label = m["label"].(string)

	if e,ok := m["error"]; ok {
		t.Err = errors.New(e.(string))
	}

	var v int64
	v, _ = strconv.ParseInt(m["count"].(string),10,64)
	t.Count = int(v)
	v, _ = strconv.ParseInt(m["active"].(string),10,64)
	t.Active = time.Duration(v)
	v, _ = strconv.ParseInt(m["total"].(string),10,64)
	t.Total = time.Duration(v)

	var a []interface{}

	t.Children = list.New()
	if v,ok := m["children"]; ok {
		a = v.([]interface{})
		for _, x := range a {
			t0 := &T{}
			t0.fromMap(x.(map[string]interface{}))
			t.Children.PushBack(t0)
		}
	}

	t.Messages = list.New()
	if v,ok := m["messages"]; ok {
		a = v.([]interface{})
		for _, x := range a {
			y := x.(map[string]interface{})
			m0 := &Message{}
			m0.Text = y["text"].(string)
			m0.Kind.fromString(y["kind"].(string))
			t.Messages.PushBack(m0)
		}
	}

	return nil
}

func (t *T) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	if err := json.Unmarshal(b,&m); err != nil {
		return err
	}
	return t.fromMap(m)
}

func (t *T) WriteJson(wr io.Writer) (int,error) {
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
		return json.Unmarshal(b,t)
	}
}

