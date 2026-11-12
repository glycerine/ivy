package sbd

import (
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func Sbd_CutMergeSimple(_ *Sbd_Man_t, a, b, out []int32) int32 {
	values := append(slices.Clone(a[1:a[0]+1]), b[1:b[0]+1]...)
	slices.Sort(values)
	values = slices.Compact(values)
	out[0] = int32(len(values))
	copy(out[1:], values)
	return out[0]
}
func Sbd_ManMergeCuts(p *Sbd_Man_t, id int32) int32 {
	node := p.pGia.object(id)
	a, b := int32(node.Fanins[0].NodeID()), int32(node.Fanins[1].NodeID())
	la, lb := max(1, vec.Vec_IntEntry(p.vLutLevs, int(a))), max(1, vec.Vec_IntEntry(p.vLutLevs, int(b)))
	cutsA, cutsB := [][]int32{Sbd_ObjCut(p, a), {1, a}}, [][]int32{Sbd_ObjCut(p, b), {1, b}}
	var best []int32
	bestLevel := int32(math.MaxInt32)
	for _, choice := range [][2]int{{0, 0}, {1, 0}, {0, 1}, {1, 1}} {
		cut := make([]int32, 2*p.pPars.LUTSize+1)
		size := Sbd_CutMergeSimple(p, cutsA[choice[0]], cutsB[choice[1]], cut)
		level := int32(math.MaxInt32)
		if size <= p.pPars.LUTSize {
			level = max(la+int32(choice[0]), lb+int32(choice[1]))
		}
		if best == nil || level < bestLevel || level == bestLevel && size < best[0] {
			best, bestLevel = cut, level
		}
	}
	if best[0] > p.pPars.LUTSize {
		panic("no feasible merged LUT cut")
	}
	vec.Vec_IntWriteEntry(p.vLutLevs, int(id), bestLevel)
	copy(Sbd_ObjCut(p, id), best[:best[0]+1])
	if bestLevel == 1 {
		return 1
	}
	return 0
}
func Sbd_ManDelay(p *Sbd_Man_t) int32 {
	var delay int32
	for _, co := range p.pGia.Graph.Outputs() {
		id := p.pGia.outputDriver(int32(co.NodeID()))
		delay = max(delay, vec.Vec_IntEntry(p.vLutLevs, int(id)))
	}
	return delay
}
func Sbd_ManMergeTest(p *Sbd_Man_t) {
	for _, id := range sbdCoreOrder(p.pGia, false) {
		kind := p.pGia.object(id).Kind
		if kind == aig.And || kind == aig.ExclusiveOr {
			Sbd_ManMergeCuts(p, id)
		}
	}
	fmt.Printf("Delay%d.\n", Sbd_ManDelay(p))
}
func Sbd_ManFindCut_rec(p *Sbd_Man_t, id int32) {
	if p.reached[id] {
		return
	}
	p.reached[id] = true
	if p.cutBoundary[id] || id == 0 {
		return
	}
	node := p.pGia.object(id)
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("new function escapes chosen cut")
	}
	for _, edge := range node.Fanins {
		Sbd_ManFindCut_rec(p, int32(edge.NodeID()))
	}
}
func Sbd_ManFindCutUnmark_rec(p *Sbd_Man_t, id int32) {
	if !p.reached[id] {
		return
	}
	delete(p.reached, id)
	if p.cutBoundary[id] || id == 0 {
		return
	}
	node := p.pGia.object(id)
	for _, edge := range node.Fanins {
		Sbd_ManFindCutUnmark_rec(p, int32(edge.NodeID()))
	}
}
func Sbd_ManFindCut(p *Sbd_Man_t, id int32, literals *vec.Vec_Int_t) {
	p.cutBoundary = map[int32]bool{}
	for _, lit := range vec.Vec_IntArray(literals) {
		p.cutBoundary[lit>>1] = true
	}
	p.reached = map[int32]bool{}
	Sbd_ManFindCut_rec(p, id)
	cut := Sbd_ObjCut(p, id)
	cut[0] = 0
	maximum := int32(0)
	for _, lit := range vec.Vec_IntArray(literals) {
		leaf := lit >> 1
		if p.reached[leaf] {
			cut[0]++
			cut[cut[0]] = leaf
			maximum = max(maximum, vec.Vec_IntEntry(p.vLutLevs, int(leaf)))
		}
	}
	vec.Vec_IntWriteEntry(p.vLutLevs, int(id), maximum+1)
	Sbd_ManFindCutUnmark_rec(p, id)
}
func sbdImplementationLeaves(p *Sbd_Man_t) []aig.Edge {
	leaves := make([]aig.Edge, vec.Vec_IntSize(p.vDivSet))
	for i, variable := range vec.Vec_IntArray(p.vDivSet) {
		id := vec.Vec_IntEntry(p.vWinObjs, int(variable))
		literal := 2 * id
		if replacement := vec.Vec_IntEntry(p.vMirrors, int(id)); replacement >= 0 {
			literal = replacement
		}
		leaves[i] = graphLiteral(literal)
	}
	return leaves
}
func sbdExtendImplementation(p *Sbd_Man_t, previous int, pivot int32) {
	count := len(p.pGia.Graph.Nodes())
	for id := previous; id < count; id++ {
		vec.Vec_IntPush(p.vMirrors, -1)
		mirror := int32(-1)
		if id == count-1 {
			mirror = pivot
		}
		Sbd_StoRefObj(p.pSto, int32(id), mirror)
	}
	for id := previous; id < count; id++ {
		delay := Sbd_StoComputeCutsNode(p.pSto, int32(id))
		vec.Vec_IntPush(p.vLutLevs, delay)
		vec.Vec_IntPush(p.vObj2Var, 0)
		for i := int32(0); i < p.pPars.LUTSize+1; i++ {
			vec.Vec_IntPush(p.vLutCuts, 0)
		}
		Sbd_StoSaveBestDelayCut(p.pSto, int32(id), Sbd_ObjCut(p, int32(id)))
		for which := range p.vSims {
			for w := int32(0); w < p.pPars.SimulationWords; w++ {
				vec.Vec_WrdPush(p.vSims[which], 0)
			}
		}
	}
	p.pGia.resize()
}
func Sbd_ManImplement(p *Sbd_Man_t, pivot int32, truth uint64) int32 {
	inputs := make([]int32, vec.Vec_IntSize(p.vDivSet))
	for i := range inputs {
		inputs[i] = int32(i)
	}
	structure := NewLUTStructure(inputs)
	structure.Res = truth
	return Sbd_ManImplement2(p, pivot, 1, []Sbd_Str_t{structure})
}
func Sbd_ManImplement2(p *Sbd_Man_t, pivot, count int32, structures []Sbd_Str_t) int32 {
	previous := len(p.pGia.Graph.Nodes())
	oldLevel := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	literals := sbdImplementationLeaves(p)
	divs := len(literals)
	literals = append(literals, make([]aig.Edge, count)...)
	for i := count - 1; i >= 0; i-- {
		structure := &structures[i]
		var edge aig.Edge
		if structure.fLut != 0 {
			inputs := make([]aig.Edge, structure.nVarIns)
			for k, index := range structure.VarIns[:structure.nVarIns] {
				inputs[k] = literals[index]
			}
			edge = dau.Dsm_ManTruthToGia(p.pGia.Graph, []uint64{structure.Res}, inputs, p.vCover)
		} else {
			edge = literals[structure.Res]
		}
		literals[divs+int(i)] = edge
	}
	result := sourceLiteral(literals[divs])
	if vec.Vec_IntEntry(p.vMirrors, int(pivot)) != -1 {
		panic("pivot already replaced")
	}
	// Structural hashing may return the original pivot (or a node containing
	// it). Such a replacement would introduce a mirror cycle, not a rewrite.
	seen := map[int32]bool{}
	var containsPivot func(int32) bool
	containsPivot = func(id int32) bool {
		if id == pivot {
			return true
		}
		if id == 0 || seen[id] {
			return false
		}
		seen[id] = true
		if int(id) < vec.Vec_IntSize(p.vMirrors) {
			if mirror := vec.Vec_IntEntry(p.vMirrors, int(id)); mirror >= 0 {
				return containsPivot(mirror >> 1)
			}
		}
		node := p.pGia.object(id)
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			return containsPivot(int32(node.Fanins[0].NodeID())) || containsPivot(int32(node.Fanins[1].NodeID()))
		case aig.Buffer:
			return containsPivot(int32(node.Fanins[0].NodeID()))
		}
		return false
	}
	cyclic := containsPivot(result >> 1)
	sbdExtendImplementation(p, previous, -1)
	if cyclic {
		return 0
	}
	newLevel := vec.Vec_IntEntry(p.vLutLevs, int(result>>1))
	if newLevel != 0 && newLevel >= oldLevel {
		return 0
	}
	vec.Vec_IntWriteEntry(p.vMirrors, int(pivot), result)
	if int(result>>1) >= previous {
		vec.Vec_IntWriteEntry(p.pSto.vRefs, int(result>>1), vec.Vec_IntEntry(p.pSto.vRefs, int(pivot)))
		vec.Vec_IntWriteEntry(p.pSto.vRefs, int(pivot), 1)
	}
	Sbd_StoDerefObj(p.pSto, pivot)
	vec.Vec_IntWriteEntry(p.vLutLevs, int(pivot), newLevel)
	return 0
}
func Sbd_ManDerive_rec(target *Network, source *Network, id int32, mirrors *vec.Vec_Int_t, copies map[int32]int32) {
	if _, ok := copies[id]; ok {
		return
	}
	real := sbdMirrorObject(mirrors, id)
	if real != id {
		Sbd_ManDerive_rec(target, source, real, mirrors, copies)
		copies[id] = copies[real] ^ (vec.Vec_IntEntry(mirrors, int(id)) & 1)
		return
	}
	node := source.object(id)
	copyEdge := func(edge aig.Edge) aig.Edge {
		literal := sourceLiteral(edge)
		Sbd_ManDerive_rec(target, source, literal>>1, mirrors, copies)
		return graphLiteral(copies[literal>>1] ^ (literal & 1))
	}
	switch node.Kind {
	case aig.And:
		copies[id] = sourceLiteral(target.Graph.AND(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1])))
	case aig.ExclusiveOr:
		copies[id] = sourceLiteral(target.Graph.NewXor(copyEdge(node.Fanins[0]), copyEdge(node.Fanins[1])))
	case aig.Buffer:
		copies[id] = sourceLiteral(copyEdge(node.Fanins[0]))
	default:
		panic("unseeded SBD graph boundary")
	}
}
func Sbd_ManDeriveMapping_rec(p *Sbd_Man_t, target *Network, id int32) {
	if id == 0 {
		return
	}
	if _, ok := target.LUTs[id]; ok {
		return
	}
	node := target.object(id)
	if node.Kind == aig.Input {
		return
	}
	cut := Sbd_ObjCut2(p, id)
	leaves := slices.Clone(cut[1 : cut[0]+1])
	for _, leaf := range leaves {
		Sbd_ManDeriveMapping_rec(p, target, leaf)
	}
	edges := make([]aig.Edge, len(leaves))
	for i, leaf := range leaves {
		edges[i] = aig.Edge(leaf) << 1
	}
	truth, err := aig.Aig_ManCutTruth(target.Graph, aig.Edge(id)<<1, edges)
	if err != nil {
		panic(err)
	}
	words := make([]uint64, max(1, (len(truth)+1)/2))
	for i, value := range truth {
		words[i/2] |= uint64(value) << uint(32*(i&1))
	}
	if len(truth) == 1 {
		words[0] |= uint64(truth[0]) << 32
	}
	target.SetLUT(id, LUTCut{Leaves: leaves, Truth: words})
}
func Sbd_ManDeriveMapping(p *Sbd_Man_t, target *Network) {
	p.vLutCuts2 = vec.Vec_IntStart(len(target.Graph.Nodes()) * int(p.pPars.LUTSize+1))
	target.LUTs = map[int32]LUTCut{}
	for _, edge := range p.pGia.Graph.LiveEdges() {
		id := int32(edge.NodeID())
		kind := p.pGia.object(id).Kind
		if kind != aig.And && kind != aig.ExclusiveOr || vec.Vec_IntEntry(p.vMirrors, int(id)) >= 0 {
			continue
		}
		copyLiteral, ok := p.copies[id]
		if !ok || copyLiteral>>1 == 0 {
			continue
		}
		newID := copyLiteral >> 1
		newKind := target.object(newID).Kind
		if newKind != aig.And && newKind != aig.ExclusiveOr {
			continue
		}
		oldCut := Sbd_ObjCut(p, id)
		var leaves []int32
		for _, leaf := range oldCut[1 : oldCut[0]+1] {
			real := sbdMirrorObject(p.vMirrors, leaf)
			literal, exists := p.copies[real]
			if !exists {
				continue
			}
			mapped := literal >> 1
			if mapped != 0 && mapped != newID {
				leaves = append(leaves, mapped)
			}
		}
		slices.Sort(leaves)
		leaves = slices.Compact(leaves)
		cut := Sbd_ObjCut2(p, newID)
		cut[0] = int32(len(leaves))
		copy(cut[1:], leaves)
	}
	for _, output := range target.Graph.Outputs() {
		Sbd_ManDeriveMapping_rec(p, target, target.outputDriver(int32(output.NodeID())))
	}
	p.vLutCuts2 = nil
	target.resize()
	clear(target.LUTRefs)
	visited := map[int32]bool{}
	var refs func(int32)
	refs = func(id int32) {
		target.LUTRefs[id]++
		if visited[id] {
			return
		}
		visited[id] = true
		if cut, ok := target.LUTs[id]; ok {
			for _, leaf := range cut.Leaves {
				refs(leaf)
			}
		}
	}
	for _, output := range target.Graph.Outputs() {
		refs(target.outputDriver(int32(output.NodeID())))
	}
}
func Sbd_ManDerive(p *Sbd_Man_t, source *Network, mirrors *vec.Vec_Int_t) *Network {
	graph := aig.NewGraph()
	graph.Name, graph.Spec, graph.Timing = source.Graph.Name, source.Graph.Spec, source.Graph.Timing
	target := &Network{Graph: graph, LUTs: map[int32]LUTCut{}, BoxInputs: map[int32][]int32{}, BoxLevelInputs: map[int32][]int32{}}
	p.copies = map[int32]int32{0: 0}
	for _, input := range source.Graph.Inputs() {
		p.copies[int32(input.NodeID())] = sourceLiteral(graph.NewInput())
	}
	for _, output := range source.Graph.Outputs() {
		driver := source.object(int32(output.NodeID())).Fanins[0]
		literal := sourceLiteral(driver)
		Sbd_ManDerive_rec(target, source, literal>>1, mirrors, p.copies)
		edge := graph.AddOutput(graphLiteral(p.copies[literal>>1] ^ (literal & 1)))
		p.copies[int32(output.NodeID())] = sourceLiteral(edge)
	}
	aig.Aig_ManSetRegNum(graph, aig.Aig_ManRegNum(source.Graph))
	aig.Aig_ManSetConstrNum(graph, aig.Aig_ManConstrNum(source.Graph))
	target.resize()
	for ci, inputs := range source.BoxInputs {
		newCI := p.copies[ci] >> 1
		mapped := make([]int32, len(inputs))
		for i, co := range inputs {
			mapped[i] = p.copies[co] >> 1
		}
		var enabled []bool
		if active, ok := source.BoxLevelInputs[ci]; ok {
			enabled = make([]bool, len(inputs))
			for i, co := range inputs {
				enabled[i] = slices.Contains(active, co)
			}
		}
		target.SetBoxOutput(newCI, mapped, enabled)
	}
	if p.pPars.Mapping {
		Sbd_ManDeriveMapping(p, target)
	}
	return target
}
func Sbd_NtkPerformOne(p *Sbd_Man_t, pivot int32) {
	if p.pSto == nil && Sbd_ManMergeCuts(p, pivot) != 0 {
		return
	}
	if Sbd_ManWindow(p, pivot) == 0 {
		return
	}
	p.nTried++
	p.nUsed++
	constant := Sbd_ManCheckConst(p, pivot)
	if constant >= 0 {
		vec.Vec_IntWriteEntry(p.vMirrors, int(pivot), constant)
		return
	}
	var truth uint64
	structures := make([]Sbd_Str_t, 10)
	if p.pPars.FindDivisors && p.pPars.LUTCount >= 1 && Sbd_ManExplore2(p, pivot, &truth) != 0 {
		inputs := make([]int32, vec.Vec_IntSize(p.vDivSet))
		for i := range inputs {
			inputs[i] = int32(i)
		}
		structures[0] = NewLUTStructure(inputs)
		structures[0].Res = truth
		Sbd_ManImplement2(p, pivot, 1, structures)
		if vec.Vec_IntEntry(p.vMirrors, int(pivot)) < 0 {
			p.nUsed--
		}
		return
	}
	var count int32
	if p.pPars.LUTCount >= 2 && Sbd_ManExplore3(p, pivot, &count, structures) != 0 {
		Sbd_ManImplement2(p, pivot, count, structures)
		if vec.Vec_IntEntry(p.vMirrors, int(pivot)) < 0 {
			p.nUsed--
		}
		return
	}
	p.nUsed--
}
func Sbd_NtkPerform(network *Network, parameters *Parameters) *Network {
	p := Sbd_ManStart(network, parameters)
	defer Sbd_ManStop(p)
	original := len(network.Graph.Nodes())
	for id := 0; id < original; id++ {
		if network.Graph.ValidEdge(aig.Edge(id) << 1) {
			Sbd_StoRefObj(p.pSto, int32(id), -1)
		} else {
			vec.Vec_IntPush(p.pSto.vRefs, 0)
		}
	}
	var path *vec.Vec_Bit_t
	if parameters.CriticalPathOnly && len(network.LUTs) > 0 {
		path = Sbc_ManCriticalPath(network)
	}
	Sbd_StoComputeCutsConst0(p.pSto, 0)
	for _, id := range sbdCoreOrder(network, len(network.BoxInputs) > 0) {
		if int(id) >= original {
			continue
		}
		node := network.object(id)
		switch node.Kind {
		case aig.Input:
			arrival := int32(0)
			inputs := network.BoxInputs[id]
			if enabled, ok := network.BoxLevelInputs[id]; ok {
				inputs = enabled
			}
			for _, co := range inputs {
				arrival = max(arrival, vec.Vec_IntEntry(p.vLutLevs, int(network.outputDriver(co)))+1)
			}
			vec.Vec_IntWriteEntry(p.vLutLevs, int(id), arrival)
			Sbd_StoComputeCutsCi(p.pSto, id, arrival, arrival)
		case aig.And, aig.ExclusiveOr:
			delay := Sbd_StoComputeCutsNode(p.pSto, id)
			Sbd_StoSaveBestDelayCut(p.pSto, id, Sbd_ObjCut(p, id))
			vec.Vec_IntWriteEntry(p.vLutLevs, int(id), delay)
			if delay > 1 && (path == nil || vec.Vec_BitEntry(path, int(id)) != 0) {
				Sbd_NtkPerformOne(p, id)
			}
		case aig.Buffer:
			literal := sourceLiteral(node.Fanins[0])
			if mirror := vec.Vec_IntEntry(p.vMirrors, int(literal>>1)); mirror >= 0 {
				literal = mirror ^ (literal & 1)
			}
			vec.Vec_IntWriteEntry(p.vMirrors, int(id), literal)
			level := vec.Vec_IntEntry(p.vLutLevs, int(literal>>1))
			vec.Vec_IntWriteEntry(p.vLutLevs, int(id), level)
			Sbd_StoComputeCutsObj(p.pSto, id, level, level)
		}
	}
	if parameters.Verbose {
		fmt.Printf("K=%d S=%d N=%d P=%d Try=%d Use=%d C=%d 1=%d 2=%d 3a=%d 3b=%d Lev=%d Time=%v\n", parameters.LUTSize, parameters.LUTCount, parameters.CutSize, parameters.CutCount, p.nTried, p.nUsed, p.nLuts[0], p.nLuts[1], p.nLuts[2], p.nLuts[3], p.nLuts[4], Sbd_ManDelay(p), time.Since(p.started))
	}
	return Sbd_ManDerive(p, network, p.vMirrors)
}
