package common

import (
	"path"
	"path/filepath"
	"runtime"
)

const TRUNC_LENGTH int = 256

// const TRUNC_LENGTH int = 10000000

func TruncStr(src []byte) []byte {
	if len(src) > TRUNC_LENGTH {
		result := string(src[0:TRUNC_LENGTH]) + "..."
		return []byte(result)
	} else {
		return src
	}
}

// CurrentDir returns current directory of the caller.
func CurrentDir() string {
	_, current, _, _ := runtime.Caller(1)
	return filepath.Join(path.Dir(current))
}
