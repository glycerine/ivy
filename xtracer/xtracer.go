//go:build !xtracer_off

// Package xtracer provides execution tracing for parallel conformance
// testing between Go goivy and Python ivy. Trace() prints timestamped
// messages to stdout by default, matching Python's `if __debug__:` guards.
// To disable tracing, build with the "xtracer_off" build tag:
//
//	go build -tags xtracer_off ./cmd/goivy_check/
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

var _ = os.Getenv

// Enabled is true when xtracer is active (the default).
var Enabled = true

// HashVerbose causes HASH trace lines to include
// the full canonical string.
// Set by XTRACE_HASH_VERBOSE=1 environment variable.
// Update: always true now.
var HashVerbose bool = true

func init() {
	//HashVerbose = os.Getenv("XTRACE_HASH_VERBOSE") == "1"
}

// Trace prints an execution trace line to stdout.
// Format: "XTRACE: " + fmt.Sprintf(format, args...) + "\n"
func Trace(format string, args ...interface{}) {
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
