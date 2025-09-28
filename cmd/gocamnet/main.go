package main

import (
	"context"
	"flag"
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"time"

	"gocamnet/internal/pydetect"
	"gocv.io/x/gocv"
)

func main() {
	var sampleImage bool
	var showCamera bool
	var listCameras bool
	var cameraIndex int

	flag.BoolVar(&sampleImage, "sample-image", false, "Output a sample image using OpenCV and exit")
	flag.BoolVar(&showCamera, "show-camera", false, "Display the primary camera feed; press any key to exit")
	flag.BoolVar(&listCameras, "list-cameras", false, "List available cameras and exit")
	flag.IntVar(&cameraIndex, "camera", 0, "Camera index to use (use --list-cameras to see available cameras)")
	flag.Parse()

	if sampleImage {
		runSampleImage()
		return
	}
	if listCameras {
		listAvailableCameras()
		return
	}
	if showCamera {
		runShowCamera(cameraIndex)
		return
	}

	flag.Usage()
}

// runSampleImage writes a simple sample image using OpenCV
func runSampleImage() {
	fmt.Println("Generating sample image...")

	img := gocv.NewMatWithSize(200, 300, gocv.MatTypeCV8UC3)
	defer img.Close()

	img.SetTo(gocv.NewScalar(255, 255, 255, 0))
	gocv.Rectangle(&img, image.Rect(50, 50, 250, 150), color.RGBA{G: 255, A: 255}, -1)

	outputFile := "output.jpg"
	if ok := gocv.IMWrite(outputFile, img); !ok {
		fmt.Printf("Error writing image: %s\n", outputFile)
		os.Exit(1)
	}
	fmt.Printf("Image successfully written to %s\n", outputFile)
}

// listAvailableCameras tests cameras 0-9 and lists which ones are available
func listAvailableCameras() {
	fmt.Println("Scanning for available cameras...")
	fmt.Println()

	availableCameras := []int{}

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

	if runtime.GOOS == "darwin" {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()
	}

	displayAvailable := hasDisplay()
	if !displayAvailable {
		if os.Getenv("QT_QPA_PLATFORM") == "" {
			_ = os.Setenv("QT_QPA_PLATFORM", "offscreen")
		}
		fmt.Println("No display detected; running headless. Preview window disabled.")
		fmt.Println("Press Ctrl+C to stop.")
	}

	// Start managed PyTorch detector
	projectRoot, _ := os.Getwd()
	detector, err := pydetect.New(context.Background(), projectRoot, 0.25)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start PyTorch detector: %v\n", err)
		os.Exit(1)
	}
	defer detector.Close()
	fmt.Println("Detector ready.")

	var win *gocv.Window
	if displayAvailable {
		win = gocv.NewWindow("GoCamNet Camera")
		defer win.Close()
	}

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

	webcam.Set(gocv.VideoCaptureFPS, 30)
	webcam.Set(gocv.VideoCaptureFrameWidth, 640)
	webcam.Set(gocv.VideoCaptureFrameHeight, 480)
	webcam.Set(gocv.VideoCaptureBufferSize, 1)

	img := gocv.NewMat()
	defer img.Close()

	if displayAvailable {
		fmt.Println("Camera opened successfully. Press any key to exit...")
	} else {
		fmt.Println("Camera opened successfully. Streaming without preview...")
	}

	frameCount := 0
	lastReportTime := time.Now()
	var totalInferenceTime time.Duration
	inferenceCount := 0

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Fprintln(os.Stderr, "Device closed or frame read failed")
			break
		}
		if img.Empty() {
			continue
		}

		inferenceStart := time.Now()
		detections, err := detector.Detect(img, uint64(frameCount))
		totalInferenceTime += time.Since(inferenceStart)
		inferenceCount++

		if err != nil {
			fmt.Fprintf(os.Stderr, "Detection error: %v\n", err)
			continue
		}

		personColor := color.RGBA{0, 255, 0, 255}
		for _, d := range detections {
			if d.ClassID == 0 || d.ClassName == "person" {
				gocv.Rectangle(&img, d.Box, personColor, 2)
			}
		}

		if win != nil {
			win.IMShow(img)
		}

		frameCount++

		currentTime := time.Now()
		if currentTime.Sub(lastReportTime) >= 5*time.Second {
			elapsed := currentTime.Sub(lastReportTime)
			fps := float64(frameCount) / elapsed.Seconds()
			avgInferenceMs := float64(0)
			if inferenceCount > 0 {
				avgInferenceMs = float64(totalInferenceTime.Milliseconds()) / float64(inferenceCount)
			}
			fmt.Printf("FPS: %.2f (Avg Inference: %.2f ms)\n", fps, avgInferenceMs)
			frameCount = 0
			inferenceCount = 0
			totalInferenceTime = 0
			lastReportTime = currentTime
		}

		if win != nil {
			if key := win.WaitKey(1); key >= 0 {
				break
			}
		} else {
			time.Sleep(time.Millisecond * 33)
		}
	}

	if win != nil {
		fmt.Println("Camera window closed")
	} else {
		fmt.Println("Camera session ended")
	}
}

func hasDisplay() bool {
	if os.Getenv("QT_QPA_PLATFORM") == "offscreen" {
		return false
	}
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		return true
	}
	if os.Getenv("DISPLAY") != "" {
		return true
	}
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}
	return false
}
