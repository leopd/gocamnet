package main

import (
    "os"
    "testing"

    "gocv.io/x/gocv"
)

func TestRunSampleImage(t *testing.T) {
    os.Remove("output.jpg")
    defer os.Remove("output.jpg")

    runSampleImage()

    if _, err := os.Stat("output.jpg"); os.IsNotExist(err) {
        t.Fatal("output.jpg was not created")
    }

    img := gocv.IMRead("output.jpg", gocv.IMReadColor)
    if img.Empty() {
        t.Fatal("Failed to load output.jpg as a valid image")
    }
    defer img.Close()

    if img.Rows() != 200 || img.Cols() != 300 {
        t.Errorf("Expected image size 200x300, got %dx%d", img.Rows(), img.Cols())
    }
    if img.Channels() != 3 {
        t.Errorf("Expected 3 channels, got %d", img.Channels())
    }
}

func TestListAvailableCameras(t *testing.T) {
    defer func() {
        if r := recover(); r != nil {
            t.Errorf("listAvailableCameras panicked: %v", r)
        }
    }()
    listAvailableCameras()
}


