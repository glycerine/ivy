//go:build web

package webui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	goivy "github.com/glycerine/ivy/goivy"
)

func TestRaftWrapperInductionDoesNotStallOnLogMatching(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir: %v", err)
	}
	path := filepath.Join(home, "ivy", "ivy-lang-examples", "examples", "raft", "raft_no_assume_test.ivy")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("raft wrapper not present at %s", path)
	}
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	sess := NewSession(goivy.NewConfig(), "raft-wrapper-log-matching-regression")
	progress := make(chan string, 256)
	sess.SetEventSink(func(event Event) {
		if event.Type != "check_progress" {
			return
		}
		data, ok := event.Data.(map[string]interface{})
		if !ok {
			return
		}
		msg, _ := data["message"].(string)
		if msg == "" {
			return
		}
		select {
		case progress <- msg:
		default:
		}
	})

	if err := sess.LoadFileContent(path, content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultCh := make(chan *WebUICheckResult, 1)
	go func() {
		resultCh <- sess.RunCheckWithOptions("induction", CheckOptions{Context: ctx})
	}()

	var logMatchingAt time.Time
	var heartbeatDeadline <-chan time.Time
	heartbeatSeen := false
	overall := time.After(30 * time.Second)
	for {
		select {
		case msg := <-progress:
			t.Log(msg)
			if strings.Contains(msg, "Checking conjecture 21 of 46: log_matching") {
				logMatchingAt = time.Now()
				heartbeatDeadline = time.After(6 * time.Second)
				continue
			}
			if !logMatchingAt.IsZero() && strings.Contains(msg, "Still checking conjecture 21 of 46: log_matching") {
				heartbeatSeen = true
				heartbeatDeadline = nil
				continue
			}
			if !logMatchingAt.IsZero() && strings.Contains(msg, "Checking conjecture 22 of 46:") {
				cancel()
				select {
				case <-resultCh:
				case <-time.After(5 * time.Second):
					t.Fatal("check did not stop promptly after cancellation")
				}
				if !heartbeatSeen && time.Since(logMatchingAt) > 3*time.Second {
					t.Fatalf("log_matching advanced without heartbeat only after %s", time.Since(logMatchingAt))
				}
				return
			}
		case result := <-resultCh:
			if logMatchingAt.IsZero() {
				t.Fatalf("check ended before reaching log_matching: %+v", result)
			}
			t.Fatalf("check ended after log_matching but before conjecture 22: %+v", result)
		case <-heartbeatDeadline:
			cancel()
			t.Fatal("timed out waiting for a log_matching heartbeat")
		case <-overall:
			cancel()
			t.Fatal("timed out waiting for induction progress past log_matching")
		}
	}
}
