package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"time"

	"gocv.io/x/gocv"
)

func main() {
	var sampleImage bool
	var showWindow bool
	var showCamera bool
	var seconds int
	
	flag.BoolVar(&sampleImage, "sample-image", false, "Output a sample image using OpenCV and exit")
	flag.BoolVar(&showWindow, "show-window", false, "Display a window using OpenCV")
	flag.BoolVar(&showCamera, "show-camera", false, "Display the primary camera feed; press any key to exit")
	flag.IntVar(&seconds, "seconds", 10, "Number of seconds to show window")
	flag.IntVar(&seconds, "s", 10, "Number of seconds to show window (short form)")
	flag.Parse()

	if sampleImage {
		runSampleImage()
		return
	}
	if showWindow {
		if seconds <= 0 {
			fmt.Fprintf(os.Stderr, "Error: seconds must be a positive integer, got %d\n", seconds)
			os.Exit(1)
		}
		runShowWindow(seconds)
		return
	}
	if showCamera {
		runShowCamera()
		return
	}

	// Default behavior: show help
	flag.Usage()
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

// runShowCamera opens the default camera and displays its feed until a key is pressed
func runShowCamera() {
	fmt.Println("Opening default camera (device 0)...")

	// On macOS, ensure we're on the main thread for UI operations
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	// Initialize OpenCV window first to ensure it's created on the main thread
	win := gocv.NewWindow("GoCamNet Camera")
	defer win.Close()

	webcam, err := gocv.OpenVideoCapture(0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening video capture device: %v\n", err)
		os.Exit(1)
	}
	defer webcam.Close()

	if !webcam.IsOpened() {
		fmt.Fprintln(os.Stderr, "Error: video capture device not opened")
		os.Exit(1)
	}

	img := gocv.NewMat()
	defer img.Close()

	fmt.Println("Camera opened successfully. Press any key to exit...")

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Fprintln(os.Stderr, "Device closed or frame read failed")
			break
		}
		if img.Empty() {
			continue
		}

		win.IMShow(img)
		// WaitKey returns the key code if a key was pressed, -1 otherwise
		// Use a longer wait time to reduce CPU usage
		if key := win.WaitKey(30); key >= 0 {
			break
		}
	}

	fmt.Println("Camera window closed")
}
