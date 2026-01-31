package main

import (
	"os"

	"github.com/queelius/deets/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
