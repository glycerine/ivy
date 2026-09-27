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

func raftFixture(t *testing.T, name string) (string, []byte) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "test_vectors", name))
	if err != nil {
		t.Fatalf("Abs: %v", err)
	}
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		t.Skipf("raft fixture not present at %s", path)
	}
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	return path, content
}

func TestRaftWrapperInductionUsesAbstractIncludedProofModel(t *testing.T) {
	path, content := raftFixture(t, "raft_no_assume_test.ivy")

	sess := NewSession(goivy.NewConfig(), "raft-wrapper-induction-module-regression")
	if err := sess.LoadFileContent(path, content); err != nil {
		t.Fatalf("LoadFileContent: %v", err)
	}
	if sess.CompiledModule == nil || sess.CompiledModule.Sig == nil {
		t.Fatal("CompiledModule should be populated")
	}
	if sess.InductionModule == nil || sess.InductionModule.Sig == nil {
		t.Fatal("InductionModule should be populated")
	}
	if _, ok := sess.CompiledModule.Sig.Interp["node"]; !ok {
		t.Fatal("concrete wrapper fixture should interpret node for randomized testing")
	}
	if _, ok := sess.InductionModule.Sig.Interp["node"]; ok {
		t.Fatal("induction should use the included abstract Raft proof model, not the wrapper's finite node interpretation")
	}
	if got, want := len(sess.InductionModule.LabeledConjs), len(sess.CompiledModule.LabeledConjs); got != want {
		t.Fatalf("induction conjectures = %d, concrete conjectures = %d", got, want)
	}
}

func TestRaftWrapperInductionDoesNotStallOnLogMatching(t *testing.T) {
	path, content := raftFixture(t, "raft_no_assume_test.ivy")

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
	reachedConjecture22 := false
	overall := time.After(45 * time.Second)
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
				reachedConjecture22 = true
				heartbeatDeadline = nil
				if !heartbeatSeen && time.Since(logMatchingAt) > 3*time.Second {
					t.Fatalf("log_matching advanced without heartbeat only after %s", time.Since(logMatchingAt))
				}
				continue
			}
		case result := <-resultCh:
			if logMatchingAt.IsZero() {
				t.Fatalf("check ended before reaching log_matching: %+v", result)
			}
			if !reachedConjecture22 {
				t.Fatalf("check ended after log_matching but before conjecture 22: %+v", result)
			}
			if result.Result != "pass" {
				t.Fatalf("induction result = %q, want pass: %s", result.Result, result.Message)
			}
			return
		case <-heartbeatDeadline:
			cancel()
			t.Fatal("timed out waiting for a log_matching heartbeat")
		case <-overall:
			cancel()
			t.Fatal("timed out waiting for induction progress past log_matching")
		}
	}
}
