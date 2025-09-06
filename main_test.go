package main

import (
	"testing"
	"time"
)

func TestRunCountdown(t *testing.T) {
	// Test with 2 seconds
	seconds := 2
	
	// Capture start time
	start := time.Now()
	
	// Run the countdown
	runCountdown(seconds)
	
	// Calculate elapsed time
	elapsed := time.Since(start)
	
	// Check that elapsed time is between 1.9 and 2.1 seconds
	// For 2 seconds: count 1, sleep 1s, count 2, sleep 1s = ~2 seconds total
	expectedMin := 1900 * time.Millisecond
	expectedMax := 2100 * time.Millisecond
	
	if elapsed < expectedMin {
		t.Errorf("Countdown took %v, expected at least %v", elapsed, expectedMin)
	}
	
	if elapsed > expectedMax {
		t.Errorf("Countdown took %v, expected at most %v", elapsed, expectedMax)
	}
	
	t.Logf("Countdown completed in %v (expected ~2 seconds)", elapsed)
}
