package utils

import (
	"os"
	"path"
)

func DirExist(name string) bool {
	stat, err := os.Stat(name)
	if err != nil {
		return false
	}
	return stat.IsDir()
}
func ParentDirExist(name string) bool {
	dir := path.Dir(name)
	return DirExist(dir)
}
func FileExist(name string) bool {
	stat, err := os.Stat(name)
	if err != nil {
		return false
	}
	return !stat.IsDir()
}

