package sbd

import (
	"fmt"
	"slices"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/bsat"
)

const SBD_MAX_LUTSIZE = 6

type Sbd_Man_t struct {
	pPars                                                           *Parameters
	pGia                                                            *Network
	vTfos, vDivLevels                                               *vec.Vec_Wec_t
	vLutLevs, vLutCuts, vLutCuts2, vMirrors                         *vec.Vec_Int_t
	vSims                                                           [4]*vec.Vec_Wrd_t
	vCover, vLits, vLits2                                           *vec.Vec_Int_t
	nLuts                                                           [6]int32
	nTried, nUsed, Pivot, DivCutoff                                 int32
	pSto                                                            *Sbd_Sto_t
	pSrv                                                            *Sbd_Srv_t
	vTfo, vRoots, vWinObjs, vObj2Var, vDivSet, vDivVars, vDivValues *vec.Vec_Int_t
	vCounts                                                         [2]*vec.Vec_Int_t
	vMatrix                                                         *vec.Vec_Wrd_t
	pSat                                                            *bsat.Solver
	seen, marks, cutBoundary, reached                               map[int32]bool
	copies                                                          map[int32]int32
	random                                                          *aig.AigRandomSource
	started                                                         time.Time
}

func Sbd_ParSetDefault(p *Parameters) {
	*p = Parameters{LUTSize: 4, LUTCount: 3, CutSize: 10, CutCount: 128, TFOLevels: 5, TFOFanoutMax: 4, WindowSizeMax: 2000, SimulationWords: 1, Mapping: true}
}
func Sbd_ObjCut(p *Sbd_Man_t, id int32) []int32 {
	width := int(p.pPars.LUTSize + 1)
	return vec.Vec_IntArray(p.vLutCuts)[int(id)*width : (int(id)+1)*width]
}
func Sbd_ObjCut2(p *Sbd_Man_t, id int32) []int32 {
	width := int(p.pPars.LUTSize + 1)
	return vec.Vec_IntArray(p.vLutCuts2)[int(id)*width : (int(id)+1)*width]
}
func sbdCoreSim(p *Sbd_Man_t, which int, id int32) []uint64 {
	width := int(p.pPars.SimulationWords)
	return vec.Vec_WrdArray(p.vSims[which])[int(id)*width : (int(id)+1)*width]
}
func Sbd_ObjSim0(p *Sbd_Man_t, id int32) []uint64 { return sbdCoreSim(p, 0, id) }
func Sbd_ObjSim1(p *Sbd_Man_t, id int32) []uint64 { return sbdCoreSim(p, 1, id) }
func Sbd_ObjSim2(p *Sbd_Man_t, id int32) []uint64 { return sbdCoreSim(p, 2, id) }
func Sbd_ObjSim3(p *Sbd_Man_t, id int32) []uint64 { return sbdCoreSim(p, 3, id) }
func sbdCoreOrder(network *Network, boxes bool) []int32 {
	seen := map[int32]byte{}
	var result []int32
	var visit func(int32)
	visit = func(id int32) {
		if seen[id] == 2 {
			return
		}
		if seen[id] == 1 {
			panic("cyclic SBD graph or box metadata")
		}
		seen[id] = 1
		node := network.object(id)
		switch node.Kind {
		case aig.Input:
			if boxes {
				for _, co := range network.BoxInputs[id] {
					visit(co)
				}
			}
		case aig.And, aig.ExclusiveOr:
			visit(int32(node.Fanins[0].NodeID()))
			visit(int32(node.Fanins[1].NodeID()))
		case aig.Buffer, aig.Output:
			visit(int32(node.Fanins[0].NodeID()))
		}
		seen[id] = 2
		result = append(result, id)
	}
	for _, edge := range network.Graph.LiveEdges() {
		visit(int32(edge.NodeID()))
	}
	return result
}
func Sbd_ManWindowRoots(network *Network, tfoLevels, fanoutMax int32) *vec.Vec_Wec_t {
	count := len(network.Graph.Nodes())
	tfos := vec.Vec_WecStart(count)
	temporary := make([][]int32, count)
	levels, refs := make([]int32, count), make([]int32, count)
	drivers := make([]bool, count)
	for _, edge := range network.Graph.LiveEdges() {
		node := network.Graph.Node(edge)
		switch node.Kind {
		case aig.And, aig.ExclusiveOr:
			refs[node.Fanins[0].NodeID()]++
			refs[node.Fanins[1].NodeID()]++
		case aig.Buffer, aig.Output:
			refs[node.Fanins[0].NodeID()]++
		}
	}
	for _, edge := range network.Graph.Inputs() {
		id := int32(edge.NodeID())
		temporary[id] = []int32{id}
	}
	for _, co := range network.Graph.Outputs() {
		drivers[network.outputDriver(int32(co.NodeID()))] = true
	}
	// Native buffers are transparent graph boundaries. Treat their drivers as
	// observed roots so a window never drops an output reached through a wire.
	for _, edge := range network.Graph.LiveEdges() {
		node := network.Graph.Node(edge)
		if node.Kind == aig.Buffer {
			drivers[node.Fanins[0].NodeID()] = true
		}
	}
	for _, id := range sbdCoreOrder(network, false) {
		node := network.object(id)
		if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
			continue
		}
		a, b := int32(node.Fanins[0].NodeID()), int32(node.Fanins[1].NodeID())
		levels[id] = 1 + max(levels[a], levels[b])
		values := append(slices.Clone(temporary[a]), temporary[b]...)
		slices.Sort(values)
		values = slices.Compact(values)
		always := drivers[id] || refs[id] >= fanoutMax
		at := 0
		for _, fan := range values {
			root := always || levels[id]-levels[fan] >= tfoLevels
			literal := 2 * id
			if root {
				literal++
			}
			vec.Vec_WecPush(tfos, int(fan), literal)
			if !root {
				values[at] = fan
				at++
			}
		}
		temporary[id] = values[:at]
		if !always {
			temporary[id] = append(temporary[id], id)
		}
	}
	return tfos
}
func Sbd_ManStart(network *Network, parameters *Parameters) *Sbd_Man_t {
	if parameters.LUTSize < 2 || parameters.LUTSize > 6 || parameters.LUTCount < 1 || parameters.LUTCount > 3 || parameters.CutSize < parameters.LUTSize || parameters.CutSize > 10 || parameters.SimulationWords < 1 {
		panic("invalid SBD parameters")
	}
	count := len(network.Graph.Nodes())
	p := &Sbd_Man_t{pPars: parameters, pGia: network, vTfos: Sbd_ManWindowRoots(network, parameters.TFOLevels, parameters.TFOFanoutMax), vLutLevs: vec.Vec_IntStart(count), vLutCuts: vec.Vec_IntStart(count * int(parameters.LUTSize+1)), vMirrors: vec.Vec_IntStartFull(count), vCover: sbdVector(), vLits: sbdVector(), vLits2: sbdVector(), vRoots: sbdVector(), vWinObjs: sbdVector(), vObj2Var: vec.Vec_IntStart(count), vDivSet: sbdVector(), vDivVars: sbdVector(), vDivValues: sbdVector(), vDivLevels: vec.Vec_WecAlloc(100), vMatrix: vec.Vec_WrdAlloc(100), random: aig.NewAigRandomSource(), started: time.Now()}
	for i := range p.vSims {
		p.vSims[i] = vec.Vec_WrdStart(count * int(parameters.SimulationWords))
	}
	for i := range p.vCounts {
		p.vCounts[i] = sbdVector()
	}
	for _, edge := range network.Graph.Inputs() {
		id := int32(edge.NodeID())
		cut := Sbd_ObjCut(p, id)
		cut[0], cut[1] = 1, id
	}
	p.random.Uint32()
	for _, edge := range network.Graph.Inputs() {
		for w := range Sbd_ObjSim0(p, int32(edge.NodeID())) {
			high, low := p.random.Uint32(), p.random.Uint32()
			Sbd_ObjSim0(p, int32(edge.NodeID()))[w] = uint64(high)<<32 | uint64(low)
		}
	}
	cutSize := parameters.LUTSize
	if parameters.MoreCuts {
		cutSize = parameters.CutSize
	}
	minimize := int32(0)
	if !parameters.Mapping {
		minimize = 1
	}
	p.pSto = Sbd_StoAlloc(network, p.vMirrors, parameters.LUTSize, cutSize, parameters.CutCount, minimize, 0)
	if !parameters.MoreCuts {
		p.pSrv = Sbd_ManCutServerStart(network, p.vMirrors, p.vLutLevs, nil, nil, parameters.LUTSize, parameters.CutSize, parameters.CutCount, 0)
	}
	return p
}
func Sbd_ManStop(p *Sbd_Man_t) {
	if p != nil {
		Sbd_StoFree(p.pSto)
		Sbd_ManCutServerStop(p.pSrv)
		*p = Sbd_Man_t{}
	}
}
func Sbd_ManPropagateControlOne(p *Sbd_Man_t, id int32) {
	node := p.pGia.object(id)
	for _, edge := range node.Fanins {
		fan := sbdMirrorObject(p.vMirrors, int32(edge.NodeID()))
		for w := range Sbd_ObjSim2(p, id) {
			Sbd_ObjSim2(p, fan)[w] |= Sbd_ObjSim2(p, id)[w]
			Sbd_ObjSim3(p, fan)[w] |= Sbd_ObjSim3(p, id)[w]
		}
	}
}
func Sbd_ManPropagateControl(p *Sbd_Man_t, pivot int32) {
	copy(Sbd_ObjSim3(p, pivot), Sbd_ObjSim2(p, pivot))
	position := vec.Vec_IntEntry(p.vObj2Var, int(pivot))
	for _, id := range vec.Vec_IntArray(p.vWinObjs)[:position] {
		clear(Sbd_ObjSim2(p, id))
		clear(Sbd_ObjSim3(p, id))
	}
	for i := position; i >= 0; i-- {
		id := vec.Vec_IntEntry(p.vWinObjs, int(i))
		kind := p.pGia.object(id).Kind
		if kind == aig.And || kind == aig.ExclusiveOr {
			Sbd_ManPropagateControlOne(p, id)
		}
	}
}
func Sbd_ManUpdateOrder(p *Sbd_Man_t, pivot int32) {
	maximum := vec.Vec_IntEntry(p.vLutLevs, int(pivot))
	levels := make([][]int32, maximum+1)
	for _, id := range vec.Vec_IntArray(p.vWinObjs) {
		level := vec.Vec_IntEntry(p.vLutLevs, int(id))
		if level > maximum {
			panic("TFI mapped level exceeds pivot")
		}
		levels[level] = append(levels[level], id)
	}
	vec.Vec_WecClear(p.vDivLevels)
	vec.Vec_IntClear(p.vWinObjs)
	valid := 0
	for level, ids := range levels {
		slices.Sort(ids)
		row := vec.Vec_WecPushLevel(p.vDivLevels)
		for _, id := range ids {
			vec.Vec_IntPush(row, id)
			vec.Vec_IntWriteEntry(p.vObj2Var, int(id), int32(vec.Vec_IntSize(p.vWinObjs)))
			vec.Vec_IntPush(p.vWinObjs, id)
		}
		if int32(level) == maximum-2 {
			valid = vec.Vec_IntSize(p.vWinObjs)
		}
	}
	if valid == 0 {
		panic("no time-valid SBD divisor")
	}
	vec.Vec_IntClear(p.vDivVars)
	p.DivCutoff = -1
	for i := max(0, valid-63); i < valid; i++ {
		id := vec.Vec_IntEntry(p.vWinObjs, i)
		if p.DivCutoff == -1 && vec.Vec_IntEntry(p.vLutLevs, int(id)) == maximum-2 {
			p.DivCutoff = int32(vec.Vec_IntSize(p.vDivVars))
		}
		vec.Vec_IntPush(p.vDivVars, int32(i))
	}
	if p.DivCutoff == -1 {
		p.DivCutoff = 0
	}
	vec.Vec_IntFill(p.vDivValues, vec.Vec_IntSize(p.vDivVars), 0)
}
func Sbd_ManWindowSim_rec(p *Sbd_Man_t, initial int32) {
	id := sbdMirrorObject(p.vMirrors, initial)
	if p.seen[id] {
		return
	}
	p.seen[id] = true
	node := p.pGia.object(id)
	if node.Kind == aig.And || node.Kind == aig.ExclusiveOr {
		for _, edge := range node.Fanins {
			Sbd_ManWindowSim_rec(p, int32(edge.NodeID()))
		}
	}
	if !p.marks[id] {
		vec.Vec_IntWriteEntry(p.vObj2Var, int(id), int32(vec.Vec_IntSize(p.vWinObjs)))
		vec.Vec_IntPush(p.vWinObjs, id)
	}
	if node.Kind == aig.Input {
		return
	}
	if node.Kind != aig.And && node.Kind != aig.ExclusiveOr {
		panic("simulation window contains non-gate")
	}
	for which := 0; which < 2; which++ {
		if which == 1 && !p.marks[id] {
			continue
		}
		for w := range sbdCoreSim(p, which, id) {
			read := func(edge aig.Edge) uint64 {
				literal := sourceLiteral(edge)
				if replacement := vec.Vec_IntEntry(p.vMirrors, int(literal>>1)); replacement >= 0 {
					literal = replacement ^ (literal & 1)
				}
				fan := literal >> 1
				array := 0
				if which == 1 && p.marks[fan] {
					array = 1
				}
				value := sbdCoreSim(p, array, fan)[w]
				if literal&1 != 0 {
					value = ^value
				}
				return value
			}
			a, b := read(node.Fanins[0]), read(node.Fanins[1])
			value := a & b
			if node.Kind == aig.ExclusiveOr {
				value = a ^ b
			}
			sbdCoreSim(p, which, id)[w] = value
		}
	}
	if id != initial {
		for w, value := range Sbd_ObjSim0(p, id) {
			if vec.Vec_IntEntry(p.vMirrors, int(initial))&1 != 0 {
				value = ^value
			}
			Sbd_ObjSim0(p, initial)[w] = value
		}
	}
}
func Sbd_ManWindow(p *Sbd_Man_t, pivot int32) int32 {
	p.Pivot = pivot
	p.vTfo = vec.Vec_WecEntry(p.vTfos, int(pivot))
	vec.Vec_IntClear(p.vWinObjs)
	vec.Vec_IntWriteEntry(p.vObj2Var, 0, 0)
	vec.Vec_IntPush(p.vWinObjs, 0)
	p.seen = map[int32]bool{0: true}
	p.marks = map[int32]bool{}
	Sbd_ManWindowSim_rec(p, pivot)
	if p.pPars.WindowSizeMax != 0 && vec.Vec_IntSize(p.vWinObjs) > int(p.pPars.WindowSizeMax) {
		return 0
	}
	Sbd_ManUpdateOrder(p, pivot)
	p.marks[pivot] = true
	for w, value := range Sbd_ObjSim0(p, pivot) {
		Sbd_ObjSim1(p, pivot)[w] = ^value
	}
	vec.Vec_IntClear(p.vRoots)
	for _, literal := range vec.Vec_IntArray(p.vTfo) {
		p.marks[literal>>1] = true
	}
	for _, literal := range vec.Vec_IntArray(p.vTfo) {
		if literal&1 != 0 {
			Sbd_ManWindowSim_rec(p, literal>>1)
			vec.Vec_IntPush(p.vRoots, literal>>1)
		}
	}
	p.marks[pivot] = false
	for _, literal := range vec.Vec_IntArray(p.vTfo) {
		id := literal >> 1
		p.marks[id] = false
		vec.Vec_IntWriteEntry(p.vObj2Var, int(id), int32(vec.Vec_IntSize(p.vWinObjs)))
		vec.Vec_IntPush(p.vWinObjs, id)
	}
	if p.pPars.WindowSizeMax != 0 && vec.Vec_IntSize(p.vWinObjs) > int(p.pPars.WindowSizeMax) {
		return 0
	}
	for w := range Sbd_ObjSim2(p, pivot) {
		if vec.Vec_IntSize(p.vTfo) == 0 {
			Sbd_ObjSim2(p, pivot)[w] = ^uint64(0)
		} else {
			Sbd_ObjSim2(p, pivot)[w] = 0
		}
	}
	for _, literal := range vec.Vec_IntArray(p.vTfo) {
		if literal&1 != 0 {
			for w, value := range Sbd_ObjSim0(p, literal>>1) {
				Sbd_ObjSim2(p, pivot)[w] |= value ^ Sbd_ObjSim1(p, literal>>1)[w]
			}
		}
	}
	Sbd_ManPropagateControl(p, pivot)
	return 1
}
func Sbd_ManCheckConst(p *Sbd_Man_t, pivot int32) int32 {
	p.pSat = Sbd_ManSatSolver(p.pSat, p.pGia, p.vMirrors, pivot, p.vWinObjs, p.vObj2Var, p.vTfo, p.vRoots, 0)
	if p.pSat != nil {
		p.pSat.SetLimits(int64(p.pPars.ConflictLimit), 0)
	}
	if p.pSat == nil {
		vec.Vec_IntWriteEntry(p.vLutLevs, int(pivot), 0)
		p.nLuts[0]++
		return 0
	}
	counts := [2]int32{}
	vec.Vec_IntClear(p.vLits)
	sim, control := Sbd_ObjSim0(p, pivot)[0], Sbd_ObjSim2(p, pivot)[0]
	for bit := int32(0); bit < 64; bit++ {
		if control>>bit&1 != 0 {
			counts[sim>>bit&1]++
		} else {
			vec.Vec_IntPush(p.vLits, bit)
		}
	}
	if counts[0] > 0 && counts[1] > 0 {
		return -1
	}
	find := int32(0)
	if counts[0] < counts[1] {
		find = 1
	}
	for i := range counts {
		if counts[i] < 1 {
			counts[i] = 1 - counts[i]
		} else {
			counts[i] = 0
		}
	}
	need := int(counts[0] + counts[1])
	if vec.Vec_IntSize(p.vLits) > need {
		vec.Vec_IntShrink(p.vLits, need)
	}
	for bit := int32(0); bit < 64 && vec.Vec_IntSize(p.vLits) < need; bit++ {
		if int32(sim>>bit&1) == find {
			vec.Vec_IntPush(p.vLits, bit)
		}
	}
	pivotVar := vec.Vec_IntEntry(p.vObj2Var, int(pivot))
	sims := make([][]uint64, pivotVar+1)
	for i, id := range vec.Vec_IntArray(p.vWinObjs)[:pivotVar+1] {
		sims[i] = Sbd_ObjSim0(p, id)
	}
	result := Sbd_ManCollectConstants(p.pSat, counts, pivotVar, sims, p.vLits)
	if result >= 0 {
		vec.Vec_IntWriteEntry(p.vLutLevs, int(pivot), 0)
		p.nLuts[0]++
		return result
	}
	if result == -2 {
		return -1
	}
	for _, bit := range vec.Vec_IntArray(p.vLits) {
		Sbd_ObjSim2(p, pivot)[0] |= uint64(1) << bit
	}
	Sbd_ManPropagateControl(p, pivot)
	return -1
}
func Sbd_ManPrintObj(p *Sbd_Man_t, pivot int32) {
	columns := int(vec.Vec_IntEntry(p.vObj2Var, int(pivot))) + 1
	if columns > 64 {
		panic("diagnostic covering matrix exceeds 64 TFI columns")
	}
	for _, counts := range p.vCounts {
		vec.Vec_IntClear(counts)
	}
	ids := vec.Vec_IntArray(p.vWinObjs)[:columns]
	fmt.Printf("Node%d divisors=%d\n", pivot, vec.Vec_IntSize(p.vDivVars))
	for _, id := range ids {
		fmt.Printf("%d : %064b\n", id, Sbd_ObjSim0(p, id)[0])
	}
	for pattern := int32(0); pattern < p.pPars.SimulationWords*64; pattern++ {
		word, bit := pattern>>6, uint(pattern&63)
		if Sbd_ObjSim2(p, pivot)[word]>>bit&1 == 0 {
			continue
		}
		for which := 2; which <= 3; which++ {
			fmt.Printf("%3d : ", pattern)
			for position, id := range ids {
				value := Sbd_ObjSim0(p, id)[word] >> bit & 1
				control := sbdCoreSim(p, which, id)[word] >> bit & 1
				if position == columns-1 && control != 0 {
					vec.Vec_IntPush(p.vCounts[value], pattern)
				}
				if control != 0 {
					fmt.Printf("%d", value)
				} else {
					fmt.Print(".")
				}
			}
			fmt.Println()
		}
	}
	vec.Vec_WrdClear(p.vMatrix)
	seen := map[uint64]bool{}
	off, on := vec.Vec_IntArray(p.vCounts[0]), vec.Vec_IntArray(p.vCounts[1])
	for _, a := range off[:min(64, len(off))] {
		for _, b := range on[:min(64, len(on))] {
			var row uint64
			for col, id := range ids {
				ca, cb := Sbd_ObjSim2(p, id)[a>>6]>>uint(a&63)&1, Sbd_ObjSim2(p, id)[b>>6]>>uint(b&63)&1
				va, vb := Sbd_ObjSim0(p, id)[a>>6]>>uint(a&63)&1, Sbd_ObjSim0(p, id)[b>>6]>>uint(b&63)&1
				if ca != 0 && cb != 0 && va != vb {
					row |= uint64(1) << uint(col)
				}
			}
			if !seen[row] {
				seen[row] = true
				vec.Vec_WrdPush(p.vMatrix, row)
				fmt.Printf("%0*b\n", columns, row)
			}
		}
	}
}
