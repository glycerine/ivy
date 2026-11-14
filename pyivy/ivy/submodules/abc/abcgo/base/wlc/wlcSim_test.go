package wlc

import "testing"

func TestWlcNtkSimPiUsesUnshiftedRandomWords(t *testing.T) {
	manager := Gia_ManStart(1)
	defer Gia_ManStop(manager)
	Gia_ManSetSimData(manager, Vec_WrdStart(2), 2)
	Gia_ManRandomW(1)
	want := []word{Gia_ManRandomW(0), Gia_ManRandomW(0)}
	Gia_ManRandomW(1)
	wlcNtkSimPi(manager, 0)
	got := Gia_ManSimObjectData(manager, 0)
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("word %d = %#x, want %#x", i, got[i], want[i])
		}
	}
}
