package main

import (
	"fmt"
	"os"
)

func main() {
	err := rootCmd().Execute()
	if err != nil {
		fmt.Printf("Error: %s", err)
		os.Exit(0)
	}
}
