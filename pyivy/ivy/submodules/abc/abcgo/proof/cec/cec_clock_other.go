//go:build !linux && !cgo

package cec

// ABC calls the platform C clock() on non-Linux targets. Without cgo there
// is no verified equivalent to that runtime's clock, so fail explicitly.
func cecCPUClock() int64        { panic("cec: source C clock requires cgo on this platform") }
func cecPinClockThread() func() { return func() {} }
