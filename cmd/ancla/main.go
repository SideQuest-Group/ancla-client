package main

import (
	"os"

	cli "github.com/SideQuest-Group/ancla-client/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
