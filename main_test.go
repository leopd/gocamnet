package main

import (
	"os"
	"testing"

	"gocv.io/x/gocv"
)

func TestRunSampleImage(t *testing.T) {
	// Clean up any existing output file
	os.Remove("output.jpg")
	defer os.Remove("output.jpg")

	// Run the sample image function
	runSampleImage()

	// Check that the file was created
	if _, err := os.Stat("output.jpg"); os.IsNotExist(err) {
		t.Fatal("output.jpg was not created")
	}

	// Load the image using GoCV
	img := gocv.IMRead("output.jpg", gocv.IMReadColor)
	if img.Empty() {
		t.Fatal("Failed to load output.jpg as a valid image")
	}
	defer img.Close()

	// Check image dimensions
	if img.Rows() != 200 || img.Cols() != 300 {
		t.Errorf("Expected image size 200x300, got %dx%d", img.Rows(), img.Cols())
	}

	// Check that it's a 3-channel image
	if img.Channels() != 3 {
		t.Errorf("Expected 3 channels, got %d", img.Channels())
	}

	// Count green pixels (BGR format: green is channel 1)
	greenPixels := 0
	totalPixels := img.Rows() * img.Cols()

	for y := 0; y < img.Rows(); y++ {
		for x := 0; x < img.Cols(); x++ {
			// Get BGR values
			b := img.GetUCharAt(y, x*3+0)
			g := img.GetUCharAt(y, x*3+1)
			r := img.GetUCharAt(y, x*3+2)

			// Check if pixel is green (high green, low red and blue)
			if g > 200 && r < 100 && b < 100 {
				greenPixels++
			}
		}
	}

	greenPercentage := float64(greenPixels) / float64(totalPixels) * 100
	t.Logf("Green pixels: %d/%d (%.1f%%)", greenPixels, totalPixels, greenPercentage)

	// Check that at least 20% of pixels are green
	if greenPercentage < 20.0 {
		t.Errorf("Expected at least 20%% green pixels, got %.1f%%", greenPercentage)
	}
}

func TestListAvailableCameras(t *testing.T) {
	// Test that listAvailableCameras runs without panicking
	// This function scans cameras 0-9 and prints available ones
	// We just need to ensure it doesn't crash
	
	// Capture any potential panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("listAvailableCameras panicked: %v", r)
		}
	}()
	
	// Run the function - it should complete without error
	listAvailableCameras()
	
	// If we get here, the function completed successfully
	t.Log("listAvailableCameras completed without error")
}
