package main

import (
	"os"

	"ex.com/goget/goget"
)

func main() {
	goget.Run(os.Args[1:])
}
