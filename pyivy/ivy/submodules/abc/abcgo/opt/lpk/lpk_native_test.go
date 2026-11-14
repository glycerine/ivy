package lpk

import (
	"testing"

	"github.com/glycerine/abcgo/bool/kit"
)

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

func TestLpkFunctionSupportMinimizeCompactsFanins(t *testing.T) {
	fun := Lpk_FunAlloc(3)
	fun.nVars = 3
	fun.pFanins = [16]byte{4, 7, 9}
	fun.pDelays = [16]int32{2, 3, 8}
	// The function is x2, whose truth table repeats 00001111.
	Lpk_FunTruth(fun, 0)[0] = 0xF0F0F0F0
	fun.uSupp = kit.Kit_TruthSupport(Lpk_FunTruth(fun, 0), 3)

	if got := Lpk_FunSuppMinimize(fun); got != 1 {
		t.Fatalf("support minimization = %d, want 1", got)
	}
	if fun.nVars != 1 || fun.uSupp != 1 {
		t.Fatalf("minimized vars/support = %d/%b, want 1/1", fun.nVars, fun.uSupp)
	}
	if got := Lpk_FunTruth(fun, 0)[0]; got != 0xAAAAAAAA {
		t.Fatalf("minimized truth = %#08x, want variable truth %#08x", got, uint32(0xAAAAAAAA))
	}
	if fun.pFanins[0] != 9 || fun.pDelays[0] != 8 {
		t.Fatalf("retained fanin metadata = %d/%d, want 9/8", fun.pFanins[0], fun.pDelays[0])
	}
}

func TestLpkFunctionCofactorSupports(t *testing.T) {
	fun := Lpk_FunAlloc(2)
	fun.nVars = 2
	// XOR: each cofactor of x0 is x1.
	Lpk_FunTruth(fun, 0)[0] = 0x66666666
	Lpk_FunComputeCofSupps(fun)
	if fun.fSupports == 0 || fun.puSupps[0] != 2 || fun.puSupps[1] != 2 {
		t.Fatalf("x0 cofactor supports = %b/%b (computed=%d), want 10/10", fun.puSupps[0], fun.puSupps[1], fun.fSupports)
	}
}

func TestLpkPackTruthWords(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []uint32
		want []uint64
	}{
		{"empty", nil, []uint64{}},
		{"one half", []uint32{0x89abcdef}, []uint64{0x0000000089abcdef}},
		{"paired halves", []uint32{0x01234567, 0x89abcdef}, []uint64{0x89abcdef01234567}},
		{"odd tail", []uint32{1, 2, 3}, []uint64{0x0000000200000001, 3}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := lpkPackTruthWords(tc.in); !equalUint64s(got, tc.want) {
				t.Fatalf("lpkPackTruthWords(%08x) = %016x, want %016x", tc.in, got, tc.want)
			}
		})
	}
}

func equalUint64s(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
