package util

import (
	"strings"
	"testing"
)

func assertingCaller() { assert(1+1 == 3, "arithmetic") }

// A failed assertion names the file and line of the assert call.
func TestAssertReportsCallSite(t *testing.T) {
	defer func() {
		msg, _ := recover().(string)
		if !strings.HasPrefix(msg, "assert_test.go:8: ") || !strings.Contains(msg, "arithmetic") {
			t.Fatalf("message %q", msg)
		}
	}()
	assertingCaller()
}
