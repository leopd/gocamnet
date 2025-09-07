package pydetect

import (
    "context"
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


