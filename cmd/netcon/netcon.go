package main

import (
	"fmt"
	"os"

	"github.com/janog-netcon/netcon-cli/pkg/command"
)

func main() {
	command := command.NewNetconCommand()

	if err := command.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
