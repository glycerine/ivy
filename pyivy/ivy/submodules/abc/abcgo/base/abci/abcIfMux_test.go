package abci

import "testing"

func TestMuxCofactorsAndSupportMatchTruthAssignments(t *testing.T) {
	for truth := int32(0); truth < 256; truth++ {
		for v := int32(0); v < 3; v++ {
			var cof0, cof1 int32
			for assignment := int32(0); assignment < 8; assignment++ {
				bit := int32(1) << assignment
				if truth&(int32(1)<<(assignment&^(1<<v))) != 0 {
					cof0 |= bit
				}
				if truth&(int32(1)<<(assignment|(1<<v))) != 0 {
					cof1 |= bit
				}
			}
			if got := Abc_NtkFuncCof0(truth, v); got != cof0 {
				t.Fatalf("cof0(%02x,%d)=%02x want %02x", truth, v, got, cof0)
			}
			if got := Abc_NtkFuncCof1(truth, v); got != cof1 {
				t.Fatalf("cof1(%02x,%d)=%02x want %02x", truth, v, got, cof1)
			}
		}
		wantSupport := int32(0)
		for v := int32(0); v < 3; v++ {
			var changes bool
			for assignment := int32(0); assignment < 8; assignment++ {
				if assignment&(1<<v) != 0 {
					continue
				}
				a, b := truth&(int32(1)<<assignment) != 0, truth&(int32(1)<<(assignment|(1<<v))) != 0
				changes = changes || a != b
			}
			if changes {
				wantSupport++
			}
		}
		if got := Abc_NtkFuncSuppSize(truth); got != wantSupport {
			t.Fatalf("support(%02x)=%d want %d", truth, got, wantSupport)
		}
		wantMux := uint8(0)
		if wantSupport < 3 {
			wantMux = 1
		} else {
			for v := int32(0); v < 3; v++ {
				if Abc_NtkFuncSuppSize(Abc_NtkFuncCof0(truth, v)) == 1 && Abc_NtkFuncSuppSize(Abc_NtkFuncCof1(truth, v)) == 1 {
					wantMux = 1
					break
				}
			}
		}
		if abcNtkCutCostMuxLookup[truth] != wantMux {
			t.Fatalf("lookup[%02x]=%d want %d", truth, abcNtkCutCostMuxLookup[truth], wantMux)
		}
	}
}
