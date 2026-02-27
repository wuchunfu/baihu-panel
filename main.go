package main

import (
	"fmt"
	"os"

	"github.com/engigu/baihu-panel/cmd/reposync"
	"github.com/engigu/baihu-panel/internal/bootstrap"
)

func main() {
	if len(os.Args) < 2 {
		bootstrap.New().Run()
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "server":
		bootstrap.New().Run()
	case "reposync":
		reposync.Run(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		fmt.Println("Available commands: server, reposync")
		os.Exit(1)
	}
}
