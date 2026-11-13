//go:build !linux && !windows && cgo

package cec

import (
	"testing"
	"time"
)

func TestSourceCClockDoesNotChargeSleep(t *testing.T) {
	before := cecCPUClock()
	time.Sleep(60 * time.Millisecond)
	elapsed := cecCPUClock() - before
	if elapsed < 0 || elapsed >= int64(40*time.Millisecond) {
		t.Fatalf("process CPU clock advanced %s during sleep", time.Duration(elapsed))
	}
}

func TestSourceCClockIncludesOtherThreadCPU(t *testing.T) {
	stop := make(chan struct{})
	done := make(chan struct{})
	ready := make(chan struct{})
	go func() {
		close(ready)
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
		}
	}()
	<-ready
	before := cecCPUClock()
	time.Sleep(80 * time.Millisecond)
	elapsed := cecCPUClock() - before
	close(stop)
	<-done
	if elapsed < int64(5*time.Millisecond) {
		t.Fatalf("process CPU clock missed worker CPU: %s", time.Duration(elapsed))
	}
}
