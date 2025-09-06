package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	var runOpenCV bool
	flag.BoolVar(&runOpenCV, "opencv", false, "Run OpenCV example instead of countdown")

	seconds := parseCLIArgs()

	if runOpenCV {
		runOpenCVExample()
	} else {
		runCountdown(seconds)
	}
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

// runOpenCVExample demonstrates a basic OpenCV image operation
func runOpenCVExample() {
	fmt.Println("Running OpenCV example...")

	img := gocv.NewMatWithSize(200, 300, gocv.MatTypeCV8U)
	defer img.Close()

	// Draw a green rectangle
	gocv.Rectangle(&img, image.Rect(50, 50, 250, 150), color.RGBA{0, 255, 0, 255}, 2)

	outputFile := "output.jpg"
	if ok := gocv.IMWrite(outputFile, img); !ok {
		fmt.Printf("Error writing image: %s\n", outputFile)
		os.Exit(1)
	}
	fmt.Printf("Image successfully written to %s\n", outputFile)
}
