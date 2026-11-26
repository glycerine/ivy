package abcmain

import "testing"

// ABC keeps one verification status per frame, starting at -1 (unknown).
// The command-state view must read and write that same status.
func TestFrameStateSharesVerificationStatus(t *testing.T) {
	f := isolatedFrame(t)
	s := ReadABCFrameState(f)
	if status, frames := s.VerificationStatus(); status != -1 || frames != -1 {
		t.Fatalf("fresh frame status %d/%d, want -1/-1", status, frames)
	}
	s.SetVerificationStatus(1, 3)
	if Abc_FrameReadProbStatus(f) != 1 || f.NFrames != 3 {
		t.Fatalf("frame status %d/%d after state update", Abc_FrameReadProbStatus(f), f.NFrames)
	}
	Abc_FrameSetStatus(0)
	if status, _ := s.VerificationStatus(); status != 0 {
		t.Fatalf("state status %d after frame update", status)
	}
}
