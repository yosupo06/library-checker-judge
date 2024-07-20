package main

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
)

var (
	app = kingpin.New("utils", "Library checker utils")
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case rejudgeCmd.FullCommand():
		execRejudgeCmd()
	}
}
