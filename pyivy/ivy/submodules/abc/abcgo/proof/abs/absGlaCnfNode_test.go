package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

func TestGlaStaticNodeUsesCutCnf(t *testing.T) {
	for x := 0; x < 2; x++ {
		for y := 0; y < 2; y++ {
			for z := 0; z < 2; z++ {
				g := gia.Gia_ManStart(8)
				a := gia.Gia_ManAppendCi(g)
				b := gia.Gia_ManAppendCi(g)
				gate := gia.Gia_ManAppendAnd(g, a, b)
				gia.Gia_ManAppendCo(g, gate)
				pars := &Abs_Par_t{}
				Abs_ParSetDefaults(pars)
				p := Ga2_ManStart(g, pars)
				p.pSat = bsat.NewLabeledSolver2()
				p.nSatVars = 1
				for _, lit := range []int32{a, b} {
					Ga2_ManSetupNode(p, gia.Gia_ManObj(g, lit>>1), 0)
				}
				obj := gia.Gia_ManObj(g, gate>>1)
				Ga2_ManSetupNode(p, obj, 1)
				p.vId2Lit = append(p.vId2Lit, vec.Vec_IntStartFull(vec.Vec_IntSize(p.vValues)))
				Ga2_ManAddToAbsOneStatic(p, obj, 0, 1)
				ax := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, a>>1), 0)
				by := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, b>>1), 0)
				out := Ga2_ObjFindLit(p, obj, 0)
				got := p.pSat.Solve(bsat.Literal(ax|int32(x^1)), bsat.Literal(by|int32(y^1)), bsat.Literal(out|int32(z^1))) == bsat.Satisfiable
				Ga2_ManStop(p)
				if got != (z == x*y) {
					t.Fatalf("%d & %d = %d: sat %v", x, y, z, got)
				}
			}
		}
	}
}

func TestGlaDynamicNodeUsesCutCnf(t *testing.T) {
	for _, skip := range []int32{0, 1} {
		for x := 0; x < 2; x++ {
			for y := 0; y < 2; y++ {
				for z := 0; z < 2; z++ {
					g := gia.Gia_ManStart(8)
					a := gia.Gia_ManAppendCi(g)
					b := gia.Gia_ManAppendCi(g)
					gate := gia.Gia_ManAppendAnd(g, a, b)
					gia.Gia_ManAppendCo(g, gate)
					pars := &Abs_Par_t{}
					Abs_ParSetDefaults(pars)
					pars.fSkipHash = skip
					p := Ga2_ManStart(g, pars)
					p.pSat = bsat.NewLabeledSolver2()
					p.nSatVars = 1
					obj := gia.Gia_ManObj(g, gate>>1)
					Ga2_ManSetupNode(p, obj, 1)
					p.LimAbs = int32(vec.Vec_IntSize(p.vValues))
					for _, lit := range []int32{a, b} {
						Ga2_ManSetupNode(p, gia.Gia_ManObj(g, lit>>1), 0)
					}
					p.LimPpi = int32(vec.Vec_IntSize(p.vValues))
					p.vId2Lit = append(p.vId2Lit, vec.Vec_IntStartFull(vec.Vec_IntSize(p.vValues)))
					Ga2_ManAddToAbsOneDynamic(p, gia.Gia_ManConst0(g), 0)
					Ga2_ManAddToAbsOneDynamic(p, obj, 0)
					ax := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, a>>1), 0)
					by := Ga2_ObjFindLit(p, gia.Gia_ManObj(g, b>>1), 0)
					out := Ga2_ObjFindLit(p, obj, 0)
					got := p.pSat.Solve(bsat.Literal(ax|int32(x^1)), bsat.Literal(by|int32(y^1)), bsat.Literal(out|int32(z^1))) == bsat.Satisfiable
					Ga2_ManStop(p)
					if got != (z == x*y) {
						t.Fatalf("skip=%d %d & %d = %d: sat %v", skip, x, y, z, got)
					}
				}
			}
		}
	}
}
