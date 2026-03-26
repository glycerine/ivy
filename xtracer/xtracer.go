//go:build !xtracer_off

// Package xtracer provides execution tracing for parallel conformance
// testing between Go goivy and Python ivy. Trace() prints timestamped
// messages to stdout by default, matching Python's `if __debug__:` guards.
//
// To disable at build time: go build -tags xtracer_off ./cmd/goivy_check/
// To disable at runtime:    XTRACE_OFF=1 go test ./...
//
// Both Go and Python emit the same format:
//
//	XTRACE: <category>.<function> <ENTER|EXIT> [<detail>]
//
// so that `diff` on the two outputs reveals the first divergence.
package xtracer

import (
	"fmt"
	"os"
)

// Enabled is true when xtracer is active (the default).
// Set XTRACE_OFF=1 to suppress output at runtime.
const Enabled = true

// HashVerbose causes HASH trace lines to include
// the full canonical string.
const HashVerbose bool = true

// suppressed is set by XTRACE_OFF=1 to silence output without
// changing Enabled (so `if xtracer.Enabled` guards still compile away).
var suppressed bool

func init() {
	if os.Getenv("XTRACE_OFF") == "1" {
		suppressed = true
	}
}

// Trace prints an execution trace line to stdout.
// Format: "XTRACE: " + fmt.Sprintf(format, args...) + "\n"
func Trace(format string, args ...interface{}) {
	if suppressed {
		return
	}
	// replace true/false with True/False to match python
	// and avoid spurious diffs.
	for i, a := range args {
		switch b := a.(type) {
		case bool:
			if b {
				args[i] = "True"
			} else {
				args[i] = "False"
			}
		}
	}
	fmt.Printf("XTRACE: "+format+"\n", args...)
}
