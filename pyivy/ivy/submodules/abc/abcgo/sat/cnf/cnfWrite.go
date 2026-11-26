package cnf

import (
	"math/bits"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Cnf_ManWriteCnfMapping(manager *Cnf_Man_t, mapped *Vec_Ptr_t) *Vec_Int_t {
	if manager == nil || manager.pManAig == nil {
		return vec.Vec_IntAllocExact(0)
	}
	result := vec.Vec_IntStart(aigObjectCount(manager.pManAig))
	for _, item := range vec.Vec_PtrArray(mapped) {
		edge, ok := item.(Aig_Obj_t)
		if !ok || edge.NodeID() >= len(manager.selectedCuts) {
			continue
		}
		cut := manager.selectedCuts[edge.NodeID()]
		if cut == nil || len(cut.truth) == 0 || len(cut.pFanins) > 4 {
			continue
		}
		vec.Vec_IntWriteEntry(result, edge.NodeID(), int32(vec.Vec_IntSize(result)))
		truth := uint32(smallCutTruth(cut.truth[0], len(cut.pFanins)))
		vec.Vec_IntPush(result, int32(truth|truth<<16))
		vec.Vec_IntPushArray(result, cut.pFanins, len(cut.pFanins))
		for i := len(cut.pFanins); i < 4; i++ {
			vec.Vec_IntPush(result, -1)
		}
	}
	return result
}

// ABC stores four ternary digits per SOP byte, with the first variable in
// the most significant digit. ISOP instead places that variable in bits 0–1.
type sopCube struct {
	packed   byte
	literals [5]byte
}

var sopCubes = func() [256]sopCube {
	var table [256]sopCube
	for code := range table {
		cube := code
		var packed byte
		for range 4 {
			packed = packed<<2 | byte((cube%3+1)%3)
			cube /= 3
		}
		entry := &table[code]
		entry.packed = packed
		for i := range 4 {
			entry.literals[i+1] = entry.literals[i]
			if packed>>(2*i)&3 != 0 {
				entry.literals[i+1]++
			}
		}
	}
	return table
}()

func Cnf_SopConvertToVector(sop string, cubes int32, cover *Vec_Int_t) {
	if cover == nil {
		return
	}
	vec.Vec_IntClear(cover)
	for i := int32(0); i < cubes; i++ {
		vec.Vec_IntPush(cover, int32(sopCubes[sop[i]].packed))
	}
}
func Cnf_SopCountLiterals(sop string, cubes int32) int32 {
	var count int32
	for i := int32(0); i < cubes; i++ {
		count += int32(sopCubes[sop[i]].literals[4])
	}
	return count
}
func Cnf_IsopCountLiterals(isop *Vec_Int_t, variables int32) int32 {
	if isop == nil || variables <= 0 {
		return 0
	}
	var count int32
	mask := uint32(0x55555555) & (uint32(1)<<uint(2*min(variables, 16)) - 1)
	for _, cube := range vec.Vec_IntArray(isop) {
		// The pair encodes a literal exactly when its two bits differ.
		u := uint32(cube)
		count += int32(bits.OnesCount32((u ^ u>>1) & mask))
	}
	return count
}
func Cnf_IsopWriteCube(cube int32, variables int32, vars, literals []int32) int32 {
	if variables < 0 || variables > 16 || vars == nil || literals == nil {
		return 0
	}
	count := int32(0)
	for i := int32(0); i < variables; i++ {
		state := cube & 3
		cube >>= 2
		if state != 1 && state != 2 {
			continue
		}
		literals[count] = vars[i]*2 + state - 1
		count++
	}
	return count
}

func Cnf_ManWriteCnf(manager *Cnf_Man_t, mapped *Vec_Ptr_t, outputs int32) *Cnf_Dat_t {
	if manager != nil && manager.pManAig != nil {
		graph := manager.pManAig
		count := int32(graph.OutputCount())
		if outputs < 0 || outputs > count || outputs != 0 && outputs != count && (graph.RegisterCount() == 0 || outputs != graph.RegisterCount()) {
			panic("cnf: invalid free output count")
		}
	}
	result, err := cnfWriteMapped(manager, mapped, outputs)
	if err != nil {
		return nil
	}
	return result
}
func Cnf_ManWriteCnfOther(manager *Cnf_Man_t, mapped *Vec_Ptr_t) *Cnf_Dat_t {
	if manager == nil {
		return nil
	}
	result, err := writeOtherMapped(manager, mapped)
	if err != nil {
		return nil
	}
	return result
}
// Cnf_DeriveSimple is a line-by-line port of cnfWrite.c: it derives a simple
// CNF for the AIG. The last nOutputs COs are not asserted; instead each gets a
// new variable equivalent to its driver. Variable 0 is unused; COs (when
// nOutputs != 0), AND nodes, CIs and the constant are numbered in that order.
func Cnf_DeriveSimple(p *Aig_Man_t, nOutputs int32) *Cnf_Dat_t {
	if p == nil || nOutputs < 0 || nOutputs > aig.Aig_ManCoNum(p) {
		return nil
	}
	nObjs := aig.Aig_ManObjNumMax(p)
	isNode := func(i int32) bool {
		t := aig.Aig_ObjType(p, aig.Aig_ManObj(p, i))
		return t == aig.AIG_OBJ_AND || t == aig.AIG_OBJ_EXOR
	}
	isExor := func(i int32) bool { return aig.Aig_ObjType(p, aig.Aig_ManObj(p, i)) == aig.AIG_OBJ_EXOR }
	nNodes, nExors := int32(0), int32(0)
	for i := int32(0); i < nObjs; i++ {
		if isNode(i) {
			nNodes++
		}
		if isExor(i) {
			nExors++
		}
	}
	nCos := aig.Aig_ManCoNum(p)
	// count the number of literals and clauses
	// (C encodes every node as an AND; native AIGs passed here never contain
	// EXOR nodes. Go callers may, so EXOR nodes get their four XOR clauses.)
	nLiterals := 1 + 7*(nNodes-nExors) + 12*nExors + nCos + 3*nOutputs
	nClauses := 1 + 3*(nNodes-nExors) + 4*nExors + nCos + nOutputs

	// create room for variable numbers
	pVarNums := make([]int32, nObjs)
	for i := range pVarNums {
		pVarNums[i] = -1
	}
	// assign variables to the last (nOutputs) POs
	Number := int32(1)
	if nOutputs != 0 {
		for i := int32(0); i < nCos; i++ {
			pVarNums[aig.Aig_ManCo(p, i).NodeID()] = Number
			Number++
		}
	}
	// assign variables to the internal nodes
	for i := int32(0); i < nObjs; i++ {
		if isNode(i) {
			pVarNums[i] = Number
			Number++
		}
	}
	// assign variables to the PIs and constant node
	for i := int32(0); i < aig.Aig_ManCiNum(p); i++ {
		pVarNums[aig.Aig_ManCi(p, i).NodeID()] = Number
		Number++
	}
	pVarNums[aig.Aig_ManConst1(p).NodeID()] = Number
	Number++
	nVars := Number

	// assign the clauses
	storage := make([]int32, 0, nLiterals)
	offsets := make([]int32, 0, nClauses+1)
	clause := func(lits ...int32) {
		offsets = append(offsets, int32(len(storage)))
		storage = append(storage, lits...)
	}
	b2i := func(c int32) int32 { return c & 1 }
	for i := int32(0); i < nObjs; i++ {
		if !isNode(i) {
			continue
		}
		pObj := aig.Aig_ManObj(p, i)
		OutVar := pVarNums[i]
		v0 := pVarNums[aig.Aig_ObjFanin0(p, pObj).NodeID()]
		v1 := pVarNums[aig.Aig_ObjFanin1(p, pObj).NodeID()]
		c0 := aig.Aig_ObjFaninC0(p, pObj)
		c1 := aig.Aig_ObjFaninC1(p, pObj)
		if isExor(i) {
			l0, l1 := 2*v0+c0, 2*v1+c1
			clause(2*OutVar+1, l0, l1)
			clause(2*OutVar+1, l0^1, l1^1)
			clause(2*OutVar, l0^1, l1)
			clause(2*OutVar, l0, l1^1)
			continue
		}
		// positive phase
		clause(2*OutVar, 2*v0+b2i(c0^1), 2*v1+b2i(c1^1))
		// negative phase
		clause(2*OutVar+1, 2*v0+c0)
		clause(2*OutVar+1, 2*v1+c1)
	}
	// write the constant literal
	OutVar := pVarNums[aig.Aig_ManConst1(p).NodeID()]
	clause(2 * OutVar)
	// write the output literals
	for i := int32(0); i < nCos; i++ {
		pObj := aig.Aig_ManCo(p, i)
		OutVar = pVarNums[aig.Aig_ObjFanin0(p, pObj).NodeID()]
		c0 := aig.Aig_ObjFaninC0(p, pObj)
		if i < nCos-nOutputs {
			clause(2*OutVar + c0)
		} else {
			PoVar := pVarNums[pObj.NodeID()]
			// first clause
			clause(2*PoVar, 2*OutVar+b2i(c0^1))
			// second clause
			clause(2*PoVar+1, 2*OutVar+c0)
		}
	}
	offsets = append(offsets, int32(len(storage)))
	assert(int32(len(storage)) == nLiterals, "pLits - pCnf->pClauses[0] == nLiterals")
	assert(int32(len(offsets)) == nClauses+1, "pClas - pCnf->pClauses == nClauses")

	pCnf := &Cnf_Dat_t{pMan: p}
	Cnf_DataSetPackedClauses(pCnf, nVars, offsets, storage)
	pCnf.pVarNums = pVarNums
	pCnf.VarNums = append([]int32(nil), pVarNums...)
	return pCnf
}
func Cnf_DeriveSimpleForRetiming(graph *Aig_Man_t) *Cnf_Dat_t {
	result, err := EncodeAIG(graph, 0)
	if err != nil {
		return nil
	}
	return result
}
