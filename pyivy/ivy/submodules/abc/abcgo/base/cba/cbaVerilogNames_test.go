package cba

import "testing"

func TestSourceVerilogNameLegality(t *testing.T) {
	for _, tc := range []struct {
		name     string
		id, want int32
	}{
		{"wire", 12, 0}, {"wire", 13, 1}, {"_x9$", 13, 1}, {"9wire", 13, 0},
		{"a-b", 13, 0}, {"\\escaped name", 1, 1}, {"a\x00-b", 13, 1},
	} {
		for _, check := range []func(string, int32) int32{Cba_NameIsLegalInVerilog, Prs_NameIsLegalInVerilog} {
			if got := check(tc.name, tc.id); got != tc.want {
				t.Errorf("%q id=%d => %d want %d", tc.name, tc.id, got, tc.want)
			}
		}
	}
}
func TestSourceVerilogNameEmptyAsserts(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected source empty-name assertion")
		}
	}()
	Cba_NameIsLegalInVerilog("", 13)
}
