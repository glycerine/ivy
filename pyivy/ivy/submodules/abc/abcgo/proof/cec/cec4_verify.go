package cec

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/glucose2"
)

// cec4VerifyRec follows the source verification traversal while CI values
// are supplied by either the private SAT model or the packed simulation.
func cec4VerifyRec(g *gia.Gia_Man_t, id int32, input func(int32) int32) int32 {
	if id == 0 {
		return 0
	}
	obj := gia.Gia_ManObj(g, id)
	if gia.Gia_ObjIsTravIdCurrentId(g, id) != 0 {
		return gia.Gia_ObjMark1(obj)
	}
	gia.Gia_ObjSetTravIdCurrentId(g, id)
	a, b, c, kind := g.Gate(id)
	var value int32
	eval := func(lit int32) int32 { return cec4VerifyRec(g, lit>>1, input) ^ (lit & 1) }
	switch kind {
	case 1:
		value = input(id)
	case 2:
		value = eval(a) & eval(b)
	case 3:
		value = eval(a) ^ eval(b)
	case 4:
		if eval(a) != 0 {
			value = eval(b)
		} else {
			value = eval(c)
		}
	case 5, 6:
		value = eval(a)
	default:
		panic("cec4: unsupported verification node")
	}
	gia.Gia_ObjSetMark1(obj, value)
	return value
}
func Cec4_ManVerify_rec(g *gia.Gia_Man_t, id int32, cnf *cec4CNF) int32 {
	if cnf == nil || cnf.graph != g {
		panic("cec4: SAT verification graph mismatch")
	}
	return cec4VerifyRec(g, id, func(input int32) int32 {
		v, ok := cnf.ids[input]
		if !ok || int(v) >= len(cnf.solver.Model) {
			panic("cec4: SAT input has no model value")
		}
		if cnf.solver.ModelValueVar(glucose2.Var(v)) == glucose2.LTrue {
			return 1
		}
		return 0
	})
}
func Cec4_ManVerify(g *gia.Gia_Man_t, a, b, phase int32, cnf *cec4CNF) {
	gia.Gia_ManIncrementTravId(g)
	if Cec4_ManVerify_rec(g, a, cnf)^Cec4_ManVerify_rec(g, b, cnf) == phase {
		fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", a, b)
	}
}
func Cec4_ManCexVerify_rec(g *gia.Gia_Man_t, id int32, sim *cec3Simulation, pattern int) int32 {
	if sim == nil || pattern < 0 || pattern >= 64*sim.words {
		panic("cec4: invalid verification pattern")
	}
	return cec4VerifyRec(g, id, func(input int32) int32 { return int32((sim.object(input)[pattern/64] >> uint(pattern%64)) & 1) })
}
func Cec4_ManCexVerify(g *gia.Gia_Man_t, a, b, phase int32, sim *cec3Simulation, pattern int) {
	gia.Gia_ManIncrementTravId(g)
	if Cec4_ManCexVerify_rec(g, a, sim, pattern)^Cec4_ManCexVerify_rec(g, b, sim, pattern) == phase {
		fmt.Printf("CEX verification FAILED for obj %d and obj %d.\n", a, b)
	}
}
