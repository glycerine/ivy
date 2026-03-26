//go:build xtracer_off

// When the xtracer_off build tag IS set, all trace calls are no-ops.
// The compiler inlines empty functions, so there is zero runtime cost.
package xtracer

// Enabled is false when xtracer is disabled via build tag.
var Enabled = false

// HashVerbose is unused when xtracer is disabled.
var HashVerbose = false

// Trace is a no-op when xtracer is disabled.
func Trace(format string, args ...interface{}) {}
