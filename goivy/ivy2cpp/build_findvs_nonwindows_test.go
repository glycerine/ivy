//go:build !windows

package ivy2cpp

import (
	"strings"
	"testing"
)

func TestFindVsLinuxStubIsNoOp(t *testing.T) {
	_, err := defaultFindVS()
	if err == nil || !strings.Contains(err.Error(), "vswhere unavailable on non-Windows") {
		t.Fatalf("defaultFindVS() error = %v, want non-Windows vswhere diagnostic", err)
	}
}
