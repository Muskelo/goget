package utils

import (
	"os"
)

// using to validate and parse filepath from string

func DirExist(name string) bool {
	stat, err := os.Stat(name)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
func FileExist(name string) bool {
	stat, err := os.Stat(name)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}
