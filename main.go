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
	case "script":
		if len(os.Args) < 3 {
			fmt.Println("Error: script command requires a script_hex argument")
			fmt.Println("Usage: noa script <script_hex>")
			os.Exit(1)
		}
		scriptHex := os.Args[2]
		if err := command.RunScript(scriptHex); err != nil {
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
	fmt.Println("Usage: noa <command> [arguments]")
	fmt.Println("\nAvailable commands:")
	fmt.Println("  address <address_ark>")
	fmt.Println("  vtxos <address_ark> [indexer_url]  (default indexer_url: https://arkade.computer)")
	fmt.Println("  script <script_hex>")
}
