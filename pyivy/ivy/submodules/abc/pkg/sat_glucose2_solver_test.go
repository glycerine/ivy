package abc

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

type gluco2TestClause []int

// bruteSat decides a CNF over n variables by enumeration.
func gluco2BruteSat(n int, cls []gluco2TestClause) bool {
	for m := 0; m < 1<<n; m++ {
		ok := true
		for _, c := range cls {
			sat := false
			for _, l := range c {
				v := l
				if v < 0 {
					v = -v
				}
				val := (m>>(v-1))&1 == 1
				if (l > 0) == val {
					sat = true
					break
				}
			}
			if !sat {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func gluco2TestLit(l int) Gluco2Lit {
	if l > 0 {
		return Gluco2MkLit(l-1, false)
	}
	return Gluco2MkLit(-l-1, true)
}

func TestGluco2SolversAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for iter := 0; iter < 300; iter++ {
		n := 3 + rng.Intn(8)
		m := 1 + rng.Intn(5*n)
		var cls []gluco2TestClause
		for i := 0; i < m; i++ {
			k := 1 + rng.Intn(3)
			var c gluco2TestClause
			for j := 0; j < k; j++ {
				v := 1 + rng.Intn(n)
				if rng.Intn(2) == 0 {
					v = -v
				}
				c = append(c, v)
			}
			cls = append(cls, c)
		}
		want := gluco2BruteSat(n, cls)

		s := NewGluco2Solver()
		ss := NewGluco2SimpSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			ss.NewVar(true, true)
		}
		okS, okSS := true, true
		for _, c := range cls {
			var ps, ps2 Gluco2Vec[Gluco2Lit]
			for _, l := range c {
				ps.PushElem(gluco2TestLit(l))
				ps2.PushElem(gluco2TestLit(l))
			}
			okS = s.AddClause_(&ps) && okS
			okSS = ss.AddClause_(&ps2) && okSS
		}
		gotS := okS && s.SolveNoAssumps()
		gotSS := okSS && ss.SolveNoAssumps(true, false)
		if gotS != want || gotSS != want {
			t.Fatalf("iter %d: n=%d cls=%v want %v, Solver %v, SimpSolver %v", iter, n, cls, want, gotS, gotSS)
		}
		if want {
			for _, sol := range []struct {
				name  string
				model func(int) Gluco2Lbool
			}{{"Solver", func(v int) Gluco2Lbool { return s.ModelValue(v) }}, {"SimpSolver", func(v int) Gluco2Lbool { return ss.ModelValue(v) }}} {
				for _, c := range cls {
					sat := false
					for _, l := range c {
						v := l
						if v < 0 {
							v = -v
						}
						mv := sol.model(v - 1)
						if (l > 0 && mv == Gluco2LTrue) || (l < 0 && mv == Gluco2LFalse) {
							sat = true
						}
					}
					if !sat {
						t.Fatalf("iter %d: %s model violates clause %v", iter, sol.name, c)
					}
				}
			}
		}
	}
}

func TestGluco2PigeonholeAndDimacs(t *testing.T) {
	// PHP(5,4): 5 pigeons, 4 holes - unsatisfiable.
	dir := t.TempDir()
	path := filepath.Join(dir, "php.cnf")
	pig, hol := 5, 4
	v := func(p, h int) int { return p*hol + h + 1 }
	var body string
	nc := 0
	for p := 0; p < pig; p++ {
		line := ""
		for h := 0; h < hol; h++ {
			line += itoaGluco2(v(p, h)) + " "
		}
		body += line + "0\n"
		nc++
	}
	for h := 0; h < hol; h++ {
		for p1 := 0; p1 < pig; p1++ {
			for p2 := p1 + 1; p2 < pig; p2++ {
				body += "-" + itoaGluco2(v(p1, h)) + " -" + itoaGluco2(v(p2, h)) + " 0\n"
				nc++
			}
		}
	}
	hdr := "c pigeonhole\np cnf " + itoaGluco2(pig*hol) + " " + itoaGluco2(nc) + "\n"
	if err := os.WriteFile(path, []byte(hdr+body), 0o644); err != nil {
		t.Fatal(err)
	}
	in, err := OpenGZIPReader(path, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := NewGluco2Solver()
	gluco2ParseDIMACS(in, s)
	in.Close()
	if s.NVars() != pig*hol || s.NClauses() == 0 {
		t.Fatalf("parsed %d vars, %d clauses", s.NVars(), s.NClauses())
	}
	if s.SolveNoAssumps() {
		t.Fatal("PHP(5,4) reported satisfiable")
	}
}

func itoaGluco2(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestGluco2LargerRandom cross-checks Solver against SimpSolver on instances big
// enough to exercise restarts, learnt clause reduction and garbage collection.
func TestGluco2LargerRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for iter := 0; iter < 40; iter++ {
		n := 60 + rng.Intn(40)
		m := int(float64(n) * 4.3)
		s := NewGluco2Solver()
		ss := NewGluco2SimpSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			ss.NewVar(true, true)
		}
		var cls []gluco2TestClause
		okS, okSS := true, true
		for i := 0; i < m; i++ {
			var c gluco2TestClause
			var ps, ps2 Gluco2Vec[Gluco2Lit]
			for j := 0; j < 3; j++ {
				v := 1 + rng.Intn(n)
				if rng.Intn(2) == 0 {
					v = -v
				}
				c = append(c, v)
				ps.PushElem(gluco2TestLit(v))
				ps2.PushElem(gluco2TestLit(v))
			}
			cls = append(cls, c)
			okS = s.AddClause_(&ps) && okS
			okSS = ss.AddClause_(&ps2) && okSS
		}
		a := okS && s.SolveNoAssumps()
		b := okSS && ss.SolveNoAssumps(true, false)
		if a != b {
			t.Fatalf("iter %d: Solver %v, SimpSolver %v", iter, a, b)
		}
		if a {
			for _, c := range cls {
				sat := false
				for _, l := range c {
					v := l
					if v < 0 {
						v = -v
					}
					mv := s.ModelValue(v - 1)
					if (l > 0 && mv == Gluco2LTrue) || (l < 0 && mv == Gluco2LFalse) {
						sat = true
					}
				}
				if !sat {
					t.Fatalf("iter %d: model violates %v", iter, c)
				}
			}
		}
	}
}
