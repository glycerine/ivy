package abs

import (
	"runtime"
	"testing"
	"time"
)

func waitAbsProof(t *testing.T, c *absProverControl) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if c.check() == 1 {
			return
		}
		runtime.Gosched()
	}
	t.Fatal("proof was not published")
}

func TestAbsProverProofIsConsumedOnce(t *testing.T) {
	var c absProverControl
	c.start(func(int32) int32 { return 1 })
	waitAbsProof(t, &c)
	if c.check() != 0 {
		t.Fatal("proof was published twice")
	}
}

// ABC's pthread worker publishes a successful result even if its run ID was
// superseded while solving. This characterizes that source behavior.
func TestAbsProverSupersededProofCanPublish(t *testing.T) {
	var c absProverControl
	started, release := make(chan struct{}), make(chan struct{})
	old := c.start(func(int32) int32 { close(started); <-release; return 1 })
	<-started
	current := c.start(func(int32) int32 { return 0 })
	if !c.stopped(old) || c.stopped(current) {
		t.Fatal("incorrect cancellation state")
	}
	close(release)
	waitAbsProof(t, &c)
}

func TestAbsProverCancelInvalidatesRun(t *testing.T) {
	var c absProverControl
	started, release := make(chan struct{}), make(chan struct{})
	id := c.start(func(int32) int32 { close(started); <-release; return 1 })
	<-started
	c.cancel()
	if !c.stopped(id) {
		t.Fatal("cancel did not stop run")
	}
	close(release)
}

func TestAbsProverCancelKeepsPublishedProof(t *testing.T) {
	var c absProverControl
	c.start(func(int32) int32 { return 1 })
	deadline := time.Now().Add(time.Second)
	for {
		c.mu.Lock()
		proved := c.proved
		c.mu.Unlock()
		if proved {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("proof was not published")
		}
		runtime.Gosched()
	}
	c.cancel()
	if c.check() != 1 || c.check() != 0 {
		t.Fatal("C oracle expects the published proof once after cancellation")
	}
}
