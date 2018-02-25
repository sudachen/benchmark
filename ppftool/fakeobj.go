package ppftool

import(
	"errors"

	"github.com/google/pprof/driver"
)

type objtool struct {}

func (o objtool) Open(file string, start, limit, offset uint64) (driver.ObjFile, error) {
	return nil, errors.New("nofile")
}

func (o objtool) Disasm(file string, start, end uint64) ([]driver.Inst, error) {
	return nil, errors.New("nofile")
}
