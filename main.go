package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	seconds := parseCLIArgs()
	runCountdown(seconds)
}

// parseCLIArgs handles command line argument parsing and validation
func parseCLIArgs() int {
	var seconds int
	flag.IntVar(&seconds, "seconds", 10, "Number of seconds to count")
	flag.IntVar(&seconds, "s", 10, "Number of seconds to count (short form)")
	flag.Parse()

	if seconds <= 0 {
		fmt.Fprintf(os.Stderr, "Error: seconds must be a positive integer, got %d\n", seconds)
		os.Exit(1)
	}

	return seconds
}

// runCountdown performs the actual countdown and printing
func runCountdown(seconds int) {
	fmt.Printf("Starting countdown for %d seconds...\n", seconds)

	for i := 1; i <= seconds; i++ {
		fmt.Printf("Count: %d\n", i)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("Countdown complete!")
}
