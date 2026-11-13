package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func glaOldHasSpecialGates(g *gia.Gia_Man_t) bool {
	for id := int32(1); id < gia.Gia_ManObjNum(g); id++ {
		_, _, _, kind := g.Gate(id)
		if kind == 3 || kind == 4 {
			return true
		}
	}
	return false
}

// glaOldLower replaces each native XOR/MUX with an AND-inverter cone.  The
// interface order is unchanged, so sequential counterexamples remain valid.
// owner records the original object represented by each newly created node.
func glaOldLower(g *gia.Gia_Man_t) (*gia.Gia_Man_t, []int32, []int32) {
	dst := gia.Gia_ManStart(3 * gia.Gia_ManObjNum(g))
	gia.Gia_ManCopyNameSpec(dst, g)
	copies := make([]int32, gia.Gia_ManObjNum(g))
	owner := []int32{0}
	copyLit := func(lit int32) int32 { return copies[lit>>1] ^ (lit & 1) }
	for id := int32(1); id < gia.Gia_ManObjNum(g); id++ {
		a, b, c, kind := g.Gate(id)
		before := gia.Gia_ManObjNum(dst)
		switch kind {
		case 1:
			copies[id] = gia.Gia_ManAppendCi(dst)
		case 2:
			copies[id] = gia.Gia_ManAppendAnd2(dst, copyLit(a), copyLit(b))
		case 3:
			copies[id] = gia.Gia_ManAppendXor2(dst, copyLit(a), copyLit(b))
		case 4:
			copies[id] = gia.Gia_ManAppendMux2(dst, copyLit(a), copyLit(b), copyLit(c))
		case 5:
			copies[id] = gia.Gia_ManAppendBuf(dst, copyLit(a))
		case 6:
			copies[id] = gia.Gia_ManAppendCo(dst, copyLit(a))
		default:
			panic("old GLA: invalid GIA object")
		}
		for j := before; j < gia.Gia_ManObjNum(dst); j++ {
			owner = append(owner, id)
		}
	}
	gia.Gia_ManSetRegNum(dst, gia.Gia_ManRegNum(g))
	gia.Gia_ManSetConstrNum(dst, gia.Gia_ManConstrNum(g))
	return dst, owner, copies
}

// glaOldPerformLowered translates abstraction classes in both directions.
// If only part of a lowered cone is selected, selecting its original gate
// strengthens the returned abstraction and is therefore conservative.
func glaOldPerformLowered(g *gia.Gia_Man_t, pars *Abs_Par_t, startVta int32) int32 {
	if pars == nil {
		panic("invalid old GLA arguments")
	}
	lowered, owner, copies := glaOldLower(g)
	defer gia.Gia_ManStop(lowered)
	if classes := gia.Gia_ManGateClasses(g); classes != nil {
		mapped := vec.Vec_IntStart(int(gia.Gia_ManObjNum(lowered)))
		for id, original := range owner {
			vec.Vec_IntWriteEntry(mapped, id, vec.Vec_IntEntry(classes, int(original)))
		}
		// A simplified cone may have no newly allocated root. Retain any
		// requested boundary at the shared resulting object as well.
		for id, lit := range copies {
			if value := vec.Vec_IntEntry(classes, id); value > vec.Vec_IntEntry(mapped, int(lit>>1)) {
				vec.Vec_IntWriteEntry(mapped, int(lit>>1), value)
			}
		}
		gia.Gia_ManSetGateClasses(lowered, mapped)
	}
	g.ClearSequentialCounterexample()
	result := Gia_ManPerformGlaOld(lowered, pars, startVta)
	if witness := lowered.TakeSequentialCounterexample(); witness != nil {
		if gia.Gia_ManVerifyCex(g, witness, 0) == 0 {
			panic("old GLA: lowered counterexample does not replay on original graph")
		}
		if err := g.SetSequentialCounterexample(witness); err != nil {
			panic(err)
		}
		gia.Gia_ManSetGateClasses(g, nil)
	} else if classes := gia.Gia_ManGateClasses(lowered); classes != nil {
		mapped := vec.Vec_IntStart(int(gia.Gia_ManObjNum(g)))
		for id, original := range owner {
			if value := vec.Vec_IntEntry(classes, id); value > vec.Vec_IntEntry(mapped, int(original)) {
				vec.Vec_IntWriteEntry(mapped, int(original), value)
			}
		}
		// Lowering may simplify a gate to an existing literal (including a
		// constant), so that gate has no entry of its own in owner. Mark all
		// classified original copies as well; this only strengthens the
		// returned abstraction and retains the proof for every frame.
		for id, lit := range copies {
			obj := gia.Gia_ManObj(g, int32(id))
			if gia.Gia_ObjIsConst0(obj) == 0 && gia.Gia_ObjIsRo(g, obj) == 0 && gia.Gia_ObjIsAnd(obj) == 0 {
				continue
			}
			if value := vec.Vec_IntEntry(classes, int(lit>>1)); value > vec.Vec_IntEntry(mapped, id) {
				vec.Vec_IntWriteEntry(mapped, id, value)
			}
		}
		gia.Gia_ManSetGateClasses(g, mapped)
	} else {
		gia.Gia_ManSetGateClasses(g, nil)
	}
	return result
}
