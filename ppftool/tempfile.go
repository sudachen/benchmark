package ppftool

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
)

func TempFileName() string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), "pprof."+hex.EncodeToString(randBytes)+".out")
}
