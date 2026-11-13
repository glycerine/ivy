package bsat

import "testing"

func TestGlobalVariableBiasOwnershipAndValidation(t *testing.T) {
	s := NewSolver()
	s.NewVariable()
	s.NewVariable()
	input := []int32{1}
	s.SetGlobalVariableBias(input)
	input[0] = 0
	if s.pGlobalVars[0] != 0 || s.pGlobalVars[1] != 1 {
		t.Fatal("bias aliases input")
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Error("invalid bias accepted")
			}
		}()
		s.SetGlobalVariableBias([]int32{0, 2})
	}()
	if s.pGlobalVars[0] != 0 || s.pGlobalVars[1] != 1 {
		t.Fatal("invalid update changed bias")
	}
	s.SetGlobalVariableBias(nil)
	if s.pGlobalVars != nil {
		t.Fatal("bias not disabled")
	}
}
