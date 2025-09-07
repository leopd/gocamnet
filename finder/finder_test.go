package finder

import (
    "image"
    "os"
    "testing"

    "gocv.io/x/gocv"
)

// TestDetectorLoadsAndRuns ensures we can load the YOLO ONNX model and run a forward pass without error.
// Assumes `make model` has downloaded models/yolov8n.onnx.
func TestDetectorLoadsAndRuns(t *testing.T) {
    if os.Getenv("RUN_DNN_TEST") == "" {
        t.Skip("skipping DNN test by default; set RUN_DNN_TEST=1 to enable")
    }
    modelPath := "models/yolov5s.onnx"
    if _, err := os.Stat(modelPath); err != nil {
        t.Skip("model not found; run `make model` first: ", modelPath)
    }

    det, err := NewDetector(Options{ ModelPath: modelPath, InputSize: image.Pt(640, 640) })
    if err != nil { t.Fatalf("NewDetector error: %v", err) }
    defer det.Close()

    // Create a blank frame; we only test that the network runs without crashing.
    frame := gocv.NewMatWithSize(480, 640, gocv.MatTypeCV8UC3)
    defer frame.Close()

    dets, err := det.Detect(frame)
    if err != nil { t.Fatalf("Detect error: %v", err) }

    // No assertion on content; just ensure it returns (possibly zero detections).
    t.Logf("detections returned: %d", len(dets))
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


