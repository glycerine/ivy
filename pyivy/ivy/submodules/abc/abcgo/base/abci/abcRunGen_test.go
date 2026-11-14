package abci

import (
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
)

func TestAcbNtkRunGenMatchesEmptyCStub(t *testing.T) {
	// C's Acb_NtkRunGen body is empty. Exercise the boundary shapes accepted
	// by the Go ABI and ensure no parsing, I/O, or network access is invented.
	Acb_NtkRunGen(nil, 0)
	Acb_NtkRunGen([]string{}, 1)
	Acb_NtkRunGen([]string{"first", "second"}, 1)
}

func TestAbcCommandRunGenDelegatesWithoutChangingFrameState(t *testing.T) {
	frame := abcmain.NewFrame()
	state := abcmain.ReadABCFrameState(frame)
	sentinel := &struct{ name string }{"gia state"}
	state.SetGia(sentinel)

	if got := Abc_CommandRunGen(frame, 3, []string{"rungen", "input.blif", "output.blif"}); got != 0 {
		t.Fatalf("valid rungen command returned %d, want 0", got)
	}
	if got := state.Gia(); got != sentinel {
		t.Fatalf("rungen changed Gia state to %v, want original sentinel", got)
	}
	if got := Abc_FrameReadNtk(frame); got != nil {
		t.Fatalf("rungen installed a network %p despite the empty C stub", got)
	}
}

func TestAbcCommandRunGenHelpUsesUsageStatus(t *testing.T) {
	if got := Abc_CommandRunGen(nil, 2, []string{"rungen", "-h"}); got != 1 {
		t.Fatalf("rungen help returned %d, want usage status 1", got)
	}
}
