package finder

import (
    "image"
    "os"
    "path/filepath"
    "testing"

    "gocv.io/x/gocv"
)

// TestDetectorLoadsAndRuns ensures we can load a DNN model and run a forward pass without error.
// NOTE: This test is skipped by default as it requires manually downloaded model files.
//
// To run this test:
// 1. Download "MobileNetSSD_deploy.prototxt" and "MobileNetSSD_deploy.caffemodel".
// 2. A reliable source is: https://github.com/chuanqi305/MobileNet-SSD
// 3. Place both files in the "models/" directory at the root of this repository.
// 4. Run `make test`.
func TestDetectorLoadsAndRuns(t *testing.T) {
	// Tests in this package run with working directory set to `finder/`.
	// The models are stored at repo root `models/`. Resolve via parent.
	protoPath := filepath.Join("..", "models", "MobileNetSSD_deploy.prototxt")
	modelPath := filepath.Join("..", "models", "MobileNetSSD_deploy.caffemodel")

	if _, err := os.Stat(protoPath); err != nil {
		t.Fatalf("DNN test failed: Caffe prototxt not found. Please download it and place it in models/: %v", err)
	}
	if _, err := os.Stat(modelPath); err != nil {
		t.Fatalf("DNN test failed: Caffe model not found. Please download it and place it in models/: %v", err)
	}

	det, err := NewDetector(Options{
		ModelProto:   protoPath,
		ModelWeights: modelPath,
		InputSize:    image.Pt(300, 300), // MobileNet-SSD default
	})
	if err != nil {
		t.Fatalf("NewDetector error: %v", err)
	}
	defer det.Close()

	// Create a blank frame; we only test that the network loads and runs without crashing.
	frame := gocv.NewMatWithSize(300, 300, gocv.MatTypeCV8UC3)
	defer frame.Close()

	detections, err := det.Detect(frame)
	if err != nil {
		t.Fatalf("Detect error: %v", err)
	}

	// No assertion on content; just ensure it returns (possibly zero detections).
	t.Logf("detections returned: %d", len(detections))
}

// TestBasicOpenCVWorks tests that basic OpenCV functionality works without DNN
func TestBasicOpenCVWorks(t *testing.T) {
    // Test that basic OpenCV functionality works without DNN
    img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)
    defer img.Close()
    
    if img.Empty() {
        t.Fatal("failed to create OpenCV Mat")
    }
    
    // Test basic operations
    img.SetTo(gocv.NewScalar(255, 0, 0, 0)) // Blue image
    
    // Convert to Go image
    goImg, err := img.ToImage()
    if err != nil {
        t.Fatalf("failed to convert Mat to Go image: %v", err)
    }
    if goImg == nil {
        t.Fatal("failed to convert Mat to Go image")
    }
    
    t.Log("Basic OpenCV functionality works")
}


