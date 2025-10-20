package main

import (
	"fmt"
	"os"

	"github.com/golshani-mhd/grizzle-kit/cmd/grizzle-kit/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
