package main

import (
	"os"

	"ex.com/goget/goget"
)

func main() {
	goget.ExecCmd(os.Args[1:])
}
