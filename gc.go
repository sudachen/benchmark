
package benchmark

import (
	"runtime"
	"runtime/debug"
)

var gcPercent int = -1

func disableGC(){
	if gcPercent == -1 {
		gcPercent = debug.SetGCPercent(-1)
		runtime.GC()
	}
}

func enableGC(){
	if gcPercent != -1 {
		debug.SetGCPercent(gcPercent)
		gcPercent = -1
	}
}
