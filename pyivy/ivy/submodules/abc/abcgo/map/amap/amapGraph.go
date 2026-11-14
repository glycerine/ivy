package amap

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
)

func Amap_ManSetupObj(p *Amap_Man_t) *Amap_Obj_t {
	if p == nil {
		panic("amap: nil mapping manager")
	}
	obj := &Amap_Obj_t_{Id: uint32(len(p.objects)), nFouts: [2]int32{1}}
	p.objects = append(p.objects, obj)
	Vec_PtrPush(p.vObjs, obj)
	return obj
}

func Amap_ManCreateConst1(p *Amap_Man_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.Type = AMAP_OBJ_CONST1
	obj.fPhase = 1
	p.nObjs[AMAP_OBJ_CONST1]++
	return obj
}

func Amap_ManCreatePi(p *Amap_Man_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.Type = AMAP_OBJ_PI
	obj.IdPio = uint32(Vec_PtrSize(p.vPis))
	Vec_PtrPush(p.vPis, obj)
	p.nObjs[AMAP_OBJ_PI]++
	return obj
}

func Amap_ManCreatePo(p *Amap_Man_t, fan *Amap_Obj_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.IdPio = uint32(Vec_PtrSize(p.vPos))
	Vec_PtrPush(p.vPos, obj)
	obj.Type = AMAP_OBJ_PO
	obj.Fan[0] = Amap_ObjToLit(fan)
	Amap_Regular(fan).nRefs++
	obj.Level = Amap_Regular(fan).Level
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	assert(p.nLevelMax < 4094, "mapping level < 4094")
	p.nObjs[AMAP_OBJ_PO]++
	return obj
}

func Amap_ManCreateAnd(p *Amap_Man_t, fan0, fan1 *Amap_Obj_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.Type = AMAP_OBJ_AND
	obj.Fan[0], obj.Fan[1] = Amap_ObjToLit(fan0), Amap_ObjToLit(fan1)
	Amap_Regular(fan0).nRefs++
	Amap_Regular(fan1).nRefs++
	assert(Abc_Lit2Var(obj.Fan[0]) != Abc_Lit2Var(obj.Fan[1]), "AND fanins are distinct")
	obj.fPhase = uint32(Amap_ObjPhaseReal(fan0) & Amap_ObjPhaseReal(fan1))
	obj.Level = 1 + max(Amap_Regular(fan0).Level, Amap_Regular(fan1).Level)
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	assert(p.nLevelMax < 4094, "mapping level < 4094")
	p.nObjs[AMAP_OBJ_AND]++
	return obj
}

func Amap_ManCreateXor(p *Amap_Man_t, fan0, fan1 *Amap_Obj_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.Type = AMAP_OBJ_XOR
	obj.Fan[0], obj.Fan[1] = Amap_ObjToLit(fan0), Amap_ObjToLit(fan1)
	Amap_Regular(fan0).nRefs++
	Amap_Regular(fan1).nRefs++
	obj.fPhase = uint32(Amap_ObjPhaseReal(fan0) ^ Amap_ObjPhaseReal(fan1))
	obj.Level = 2 + max(Amap_Regular(fan0).Level, Amap_Regular(fan1).Level)
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	assert(p.nLevelMax < 4094, "mapping level < 4094")
	p.nObjs[AMAP_OBJ_XOR]++
	return obj
}

func Amap_ManCreateMux(p *Amap_Man_t, no, yes, control *Amap_Obj_t) *Amap_Obj_t {
	obj := Amap_ManSetupObj(p)
	obj.Type = AMAP_OBJ_MUX
	obj.Fan[0], obj.Fan[1], obj.Fan[2] = Amap_ObjToLit(no), Amap_ObjToLit(yes), Amap_ObjToLit(control)
	Amap_Regular(no).nRefs++
	Amap_Regular(yes).nRefs++
	Amap_Regular(control).nRefs++
	obj.fPhase = uint32((Amap_ObjPhaseReal(yes) & Amap_ObjPhaseReal(control)) | (Amap_ObjPhaseReal(no) & (1 ^ Amap_ObjPhaseReal(control))))
	obj.Level = 2 + max(Amap_Regular(no).Level, Amap_Regular(yes).Level, Amap_Regular(control).Level)
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	assert(p.nLevelMax < 4094, "mapping level < 4094")
	p.nObjs[AMAP_OBJ_MUX]++
	return obj
}

func Amap_ManCreateXorChoices(p *Amap_Man_t, a, b *Amap_Obj_t, choices []*Amap_Obj_t) {
	if len(choices) < 4 {
		panic("amap: four choice slots required")
	}
	choices[0] = Amap_ManCreateXor(p, a, b)
	choices[1] = Amap_ManCreateXor(p, Amap_Not(a), b)
	choices[2] = Amap_ManCreateXor(p, a, Amap_Not(b))
	choices[3] = Amap_ManCreateXor(p, Amap_Not(a), Amap_Not(b))
}

func Amap_ManCreateMuxChoices(p *Amap_Man_t, a, b, control *Amap_Obj_t, choices []*Amap_Obj_t) {
	if len(choices) < 4 {
		panic("amap: four choice slots required")
	}
	choices[0] = Amap_ManCreateMux(p, a, b, control)
	choices[1] = Amap_ManCreateMux(p, Amap_Not(a), Amap_Not(b), control)
	choices[2] = Amap_ManCreateMux(p, b, a, Amap_Not(control))
	choices[3] = Amap_ManCreateMux(p, Amap_Not(b), Amap_Not(a), Amap_Not(control))
}

func Amap_ManCreateChoice(p *Amap_Man_t, obj *Amap_Obj_t) {
	obj = Amap_Regular(obj)
	obj.fRepr = 1
	seen := make(map[uint32]bool)
	for current := obj; current != nil; current = Amap_ObjChoice(p, current) {
		current = Amap_Regular(current)
		if seen[current.Id] {
			panic("amap: cyclic choice chain")
		}
		seen[current.Id] = true
		if obj.Level < current.Level {
			obj.Level = current.Level
		}
	}
	if p.nLevelMax < int32(obj.Level) {
		p.nLevelMax = int32(obj.Level)
	}
	assert(p.nLevelMax < 4094, "mapping level < 4094")
}

func Amap_AndToObj(p *Amap_Man_t, g *aig.Graph, edge aig.Edge) *Amap_Obj_t {
	obj := p.sourceCopies[edge.NodeID()]
	if obj == nil {
		panic(fmt.Sprintf("amap: source node %d has not been mapped", edge.NodeID()))
	}
	return Amap_NotCond(obj, int32(edge&1))
}

func Amap_ManGetLast_rec(p *Amap_Man_t, obj *Amap_Obj_t) *Amap_Obj_t {
	seen := make(map[uint32]bool)
	for obj != nil {
		obj = Amap_Regular(obj)
		if obj.Equiv == 0 {
			return obj
		}
		if seen[obj.Id] {
			panic("amap: cyclic equivalence chain")
		}
		seen[obj.Id] = true
		obj = Amap_ObjChoice(p, obj)
	}
	return nil
}

// Amap_ManCreate copies an AIG graph into Amap's internal graph. The source
// uses stable node IDs and complemented-edge bits, so a local ID map replaces
// the C implementation's temporary pData fields.
func Amap_ManCreate(p *Amap_Man_t, source *aig.Graph) {
	if p == nil || source == nil {
		panic("amap: nil source graph or manager")
	}
	p.sourceCopies = make(map[int]*Amap_Obj_t, source.NodeCount())
	p.pConst1 = Amap_ManCreateConst1(p)
	p.sourceCopies[source.Const1().NodeID()] = p.pConst1
	for _, input := range source.Inputs() {
		p.sourceCopies[input.NodeID()] = Amap_ManCreatePi(p)
	}

	for _, edge := range aig.Aig_ManDfsChoices(source) {
		id := edge.NodeID()
		if p.sourceCopies[id] != nil {
			continue
		}
		node := source.Node(edge)
		fan0 := Amap_AndToObj(p, source, node.Fanins[0])
		var mapped *Amap_Obj_t
		switch node.Kind {
		case aig.And:
			if a, b, ok := aig.Aig_ObjRecognizeExor(source, edge); p.fUseXor != 0 && ok {
				var choices [4]*Amap_Obj_t
				Amap_ManCreateXorChoices(p, Amap_AndToObj(p, source, a), Amap_AndToObj(p, source, b), choices[:])
				mapped = Amap_ManCreateAnd(p, fan0, Amap_AndToObj(p, source, node.Fanins[1]))
				p.nChoicesAdded++
				Amap_ObjSetChoice(mapped, choices[0])
				for i := 0; i < 3; i++ {
					Amap_ObjSetChoice(choices[i], choices[i+1])
				}
				Amap_ManCreateChoice(p, mapped)
			} else if control, yes, no, ok := amapRecognizeMux(source, edge); p.fUseMux != 0 && ok {
				var choices [4]*Amap_Obj_t
				Amap_ManCreateMuxChoices(p, Amap_AndToObj(p, source, no), Amap_AndToObj(p, source, yes), Amap_AndToObj(p, source, control), choices[:])
				mapped = Amap_ManCreateAnd(p, fan0, Amap_AndToObj(p, source, node.Fanins[1]))
				p.nChoicesAdded++
				Amap_ObjSetChoice(mapped, choices[0])
				for i := 0; i < 3; i++ {
					Amap_ObjSetChoice(choices[i], choices[i+1])
				}
				Amap_ManCreateChoice(p, mapped)
			} else {
				mapped = Amap_ManCreateAnd(p, fan0, Amap_AndToObj(p, source, node.Fanins[1]))
			}
		case aig.ExclusiveOr:
			fan1 := Amap_AndToObj(p, source, node.Fanins[1])
			if edge.Complemented() {
				mapped = Amap_Not(Amap_ManCreateXor(p, fan0, fan1))
			} else {
				mapped = Amap_ManCreateXor(p, fan0, fan1)
			}
		default:
			continue
		}
		p.sourceCopies[id] = Amap_Regular(mapped)
		if Amap_IsComplement(mapped) != 0 {
			// The source copy table stores the regular result; the edge's phase is
			// reapplied by Amap_AndToObj at each use.
			p.sourceCopies[id] = Amap_Regular(mapped)
		}

		// Preserve source equivalence classes as ordered choice chains.
		alt := aig.Aig_ObjEquiv(source, aig.Aig_Regular(edge))
		if alt != aig.InvalidEdge {
			p.nChoicesGiven++
			prev := Amap_Regular(mapped)
			for alt != aig.InvalidEdge {
				alternative := p.sourceCopies[alt.NodeID()]
				if alternative == nil {
					break
				}
				prev.fRepr = 0
				Amap_ObjSetChoice(Amap_ManGetLast_rec(p, prev), alternative)
				prev = alternative
				alt = aig.Aig_ObjEquiv(source, aig.Aig_Regular(alt))
			}
			Amap_ManCreateChoice(p, Amap_Regular(mapped))
		}
	}

	for _, output := range source.Outputs() {
		driver := source.Node(output).Fanins[0]
		Amap_ManCreatePo(p, Amap_AndToObj(p, source, driver))
	}
	if p.pPars != nil && p.pPars.FVerbose != 0 {
		fmt.Printf("Performing mapping with %d given and %d created choices.\n", p.nChoicesGiven, p.nChoicesAdded)
	}
}

func amapRecognizeMux(g *aig.Graph, edge aig.Edge) (control, yes, no aig.Edge, ok bool) {
	if !aig.Aig_ObjIsMuxType(g, edge) {
		return aig.InvalidEdge, aig.InvalidEdge, aig.InvalidEdge, false
	}
	control, yes, no = aig.Aig_ObjRecognizeMux(g, edge)
	return control, yes, no, true
}
