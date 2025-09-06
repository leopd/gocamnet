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
	var listCameras bool
	var cameraIndex int
	var seconds int
	
	flag.BoolVar(&sampleImage, "sample-image", false, "Output a sample image using OpenCV and exit")
	flag.BoolVar(&showWindow, "show-window", false, "Display a window using OpenCV")
	flag.BoolVar(&showCamera, "show-camera", false, "Display the primary camera feed; press any key to exit")
	flag.BoolVar(&listCameras, "list-cameras", false, "List available cameras and exit")
	flag.IntVar(&cameraIndex, "camera", 0, "Camera index to use (use --list-cameras to see available cameras)")
	flag.IntVar(&seconds, "seconds", 10, "Number of seconds to show window")
	flag.IntVar(&seconds, "s", 10, "Number of seconds to show window (short form)")
	flag.Parse()

	if sampleImage {
		runSampleImage()
		return
	}
	if listCameras {
		listAvailableCameras()
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
		runShowCamera(cameraIndex)
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

// listAvailableCameras tests cameras 0-9 and lists which ones are available
func listAvailableCameras() {
	fmt.Println("Scanning for available cameras...")
	fmt.Println()
	
	availableCameras := []int{}
	
	// Test cameras 0-9
	for i := 0; i < 10; i++ {
		webcam, err := gocv.OpenVideoCapture(i)
		if err == nil && webcam.IsOpened() {
			availableCameras = append(availableCameras, i)
			webcam.Close()
		}
	}
	
	if len(availableCameras) == 0 {
		fmt.Println("No cameras found.")
		return
	}
	
	fmt.Printf("Found %d available camera(s):\n", len(availableCameras))
	for _, index := range availableCameras {
		fmt.Printf("  Camera %d\n", index)
	}
	fmt.Println()
	fmt.Println("Use --camera <index> to select a specific camera.")
	fmt.Println("Example: ./gocamnet --show-camera --camera 1")
}

// runShowCamera opens the specified camera and displays its feed until a key is pressed
func runShowCamera(cameraIndex int) {
	fmt.Printf("Opening camera %d...\n", cameraIndex)

	// On macOS, ensure we're on the main thread for UI operations
	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	// Initialize OpenCV window first to ensure it's created on the main thread
	win := gocv.NewWindow("GoCamNet Camera")
	defer win.Close()

	webcam, err := gocv.OpenVideoCapture(cameraIndex)
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

	// FPS tracking variables
	frameCount := 0
	lastReportTime := time.Now()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Fprintln(os.Stderr, "Device closed or frame read failed")
			break
		}
		if img.Empty() {
			continue
		}

		win.IMShow(img)
		
		// Increment frame count
		frameCount++
		
		// Check if 5 seconds have passed since last report
		currentTime := time.Now()
		if currentTime.Sub(lastReportTime) >= 5*time.Second {
			elapsed := currentTime.Sub(lastReportTime)
			fps := float64(frameCount) / elapsed.Seconds()
			fmt.Printf("FPS: %.2f (frames: %d, elapsed: %.1fs)\n", fps, frameCount, elapsed.Seconds())
			// Reset frame count and update last report time
			frameCount = 0
			lastReportTime = currentTime
		}
		
		// WaitKey returns the key code if a key was pressed, -1 otherwise
		// Use a longer wait time to reduce CPU usage
		if key := win.WaitKey(30); key >= 0 {
			break
		}
	}

	fmt.Println("Camera window closed")
}
