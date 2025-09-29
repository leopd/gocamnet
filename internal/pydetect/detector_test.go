package pydetect

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"gocv.io/x/gocv"
)

// TestIPCDetectMock spins up the python server in mock mode and verifies a basic roundtrip.
func TestIPCDetectMock(t *testing.T) {
    t.Setenv("PYDETECT_MOCK", "1")

    // project root: repo root two dirs up from this file
    _, file, _, _ := runtime.Caller(0)
    root := filepath.Join(filepath.Dir(file), "..", "..")

    det, err := New(context.Background(), root, 0.25)
    if err != nil {
        t.Fatalf("failed to start detector: %v", err)
    }
    defer det.Close()

    img := gocv.NewMatWithSize(64, 64, gocv.MatTypeCV8UC3)
    defer img.Close()
    img.SetTo(gocv.NewScalar(0, 0, 0, 0))

    time.Sleep(100 * time.Millisecond)

    dets, err := det.Detect(img, 1)
    if err != nil {
        t.Fatalf("detect error: %v", err)
    }
	if len(dets) == 0 {
		t.Fatalf("expected mock detection, got 0")
	}
}

// TestPersonDetectionWithFixture tests detection on the soccer-people.jpeg fixture
func TestPersonDetectionWithFixture(t *testing.T) {
    if os.Getenv("PYDETECT_RUN_REAL") != "1" {
        t.Skip("Set PYDETECT_RUN_REAL=1 to exercise the real YOLO pipeline")
    }
    if os.Getenv("PYDETECT_MOCK") == "1" {
        t.Skip("Skipping real detection test in mock mode")
    }

	// project root: repo root two dirs up from this file
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(file), "..", "..")

	det, err := New(context.Background(), root, 0.2) // Lower threshold for fixture
	if err != nil {
		t.Fatalf("failed to start detector: %v", err)
	}
	defer det.Close()

	// Load the fixture image
	fixturePath := filepath.Join(root, "fixtures", "soccer-people.jpeg")
	img := gocv.IMRead(fixturePath, gocv.IMReadColor)
	if img.Empty() {
		t.Fatalf("failed to load fixture image: %s", fixturePath)
	}
	defer img.Close()

	time.Sleep(100 * time.Millisecond) // Allow server to be ready

	detections, err := det.Detect(img, 1)
	if err != nil {
		t.Fatalf("detect error: %v", err)
	}

	// Count person detections (class_id == 0 or class_name == "person")
	personCount := 0
	for _, d := range detections {
		if d.ClassID == 0 || d.ClassName == "person" {
			personCount++
			t.Logf("Found person: confidence=%.3f, box=%v", d.Score, d.Box)
		}
	}

	if personCount != 3 {
		t.Errorf("Expected 3 people, found %d", personCount)
	} else {
		t.Logf("Successfully detected %d people in fixture image", personCount)
	}
}


