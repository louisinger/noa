package main

import (
	"fmt"
	"os"

	"github.com/louisinger/noa/command"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "address":
		if len(os.Args) < 3 {
			fmt.Println("Error: address command requires an address_ark argument")
			fmt.Println("Usage: noa address <address_ark>")
			os.Exit(1)
		}
		arkAddress := os.Args[2]
		if err := command.RunAddress(arkAddress); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: noa address <address_ark>")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  address <address_ark>")
}
