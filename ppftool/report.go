package ppftool

import (
	"errors"
	"fmt"
	"strings"
)

type Report struct {
	Unit
	Rows
	Errors []string
	Label  string
	Image  string
}

type Rows []*Row

type Row struct {
	Function string

	Flat, FlatPercent, SumPercent, Cum, CumPercent float64
}

type Unit byte

const (
	Second      Unit = 0
	Millisecond Unit = 1
	Microsecond Unit = 2
	Megabyte    Unit = 3
	Kilobyte    Unit = 4
	Byte        Unit = 5
	None        Unit = 6
)

var DefaultUnit = Second

func (u Unit) String() string {
	switch u {
	case Microsecond:
		return "us"
	case Millisecond:
		return "ms"
	case Second:
		return "s"
	case Megabyte:
		return "mb"
	case Kilobyte:
		return "kb"
	case Byte:
		return "b"
	case None:
		return ""
	default:
		return DefaultUnit.String()
	}
}

func (u *Unit) FromString(s string) error {
	switch s {
	case "us":
		*u = Microsecond
		return nil
	case "ms":
		*u = Millisecond
		return nil
	case "s":
		*u = Second
		return nil
	case "mb":
		*u = Megabyte
		return nil
	case "kb":
		*u = Kilobyte
		return nil
	case "b":
		*u = Byte
		return nil
	case "":
		*u = None
		return nil
	}
	return errors.New("invalid unit string")
}

func (r *Report) Write(b []byte) (int, error) {

	if r.Rows == nil {
		r.Rows = make(Rows, 0)
	}

	tf := func(s string) (x []string) {
		x = make([]string, 0, len(s))
		for _, v := range strings.Fields(s) {
			if s := strings.TrimSpace(v); len(s) > 0 {
				x = append(x, s)
			}
		}
		return
	}

	skip := true
	for _, l := range strings.Split(string(b), "\n") {
		a := tf(l)
		if skip && "flat flat% sum% cum cum%" == strings.Join(a, " ") {
			skip = false
		}
		if !skip && len(a) > 5 {
			i := &Row{}
			fmt.Sscanf(a[0], "%f", &i.Flat)
			fmt.Sscanf(a[1], "%f", &i.FlatPercent)
			fmt.Sscanf(a[2], "%f", &i.SumPercent)
			fmt.Sscanf(a[3], "%f", &i.Cum)
			fmt.Sscanf(a[4], "%f", &i.CumPercent)
			i.Function = a[5]
			r.Rows = append(r.Rows, i)
		}
	}

	return len(b), nil
}
