package abc

import (
	"encoding/json"
	"errors"
	"os"
	"testing"
)

func TestGiaSimplifyingConstructorsTruthAndAllocation(t *testing.T) {
	for kind := 0; kind < 6; kind++ {
		for ai := 0; ai < 6; ai++ {
			for bi := 0; bi < 6; bi++ {
				for ci := 0; ci < 6; ci++ {
					m := NewGiaManager()
					m.AppendCI()
					m.AppendCI()
					a, b, c := GiaLit(ai), GiaLit(bi), GiaLit(ci)
					var l GiaLit
					var e error
					switch kind {
					case 0:
						l, e = m.AppendAndSimplified(a, b)
					case 1:
						l, e = m.AppendOrSimplified(a, b)
					case 2:
						l, e = m.AppendMuxSimplified(a, b, c)
					case 3:
						l, e = m.AppendMajoritySimplified(a, b, c)
					case 4:
						l, e = m.AppendXorSimplified(a, b)
					case 5:
						l, e = m.AppendXorRealSimplified(a, b)
					}
					if e != nil {
						t.Fatal(kind, ai, bi, ci, e)
					}
					if _, e = m.AppendCO(l); e != nil {
						t.Fatal(e)
					}
					if e = m.Check(); e != nil {
						t.Fatal(e)
					}
					for bits := 0; bits < 4; bits++ {
						read := func(l GiaLit) bool {
							v := false
							if l.ID() > 0 {
								v = bits&(1<<(l.ID()-1)) != 0
							}
							return v != l.Complemented()
						}
						av, bv, cv := read(a), read(b), read(c)
						want := false
						switch kind {
						case 0:
							want = av && bv
						case 1:
							want = av || bv
						case 2:
							want = cv
							if av {
								want = bv
							}
						case 3:
							want = (av && bv) || (av && cv) || (bv && cv)
						case 4, 5:
							want = av != bv
						}
						got, e := m.Evaluate([]bool{bits&1 != 0, bits&2 != 0})
						if e != nil || got[0] != want {
							t.Fatal(kind, ai, bi, ci, bits, got, want, e)
						}
					}
				}
			}
		}
	}
	m, _ := NewGiaManagerWithLimit(3)
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	if l, e := m.AppendAndSimplified(a, a.Not()); e != nil || l != GiaConst0 {
		t.Fatal(l, e)
	}
	if _, e := m.AppendAndSimplified(a, b); !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	m = NewGiaManager()
	a, _ = m.AppendCI()
	b, _ = m.AppendCI()
	x, _ := m.AppendAndSimplified(a, b)
	y, _ := m.AppendAndSimplified(a, b)
	if x == y {
		t.Fatal("unexpected strashing")
	}
	before := m.ObjectCount()
	if _, e := m.AppendMuxSimplified(a, b, InvalidGiaLit); e == nil || m.ObjectCount() != before {
		t.Fatal("invalid signal mutated graph", e)
	}
}
func FuzzGiaSimplifiedMux(f *testing.F) {
	f.Add(uint8(2), uint8(3), uint8(4), uint8(1))
	f.Fuzz(func(t *testing.T, ai, bi, ci, assignment uint8) {
		m := NewGiaManager()
		m.AppendCI()
		m.AppendCI()
		a, b, c := GiaLit(ai%6), GiaLit(bi%6), GiaLit(ci%6)
		l, e := m.AppendMuxSimplified(a, b, c)
		if e != nil {
			t.Fatal(e)
		}
		m.AppendCO(l)
		read := func(l GiaLit) bool {
			v := false
			if l.ID() > 0 {
				v = assignment&(1<<(l.ID()-1)) != 0
			}
			return v != l.Complemented()
		}
		want := read(c)
		if read(a) {
			want = read(b)
		}
		got, e := m.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		if e != nil || got[0] != want {
			t.Fatal(a, b, c, got, want, e)
		}
	})
}

func TestGiaSimplifyingConstructorsNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_append_simplified.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows [][2]int
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 1296 {
		t.Fatal(e)
	}
	index := 0
	for kind := 0; kind < 6; kind++ {
		for a := 0; a < 6; a++ {
			for b := 0; b < 6; b++ {
				for c := 0; c < 6; c++ {
					m := NewGiaManager()
					m.AppendCI()
					m.AppendCI()
					var l GiaLit
					var e error
					switch kind {
					case 0:
						l, e = m.AppendAndSimplified(GiaLit(a), GiaLit(b))
					case 1:
						l, e = m.AppendOrSimplified(GiaLit(a), GiaLit(b))
					case 2:
						l, e = m.AppendMuxSimplified(GiaLit(a), GiaLit(b), GiaLit(c))
					case 3:
						l, e = m.AppendMajoritySimplified(GiaLit(a), GiaLit(b), GiaLit(c))
					case 4:
						l, e = m.AppendXorSimplified(GiaLit(a), GiaLit(b))
					case 5:
						l, e = m.AppendXorRealSimplified(GiaLit(a), GiaLit(b))
					}
					row := rows[index]
					if e != nil || m.ObjectCount() != row[0] || int(l) != row[1] {
						t.Fatal(kind, a, b, c, m.ObjectCount(), l, row, e)
					}
					index++
				}
			}
		}
	}
}
