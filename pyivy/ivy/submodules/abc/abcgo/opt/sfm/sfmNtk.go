package sfm

// Port of src/opt/sfm/sfmNtk.c.

import (
	"fmt"

	"github.com/glycerine/abcgo/sat/bsat"
)

// Sfm_CheckConsistency reports the conditions C asserts in
// Sfm_CheckConsistency.
func Sfm_CheckConsistency(vFanins [][]int, nPis, nPos int, vFixed []bool) error {
	for i, vArray := range vFanins {
		// PIs have no fanins
		if i < nPis && (len(vArray) != 0 || vFixed[i]) {
			return fmt.Errorf("sfm: PI %d has fanins or is fixed", i)
		}
		// nodes are in a topo order; POs cannot be fanins
		for _, Fanin := range vArray {
			if !(Fanin+nPos < len(vFanins)) {
				return fmt.Errorf("sfm: object %d has PO fanin %d", i, Fanin)
			}
		}
		// POs have one fanout
		if i+nPos >= len(vFanins) && (len(vArray) != 1 || vFixed[i]) {
			return fmt.Errorf("sfm: PO %d has %d fanins or is fixed", i, len(vArray))
		}
	}
	return nil
}

// Sfm_CreateFanout builds the fanout arrays. As in C, each array is allocated
// with its exact size, then filled in increasing object order.
func Sfm_CreateFanout(vFanins [][]int) [][]int {
	counts := make([]int, len(vFanins))
	total := 0
	for _, vArray := range vFanins {
		for _, Fanin := range vArray {
			counts[Fanin]++
			total++
		}
	}
	vFanouts := make([][]int, len(vFanins))
	storage := make([]int, total)
	for i, k := range counts {
		vFanouts[i] = storage[:0:k]
		storage = storage[k:]
	}
	for i, vArray := range vFanins {
		for _, Fanin := range vArray {
			vFanouts[Fanin] = append(vFanouts[Fanin], i)
		}
	}
	return vFanouts
}

func Sfm_ObjLevelNew(vArray []int, vLevels []int, fAddLevel bool) int {
	Level := 0
	for _, Fanin := range vArray {
		Level = max(Level, vLevels[Fanin])
	}
	if fAddLevel {
		Level++
	}
	return Level
}
func Sfm_CreateLevel(vFanins [][]int, vEmpty []bool) []int {
	vLevels := make([]int, len(vFanins))
	for i, vArray := range vFanins {
		vLevels[i] = Sfm_ObjLevelNew(vArray, vLevels, Sfm_ObjAddsLevelArray(vEmpty, i))
	}
	return vLevels
}

func Sfm_ObjLevelNewR(vArray []int, vLevelsR []int, fAddLevel bool) int {
	LevelR := 0
	for _, Fanout := range vArray {
		LevelR = max(LevelR, vLevelsR[Fanout])
	}
	if fAddLevel {
		LevelR++
	}
	return LevelR
}
func Sfm_CreateLevelR(vFanouts [][]int, vEmpty []bool) []int {
	vLevelsR := make([]int, len(vFanouts))
	for i := len(vFanouts) - 1; i >= 0; i-- {
		vLevelsR[i] = Sfm_ObjLevelNewR(vFanouts[i], vLevelsR, Sfm_ObjAddsLevelArray(vEmpty, i))
	}
	return vLevelsR
}

// Sfm_NtkConstruct takes ownership of its arguments, as C does. vFixed may
// be nil (no fixed objects); vEmpty may be nil (no transparent objects).
// vTruths holds one word per object; objects with more than six fanins take
// their truth table from vTruths2 at offset vStarts[i].
func Sfm_NtkConstruct(vFanins [][]int, nPis, nPos int, vFixed, vEmpty []bool, vTruths []uint64, vStarts []int, vTruths2 []uint64) (*Sfm_Ntk_t, error) {
	if vFixed == nil {
		vFixed = make([]bool, len(vFanins))
	}
	if len(vFixed) != len(vFanins) || (vEmpty != nil && len(vEmpty) != len(vFanins)) || len(vTruths) != len(vFanins) {
		return nil, fmt.Errorf("sfm: inconsistent network array sizes")
	}
	if vStarts == nil {
		vStarts = make([]int, len(vFanins))
	}
	if err := Sfm_CheckConsistency(vFanins, nPis, nPos, vFixed); err != nil {
		return nil, err
	}
	p := &Sfm_Ntk_t{}
	p.nObjs = len(vFanins)
	p.Pis = nPis
	p.Pos = nPos
	p.nNodes = p.nObjs - p.Pis - p.Pos
	// user data
	p.Fixed = vFixed
	p.Empty = vEmpty
	p.vTruths = vTruths
	p.Fanins = vFanins
	p.vStarts = vStarts
	p.vTruths2 = vTruths2
	// attributes
	p.Fanouts = Sfm_CreateFanout(p.Fanins)
	p.Levels = Sfm_CreateLevel(p.Fanins, vEmpty)
	p.Reverse = Sfm_CreateLevelR(p.Fanouts, vEmpty)
	p.vCounts = make([]int, p.nObjs)
	p.vTravIds = make([]int, p.nObjs)
	p.vTravIds2 = make([]int, p.nObjs)
	p.vId2Var = make([]int32, 2*p.nObjs)
	p.vVar2Id = make([]int, 2*p.nObjs)
	for i := range p.vId2Var {
		p.vId2Var[i] = -1
		p.vVar2Id[i] = -1
	}
	p.vCover = make([]uint32, 0, 1<<16)
	p.vCnfs = Sfm_CreateCnf(p)
	Sfm_ParSetDefault(&p.Params)
	p.pPars = &p.Params
	return p, nil
}

func Sfm_NtkPrepare(p *Sfm_Ntk_t) {
	p.LevelMax = intVecFindMax(p.Levels) + p.pPars.GrowthLevel
	p.vNodes = make([]int, 0, 1000)
	p.vDivs = make([]int, 0, 100)
	p.vRoots = make([]int, 0, 1000)
	p.vTfo = make([]int, 0, 1000)
	p.vDivCexes = make([]uint64, p.pPars.WinSizeMax)
	p.vOrder = make([]int, 0, 100)
	p.vDivVars = make([]int32, 0, 100)
	p.vDivIds = make([]int32, 0, 1000)
	p.vLits = make([]int32, 0, 100)
	p.vValues = make([]int32, 0, 100)
	p.vClauses = make([][]int32, 0, 100)
	p.vFaninMap = make([]int32, 0, 10)
	p.pSat = bsat.NewSolver()
	p.pSat.SetNumVars(int32(p.pPars.WinSizeMax))
}

// intVecFindMax mirrors Vec_IntFindMax, which returns 0 for an empty vector.
func intVecFindMax(a []int) int {
	if len(a) == 0 {
		return 0
	}
	Best := a[0]
	for _, v := range a {
		if Best < v {
			Best = v
		}
	}
	return Best
}

func Sfm_NtkFree(p *Sfm_Ntk_t) {
	if p.pSat != nil {
		p.pSat.Delete()
	}
	*p = Sfm_Ntk_t{}
}

// intVecRemove mirrors Vec_IntRemove: it deletes the first occurrence of
// Entry, keeping the order of the remaining entries, and reports success.
func intVecRemove(a *[]int, Entry int) bool {
	v := *a
	for i, e := range v {
		if e == Entry {
			copy(v[i:], v[i+1:])
			*a = v[:len(v)-1]
			return true
		}
	}
	return false
}
func intVecFind(a []int, Entry int) int {
	for i, e := range a {
		if e == Entry {
			return i
		}
	}
	return -1
}

func Sfm_NtkRemoveFanin(p *Sfm_Ntk_t, iNode, iFanin int) {
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	assert(!Sfm_ObjIsPo(p, iFanin), "!Sfm_ObjIsPo(p, iFanin)")
	RetValue := intVecRemove(&p.Fanins[iNode], iFanin)
	assert(RetValue, "RetValue")
	RetValue = intVecRemove(&p.Fanouts[iFanin], iNode)
	assert(RetValue, "RetValue")
}
func Sfm_NtkAddFanin(p *Sfm_Ntk_t, iNode, iFanin int) {
	if iFanin < 0 {
		return
	}
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	assert(!Sfm_ObjIsPo(p, iFanin), "!Sfm_ObjIsPo(p, iFanin)")
	assert(intVecFind(p.Fanins[iNode], iFanin) == -1, "Vec_IntFind( Sfm_ObjFiArray(p, iNode), iFanin ) == -1")
	assert(intVecFind(p.Fanouts[iFanin], iNode) == -1, "Vec_IntFind( Sfm_ObjFoArray(p, iFanin), iNode ) == -1")
	p.Fanins[iNode] = append(p.Fanins[iNode], iFanin)
	p.Fanouts[iFanin] = append(p.Fanouts[iFanin], iNode)
}
func Sfm_NtkDeleteObj_rec(p *Sfm_Ntk_t, iNode int) {
	if Sfm_ObjFanoutNum(p, iNode) > 0 || Sfm_ObjIsPi(p, iNode) || Sfm_ObjIsFixed(p, iNode) {
		return
	}
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	for _, iFanin := range p.Fanins[iNode] {
		RetValue := intVecRemove(&p.Fanouts[iFanin], iNode)
		assert(RetValue, "RetValue")
		Sfm_NtkDeleteObj_rec(p, iFanin)
	}
	p.Fanins[iNode] = p.Fanins[iNode][:0]
	p.vTruths[iNode] = 0
}
func Sfm_NtkUpdateLevel_rec(p *Sfm_Ntk_t, iNode int) {
	LevelNew := Sfm_ObjLevelNew(p.Fanins[iNode], p.Levels, Sfm_ObjAddsLevel(p, iNode))
	if LevelNew == Sfm_ObjLevel(p, iNode) {
		return
	}
	Sfm_ObjSetLevel(p, iNode, LevelNew)
	for _, iFanout := range p.Fanouts[iNode] {
		Sfm_NtkUpdateLevel_rec(p, iFanout)
	}
}
func Sfm_NtkUpdateLevelR_rec(p *Sfm_Ntk_t, iNode int) {
	LevelNew := Sfm_ObjLevelNewR(p.Fanouts[iNode], p.Reverse, Sfm_ObjAddsLevel(p, iNode))
	if LevelNew == Sfm_ObjLevelR(p, iNode) {
		return
	}
	Sfm_ObjSetLevelR(p, iNode, LevelNew)
	for _, iFanin := range p.Fanins[iNode] {
		Sfm_NtkUpdateLevelR_rec(p, iFanin)
	}
}
func Sfm_NtkUpdate(p *Sfm_Ntk_t, iNode, f, iFaninNew int, uTruth uint64) {
	iFanin := Sfm_ObjFanin(p, iNode, f)
	assert(Sfm_ObjIsNode(p, iNode), "Sfm_ObjIsNode(p, iNode)")
	assert(iFanin != iFaninNew, "iFanin != iFaninNew")
	assert(Sfm_ObjFaninNum(p, iNode) <= 6, "Sfm_ObjFaninNum(p, iNode) <= 6")
	if uTruth == 0 || ^uTruth == 0 {
		// Sfm_ObjForEachFanin( p, iNode, iFanin, f ) leaves iFanin set to the
		// last fanin visited; the level update below uses that value.
		for _, iFanin = range p.Fanins[iNode] {
			RetValue := intVecRemove(&p.Fanouts[iFanin], iNode)
			assert(RetValue, "RetValue")
			Sfm_NtkDeleteObj_rec(p, iFanin)
		}
		p.Fanins[iNode] = p.Fanins[iNode][:0]
	} else {
		// replace old fanin by new fanin
		Sfm_NtkRemoveFanin(p, iNode, iFanin)
		Sfm_NtkAddFanin(p, iNode, iFaninNew)
		// recursively remove MFFC
		Sfm_NtkDeleteObj_rec(p, iFanin)
	}
	// update logic level
	Sfm_NtkUpdateLevel_rec(p, iNode)
	if iFaninNew != -1 {
		Sfm_NtkUpdateLevelR_rec(p, iFaninNew)
	}
	if Sfm_ObjFanoutNum(p, iFanin) > 0 {
		Sfm_NtkUpdateLevelR_rec(p, iFanin)
	}
	// update truth table
	p.vTruths[iNode] = uTruth
	Sfm_TruthToCnf(uTruth, nil, Sfm_ObjFaninNum(p, iNode), &p.vCover, &p.vCnfs[iNode])
}

// Sfm_NodeReadFanins returns the fanin array of the object; it belongs to
// the network, as in C.
func Sfm_NodeReadFanins(p *Sfm_Ntk_t, i int) []int {
	return p.Fanins[i]
}

// Sfm_NodeReadTruth returns the truth table of the object; it belongs to the
// network. Objects with up to six fanins have one word.
func Sfm_NodeReadTruth(p *Sfm_Ntk_t, i int) []uint64 {
	if Sfm_ObjFaninNum(p, i) <= 6 {
		return p.vTruths[i : i+1]
	}
	return p.vTruths2[p.vStarts[i]:]
}
func Sfm_NodeReadFixed(p *Sfm_Ntk_t, i int) bool {
	return p.Fixed[i]
}
func Sfm_NodeReadUsed(p *Sfm_Ntk_t, i int) bool {
	return (Sfm_ObjFaninNum(p, i) > 0) || (Sfm_ObjFanoutNum(p, i) > 0)
}
