package utils

import "os"

// copied from internet
func PipeExist() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		return false
	} else {
		return true
	}
}
