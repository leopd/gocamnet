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
	var sampleImage bool
	var showWindow bool
	flag.BoolVar(&sampleImage, "sample-image", false, "Output a sample image using OpenCV and exit")
	flag.BoolVar(&showWindow, "show-window", false, "Display a window using OpenCV")

	seconds := parseCLIArgs()

	if sampleImage {
		runSampleImage()
		return
	}
	if showWindow {
		runShowWindow(seconds)
		return
	}

	runCountdown(seconds)
}

// parseCLIArgs handles command line argument parsing and validation
func parseCLIArgs() int {
	var seconds int
	flag.IntVar(&seconds, "seconds", 10, "Number of seconds to count or show window")
	flag.IntVar(&seconds, "s", 10, "Number of seconds to count or show window (short form)")
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

// runSampleImage writes a simple sample image using OpenCV
func runSampleImage() {
	fmt.Println("Generating sample image...")

	// Create a 3-channel (BGR) image so colors render correctly
	img := gocv.NewMatWithSize(200, 300, gocv.MatTypeCV8UC3)
	defer img.Close()

	// Set a white background for visibility
	img.SetTo(gocv.NewScalar(255, 255, 255, 0)) // B, G, R, A

	// Draw a filled green rectangle
	gocv.Rectangle(&img, image.Rect(50, 50, 250, 150), color.RGBA{G: 255, A: 255}, -1)

	outputFile := "output.jpg"
	if ok := gocv.IMWrite(outputFile, img); !ok {
		fmt.Printf("Error writing image: %s\n", outputFile)
		os.Exit(1)
	}
	fmt.Printf("Image successfully written to %s\n", outputFile)
}

// runShowWindow opens a window and displays a sample image for the given duration
func runShowWindow(seconds int) {
	fmt.Printf("Showing window for %d seconds...\n", seconds)

	img := gocv.NewMatWithSize(300, 480, gocv.MatTypeCV8UC3)
	defer img.Close()
	img.SetTo(gocv.NewScalar(255, 255, 255, 0))
	gocv.Rectangle(&img, image.Rect(80, 80, 400, 220), color.RGBA{B: 255, A: 255}, 3)
	gocv.PutText(&img, "GoCamNet", image.Pt(120, 160), gocv.FontHersheySimplex, 1.0, color.RGBA{R: 255, A: 255}, 2)

	win := gocv.NewWindow("GoCamNet Sample")
	defer win.Close()

	end := time.Now().Add(time.Duration(seconds) * time.Second)
	for time.Now().Before(end) {
		win.IMShow(img)
		if key := win.WaitKey(10); key >= 0 {
			break
		}
	}

	fmt.Println("Window closed")
}
