package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

func glaContradictoryProperty() *aig.Graph {
	g := aig.Aig_ManStart(32)
	a, b := aig.Aig_ObjCreateCi(g), aig.Aig_ObjCreateCi(g)
	clauses := []aig.Edge{
		aig.Aig_Or(g, a, b),
		aig.Aig_Or(g, aig.Aig_Not(a), b),
		aig.Aig_Or(g, a, aig.Aig_Not(b)),
		aig.Aig_Or(g, aig.Aig_Not(a), aig.Aig_Not(b)),
	}
	bad := aig.Aig_And(g, aig.Aig_And(g, clauses[0], clauses[1]), aig.Aig_And(g, clauses[2], clauses[3]))
	aig.Aig_ObjCreateCo(g, bad)
	aig.Aig_ManSetRegNum(g, 0)
	return g
}

func TestGlaPbaUnsatCoreSelectsCone(t *testing.T) {
	for name, perform := range map[string]func(*Aig_Man_t, int32, int32, int32, int32, int32, int32) *Vec_Int_t{
		"gla2": Aig_Gla2ManPerform,
		"gla3": Aig_Gla3ManPerform,
	} {
		t.Run(name, func(t *testing.T) {
			g := glaContradictoryProperty()
			defer aig.Aig_ManStop(g)
			got := perform(g, 0, 2, 0, 0, 0, 0)
			if got == nil {
				t.Fatal("expected UNSAT proof core")
			}
			if vecEntry(got, 0) == 0 {
				t.Fatal("proof abstraction omits the constant node")
			}
			selected := 0
			for _, value := range vec.Vec_IntArray(got) {
				if value != 0 {
					selected++
				}
			}
			if selected < 2 {
				t.Fatalf("proof core selected no logic nodes: %v", vec.Vec_IntArray(got))
			}
		})
	}
}

func TestGlaPbaSatPropertyReturnsNil(t *testing.T) {
	for name, perform := range map[string]func(*Aig_Man_t, int32, int32, int32, int32, int32, int32) *Vec_Int_t{
		"gla2": Aig_Gla2ManPerform,
		"gla3": Aig_Gla3ManPerform,
	} {
		t.Run(name, func(t *testing.T) {
			g := aig.Aig_ManStart(4)
			aig.Aig_ObjCreateCo(g, aig.Aig_ManConst1(g))
			aig.Aig_ManSetRegNum(g, 0)
			defer aig.Aig_ManStop(g)
			if got := perform(g, 0, 1, 0, 0, 0, 0); got != nil {
				t.Fatalf("SAT property returned abstraction %v", got)
			}
		})
	}
}

func vecEntry(v *Vec_Int_t, id int32) int32 {
	return vec.Vec_IntEntry(v, int(id))
}

func TestGlaPbaGateCNFMatchesBooleanFunctions(t *testing.T) {
	for complement := int32(0); complement <= 1; complement++ {
		for x := int32(0); x <= 1; x++ {
			for z := int32(0); z <= 1; z++ {
				s := bsat.NewSolver2()
				Aig_Gla2AddBuffer(s, 0, 1, complement)
				force := func(v, value int32) {
					lit := bsat.Negative(v)
					if value != 0 {
						lit = bsat.Positive(v)
					}
					s.AddClause(lit)
				}
				force(0, z)
				force(1, x)
				wantSat := z == (x ^ complement)
				if got := s.Solve() == bsat.Satisfiable; got != wantSat {
					t.Fatalf("buffer z=%d x=%d complement=%d: sat=%v want %v", z, x, complement, got, wantSat)
				}
			}
		}
	}
	for ca := int32(0); ca <= 1; ca++ {
		for cb := int32(0); cb <= 1; cb++ {
			for a := int32(0); a <= 1; a++ {
				for b := int32(0); b <= 1; b++ {
					for z := int32(0); z <= 1; z++ {
						s := bsat.NewSolver2()
						Aig_Gla2AddNode(s, 0, 1, 2, ca, cb)
						force := func(v, value int32) {
							lit := bsat.Negative(v)
							if value != 0 {
								lit = bsat.Positive(v)
							}
							s.AddClause(lit)
						}
						force(0, z)
						force(1, a)
						force(2, b)
						wantSat := z == ((a ^ ca) & (b ^ cb))
						if got := s.Solve() == bsat.Satisfiable; got != wantSat {
							t.Fatalf("AND z=%d a=%d b=%d complements=%d,%d: sat=%v want %v", z, a, b, ca, cb, got, wantSat)
						}
					}
				}
			}
		}
	}
}

func glaOneRegister(next aig.Edge) *aig.Graph {
	g := aig.Aig_ManStart(8)
	lo := aig.Aig_ObjCreateCi(g)
	aig.Aig_ObjCreateCo(g, lo) // the bad property is the latch value
	aig.Aig_ObjCreateCo(g, next)
	aig.Aig_ManSetRegNum(g, 1)
	return g
}

func TestGla3FrameTransitionComplement(t *testing.T) {
	for _, tc := range []struct {
		name    string
		next    aig.Edge
		wantSAT bool
	}{
		{name: "constant-zero-next-state", next: 1, wantSAT: false},
		{name: "constant-one-next-state", next: 0, wantSAT: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := glaOneRegister(tc.next)
			defer aig.Aig_ManStop(g)
			got := Aig_Gla3ManPerform(g, 1, 2, 0, 0, 0, 0)
			if (got != nil) != (!tc.wantSAT) {
				t.Fatalf("core=%v, wantSAT=%v", got, tc.wantSAT)
			}
		})
	}
}
