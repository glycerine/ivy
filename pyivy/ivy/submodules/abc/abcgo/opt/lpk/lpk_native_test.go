package lpk

import "testing"

func TestLpkLutSizing(t *testing.T) {
	for _, tc := range []struct{ luts, k, wantVars, wantLuts int32 }{
		{1, 4, 4, 1}, {2, 4, 7, 2}, {3, 6, 16, 3}, {5, 3, 11, 5},
	} {
		if got := Lpk_LutNumVars(tc.luts, tc.k); got != tc.wantVars {
			t.Errorf("Lpk_LutNumVars(%d,%d)=%d want %d", tc.luts, tc.k, got, tc.wantVars)
		}
		if got := Lpk_LutNumLuts(tc.wantVars, tc.k); got != tc.wantLuts {
			t.Errorf("Lpk_LutNumLuts(%d,%d)=%d want %d", tc.wantVars, tc.k, got, tc.wantLuts)
		}
	}
}

func TestLpkSupportDelayAndVariables(t *testing.T) {
	delays := []int32{2, 5, 3, 9}
	if got := Lpk_SuppDelay(0b1010, delays); got != 10 {
		t.Fatalf("support delay=%d want 10", got)
	}
	var vars [4]byte
	if got := Lpk_SuppToVars(0b10101, vars[:]); got != 3 {
		t.Fatalf("support variable count=%d want 3", got)
	}
	if want := [4]byte{0, 2, 4, 0}; vars != want {
		t.Fatalf("support variables=%v want %v", vars, want)
	}
}

func TestLpkFunctionTruthStorage(t *testing.T) {
	fun := Lpk_FunAlloc(6)
	if fun == nil {
		t.Fatal("Lpk_FunAlloc(6) returned nil")
	}
	base, workspace0, workspace1 := Lpk_FunTruth(fun, 0), Lpk_FunTruth(fun, 1), Lpk_FunTruth(fun, 2)
	if len(base) != 2 || len(workspace0) != 2 || len(workspace1) != 2 {
		t.Fatalf("truth table word counts = %d/%d/%d, want 2 each", len(base), len(workspace0), len(workspace1))
	}
	base[0] = 0x1234
	workspace0[0] = 0x5678
	if base[0] != 0x1234 || workspace0[0] != 0x5678 || workspace1[0] != 0 {
		t.Fatalf("truth table workspaces alias: %x/%x/%x", base[0], workspace0[0], workspace1[0])
	}
}
