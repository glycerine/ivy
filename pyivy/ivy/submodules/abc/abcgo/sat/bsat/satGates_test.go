package bsat

import "testing"

func fixedModelAssumptions(model, n int) []lit {
	out := make([]lit, n)
	for v := range out {
		sign := int32(1)
		if model&(1<<uint(v)) != 0 {
			sign = 0
		}
		out[v] = toLitCond(int32(v), sign)
	}
	return out
}
func TestSolverAndGateComplements(t *testing.T) {
	for flags := 0; flags < 8; flags++ {
		s := sat_solver_new()
		sat_solver_add_and(s, 0, 1, 2, int32(flags&1), int32((flags>>1)&1), int32((flags>>2)&1))
		for model := 0; model < 8; model++ {
			a := (model >> 1) & 1
			b := (model >> 2) & 1
			want := (a ^ (flags & 1)) & (b ^ ((flags >> 1) & 1))
			want ^= (flags >> 2) & 1
			status := sat_solver_solve(s, fixedModelAssumptions(model, 3), 0, 0, 0, 0)
			if (status == l_True) != (model&1 == want) {
				t.Fatal(flags, model, status, want)
			}
		}
	}
}
func TestSolverMuxGateComplements(t *testing.T) {
	for flags := 0; flags < 16; flags++ {
		s := sat_solver_new()
		sat_solver_add_mux(s, 0, 1, 2, 3, int32(flags&1), int32((flags>>1)&1), int32((flags>>2)&1), int32((flags>>3)&1))
		for model := 0; model < 16; model++ {
			control := ((model >> 1) & 1) ^ (flags & 1)
			then := ((model >> 2) & 1) ^ ((flags >> 1) & 1)
			otherwise := ((model >> 3) & 1) ^ ((flags >> 2) & 1)
			want := otherwise
			if control != 0 {
				want = then
			}
			want ^= (flags >> 3) & 1
			status := sat_solver_solve(s, fixedModelAssumptions(model, 4), 0, 0, 0, 0)
			if (status == l_True) != (model&1 == want) {
				t.Fatal("mux truth table", flags, model, status, want)
			}
		}
	}
}
func TestSolverXorGateComplements(t *testing.T) {
	for flag := int32(0); flag < 2; flag++ {
		s := sat_solver_new()
		sat_solver_add_xor(s, 0, 1, 2, flag)
		for model := 0; model < 8; model++ {
			want := ((model >> 1) & 1) ^ ((model >> 2) & 1) ^ int(flag)
			status := sat_solver_solve(s, fixedModelAssumptions(model, 3), 0, 0, 0, 0)
			if (status == l_True) != (model&1 == want) {
				t.Fatal(flag, model, status)
			}
		}
	}
}
