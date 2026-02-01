package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/queelius/deets/internal/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		var exitErr *commands.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Message != "" {
				fmt.Fprintln(os.Stderr, exitErr.Message)
			}
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}
