package benchmark

import (
	"testing"
	"errors"
	"bytes"
	"bufio"
)

type F struct {
	Label string
	Func  func(*T)error
}

var funcs = []*F{
	&F{"ItFails",func(t *T)error { return errors.New("it fails always")}},
	&F{"ItsSuccessful",func(t *T)error { return nil}},
	&F{"ItWritesMessages",func(t *T)error {
		t.Info("hello!")
		t.Debugf("benchmarking in test %s",t.Label)
		t.Error("shit happens")
		return nil
	}},
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func Test1(t *testing.T) {

	t0 := Run(".",func(t1 *T)error{
		for _, f := range funcs {
			t1.Run(f.Label,f.Func)
		}
		return nil
	})

	bf := &bytes.Buffer{}
	wr := bufio.NewWriter(bf)
	t0.WriteJson(wr)
	wr.Flush()

	s := bf.String()

	rd := bufio.NewReader(bf)
	t1 := &T{}
	t1.ReadJson(rd)

	bf = &bytes.Buffer{}
	wr = bufio.NewWriter(bf)
	t1.WriteJson(wr)
	wr.Flush()

	s1 := bf.String()
	if s1 != s {
		if len(s1) != len(s) {
			t.Errorf("length is not matched %v != %v",len(s1),len(s))
		}
		L := min(len(s1),len(s))
		for i := 0; i < L; i++ {
			if s1[i] != s[i] {
				t.Errorf("found difference at %d",i)
				t.Errorf("\t s  => %s",s[i:i+min(L-i,20)])
				t.Errorf("\t s1 => %s",s1[i:i+min(L-i,20)])
				break
			}
		}

		t.Log(s)
		t.Log(s1)
		t.Error("write/read/write failed")
	}

}