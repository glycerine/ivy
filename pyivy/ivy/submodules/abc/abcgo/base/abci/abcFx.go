package abci

import (
	"fmt"

	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Fx_Man_t is ABC's traditional fast_extract manager. Cubes are kept in
// source order and literals use ABC's 2*variable+complement encoding.
type Fx_Man_t struct {
	vCubes                                                 *vec.Vec_Wec_t
	LitCountMax, fCanonDivs                                int32
	vLits                                                  *vec.Vec_Wec_t
	vCounts                                                *vec.Vec_Int_t
	pHash                                                  *vec.Hsh_VecMan_t
	vWeights                                               *vec.Vec_Flt_t
	vPrio                                                  *vec.Vec_Que_t
	vVarCube, vLevels                                      *vec.Vec_Int_t
	vCubesS, vCubesD, vCompls, vCubeFree, vDiv, vSCC       *vec.Vec_Int_t
	timeStart                                              int64
	nVars, nLits, nDivs, nCompls, nPairsS, nPairsD, nDivsS int32
	nDivMux                                                [3]int32
}

func Fx_ManGetFirstVarCube(p *Fx_Man_t, cube *vec.Vec_Int_t) int32 {
	return vec.Vec_IntEntry(p.vVarCube, int(vec.Vec_IntEntry(cube, 0)))
}

func Abc_NtkFxRetrieve(ntk *Abc_Ntk_t) *vec.Vec_Wec_t {
	assert(baseabc.Abc_NtkIsSopLogic(ntk) != 0, "Abc_NtkIsSopLogic(pNtk)")
	cubes := vec.Vec_WecAlloc(1000)
	for _, node := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		sop, ok := node.Data.(string)
		assert(ok, "SOP nodes carry string functions")
		nVars := baseabc.Abc_SopGetVarNum(sop)
		assert(nVars == baseabc.Abc_ObjFaninNum(node), "SOP variable count matches fanins")
		stride := int(nVars + 3)
		for offset := 0; offset < len(sop) && sop[offset] != 0; offset += stride {
			cube := vec.Vec_WecPushLevel(cubes)
			vec.Vec_IntPush(cube, int32(node.ID))
			for pin := int32(0); pin < nVars; pin++ {
				ch := sop[offset+int(pin)]
				if ch == '0' {
					vec.Vec_IntPush(cube, util.Abc_Var2Lit(int32(baseabc.Abc_ObjFanin(node, pin).ID), 1))
				}
				if ch == '1' {
					vec.Vec_IntPush(cube, util.Abc_Var2Lit(int32(baseabc.Abc_ObjFanin(node, pin).ID), 0))
				}
			}
			lits := vec.Vec_IntArray(cube)
			vec.Vec_IntSelectSort(lits[1:], len(lits)-1)
		}
	}
	return cubes
}

func Abc_NtkFxInsert(ntk *Abc_Ntk_t, cubes *vec.Vec_Wec_t) {
	assert(baseabc.Abc_NtkIsSopLogic(ntk) != 0, "Abc_NtkIsSopLogic(pNtk)")
	maxID, last := int32(0), int32(-1)
	for i := 0; i < vec.Vec_WecSize(cubes); i++ {
		cube := vec.Vec_WecEntry(cubes, i)
		assert(vec.Vec_IntSize(cube) > 0, "nonempty fast_extract cube")
		id := vec.Vec_IntEntry(cube, 0)
		assert(last <= id, "cubes ordered by node")
		last = id
		if id > maxID {
			maxID = id
		}
	}
	if maxID < baseabc.Abc_NtkObjNumMax(ntk) {
		fmt.Printf("The network is unchanged by fast extract.\n")
		return
	}
	for id := baseabc.Abc_NtkObjNumMax(ntk); id <= maxID; id++ {
		node := baseabc.Abc_NtkCreateNode(ntk)
		assert(int32(node.ID) == id, "new node ID follows cube ID")
	}
	first := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNumMax(ntk)))
	count := vec.Vec_IntStart(int(baseabc.Abc_NtkObjNumMax(ntk)))
	for i := 0; i < vec.Vec_WecSize(cubes); i++ {
		cube := vec.Vec_WecEntry(cubes, i)
		id := int(vec.Vec_IntEntry(cube, 0))
		if vec.Vec_IntEntry(count, id) == 0 {
			vec.Vec_IntWriteEntry(first, id, int32(i))
		}
		vec.Vec_IntAddToEntry(count, id, 1)
	}
	present := vec.Vec_IntStartFull(int(baseabc.Abc_NtkObjNumMax(ntk)))
	for _, node := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		oldSop, _ := node.Data.(string)
		baseabc.Abc_ObjRemoveFanins(node)
		n := vec.Vec_IntEntry(count, node.ID)
		assert(n > 0, "each logic node has cubes")
		for k := int32(0); k < n; k++ {
			cube := vec.Vec_WecEntry(cubes, int(vec.Vec_IntEntry(first, node.ID)+k))
			for j := 1; j < vec.Vec_IntSize(cube); j++ {
				lit := vec.Vec_IntEntry(cube, j)
				fanin := ntk.Object(int(util.Abc_Lit2Var(lit)))
				if vec.Vec_IntEntry(present, fanin.ID) >= 0 {
					continue
				}
				vec.Vec_IntWriteEntry(present, fanin.ID, baseabc.Abc_ObjFaninNum(node))
				baseabc.Abc_ObjAddFanin(node, fanin)
			}
		}
		faninNum := baseabc.Abc_ObjFaninNum(node)
		text := []byte(baseabc.Abc_SopStart(nil, n, faninNum))
		stride := int(faninNum + 3)
		for k := int32(0); k < n; k++ {
			cube := vec.Vec_WecEntry(cubes, int(vec.Vec_IntEntry(first, node.ID)+k))
			for j := 1; j < vec.Vec_IntSize(cube); j++ {
				lit := vec.Vec_IntEntry(cube, j)
				fanin := ntk.Object(int(util.Abc_Lit2Var(lit)))
				pin := vec.Vec_IntEntry(present, fanin.ID)
				assert(pin >= 0 && pin < faninNum, "fanin is in the compact fanin list")
				text[int(k)*stride+int(pin)] = '1'
				if util.Abc_LitIsCompl(lit) != 0 {
					text[int(k)*stride+int(pin)] = '0'
				}
			}
		}
		sop := string(text)
		if oldSop != "" && baseabc.Abc_SopIsComplement(oldSop) != 0 {
			sop = baseabc.Abc_SopComplement(sop)
		}
		node.Data = sop
		for _, edge := range node.Fanins() {
			vec.Vec_IntWriteEntry(present, edge.Object.ID, -1)
		}
	}
}

func Abc_NtkFxCheck(ntk *Abc_Ntk_t) int32 {
	for _, node := range ntk.ObjectsOfKind(baseabc.ObjectNode) {
		seen := map[int]bool{}
		for _, edge := range node.Fanins() {
			if seen[edge.Object.ID] {
				return 0
			}
			seen[edge.Object.ID] = true
		}
	}
	return 1
}

func Abc_NtkFxPerform(ntk *Abc_Ntk_t, maxNew, litMax, canonical, verbose, veryVerbose int32) int32 {
	assert(baseabc.Abc_NtkIsSopLogic(ntk) != 0, "Abc_NtkIsSopLogic(pNtk)")
	if Abc_NtkFxCheck(ntk) == 0 {
		fmt.Printf("Abc_NtkFastExtract: Nodes have duplicated fanins. FX is not performed.\n")
		return 0
	}
	cubes := Abc_NtkFxRetrieve(ntk)
	if Fx_FastExtract(cubes, baseabc.Abc_NtkObjNumMax(ntk), maxNew, litMax, canonical, verbose, veryVerbose) > 0 {
		Abc_NtkFxInsert(ntk, cubes)
		vec.Vec_WecFree(cubes)
		if baseabc.Abc_NtkCheck(ntk) == 0 {
			fmt.Printf("Abc_NtkFxPerform: The network check has failed.\n")
		}
		return 1
	}
	fmt.Printf("Warning: The network has not been changed by \"fx\".\n")
	vec.Vec_WecFree(cubes)
	return 0
}

func Fx_ManStart(cubes *vec.Vec_Wec_t) *Fx_Man_t {
	return &Fx_Man_t{vCubes: cubes, vCubesS: vec.Vec_IntAlloc(100), vCubesD: vec.Vec_IntAlloc(100), vCompls: vec.Vec_IntAlloc(100), vCubeFree: vec.Vec_IntAlloc(100), vDiv: vec.Vec_IntAlloc(100), vSCC: vec.Vec_IntAlloc(100)}
}
func Fx_ManStop(p *Fx_Man_t) {
	vec.Vec_WecFree(p.vLits)
	vec.Vec_IntFree(p.vCounts)
	vec.Hsh_VecManStop(p.pHash)
	vec.Vec_FltFree(p.vWeights)
	vec.Vec_QueFree(p.vPrio)
	vec.Vec_IntFree(p.vVarCube)
	vec.Vec_IntFree(p.vLevels)
	vec.Vec_IntFree(p.vCubesS)
	vec.Vec_IntFree(p.vCubesD)
	vec.Vec_IntFree(p.vCompls)
	vec.Vec_IntFree(p.vCubeFree)
	vec.Vec_IntFree(p.vDiv)
	vec.Vec_IntFree(p.vSCC)
}
func Fx_ManComputeLevelDiv(p *Fx_Man_t, cubeFree *vec.Vec_Int_t) int32 {
	level := int32(0)
	for i := 0; i < vec.Vec_IntSize(cubeFree); i++ {
		lit := vec.Vec_IntEntry(cubeFree, i)
		v := util.Abc_Lit2Var(util.Abc_Lit2Var(lit))
		level = max(level, vec.Vec_IntEntry(p.vLevels, int(v)))
	}
	return min(level, int32(800))
}
func Fx_ManComputeLevelCube(p *Fx_Man_t, cube *vec.Vec_Int_t) int32 {
	level := int32(0)
	for i := 1; i < vec.Vec_IntSize(cube); i++ {
		v := util.Abc_Lit2Var(vec.Vec_IntEntry(cube, i))
		level = max(level, vec.Vec_IntEntry(p.vLevels, int(v)))
	}
	return level
}
func Fx_ManComputeLevel(p *Fx_Man_t) {
	assert(vec.Vec_WecSize(p.vCubes) > 0, "nonempty cube set")
	first := vec.Vec_WecEntry(p.vCubes, 0)
	v := vec.Vec_IntEntry(first, 0)
	p.vLevels = vec.Vec_IntStart(int(p.nVars))
	cubeFirst := 0
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		owner := vec.Vec_IntEntry(cube, 0)
		if owner != v {
			vec.Vec_IntAddToEntry(p.vLevels, int(v), int32(i-cubeFirst))
			v = owner
			cubeFirst = i
		}
		vec.Vec_IntUpdateEntry(p.vLevels, int(v), Fx_ManComputeLevelCube(p, cube))
	}
}
func Fx_PrintDivArray(div *vec.Vec_Int_t) {
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		if util.Abc_LitIsCompl(lit) == 0 {
			fmt.Printf("%d(1)", util.Abc_Lit2Var(lit))
		}
	}
	fmt.Printf(" + ")
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		if util.Abc_LitIsCompl(lit) != 0 {
			fmt.Printf("%d(2)", util.Abc_Lit2Var(lit))
		}
	}
}
func Fx_PrintDiv(p *Fx_Man_t, id int32) {
	fmt.Printf("%4d : Div %7d : Weight %12.5f  ", p.nDivs, id, vec.Vec_FltEntry(p.vWeights, int(id)))
	div := vec.Hsh_VecReadEntry(p.pHash, id)
	Fx_PrintDivArray(div)
	for i := vec.Vec_IntSize(div) + 3; i < 16; i++ {
		fmt.Print(" ")
	}
	fmt.Printf("Lits =%7d  Divs =%8d  ", p.nLits, vec.Hsh_VecSize(p.pHash))
	util.Abc_PrintTime(1, "Time", util.Abc_Clock()-p.timeStart)
}
func Fx_PrintDivisors(p *Fx_Man_t) {
	for i := 0; i < vec.Vec_FltSize(p.vWeights); i++ {
		Fx_PrintDiv(p, int32(i))
	}
}
func Fx_PrintStats(p *Fx_Man_t, elapsed int64) {
	fmt.Printf("Cubes =%8d  Lits  =%8d  Divs  =%8d  Divs+ =%8d  Compl =%8d  Extr  =%7d  ", vec.Vec_WecSizeUsed(p.vCubes), vec.Vec_WecSizeUsed(p.vLits), vec.Hsh_VecSize(p.pHash), vec.Vec_QueSize(p.vPrio), p.nDivMux[1], p.nDivs)
	util.Abc_PrintTime(1, "Time", elapsed)
}

func Fx_ManDivNormalize(cubeFree *vec.Vec_Int_t) int32 {
	l := vec.Vec_IntArray(cubeFree)
	assert(len(l) == 4, "four-literal canonical divisor")
	var a0, b0, a1, b1 int32 = -1, -1, -1, -1
	compl := func(x int32) bool { return util.Abc_LitIsCompl(x) != 0 }
	switch {
	case compl(l[0]) != compl(l[1]) && l[0]>>2 == l[1]>>2:
		if compl(l[2]) == compl(l[3]) {
			return -1
		}
		a0, b0 = util.Abc_Lit2Var(l[0]), util.Abc_Lit2Var(l[1])
		if compl(l[0]) == compl(l[2]) {
			a1, b1 = util.Abc_Lit2Var(l[2]), util.Abc_Lit2Var(l[3])
		} else {
			assert(compl(l[0]) == compl(l[3]) && compl(l[1]) == compl(l[2]), "MUX phase layout")
			a1, b1 = util.Abc_Lit2Var(l[3]), util.Abc_Lit2Var(l[2])
		}
	case compl(l[1]) != compl(l[2]) && l[1]>>2 == l[2]>>2:
		if compl(l[0]) == compl(l[3]) {
			return -1
		}
		a0, b0 = util.Abc_Lit2Var(l[1]), util.Abc_Lit2Var(l[2])
		if compl(l[1]) == compl(l[0]) {
			a1, b1 = util.Abc_Lit2Var(l[0]), util.Abc_Lit2Var(l[3])
		} else {
			a1, b1 = util.Abc_Lit2Var(l[3]), util.Abc_Lit2Var(l[0])
		}
	case compl(l[2]) != compl(l[3]) && l[2]>>2 == l[3]>>2:
		if compl(l[0]) == compl(l[1]) {
			return -1
		}
		a0, b0 = util.Abc_Lit2Var(l[2]), util.Abc_Lit2Var(l[3])
		if compl(l[2]) == compl(l[0]) {
			a1, b1 = util.Abc_Lit2Var(l[0]), util.Abc_Lit2Var(l[1])
		} else {
			a1, b1 = util.Abc_Lit2Var(l[1]), util.Abc_Lit2Var(l[0])
		}
	default:
		return -1
	}
	assert(a0 == util.Abc_LitNot(b0), "control literals complementary")
	if compl(a0) {
		a0, b0 = b0, a0
		a1, b1 = b1, a1
	}
	assert(!compl(a0), "positive control phase")
	ret := int32(0)
	if compl(a1) {
		a1 = util.Abc_LitNot(a1)
		b1 = util.Abc_LitNot(b1)
		ret = 1
	}
	assert(!compl(a1), "positive data phase")
	l[0] = util.Abc_Var2Lit(a0, 0)
	l[1] = util.Abc_Var2Lit(b0, 1)
	l[2] = util.Abc_Var2Lit(a1, 0)
	l[3] = util.Abc_Var2Lit(b1, 1)
	return ret
}

func Fx_ManDivFindCubeFree(a, b, free *vec.Vec_Int_t, warning *int32) int32 {
	x := vec.Vec_IntArray(a)[1:]
	y := vec.Vec_IntArray(b)[1:]
	i, j := 0, 0
	counter := int32(0)
	attr0, attr1 := int32(0), int32(1)
	vec.Vec_IntClear(free)
	for i < len(x) && j < len(y) {
		if x[i] == y[j] {
			i++
			j++
			counter++
		} else if x[i] < y[j] {
			vec.Vec_IntPush(free, util.Abc_Var2Lit(x[i], attr0))
			i++
		} else {
			if vec.Vec_IntSize(free) == 0 {
				attr0, attr1 = 1, 0
			}
			vec.Vec_IntPush(free, util.Abc_Var2Lit(y[j], attr1))
			j++
		}
	}
	for i < len(x) {
		vec.Vec_IntPush(free, util.Abc_Var2Lit(x[i], attr0))
		i++
	}
	for j < len(y) {
		vec.Vec_IntPush(free, util.Abc_Var2Lit(y[j], attr1))
		j++
	}
	switch vec.Vec_IntSize(free) {
	case 0:
		fmt.Printf("The SOP has duplicated cubes.\n")
	case 1:
		return -1
	case 3:
		l := vec.Vec_IntArray(free)
		drop := -1
		if util.Abc_Lit2Var(l[0]) == util.Abc_LitNot(util.Abc_Lit2Var(l[1])) {
			if util.Abc_LitIsCompl(l[0]) == util.Abc_LitIsCompl(l[2]) {
				drop = 0
			} else {
				drop = 1
			}
		} else if util.Abc_Lit2Var(l[1]) == util.Abc_LitNot(util.Abc_Lit2Var(l[2])) {
			if util.Abc_LitIsCompl(l[1]) == util.Abc_LitIsCompl(l[0]) {
				drop = 1
			} else {
				drop = 2
			}
		}
		if drop >= 0 {
			vec.Vec_IntDrop(free, drop)
		}
		if vec.Vec_IntSize(free) == 2 {
			l = vec.Vec_IntArray(free)
			u, v := util.Abc_Lit2Var(l[0]), util.Abc_Lit2Var(l[1])
			if u > v {
				u, v = v, u
			}
			vec.Vec_IntWriteEntry(free, 0, util.Abc_Var2Lit(u, 0))
			vec.Vec_IntWriteEntry(free, 1, util.Abc_Var2Lit(v, 1))
		}
	}
	assert(util.Abc_LitIsCompl(vec.Vec_IntEntry(free, 0)) == 0, "normalized leading phase")
	return counter
}

func Fx_ManDivFindPivots(div *vec.Vec_Int_t) (int32, int32) {
	a, b := int32(-1), int32(-1)
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		if util.Abc_LitIsCompl(lit) != 0 {
			if b < 0 {
				b = util.Abc_Lit2Var(lit)
			}
		} else if a < 0 {
			a = util.Abc_Lit2Var(lit)
		}
		if a >= 0 && b >= 0 {
			return a, b
		}
	}
	return a, b
}
func Fx_ManDivRemoveLits(cube, div *vec.Vec_Int_t, complemented int32) int32 {
	assert(complemented == 0 || vec.Vec_IntSize(div) == 4, "complemented literals require 4-cube divisor")
	count := int32(0)
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		v := util.Abc_Lit2Var(lit)
		if complemented != 0 && i > 1 {
			v = util.Abc_LitNot(v)
		}
		count += int32(vec.Vec_IntRemove1(cube, v))
		if vec.Vec_IntSize(div) == 2 {
			count += int32(vec.Vec_IntRemove1(cube, util.Abc_LitNot(util.Abc_Lit2Var(lit))))
		}
	}
	return count
}
func Fx_ManDivAddLits(cube, cube2, div *vec.Vec_Int_t) {
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		if util.Abc_LitIsCompl(lit) != 0 {
			vec.Vec_IntPush(cube2, util.Abc_Lit2Var(lit))
		} else {
			vec.Vec_IntPush(cube, util.Abc_Lit2Var(lit))
		}
	}
	if vec.Vec_IntSize(div) == 4 && vec.Vec_IntSize(cube) == 3 {
		a := vec.Vec_IntArray(cube)
		if a[1] > a[2] {
			a[1], a[2] = a[2], a[1]
		}
		b := vec.Vec_IntArray(cube2)
		if b[1] > b[2] {
			b[1], b[2] = b[2], b[1]
		}
	}
}

func Fx_ManCreateLiterals(p *Fx_Man_t, nVars int32) {
	p.nVars, p.nLits = 0, 0
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		assert(vec.Vec_IntSize(cube) > 0, "nonempty cube")
		p.nVars = max(p.nVars, vec.Vec_IntEntry(cube, 0))
		p.nLits += int32(vec.Vec_IntSize(cube) - 1)
		for k := 1; k < vec.Vec_IntSize(cube); k++ {
			p.nVars = max(p.nVars, util.Abc_Lit2Var(vec.Vec_IntEntry(cube, k)))
		}
	}
	assert(p.nVars < nVars, "ObjIdMax exceeds all cube object IDs")
	p.nVars = nVars
	p.vCounts = vec.Vec_IntStart(int(2 * p.nVars))
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		for k := 1; k < vec.Vec_IntSize(cube); k++ {
			vec.Vec_IntAddToEntry(p.vCounts, int(vec.Vec_IntEntry(cube, k)), 1)
		}
	}
	p.vLits = vec.Vec_WecStart(int(2 * p.nVars))
	for lit := 0; lit < vec.Vec_IntSize(p.vCounts); lit++ {
		vec.Vec_IntGrow(vec.Vec_WecEntry(p.vLits, lit), int(vec.Vec_IntEntry(p.vCounts, lit)))
	}
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		for k := 1; k < vec.Vec_IntSize(cube); k++ {
			vec.Vec_WecPush(p.vLits, int(vec.Vec_IntEntry(cube, k)), int32(i))
		}
	}
	p.vVarCube = vec.Vec_IntStartFull(int(p.nVars))
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		owner := vec.Vec_IntEntry(cube, 0)
		if vec.Vec_IntEntry(p.vVarCube, int(owner)) == -1 {
			vec.Vec_IntWriteEntry(p.vVarCube, int(owner), int32(i))
		}
	}
}
func Fx_ManCubeSingleCubeDivisors(p *Fx_Man_t, pivot *vec.Vec_Int_t, remove, update int32) int32 {
	n := vec.Vec_IntSize(pivot)
	if n < 2 {
		return 0
	}
	for k := 1; k < n; k++ {
		lit := vec.Vec_IntEntry(pivot, k)
		for j := k + 1; j < n; j++ {
			lit2 := vec.Vec_IntEntry(pivot, j)
			assert(lit < lit2, "sorted pivot literals")
			vec.Vec_IntClear(p.vCubeFree)
			vec.Vec_IntPush(p.vCubeFree, util.Abc_Var2Lit(util.Abc_LitNot(lit), 0))
			vec.Vec_IntPush(p.vCubeFree, util.Abc_Var2Lit(util.Abc_LitNot(lit2), 1))
			id := vec.Hsh_VecManAdd(p.pHash, p.vCubeFree)
			if remove == 0 {
				if vec.Vec_FltSize(p.vWeights) == int(id) {
					vec.Vec_FltPush(p.vWeights, -2+0.9-0.001*float32(Fx_ManComputeLevelDiv(p, p.vCubeFree)))
					p.nDivsS++
				}
				assert(int(id) < vec.Vec_FltSize(p.vWeights), "weight array aligned with hash")
				vec.Vec_FltAddToEntry(p.vWeights, int(id), 1)
				p.nPairsS++
			} else {
				assert(int(id) < vec.Vec_FltSize(p.vWeights), "known divisor weight")
				vec.Vec_FltAddToEntry(p.vWeights, int(id), -1)
				p.nPairsS--
			}
			if update != 0 {
				if vec.Vec_QueIsMember(p.vPrio, int(id)) != 0 {
					vec.Vec_QueUpdate(p.vPrio, int(id))
				} else if remove == 0 {
					vec.Vec_QuePush(p.vPrio, int(id))
				}
			}
		}
	}
	return int32(n * (n - 1) / 2)
}
func Fx_ManCubeDoubleCubeDivisors(p *Fx_Man_t, first int32, pivot *vec.Vec_Int_t, remove, update int32, warning *int32) {
	owner := vec.Vec_IntEntry(pivot, 0)
	for i := int(first); i < vec.Vec_WecSize(p.vCubes); i++ {
		cube := vec.Vec_WecEntry(p.vCubes, i)
		if vec.Vec_IntSize(cube) == 0 || cube == pivot {
			continue
		}
		if vec.Vec_WecIntHasMark(cube) != 0 && vec.Vec_WecIntHasMark(pivot) != 0 && i > vec.Vec_WecLevelId(p.vCubes, pivot) {
			continue
		}
		if vec.Vec_IntEntry(cube, 0) != owner {
			break
		}
		base := Fx_ManDivFindCubeFree(cube, pivot, p.vCubeFree, warning)
		if base == -1 {
			if remove == 0 {
				if vec.Vec_IntSize(cube) > vec.Vec_IntSize(pivot) {
					vec.Vec_IntPush(p.vSCC, int32(vec.Vec_WecLevelId(p.vCubes, cube)))
				} else {
					vec.Vec_IntPush(p.vSCC, int32(vec.Vec_WecLevelId(p.vCubes, pivot)))
				}
			}
			continue
		}
		if vec.Vec_IntSize(p.vCubeFree) == 4 {
			value := Fx_ManDivNormalize(p.vCubeFree)
			if value == 0 {
				p.nDivMux[0]++
			} else if value == 1 {
				p.nDivMux[1]++
			} else {
				p.nDivMux[2]++
			}
			if p.fCanonDivs != 0 && value < 0 {
				continue
			}
		}
		if p.LitCountMax != 0 && p.LitCountMax < int32(vec.Vec_IntSize(p.vCubeFree)) {
			continue
		}
		if p.fCanonDivs != 0 && vec.Vec_IntSize(p.vCubeFree) == 3 {
			continue
		}
		id := vec.Hsh_VecManAdd(p.pHash, p.vCubeFree)
		cost := int32(base + int32(vec.Vec_IntSize(p.vCubeFree)) - 1)
		if remove == 0 {
			if int(id) == vec.Vec_FltSize(p.vWeights) {
				vec.Vec_FltPush(p.vWeights, -float32(vec.Vec_IntSize(p.vCubeFree))+0.9-0.0009*float32(Fx_ManComputeLevelDiv(p, p.vCubeFree)))
			}
			assert(int(id) < vec.Vec_FltSize(p.vWeights), "weight array aligned with hash")
			vec.Vec_FltAddToEntry(p.vWeights, int(id), float32(cost))
			p.nPairsD++
		} else {
			assert(int(id) < vec.Vec_FltSize(p.vWeights), "known divisor weight")
			vec.Vec_FltAddToEntry(p.vWeights, int(id), -float32(cost))
			p.nPairsD--
		}
		if update != 0 {
			if vec.Vec_QueIsMember(p.vPrio, int(id)) != 0 {
				vec.Vec_QueUpdate(p.vPrio, int(id))
			} else if remove == 0 {
				vec.Vec_QuePush(p.vPrio, int(id))
			}
		}
	}
}
func Fx_ManCreateDivisors(p *Fx_Man_t) {
	assert(p.pHash == nil, "divisor hash is initialized once")
	p.pHash = vec.Hsh_VecManStart(1000)
	p.vWeights = vec.Vec_FltAlloc(1000)
	warning := int32(0)
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		Fx_ManCubeSingleCubeDivisors(p, vec.Vec_WecEntry(p.vCubes, i), 0, 0)
	}
	assert(int32(p.nDivsS) == int32(vec.Vec_FltSize(p.vWeights)), "single divisor indexes are first")
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		Fx_ManCubeDoubleCubeDivisors(p, int32(i+1), vec.Vec_WecEntry(p.vCubes, i), 0, 0, &warning)
	}
	p.vPrio = vec.Vec_QueAlloc(vec.Vec_FltSize(p.vWeights))
	vec.Vec_QueSetPriority(p.vPrio, vec.Vec_FltArrayP(p.vWeights))
	for i := 0; i < vec.Vec_FltSize(p.vWeights); i++ {
		if vec.Vec_FltEntry(p.vWeights, i) > 0 {
			vec.Vec_QuePush(p.vPrio, i)
		}
	}
}

func Fx_ManCompressCubes(cubes *vec.Vec_Wec_t, lits *vec.Vec_Int_t) {
	k := 0
	for i := 0; i < vec.Vec_IntSize(lits); i++ {
		id := vec.Vec_IntEntry(lits, i)
		if vec.Vec_IntSize(vec.Vec_WecEntry(cubes, int(id))) > 0 {
			vec.Vec_IntWriteEntry(lits, k, id)
			k++
		}
	}
	vec.Vec_IntShrink(lits, k)
}
func fxMarkCubeIDs(cubes *vec.Vec_Wec_t, ids *vec.Vec_Int_t, marked bool) {
	for i := 0; i < vec.Vec_IntSize(ids); i++ {
		cube := vec.Vec_WecEntry(cubes, int(vec.Vec_IntEntry(ids, i)))
		isMarked := vec.Vec_WecIntHasMark(cube) != 0
		if isMarked == marked {
			continue
		}
		vec.Vec_WecIntXorMark(cube)
	}
}
func Fx_ManGetCubeVar(cubes *vec.Vec_Wec_t, id int32) int32 {
	return vec.Vec_IntEntry(vec.Vec_WecEntry(cubes, int(id)), 0)
}
func Fx_ManFindCommonPairs(cubes *vec.Vec_Wec_t, part0, part1, pairs, compls, div, free *vec.Vec_Int_t, warning *int32) {
	vec.Vec_IntClear(pairs)
	vec.Vec_IntClear(compls)
	i, j := 0, 0
	for i < vec.Vec_IntSize(part0) && j < vec.Vec_IntSize(part1) {
		id0, id1 := vec.Vec_IntEntry(part0, i), vec.Vec_IntEntry(part1, j)
		owner0, owner1 := Fx_ManGetCubeVar(cubes, id0), Fx_ManGetCubeVar(cubes, id1)
		if owner0 == owner1 {
			iEnd, jEnd := i+1, j+1
			for iEnd < vec.Vec_IntSize(part0) && Fx_ManGetCubeVar(cubes, vec.Vec_IntEntry(part0, iEnd)) == owner0 {
				iEnd++
			}
			for jEnd < vec.Vec_IntSize(part1) && Fx_ManGetCubeVar(cubes, vec.Vec_IntEntry(part1, jEnd)) == owner0 {
				jEnd++
			}
			for a := i; a < iEnd; a++ {
				for b := j; b < jEnd; b++ {
					c0, c1 := vec.Vec_IntEntry(part0, a), vec.Vec_IntEntry(part1, b)
					if c0 == c1 {
						continue
					}
					Fx_ManDivFindCubeFree(vec.Vec_WecEntry(cubes, int(c0)), vec.Vec_WecEntry(cubes, int(c1)), free, warning)
					fCompl := int32(0)
					if vec.Vec_IntSize(free) == 4 && Fx_ManDivNormalize(free) == 1 {
						fCompl = 1
					}
					if vec.Vec_IntEqual(div, free) == 0 {
						continue
					}
					vec.Vec_IntPush(pairs, c0)
					vec.Vec_IntPush(pairs, c1)
					vec.Vec_IntPush(compls, fCompl)
				}
			}
			i, j = iEnd, jEnd
		} else if owner0 < owner1 {
			i++
		} else {
			j++
		}
	}
}

func Fx_ManUpdate(p *Fx_Man_t, divID int32, warning *int32) {
	var cube, cube2, litPos, litNeg *vec.Vec_Int_t
	div := p.vDiv
	diff := vec.Vec_FltEntry(p.vWeights, int(divID)) - float32(int32(vec.Vec_FltEntry(p.vWeights, int(divID))))
	assert(diff > 0 && diff < 1, "divisor priority has fractional tie-breaker")
	p.nDivs++
	vec.Vec_IntClear(div)
	vec.Vec_IntAppend(div, vec.Hsh_VecReadEntry(p.pHash, divID))
	lit0, lit1 := Fx_ManDivFindPivots(div)
	assert(lit0 >= 0 && lit1 >= 0, "both divisor phases are present")
	vec.Vec_IntClear(p.vCubesS)
	if vec.Vec_IntSize(div) == 2 {
		vecLit0, vecLit1 := vec.Vec_WecEntry(p.vLits, int(util.Abc_LitNot(lit0))), vec.Vec_WecEntry(p.vLits, int(util.Abc_LitNot(lit1)))
		Fx_ManCompressCubes(p.vCubes, vecLit0)
		Fx_ManCompressCubes(p.vCubes, vecLit1)
		vec.Vec_IntTwoRemoveCommon(vecLit0, vecLit1, p.vCubesS)
	}
	part0, part1 := vec.Vec_WecEntry(p.vLits, int(lit0)), vec.Vec_WecEntry(p.vLits, int(lit1))
	Fx_ManCompressCubes(p.vCubes, part0)
	Fx_ManCompressCubes(p.vCubes, part1)
	Fx_ManFindCommonPairs(p.vCubes, part0, part1, p.vCubesD, p.vCompls, div, p.vCubeFree, warning)
	for i := 0; i < vec.Vec_IntSize(p.vCubesS); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesS, i)))
		Fx_ManCubeSingleCubeDivisors(p, cube, 1, 1)
	}
	for i := 0; i < vec.Vec_IntSize(p.vCubesD); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i)))
		Fx_ManCubeSingleCubeDivisors(p, cube, 1, 1)
	}
	fxMarkCubeIDs(p.vCubes, p.vCubesS, true)
	fxMarkCubeIDs(p.vCubes, p.vCubesD, true)
	for i := 0; i < vec.Vec_IntSize(p.vCubesS); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesS, i)))
		Fx_ManCubeDoubleCubeDivisors(p, Fx_ManGetFirstVarCube(p, cube), cube, 1, 1, warning)
	}
	for i := 0; i < vec.Vec_IntSize(p.vCubesD); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i)))
		Fx_ManCubeDoubleCubeDivisors(p, Fx_ManGetFirstVarCube(p, cube), cube, 1, 1, warning)
	}
	fxMarkCubeIDs(p.vCubes, p.vCubesS, false)
	fxMarkCubeIDs(p.vCubes, p.vCubesD, false)
	newVar := int32(0)
	if util.Abc_Lit2Var(lit0) == util.Abc_Lit2Var(lit1) && vec.Vec_IntSize(vec.Hsh_VecReadEntry(p.pHash, divID)) == 2 {
		goto extractPairs
	}
	newVar = int32(vec.Vec_WecSize(p.vLits) / 2)
	assert(vec.Vec_IntSize(p.vVarCube) == int(newVar), "new variable IDs append")
	vec.Vec_IntPush(p.vVarCube, int32(vec.Vec_WecSize(p.vCubes)))
	cube = vec.Vec_WecPushLevel(p.vCubes)
	vec.Vec_IntPush(cube, newVar)
	if vec.Vec_IntSize(div) == 2 {
		vec.Vec_IntPush(cube, util.Abc_LitNot(lit0))
		vec.Vec_IntPush(cube, util.Abc_LitNot(lit1))
		level := 1 + Fx_ManComputeLevelCube(p, cube)
		assert(vec.Vec_IntSize(p.vLevels) == int(newVar), "new level index appends")
		vec.Vec_IntPush(p.vLevels, level)
	} else {
		cube2 = vec.Vec_WecPushLevel(p.vCubes)
		cube = vec.Vec_WecEntry(p.vCubes, vec.Vec_WecSize(p.vCubes)-2)
		vec.Vec_IntPush(cube2, newVar)
		Fx_ManDivAddLits(cube, cube2, div)
		level := 2 + max(Fx_ManComputeLevelCube(p, cube), Fx_ManComputeLevelCube(p, cube2))
		assert(vec.Vec_IntSize(p.vLevels) == int(newVar), "new level index appends")
		vec.Vec_IntPush(p.vLevels, level)
	}
	p.nLits += int32(vec.Vec_IntSize(div))
	vec.Vec_WecPushLevel(p.vLits)
	vec.Vec_WecPushLevel(p.vLits)
	litPos = vec.Vec_WecEntry(p.vLits, vec.Vec_WecSize(p.vLits)-2)
	litNeg = vec.Vec_WecEntry(p.vLits, vec.Vec_WecSize(p.vLits)-1)
	for i := 0; i < vec.Vec_IntSize(p.vCubesS); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesS, i)))
		removed := int32(vec.Vec_IntRemove1(cube, util.Abc_LitNot(lit0))) + int32(vec.Vec_IntRemove1(cube, util.Abc_LitNot(lit1)))
		assert(removed == 2, "both single-cube divisor literals removed")
		vec.Vec_IntPush(cube, util.Abc_Var2Lit(newVar, 0))
		vec.Vec_IntPush(litPos, int32(vec.Vec_WecLevelId(p.vCubes, cube)))
		p.nLits--
	}
extractPairs:
	k := 0
	p.nCompls = 0
	assert(vec.Vec_IntSize(p.vCubesD)%2 == 0 && vec.Vec_IntSize(p.vCubesD) == 2*vec.Vec_IntSize(p.vCompls), "pair/complement arrays align")
	for i := 0; i < vec.Vec_IntSize(p.vCubesD); i += 2 {
		fCompl := vec.Vec_IntEntry(p.vCompls, i/2)
		p.nCompls += fCompl
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i)))
		cube2 = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i+1)))
		removed := Fx_ManDivRemoveLits(cube, div, fCompl) + Fx_ManDivRemoveLits(cube2, div, fCompl)
		assert(removed == int32(vec.Vec_IntSize(div)) || removed == int32(vec.Vec_IntSize(div)+1), "paired cubes contain divisor")
		if newVar > 0 {
			if vec.Vec_IntSize(div) == 2 || fCompl != 0 {
				vec.Vec_IntPush(cube, util.Abc_Var2Lit(newVar, 1))
				vec.Vec_IntPush(litNeg, int32(vec.Vec_WecLevelId(p.vCubes, cube)))
			} else {
				vec.Vec_IntPush(cube, util.Abc_Var2Lit(newVar, 0))
				vec.Vec_IntPush(litPos, int32(vec.Vec_WecLevelId(p.vCubes, cube)))
			}
		}
		p.nLits -= int32(vec.Vec_IntSize(div) + vec.Vec_IntSize(cube2) - 2)
		vec.Vec_IntWriteEntry(p.vCubesD, k, int32(vec.Vec_WecLevelId(p.vCubes, cube)))
		k++
		vec.Vec_IntClear(cube2)
	}
	assert(k == vec.Vec_IntSize(p.vCubesD)/2, "each pair leaves one cube")
	vec.Vec_IntShrink(p.vCubesD, k)
	vec.Vec_IntSort(p.vCubesD, 0)
	for i := 0; i < vec.Vec_IntSize(p.vCubesS); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesS, i)))
		Fx_ManCubeSingleCubeDivisors(p, cube, 0, 1)
	}
	for i := 0; i < vec.Vec_IntSize(p.vCubesD); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i)))
		Fx_ManCubeSingleCubeDivisors(p, cube, 0, 1)
	}
	fxMarkCubeIDs(p.vCubes, p.vCubesS, true)
	fxMarkCubeIDs(p.vCubes, p.vCubesD, true)
	for i := 0; i < vec.Vec_IntSize(p.vCubesS); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesS, i)))
		Fx_ManCubeDoubleCubeDivisors(p, Fx_ManGetFirstVarCube(p, cube), cube, 0, 1, warning)
	}
	for i := 0; i < vec.Vec_IntSize(p.vCubesD); i++ {
		cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vCubesD, i)))
		Fx_ManCubeDoubleCubeDivisors(p, Fx_ManGetFirstVarCube(p, cube), cube, 0, 1, warning)
	}
	fxMarkCubeIDs(p.vCubes, p.vCubesS, false)
	fxMarkCubeIDs(p.vCubes, p.vCubesD, false)
	if vec.Vec_IntSize(p.vSCC) > 0 {
		vec.Vec_IntUniqify(p.vSCC)
		for i := 0; i < vec.Vec_IntSize(p.vSCC); i++ {
			cube = vec.Vec_WecEntry(p.vCubes, int(vec.Vec_IntEntry(p.vSCC, i)))
			Fx_ManCubeDoubleCubeDivisors(p, Fx_ManGetFirstVarCube(p, cube), cube, 1, 1, warning)
			vec.Vec_IntClear(cube)
		}
		vec.Vec_IntClear(p.vSCC)
	}
	if vec.Vec_IntSize(div) > 2 {
		cube = vec.Vec_WecEntry(p.vCubes, vec.Vec_WecSize(p.vCubes)-2)
		cube2 = vec.Vec_WecEntry(p.vCubes, vec.Vec_WecSize(p.vCubes)-1)
		Fx_ManCubeSingleCubeDivisors(p, cube, 0, 1)
		Fx_ManCubeSingleCubeDivisors(p, cube2, 0, 1)
		for i := 1; i < vec.Vec_IntSize(cube); i++ {
			lit := vec.Vec_IntEntry(cube, i)
			vec.Vec_WecPush(p.vLits, int(lit), int32(vec.Vec_WecLevelId(p.vCubes, cube)))
		}
		for i := 1; i < vec.Vec_IntSize(cube2); i++ {
			lit := vec.Vec_IntEntry(cube2, i)
			vec.Vec_WecPush(p.vLits, int(lit), int32(vec.Vec_WecLevelId(p.vCubes, cube2)))
		}
	}
	for i := 0; i < vec.Vec_IntSize(div); i++ {
		lit := vec.Vec_IntEntry(div, i)
		vec.Vec_IntTwoRemove(vec.Vec_WecEntry(p.vLits, int(util.Abc_Lit2Var(lit))), p.vCubesD)
		if p.nCompls != 0 && i > 1 || vec.Vec_IntSize(div) == 2 {
			vec.Vec_IntTwoRemove(vec.Vec_WecEntry(p.vLits, int(util.Abc_LitNot(util.Abc_Lit2Var(lit)))), p.vCubesD)
		}
	}
}

func Fx_FastExtract(cubes *vec.Vec_Wec_t, objIDMax, newNodesMax, litMax, canonical, verbose, veryVerbose int32) int32 {
	started := util.Abc_Clock()
	p := Fx_ManStart(cubes)
	p.LitCountMax = litMax
	p.fCanonDivs = canonical
	Fx_ManCreateLiterals(p, objIDMax)
	Fx_ManComputeLevel(p)
	Fx_ManCreateDivisors(p)
	if veryVerbose != 0 {
		Fx_PrintDivisors(p)
	}
	if verbose != 0 {
		Fx_PrintStats(p, util.Abc_Clock()-started)
	}
	p.timeStart = util.Abc_Clock()
	warning := int32(0)
	for i := int32(0); i < newNodesMax && vec.Vec_QueSize(p.vPrio) > 0 && vec.Vec_QueTopPriority(p.vPrio) > 0; i++ {
		id := int32(vec.Vec_QuePop(p.vPrio))
		if veryVerbose != 0 {
			Fx_PrintDiv(p, id)
		}
		Fx_ManUpdate(p, id, &warning)
	}
	if verbose != 0 {
		Fx_PrintStats(p, util.Abc_Clock()-started)
	}
	Fx_ManStop(p)
	vec.Vec_WecRemoveEmpty(cubes)
	return 1
}
