package bsat

import "testing"

// These thresholds and strict comparison are taken from
// sat_solver_set_resource_limits and sat_solver_search in satSolver.c.
func TestResourceLimitsFollowSourceCounters(t *testing.T) {
	for _, tc := range []struct {
		name                                         string
		localConf, localProp, globalConf, globalProp int64
		wantConf, wantProp                           int64
	}{
		{"unlimited", 0, 0, 0, 0, 0, 0},
		{"local", 10, 20, 0, 0, 13, 24},
		{"global tighter", 10, 20, 8, 11, 8, 11},
		{"local tighter", 10, 20, 30, 40, 13, 24},
		{"global only", 0, 0, 8, 11, 8, 11},
		{"mixed", 10, 0, 0, 11, 13, 11},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSolver()
			s.stats.conflicts, s.stats.propagations, s.stats.inspects = 3, 4, 999
			s.SetLimits(71, 72)
			if got := s.SolveWithResourceLimits(nil, tc.localConf, tc.localProp, tc.globalConf, tc.globalProp); got != Satisfiable {
				t.Fatalf("status=%v", got)
			}
			if s.nConfLimit != tc.wantConf || s.nInsLimit != tc.wantProp {
				t.Fatalf("limits=%d,%d; want %d,%d", s.nConfLimit, s.nInsLimit, tc.wantConf, tc.wantProp)
			}
			if c, p := s.Limits(); c != 71 || p != 72 {
				t.Fatalf("persistent limits changed: %d,%d", c, p)
			}
		})
	}
}

func TestResourceLimitsGlobalCapsUseStrictSourceBoundary(t *testing.T) {
	for _, tc := range []struct {
		name                                            string
		conflicts, propagations, globalConf, globalProp int64
		want                                            Status
	}{
		{"conflict equal", 3, 0, 3, 0, Satisfiable},
		{"conflict exceeded", 4, 0, 3, 0, Unknown},
		{"propagation equal", 0, 4, 0, 4, Satisfiable},
		{"propagation exceeded", 0, 5, 0, 4, Unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewSolver()
			s.stats.conflicts, s.stats.propagations = tc.conflicts, tc.propagations
			got := s.SolveWithResourceLimits(nil, 0, 0, tc.globalConf, tc.globalProp)
			if got != tc.want {
				t.Fatalf("status=%v want=%v", got, tc.want)
			}
		})
	}
}

func TestResourceLimitsRejectNegativeValues(t *testing.T) {
	for i := 0; i < 4; i++ {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("negative limit accepted")
				}
			}()
			limits := [4]int64{}
			limits[i] = -1
			NewSolver().SolveWithResourceLimits(nil, limits[0], limits[1], limits[2], limits[3])
		})
	}
}

func TestResourceLimitsAccumulateAcrossSolves(t *testing.T) {
	s := NewSolver()
	for i := 0; i < 3; i++ {
		s.NewVariable()
	}
	if got := s.SolveWithResourceLimits(nil, 0, 0, 0, 0); got != Satisfiable {
		t.Fatalf("initial status=%v", got)
	}
	used := s.Statistics().Propagations
	if used < 2 {
		t.Fatalf("fixture did not accumulate propagations: %d", used)
	}
	if got := s.SolveWithResourceLimits(nil, 0, 100, 0, used-1); got != Unknown {
		t.Fatalf("already-exceeded cumulative cap: %v", got)
	}
	if _, ok := s.Value(0); ok {
		t.Fatal("unknown result exposed a stale model")
	}
	if got := s.SolveWithResourceLimits(nil, 0, 0, 0, 0); got != Satisfiable {
		t.Fatalf("unlimited retry status=%v", got)
	}
	if _, ok := s.Value(0); !ok {
		t.Fatal("SAT retry did not expose a model")
	}
}
