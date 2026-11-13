package cec

import "github.com/glycerine/abcgo/aig/gia"

// Cec4_ObjSatId returns the SAT variable assigned to object, or -1.
func Cec4_ObjSatId(c *cec4CNF, object int32) int32 {
	if v, ok := c.ids[object]; ok {
		return v
	}
	return -1
}

// Cec4_ObjSetSatId records an already allocated SAT variable. The CNF state
// belongs to one manager, replacing the C implementation's graph scratch arrays.
func Cec4_ObjSetSatId(c *cec4CNF, object, variable int32) int32 {
	if object < 0 || object >= gia.Gia_ManObjNum(c.graph) || variable < 0 || int(variable) >= c.solver.NVars() || Cec4_ObjSatId(c, object) != -1 {
		panic("cec4: invalid SAT ID mapping")
	}
	if c.objects[variable] != -1 {
		panic("cec4: SAT variable already mapped")
	}
	c.ids[object] = variable
	c.objects[variable] = object
	if _, _, _, kind := c.graph.Gate(object); kind == 1 {
		c.inputs[object] = variable
	}
	return variable
}

func Cec4_ObjCleanSatId(c *cec4CNF, object int32) {
	v := Cec4_ObjSatId(c, object)
	if v < 0 {
		panic("cec4: object has no SAT ID")
	}
	delete(c.ids, object)
	delete(c.inputs, object)
	c.objects[v] = -1
}

func cec4SatLiteral(c *cec4CNF, lit int32) int32 {
	v := Cec4_ObjSatId(c, lit>>1)
	if v < 0 {
		panic("cec4: fanin has no SAT ID")
	}
	return v<<1 | lit&1
}

// Cec4_AddClausesMux encodes a structurally recognized mux using ABC's four
// ITE clauses and, when the arms differ, its two strengthening clauses.
func Cec4_AddClausesMux(c *cec4CNF, node int32) {
	obj := gia.Gia_ManObj(c.graph, node)
	var then, otherwise int32
	control := gia.Gia_ObjRecognizeMuxLits(c.graph, obj, &then, &otherwise)
	i, t, e := cec4SatLiteral(c, control), cec4SatLiteral(c, then), cec4SatLiteral(c, otherwise)
	f := cec4SatLiteral(c, node<<1)
	clauses := [][]int32{{i ^ 1, t ^ 1, f}, {i ^ 1, t, f ^ 1}, {i, e ^ 1, f}, {i, e, f ^ 1}}
	if t>>1 != e>>1 {
		clauses = append(clauses, []int32{t, e, f ^ 1}, []int32{t ^ 1, e ^ 1, f})
	}
	for _, clause := range clauses {
		if !c.solver.AddABCClause(clause) {
			panic("cec4: inconsistent mux clauses")
		}
	}
}

// Cec4_AddClausesSuper encodes the AND of the supplied graph literals.
func Cec4_AddClausesSuper(c *cec4CNF, node int32, fanins []int32) {
	f := cec4SatLiteral(c, node<<1)
	last := make([]int32, len(fanins)+1)
	for i, lit := range fanins {
		x := cec4SatLiteral(c, lit)
		if !c.solver.AddABCClause([]int32{x, f ^ 1}) {
			panic("cec4: inconsistent supergate clause")
		}
		last[i] = x ^ 1
	}
	last[len(fanins)] = f
	if !c.solver.AddABCClause(last) {
		panic("cec4: inconsistent supergate clause")
	}
}

// Cec4_CollectSuper_rec uses GIA literals in place of tagged C pointers.
func Cec4_CollectSuper_rec(graph *gia.Gia_Man_t, lit int32, leaves *[]int32, first, useMuxes bool) {
	obj := gia.Gia_ManObj(graph, lit>>1)
	if obj == nil {
		panic("cec4: invalid supergate object")
	}
	refs := gia.Gia_ObjValue(obj)
	if graph.Refs != nil {
		refs = gia.Gia_ObjRefNumId(graph, lit>>1)
	}
	if lit&1 != 0 || gia.Gia_ObjIsCi(obj) != 0 || !first && refs > 1 || useMuxes && gia.Gia_ObjMark0(obj) != 0 {
		for _, entry := range *leaves {
			if entry == lit {
				return
			}
		}
		*leaves = append(*leaves, lit)
		return
	}
	if gia.Gia_ObjIsAnd(obj) == 0 {
		panic("cec4: expected AND in supergate")
	}
	Cec4_CollectSuper_rec(graph, gia.Gia_ObjFaninLit0(obj, lit>>1), leaves, false, useMuxes)
	Cec4_CollectSuper_rec(graph, gia.Gia_ObjFaninLit1(obj, lit>>1), leaves, false, useMuxes)
}

func Cec4_CollectSuper(graph *gia.Gia_Man_t, root int32, useMuxes bool) []int32 {
	if root <= 0 || root&1 != 0 {
		panic("cec4: invalid supergate root")
	}
	var leaves []int32
	Cec4_CollectSuper_rec(graph, root, &leaves, true, useMuxes)
	return leaves
}

func Cec4_ObjAddToFrontier(c *cec4CNF, object int32, frontier *[]int32) {
	if object <= 0 || object >= gia.Gia_ManObjNum(c.graph) {
		panic("cec4: invalid frontier object")
	}
	if Cec4_ObjSatId(c, object) >= 0 {
		return
	}
	Cec4_ObjSetSatId(c, object, c.newVar(-1))
	if _, _, _, kind := c.graph.Gate(object); kind == 2 || kind == 3 || kind == 4 {
		*frontier = append(*frontier, object)
	}
}
