package main

import (
	"fmt"
	"log"
	"os"

	"github.com/rs/jplot/cmd"
	"github.com/rs/jplot/window"
)

func fatalIf(err error, msg string) {
	if err != nil {
		log.Fatal(msg + fmt.Sprintf(" error=%v", err))
		os.Exit(1)
	}
}

func main() {
	if _, _, err := window.Size(); err != nil {
		log.Fatal(fmt.Sprintf("Cannot get window size error=%v", err))
		os.Exit(1)
	}

	cmd.Execute()
}
