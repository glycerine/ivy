package fraig

import "github.com/glycerine/abcgo/sat/msat"

// Fraig_OrderVariables builds the original breadth-first cone order and
// installs supergate or mux clauses when a node is first encountered.
func Fraig_OrderVariables(p *Fraig_Man_t, old, node *Fraig_Node_t) {
	assert(old != node && !Fraig_IsComplement(old) && !Fraig_IsComplement(node), "distinct regular SAT roots")
	p.nTravIds++
	msat.Msat_IntVecFill(p.vVarsUsed, msat.Msat_IntVecReadSize(p.vVarsUsed), 0)
	msat.Msat_IntVecClear(p.vVarsInt)
	for _, n := range []*Fraig_Node_t{old, node} {
		msat.Msat_IntVecPush(p.vVarsInt, int(n.Num))
		msat.Msat_IntVecWriteEntry(p.vVarsUsed, int(n.Num), 1)
		n.TravId = p.nTravIds
	}
	for i := 0; i < msat.Msat_IntVecReadSize(p.vVarsInt); i++ {
		n := p.vNodes.pArray[msat.Msat_IntVecReadEntry(p.vVarsInt, i)]
		if Fraig_NodeIsAnd(n) == 0 {
			continue
		}
		if n.vFanins == nil {
			assert(n.fClauses == 0, "new supergate has no clauses")
			if Fraig_NodeIsMuxType(n) != 0 {
				n.vFanins = Fraig_NodeVecAlloc(4)
				a, b := Fraig_Regular(n.p1), Fraig_Regular(n.p2)
				for _, edge := range []*Fraig_Node_t{a.p1, a.p2, b.p1, b.p2} {
					Fraig_NodeVecPushUnique(n.vFanins, Fraig_Regular(edge))
				}
				Fraig_SupergateAddClausesMux(p, n)
			} else {
				n.vFanins = Fraig_CollectSupergate(n, 1)
				Fraig_SupergateAddClauses(p, n, n.vFanins)
			}
			assert(Fraig_NodeVecReadSize(n.vFanins) > 1, "supergate has multiple fanins")
			n.fClauses = 1
			p.nVarsClauses++
			n.fMark2 = 1
		}
		for _, edge := range n.vFanins.pArray {
			child := Fraig_Regular(edge)
			if child.TravId == p.nTravIds {
				continue
			}
			msat.Msat_IntVecPush(p.vVarsInt, int(child.Num))
			msat.Msat_IntVecWriteEntry(p.vVarsUsed, int(child.Num), 1)
			child.TravId = p.nTravIds
		}
	}
	Fraig_SetupAdjacentMark(p, p.vVarsInt)
}
func Fraig_SetupAdjacentMark(p *Fraig_Man_t, vars *msat.Msat_IntVec_t) {
	for _, index := range msat.Msat_IntVecReadArray(vars) {
		n := p.vNodes.pArray[index]
		if n.fMark2 == 0 || Fraig_NodeIsAnd(n) == 0 {
			continue
		}
		for _, edge := range n.vFanins.pArray {
			msat.Msat_IntVecPush(p.vAdjacents[index], int(Fraig_Regular(edge).Num))
		}
	}
	for _, index := range msat.Msat_IntVecReadArray(vars) {
		n := p.vNodes.pArray[index]
		if n.fMark2 == 0 {
			continue
		}
		n.fMark2 = 0
		if Fraig_NodeIsAnd(n) == 0 {
			continue
		}
		for _, edge := range n.vFanins.pArray {
			child := Fraig_Regular(edge)
			msat.Msat_IntVecPush(p.vAdjacents[child.Num], index)
		}
	}
}
func Fraig_SupergateAddClauses(p *Fraig_Man_t, node *Fraig_Node_t, super *Fraig_NodeVec_t) {
	assert(Fraig_NodeIsAnd(node) != 0, "supergate root is AND")
	for _, edge := range super.pArray {
		fraigSatClause(p, fraigLit(Fraig_Regular(edge), Fraig_IsComplement(edge)), fraigLit(node, true))
	}
	literals := make([]int, 0, len(super.pArray)+1)
	for _, edge := range super.pArray {
		literals = append(literals, fraigLit(Fraig_Regular(edge), !Fraig_IsComplement(edge)))
	}
	literals = append(literals, fraigLit(node, false))
	fraigSatClause(p, literals...)
}
func Fraig_SupergateAddClausesExor(p *Fraig_Man_t, node *Fraig_Node_t) {
	assert(!Fraig_IsComplement(node) && Fraig_NodeIsExorType(node) != 0, "XOR root")
	a := Fraig_Regular(Fraig_Regular(node.p1).p1)
	b := Fraig_Regular(Fraig_Regular(node.p1).p2)
	comp := Fraig_NodeIsExor(node) != 0
	for _, signs := range [][3]bool{{comp, comp, comp}, {comp, !comp, !comp}, {!comp, comp, !comp}, {!comp, !comp, comp}} {
		fraigSatClause(p, fraigLit(node, signs[0]), fraigLit(a, signs[1]), fraigLit(b, signs[2]))
	}
}
func Fraig_SupergateAddClausesMux(p *Fraig_Man_t, node *Fraig_Node_t) {
	assert(!Fraig_IsComplement(node) && Fraig_NodeIsMuxType(node) != 0, "MUX root")
	var t, e *Fraig_Node_t
	c := Fraig_NodeRecognizeMux(node, &t, &e)
	x, y := Fraig_Regular(t), Fraig_Regular(e)
	tx, ey := Fraig_IsComplement(t), Fraig_IsComplement(e)
	fraigSatClause(p, fraigLit(c, true), fraigLit(x, !tx), fraigLit(node, false))
	fraigSatClause(p, fraigLit(c, true), fraigLit(x, tx), fraigLit(node, true))
	fraigSatClause(p, fraigLit(c, false), fraigLit(y, !ey), fraigLit(node, false))
	fraigSatClause(p, fraigLit(c, false), fraigLit(y, ey), fraigLit(node, true))
	if x != y {
		fraigSatClause(p, fraigLit(x, tx), fraigLit(y, ey), fraigLit(node, true))
		fraigSatClause(p, fraigLit(x, !tx), fraigLit(y, !ey), fraigLit(node, false))
	}
}
