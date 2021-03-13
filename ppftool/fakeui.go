package ppftool

import (
	"fmt"

	"github.com/google/pprof/driver"
)

type ui struct {
	*Report
}

func (u *ui) ReadLine(prompt string) (string, error) {
	return "quit", nil
}

func (u *ui) PrintErr(a ...interface{}) {
	if u.Report != nil {
		u.Errors = append(u.Errors, fmt.Sprint(a...))
	}
}

func (u *ui) IsTerminal() bool { return false }

func (u *ui) Print(a ...interface{}) {}

func (u *ui) SetAutoComplete(complete func(string) string) {}

func (u *ui) WantBrowser() bool { return false }

func FakeUi() driver.UI {
	return &ui{}
}
