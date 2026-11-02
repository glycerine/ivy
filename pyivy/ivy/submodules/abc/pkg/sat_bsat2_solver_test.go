package abc

import (
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

type minisatTestClause []int

// bruteSat decides a CNF over n variables by enumeration.
func minisatBruteSat(n int, cls []minisatTestClause) bool {
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

func minisatTestLit(l int) MinisatLit {
	if l > 0 {
		return MinisatMkLit(l-1, false)
	}
	return MinisatMkLit(-l-1, true)
}

func TestMinisatSolversAgainstBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for iter := 0; iter < 300; iter++ {
		n := 3 + rng.Intn(8)
		m := 1 + rng.Intn(5*n)
		var cls []minisatTestClause
		for i := 0; i < m; i++ {
			k := 1 + rng.Intn(3)
			var c minisatTestClause
			for j := 0; j < k; j++ {
				v := 1 + rng.Intn(n)
				if rng.Intn(2) == 0 {
					v = -v
				}
				c = append(c, v)
			}
			cls = append(cls, c)
		}
		want := minisatBruteSat(n, cls)

		s := NewMinisatSolver()
		ss := NewMinisatSimpSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			ss.NewVar(true, true)
		}
		okS, okSS := true, true
		for _, c := range cls {
			var ps, ps2 MinisatVec[MinisatLit]
			for _, l := range c {
				ps.PushElem(minisatTestLit(l))
				ps2.PushElem(minisatTestLit(l))
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
				model func(int) MinisatLbool
			}{{"Solver", func(v int) MinisatLbool { return s.ModelValue(v) }}, {"SimpSolver", func(v int) MinisatLbool { return ss.ModelValue(v) }}} {
				for _, c := range cls {
					sat := false
					for _, l := range c {
						v := l
						if v < 0 {
							v = -v
						}
						mv := sol.model(v - 1)
						if (l > 0 && mv == MinisatLTrue) || (l < 0 && mv == MinisatLFalse) {
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

func TestMinisatPigeonholeAndDimacs(t *testing.T) {
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
			line += itoaMinisat(v(p, h)) + " "
		}
		body += line + "0\n"
		nc++
	}
	for h := 0; h < hol; h++ {
		for p1 := 0; p1 < pig; p1++ {
			for p2 := p1 + 1; p2 < pig; p2++ {
				body += "-" + itoaMinisat(v(p1, h)) + " -" + itoaMinisat(v(p2, h)) + " 0\n"
				nc++
			}
		}
	}
	hdr := "c pigeonhole\np cnf " + itoaMinisat(pig*hol) + " " + itoaMinisat(nc) + "\n"
	if err := os.WriteFile(path, []byte(hdr+body), 0o644); err != nil {
		t.Fatal(err)
	}
	in, err := OpenGZIPReader(path, GZIPReaderOptions{})
	if err != nil {
		t.Fatal(err)
	}
	s := NewMinisatSolver()
	minisatParseDIMACS(in, s)
	in.Close()
	if s.NVars() != pig*hol || s.NClauses() == 0 {
		t.Fatalf("parsed %d vars, %d clauses", s.NVars(), s.NClauses())
	}
	if s.SolveNoAssumps() {
		t.Fatal("PHP(5,4) reported satisfiable")
	}
}

func itoaMinisat(n int) string {
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

// TestMinisatLargerRandom cross-checks Solver against SimpSolver on instances big
// enough to exercise restarts, learnt clause reduction and garbage collection.
func TestMinisatLargerRandom(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	for iter := 0; iter < 40; iter++ {
		n := 60 + rng.Intn(40)
		m := int(float64(n) * 4.3)
		s := NewMinisatSolver()
		ss := NewMinisatSimpSolver()
		for i := 0; i < n; i++ {
			s.NewVar(true, true)
			ss.NewVar(true, true)
		}
		var cls []minisatTestClause
		okS, okSS := true, true
		for i := 0; i < m; i++ {
			var c minisatTestClause
			var ps, ps2 MinisatVec[MinisatLit]
			for j := 0; j < 3; j++ {
				v := 1 + rng.Intn(n)
				if rng.Intn(2) == 0 {
					v = -v
				}
				c = append(c, v)
				ps.PushElem(minisatTestLit(v))
				ps2.PushElem(minisatTestLit(v))
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
					if (l > 0 && mv == MinisatLTrue) || (l < 0 && mv == MinisatLFalse) {
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
