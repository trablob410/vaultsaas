package main

import (
	"fmt"
	"os"

	"github.com/valt-dev/valt/server/cmd/valt/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
