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

func TestSolverAdditionalGateTruthTables(t *testing.T) {
	for flag := int32(0); flag < 2; flag++ {
		for _, enabled := range []bool{false, true} {
			s := sat_solver_new()
			if enabled {
				sat_solver_add_buffer_enable(s, 0, 1, 2, flag)
			} else {
				sat_solver_add_buffer(s, 0, 1, flag)
			}
			n := 2
			if enabled {
				n = 3
			}
			for model := 0; model < 1<<n; model++ {
				want := (model&1 != 0) == ((model&2 != 0) != (flag != 0))
				if enabled && model&4 == 0 {
					want = true
				}
				status := sat_solver_solve(s, fixedModelAssumptions(model, n), 0, 0, 0, 0)
				if (status == l_True) != want {
					t.Fatalf("buffer enabled=%v flag=%d model=%03b status=%v", enabled, flag, model, status)
				}
			}
		}
	}
	for _, test := range []struct {
		name  string
		add   func(*sat_solver)
		valid func(int) bool
	}{
		{"xor-and", func(s *sat_solver) { sat_solver_add_xor_and(s, 0, 1, 2, 3) }, func(m int) bool {
			return (m&1 != 0) == (((m&2 != 0) != (m&4 != 0)) && m&8 != 0)
		}},
		{"half-sorter", func(s *sat_solver) { sat_solver_add_half_sorter(s, 0, 1, 2, 3) }, func(m int) bool {
			return (m&1 != 0 || (m&4 == 0 && m&8 == 0)) && (m&2 != 0 || m&4 == 0 || m&8 == 0)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := sat_solver_new()
			test.add(s)
			for model := 0; model < 16; model++ {
				status := sat_solver_solve(s, fixedModelAssumptions(model, 4), 0, 0, 0, 0)
				if (status == l_True) != test.valid(model) {
					t.Fatalf("model=%04b status=%v", model, status)
				}
			}
		})
	}
}
