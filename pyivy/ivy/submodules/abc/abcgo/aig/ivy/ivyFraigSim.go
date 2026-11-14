// Word simulation and candidate classes ported from ABC ivyFraig.c.
package ivy

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/sat/bsat"
	"math/bits"
	"math/rand/v2"
	"slices"
	"time"
)

type Ivy_FraigList_t struct {
	pHead, pTail *Ivy_Obj_t
	nItems       int32
}
type Ivy_FraigSim_t struct {
	Type                    int32
	pNext, pFanin0, pFanin1 *Ivy_FraigSim_t
	pData                   []uint32
}
type ivyFraigNodeState struct {
	sim                                                        *Ivy_FraigSim_t
	classLink, classNext, hashNext, listNext, listPrev, mapped *Ivy_Obj_t
	satNum                                                     int32
	fanins                                                     []*Ivy_Obj_t
}
type Ivy_FraigMan_t struct {
	pSat                                                                              *bsat.Solver
	nSatVars                                                                          int32
	vPiVars                                                                           []*Ivy_Obj_t
	nBTLimitGlobal, nInsLimitGlobal                                                   int64
	nSatCalls, nSatCallsSat, nSatCallsUnsat, nSatProof, nSatFails, nSatFailsReal      int32
	nNodesMiter, nClassesZero, nClassesBeg, nClassesEnd, nPairsBeg, nPairsEnd         int32
	timeSat, timeSatSat, timeSatUnsat, timeSatFail, timeTrav, timeTotal, time1, time2 time.Duration
	pParams                                                                           *Ivy_FraigParams_t
	pManAig, pManFraig                                                                *Ivy_Man_t
	nSimWords                                                                         int32
	pSimStart                                                                         *Ivy_FraigSim_t
	pPatWords                                                                         []uint32
	pPatScores                                                                        []int32
	lClasses, lCand                                                                   Ivy_FraigList_t
	nPairs, nSimRounds                                                                int32
	timeSim, timeRef                                                                  time.Duration
	rng                                                                               *rand.Rand
}

func ivyFraigState(n *Ivy_Obj_t) *ivyFraigNodeState {
	n = Ivy_Regular(n)
	if n == nil || n.recycled {
		panic("ivy: invalid fraig node")
	}
	if n.fraig == nil {
		n.fraig = new(ivyFraigNodeState)
	}
	return n.fraig
}
func Ivy_ObjSim(n *Ivy_Obj_t) *Ivy_FraigSim_t         { return ivyFraigState(n).sim }
func Ivy_ObjClassNodeLast(n *Ivy_Obj_t) *Ivy_Obj_t    { return ivyFraigState(n).classLink }
func Ivy_ObjClassNodeRepr(n *Ivy_Obj_t) *Ivy_Obj_t    { return ivyFraigState(n).classLink }
func Ivy_ObjClassNodeNext(n *Ivy_Obj_t) *Ivy_Obj_t    { return ivyFraigState(n).classNext }
func Ivy_ObjNodeHashNext(n *Ivy_Obj_t) *Ivy_Obj_t     { return ivyFraigState(n).hashNext }
func Ivy_ObjEquivListNext(n *Ivy_Obj_t) *Ivy_Obj_t    { return ivyFraigState(n).listNext }
func Ivy_ObjEquivListPrev(n *Ivy_Obj_t) *Ivy_Obj_t    { return ivyFraigState(n).listPrev }
func Ivy_ObjFraig(n *Ivy_Obj_t) *Ivy_Obj_t            { return ivyFraigState(n).mapped }
func Ivy_ObjSatNum(n *Ivy_Obj_t) int32                { return ivyFraigState(n).satNum }
func Ivy_ObjFaninVec(n *Ivy_Obj_t) []*Ivy_Obj_t       { return ivyFraigState(n).fanins }
func Ivy_ObjSetSim(n *Ivy_Obj_t, v *Ivy_FraigSim_t)   { ivyFraigState(n).sim = v }
func Ivy_ObjSetClassNodeLast(n, v *Ivy_Obj_t)         { ivyFraigState(n).classLink = v }
func Ivy_ObjSetClassNodeRepr(n, v *Ivy_Obj_t)         { ivyFraigState(n).classLink = v }
func Ivy_ObjSetClassNodeNext(n, v *Ivy_Obj_t)         { ivyFraigState(n).classNext = v }
func Ivy_ObjSetNodeHashNext(n, v *Ivy_Obj_t)          { ivyFraigState(n).hashNext = v }
func Ivy_ObjSetEquivListNext(n, v *Ivy_Obj_t)         { ivyFraigState(n).listNext = v }
func Ivy_ObjSetEquivListPrev(n, v *Ivy_Obj_t)         { ivyFraigState(n).listPrev = v }
func Ivy_ObjSetFraig(n, v *Ivy_Obj_t)                 { ivyFraigState(n).mapped = v }
func Ivy_ObjSetSatNum(n *Ivy_Obj_t, v int32)          { ivyFraigState(n).satNum = v }
func Ivy_ObjSetFaninVec(n *Ivy_Obj_t, v []*Ivy_Obj_t) { ivyFraigState(n).fanins = v }
func Ivy_ObjRandomSim() uint32                        { return rand.Uint32() }
func Ivy_FraigParamsDefault(params *Ivy_FraigParams_t) {
	*params = Ivy_FraigParams_t{nSimWords: 32, dSimSatur: 0.005, MaxScore: 25, FDoSparse: 1, dActConeRatio: 0.3, dActConeBumpMax: 10, NBTLimitNode: 100, NBTLimitMiter: 500000}
}

// Ivy_FraigParamsSetBmc configures the FRAIG options used by ABC's bounded
// model checker. The source sets an unlimited node backtrack limit and turns
// proving and verbose reporting off.
func Ivy_FraigParamsSetBmc(params *Ivy_FraigParams_t, nodeLimit int32, verbose, prove int32) {
	if params == nil {
		panic("ivy: nil FRAIG parameters")
	}
	params.NBTLimitNode = nodeLimit
	params.FVerbose = verbose
	params.FProve = prove
}

// This allocator preserves the real graph fanout lists: FRAIG owns typed state
// instead of reinterpreting fanout pointers as simulation blocks and integers.
func ivyFraigStartSimulation(source *Ivy_Man_t, params *Ivy_FraigParams_t) *Ivy_FraigMan_t {
	if params == nil || params.nSimWords < 1 || params.nSimWords > 128 || Ivy_ManCheck(source) == 0 || Ivy_ManLatchNum(source) != 0 || Ivy_ManBufNum(source) != 0 {
		panic("ivy: invalid fraig simulation request")
	}
	for _, n := range source.vObjs {
		if n != nil && n.fraig != nil {
			panic("ivy: fraig state already in use")
		}
	}
	p := &Ivy_FraigMan_t{pParams: params, pManAig: source, pManFraig: Ivy_ManStartFrom(source), nSimWords: params.nSimWords, rng: rand.New(rand.NewPCG(0xabcabc, 0xabcabc))}
	for _, n := range source.vObjs {
		if n != nil {
			if n.fraig != nil {
				panic("ivy: fraig state already in use")
			}
			n.fraig = &ivyFraigNodeState{sim: &Ivy_FraigSim_t{pData: make([]uint32, params.nSimWords)}}
		}
	}
	var last *Ivy_FraigSim_t
	for _, n := range ivyFastMapOrder(source) {
		sim := Ivy_ObjSim(n)
		sim.pFanin0 = Ivy_ObjSim(Ivy_ObjFanin0(n))
		sim.pFanin1 = Ivy_ObjSim(Ivy_ObjFanin1(n))
		sim.Type = Ivy_ObjFaninPhase(n.pFanin0)<<2 | Ivy_ObjFaninPhase(n.pFanin1)<<1 | n.fPhase
		if n.Type == IVY_EXOR {
			sim.Type |= 8
		}
		if p.pSimStart == nil {
			p.pSimStart = sim
		} else {
			last.pNext = sim
		}
		last = sim
	}
	p.pPatWords = make([]uint32, (len(source.vPis)+31)/32)
	p.pPatScores = make([]int32, 32*params.nSimWords)
	return p
}
func ivyFraigStopSimulation(p *Ivy_FraigMan_t) {
	for _, n := range p.pManAig.vObjs {
		if n != nil {
			n.fraig = nil
		}
	}
	p.pSimStart = nil
	p.pPatWords = nil
	p.pPatScores = nil
	p.lClasses = Ivy_FraigList_t{}
	p.lCand = Ivy_FraigList_t{}
}
func Ivy_NodeAssignRandom(p *Ivy_FraigMan_t, n *Ivy_Obj_t) {
	if Ivy_ObjIsPi(n) == 0 {
		panic("ivy: simulation input expected")
	}
	for i := range Ivy_ObjSim(n).pData {
		Ivy_ObjSim(n).pData[i] = p.rng.Uint32()
	}
}
func Ivy_NodeAssignConst(p *Ivy_FraigMan_t, n *Ivy_Obj_t, value int32) {
	if Ivy_ObjIsPi(n) == 0 {
		panic("ivy: simulation input expected")
	}
	word := uint32(0)
	if value != 0 {
		word = ^word
	}
	for i := range Ivy_ObjSim(n).pData {
		Ivy_ObjSim(n).pData[i] = word
	}
}
func Ivy_FraigAssignRandom(p *Ivy_FraigMan_t) {
	for _, n := range p.pManAig.vPis {
		Ivy_NodeAssignRandom(p, n)
	}
}
func Ivy_FraigAssignDist1(p *Ivy_FraigMan_t, pattern []uint32) {
	if len(pattern)*32 < len(p.pManAig.vPis) {
		panic("ivy: simulation pattern too short")
	}
	for i, n := range p.pManAig.vPis {
		Ivy_NodeAssignConst(p, n, int32(pattern[i/32]>>uint(i%32)&1))
	}
	limit := min(len(p.pManAig.vPis), int(p.nSimWords)*32-1)
	for i := 0; i < limit; i++ {
		Ivy_ObjSim(p.pManAig.vPis[i]).pData[(i+1)/32] ^= 1 << uint((i+1)%32)
	}
}
func Ivy_NodeHasZeroSim(p *Ivy_FraigMan_t, n *Ivy_Obj_t) int32 {
	for _, word := range Ivy_ObjSim(n).pData {
		if word != 0 {
			return 0
		}
	}
	return 1
}
func Ivy_NodeComplementSim(p *Ivy_FraigMan_t, n *Ivy_Obj_t) {
	for i := range Ivy_ObjSim(n).pData {
		Ivy_ObjSim(n).pData[i] = ^Ivy_ObjSim(n).pData[i]
	}
}
func Ivy_NodeCompareSims(p *Ivy_FraigMan_t, a, b *Ivy_Obj_t) int32 {
	return ivyBool(slices.Equal(Ivy_ObjSim(a).pData, Ivy_ObjSim(b).pData))
}
func Ivy_NodeSimulateSim(p *Ivy_FraigMan_t, sim *Ivy_FraigSim_t) {
	if sim == nil || sim.pFanin0 == nil || sim.pFanin1 == nil || sim.Type < 0 || sim.Type > 15 {
		panic("ivy: invalid simulation record")
	}
	for i := range sim.pData {
		a, b := sim.pFanin0.pData[i], sim.pFanin1.pData[i]
		if sim.Type&4 != 0 {
			a = ^a
		}
		if sim.Type&2 != 0 {
			b = ^b
		}
		result := a & b
		if sim.Type&8 != 0 {
			result = a ^ b
		}
		if sim.Type&1 != 0 {
			result = ^result
		}
		sim.pData[i] = result
	}
}
func Ivy_NodeSimulate(p *Ivy_FraigMan_t, n *Ivy_Obj_t) { Ivy_NodeSimulateSim(p, Ivy_ObjSim(n)) }

func Ivy_FraigSimulateOne(p *Ivy_FraigMan_t) {
	start := time.Now()
	for _, n := range ivyFastMapOrder(p.pManAig) {
		Ivy_NodeSimulate(p, n)
	}
	p.timeSim += time.Since(start)
	p.nSimRounds++
}
func Ivy_FraigSimulateOneSim(p *Ivy_FraigMan_t) {
	start := time.Now()
	for sim := p.pSimStart; sim != nil; sim = sim.pNext {
		Ivy_NodeSimulateSim(p, sim)
	}
	p.timeSim += time.Since(start)
	p.nSimRounds++
}
func Ivy_NodeAddToClass(class, n *Ivy_Obj_t) {
	if Ivy_ObjClassNodeNext(class) == nil {
		Ivy_ObjSetClassNodeNext(class, n)
	} else {
		Ivy_ObjSetClassNodeNext(Ivy_ObjClassNodeLast(class), n)
	}
	Ivy_ObjSetClassNodeLast(class, n)
	Ivy_ObjSetClassNodeRepr(n, class)
	Ivy_ObjSetClassNodeNext(n, nil)
}
func Ivy_FraigAddClass(list *Ivy_FraigList_t, class *Ivy_Obj_t) {
	if list.pHead == nil {
		list.pHead = class
	} else {
		Ivy_ObjSetEquivListNext(list.pTail, class)
	}
	Ivy_ObjSetEquivListPrev(class, list.pTail)
	Ivy_ObjSetEquivListNext(class, nil)
	list.pTail = class
	list.nItems++
}
func Ivy_FraigInsertClass(list *Ivy_FraigList_t, base, class *Ivy_Obj_t) {
	next := Ivy_ObjEquivListNext(base)
	Ivy_ObjSetEquivListPrev(class, base)
	Ivy_ObjSetEquivListNext(class, next)
	if next != nil {
		Ivy_ObjSetEquivListPrev(next, class)
	} else {
		list.pTail = class
	}
	Ivy_ObjSetEquivListNext(base, class)
	list.nItems++
}
func Ivy_FraigRemoveClass(list *Ivy_FraigList_t, class *Ivy_Obj_t) {
	prev, next := Ivy_ObjEquivListPrev(class), Ivy_ObjEquivListNext(class)
	if prev != nil {
		Ivy_ObjSetEquivListNext(prev, next)
	} else {
		list.pHead = next
	}
	if next != nil {
		Ivy_ObjSetEquivListPrev(next, prev)
	} else {
		list.pTail = prev
	}
	Ivy_ObjSetEquivListNext(class, nil)
	Ivy_ObjSetEquivListPrev(class, nil)
	class.fMarkA = 0
	list.nItems--
}
func Ivy_FraigCountPairsClasses(p *Ivy_FraigMan_t) int32 {
	var pairs int32
	for class := p.lClasses.pHead; class != nil; class = Ivy_ObjEquivListNext(class) {
		n := Ivy_FraigCountClassNodes(class)
		pairs += n * (n - 1) / 2
	}
	return pairs
}
func Ivy_FraigCreateClasses(p *Ivy_FraigMan_t) {
	if p.lClasses.pHead != nil {
		panic("ivy: simulation classes already initialized")
	}
	groups := map[string]*Ivy_Obj_t{}
	var heads []*Ivy_Obj_t
	for _, n := range p.pManAig.vObjs {
		if n == nil || (n != p.pManAig.pConst1 && Ivy_ObjIsPi(n) == 0 && Ivy_ObjIsNode(n) == 0) {
			continue
		}
		sim := Ivy_ObjSim(n)
		encoded := make([]byte, 4*len(sim.pData))
		for i, word := range sim.pData {
			binary.LittleEndian.PutUint32(encoded[4*i:], word)
		}
		key := string(encoded)
		if class := groups[key]; class != nil {
			Ivy_NodeAddToClass(class, n)
		} else {
			groups[key] = n
			heads = append(heads, n)
		}
	}
	for _, class := range heads {
		if Ivy_ObjClassNodeNext(class) != nil {
			Ivy_ObjSetClassNodeRepr(class, nil)
			Ivy_FraigAddClass(&p.lClasses, class)
		}
	}
}
func Ivy_FraigRefineClass_rec(p *Ivy_FraigMan_t, class *Ivy_Obj_t) int32 {
	var groups [][]*Ivy_Obj_t
	seen := map[*Ivy_Obj_t]bool{}
	for n := class; n != nil; {
		if seen[n] {
			panic("ivy: cyclic fraig class")
		}
		seen[n] = true
		next := Ivy_ObjClassNodeNext(n)
		index := -1
		for i, group := range groups {
			if Ivy_NodeCompareSims(p, group[0], n) != 0 {
				index = i
				break
			}
		}
		if index < 0 {
			groups = append(groups, []*Ivy_Obj_t{n})
		} else {
			groups[index] = append(groups[index], n)
		}
		n = next
	}
	if len(groups) < 2 {
		return 0
	}
	if p.pParams.fPatScores != 0 {
		for _, group := range groups[1:] {
			Ivy_FraigAddToPatScores(p, class, group[0])
		}
	}
	previous := class
	for i, group := range groups {
		head := group[0]
		Ivy_ObjSetClassNodeRepr(head, nil)
		Ivy_ObjSetClassNodeNext(head, nil)
		for _, n := range group[1:] {
			Ivy_NodeAddToClass(head, n)
		}
		if i > 0 {
			Ivy_FraigInsertClass(&p.lClasses, previous, head)
		}
		previous = head
	}
	for _, group := range groups {
		if len(group) == 1 {
			Ivy_FraigRemoveClass(&p.lClasses, group[0])
		}
	}
	return int32(len(groups) - 1)
}
func Ivy_FraigCountClassNodes(class *Ivy_Obj_t) int32 {
	var count int32
	seen := map[*Ivy_Obj_t]bool{}
	for n := class; n != nil; n = Ivy_ObjClassNodeNext(n) {
		if seen[n] {
			panic("ivy: cyclic fraig class")
		}
		seen[n] = true
		count++
	}
	return count
}
func Ivy_FraigPrintClass(class *Ivy_Obj_t) {
	fmt.Print("Class {")
	for n := class; n != nil; n = Ivy_ObjClassNodeNext(n) {
		fmt.Printf(" %d(%d)", n.Id, n.Level)
	}
	fmt.Println(" }")
}
func Ivy_FraigPrintSimClasses(p *Ivy_FraigMan_t) {
	for class := p.lClasses.pHead; class != nil; class = Ivy_ObjEquivListNext(class) {
		fmt.Printf("%d ", Ivy_FraigCountClassNodes(class))
	}
}
func Ivy_FraigCleanPatScores(p *Ivy_FraigMan_t) { clear(p.pPatScores) }
func Ivy_FraigAddToPatScores(p *Ivy_FraigMan_t, a, b *Ivy_Obj_t) {
	for w, word := range Ivy_ObjSim(a).pData {
		difference := word ^ Ivy_ObjSim(b).pData[w]
		for difference != 0 {
			bit := bits.TrailingZeros32(difference)
			p.pPatScores[w*32+bit]++
			difference &= difference - 1
		}
	}
}
func Ivy_FraigSelectBestPat(p *Ivy_FraigMan_t) int32 {
	score, index := int32(0), -1
	for i := 1; i < len(p.pPatScores); i++ {
		if p.pPatScores[i] > score {
			score, index = p.pPatScores[i], i
		}
	}
	if index < 0 {
		return 0
	}
	clear(p.pPatWords)
	for i, n := range p.pManAig.vPis {
		if Ivy_ObjSim(n).pData[index/32]&(1<<uint(index%32)) != 0 {
			p.pPatWords[i/32] |= 1 << uint(i%32)
		}
	}
	return score
}

var ivyFraigPrimes = [128]uint32{
	1009, 1049, 1093, 1151, 1201, 1249, 1297, 1361, 1427, 1459,
	1499, 1559, 1607, 1657, 1709, 1759, 1823, 1877, 1933, 1997,
	2039, 2089, 2141, 2213, 2269, 2311, 2371, 2411, 2467, 2543,
	2609, 2663, 2699, 2741, 2797, 2851, 2909, 2969, 3037, 3089,
	3169, 3221, 3299, 3331, 3389, 3461, 3517, 3557, 3613, 3671,
	3719, 3779, 3847, 3907, 3943, 4013, 4073, 4129, 4201, 4243,
	4289, 4363, 4441, 4493, 4549, 4621, 4663, 4729, 4793, 4871,
	4933, 4973, 5021, 5087, 5153, 5227, 5281, 5351, 5417, 5471,
	5519, 5573, 5651, 5693, 5749, 5821, 5861, 5923, 6011, 6073,
	6131, 6199, 6257, 6301, 6353, 6397, 6481, 6563, 6619, 6689,
	6737, 6803, 6863, 6917, 6977, 7027, 7109, 7187, 7237, 7309,
	7393, 7477, 7523, 7561, 7607, 7681, 7727, 7817, 7877, 7933,
	8011, 8039, 8059, 8081, 8093, 8111, 8123, 8147,
}

func Ivy_NodeHash(p *Ivy_FraigMan_t, n *Ivy_Obj_t) uint32 {
	var hash uint32
	for i, word := range Ivy_ObjSim(n).pData {
		hash ^= word * ivyFraigPrimes[i]
	}
	return hash
}

func Ivy_FraigSavePattern0(p *Ivy_FraigMan_t) { clear(p.pPatWords) }
func Ivy_FraigSavePattern1(p *Ivy_FraigMan_t) {
	for i := range p.pPatWords {
		p.pPatWords[i] = ^uint32(0)
	}
}
func Ivy_FraigCheckOutputSimsSavePattern(p *Ivy_FraigMan_t, edge *Ivy_Obj_t) {
	signature := Ivy_ObjSim(edge).pData
	phase := Ivy_ObjFaninPhase(edge)
	pattern := -1
	for word, value := range signature {
		if phase != 0 {
			value = ^value
		}
		if value != 0 {
			pattern = word*32 + bits.TrailingZeros32(value)
			break
		}
	}
	if pattern < 0 {
		panic("ivy: no satisfying simulation pattern")
	}
	model := make([]int32, len(p.pManAig.vPis))
	for i, input := range p.pManAig.vPis {
		model[i] = int32(Ivy_ObjSim(input).pData[pattern/32] >> uint(pattern%32) & 1)
	}
	if p.pManFraig.pData != nil {
		panic("ivy: counterexample already attached")
	}
	p.pManFraig.pData = model
}
func Ivy_FraigCheckOutputSims(p *Ivy_FraigMan_t) int32 {
	if p.pManFraig.pData != nil {
		return 1
	}
	for _, output := range p.pManAig.vPos {
		edge := Ivy_ObjChild0(output)
		for _, word := range Ivy_ObjSim(edge).pData {
			if Ivy_ObjFaninPhase(edge) != 0 {
				word = ^word
			}
			if word != 0 {
				Ivy_FraigCheckOutputSimsSavePattern(p, edge)
				return 1
			}
		}
	}
	return 0
}
func Ivy_FraigRefineClasses(p *Ivy_FraigMan_t) int32 {
	if p.pParams.FProve != 0 {
		Ivy_FraigCheckOutputSims(p)
	}
	if p.pManFraig.pData != nil {
		return 0
	}
	start := time.Now()
	var count int32
	for class := p.lClasses.pHead; class != nil; {
		next := Ivy_ObjEquivListNext(class)
		if class.fMarkA == 0 && Ivy_FraigRefineClass_rec(p, class) > 0 {
			count++
		}
		class = next
	}
	p.timeRef += time.Since(start)
	return count
}
func Ivy_FraigSimulate(p *Ivy_FraigMan_t) {
	Ivy_FraigAssignRandom(p)
	Ivy_FraigSimulateOne(p)
	Ivy_FraigCreateClasses(p)
	for _, ones := range []bool{false, true} {
		if ones {
			Ivy_FraigSavePattern1(p)
		} else {
			Ivy_FraigSavePattern0(p)
		}
		Ivy_FraigAssignDist1(p, p.pPatWords)
		Ivy_FraigSimulateOne(p)
		Ivy_FraigRefineClasses(p)
		if p.pManFraig.pData != nil {
			return
		}
	}
	for {
		before := p.lClasses.nItems
		Ivy_FraigAssignRandom(p)
		Ivy_FraigSimulateOne(p)
		changed := Ivy_FraigRefineClasses(p)
		if p.pManFraig.pData != nil || before == 0 || float64(changed)/float64(before) <= p.pParams.dSimSatur {
			return
		}
	}
}
func Ivy_FraigResimulate(p *Ivy_FraigMan_t) {
	Ivy_FraigAssignDist1(p, p.pPatWords)
	Ivy_FraigSimulateOne(p)
	if p.pParams.fPatScores != 0 {
		Ivy_FraigCleanPatScores(p)
	}
	changed := Ivy_FraigRefineClasses(p)
	if p.pManFraig.pData != nil {
		return
	}
	if changed < 1 {
		panic("ivy: counterexample did not refine simulation classes")
	}
	if p.pParams.fPatScores == 0 {
		return
	}
	for Ivy_FraigSelectBestPat(p) > p.pParams.MaxScore {
		Ivy_FraigAssignDist1(p, p.pPatWords)
		Ivy_FraigSimulateOne(p)
		Ivy_FraigCleanPatScores(p)
		if Ivy_FraigRefineClasses(p) == 0 || p.pManFraig.pData != nil {
			return
		}
	}
}
