package saig

// Automatic phase abstraction follows saigPhase.c and Bjesse/Kukula's
// ICCAD 2005 algorithm, using owned packed states instead of pointer overlays.
import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
	"os"
)

const (
	TSIM_MAX_ROUNDS       = 10000
	TSIM_ONE_SERIES       = 3000
	SAIG_XVS0       int32 = 1
	SAIG_XVS1       int32 = 2
	SAIG_XVSX       int32 = 3
)

func Saig_XsimConvertValue(value int32) int32 {
	if value >= 0 && value <= 2 {
		return value + 1
	}
	return -1
}
func Saig_ObjSetXsim(graph *aig.Graph, obj aig.Edge, value int32) {
	aig.Aig_ObjSetXsim(graph, obj, value)
}
func Saig_ObjGetXsim(graph *aig.Graph, obj aig.Edge) int32 { return aig.Aig_ObjGetXsim(graph, obj) }
func Saig_XsimInv(value int32) int32                       { return aig.Aig_XsimInv(value) }
func Saig_XsimAnd(a, b int32) int32                        { return aig.Aig_XsimAnd(a, b) }
func Saig_XsimRand2() int32                                { return 1 + int32(aig.Aig_ManRandom(0)&1) }
func Saig_XsimRand3() int32 {
	for {
		value := int32(aig.Aig_ManRandom(0) & 3)
		if value != 0 {
			return value
		}
	}
}
func Saig_ObjGetXsimFanin0(graph *aig.Graph, obj aig.Edge) int32 {
	return aig.Aig_ObjGetXsimFanin0(graph, obj)
}
func Saig_ObjGetXsimFanin1(graph *aig.Graph, obj aig.Edge) int32 {
	return aig.Aig_ObjGetXsimFanin1(graph, obj)
}
func Saig_XsimPrint(writer io.Writer, value int32) { aig.Aig_XsimPrint(writer, value) }

type Saig_Tsim_t struct {
	pAig            *aig.Graph
	fast            *Faig_Man_t
	states          [][]uint32
	seen            map[string]int
	nPrefix, nCycle int32
	vNonXRegs       *vec.Vec_Int_t
}

func Saig_TsiStart(graph *aig.Graph) *Saig_Tsim_t {
	return &Saig_Tsim_t{pAig: graph, fast: Faig_ManCreate(graph), seen: make(map[string]int)}
}
func Saig_TsiStop(p *Saig_Tsim_t) {
	if p != nil {
		p.states = nil
		p.seen = nil
		p.vNonXRegs = nil
	}
}
func (p *Saig_Tsim_t) words() int { return (2*int(aig.Aig_ManRegNum(p.pAig)) + 31) / 32 }
func phaseKey(state []uint32, words int32) string {
	if words < 0 || int(words) > len(state) {
		panic("saig: invalid ternary state width")
	}
	key := make([]byte, int(words)*4)
	for i, value := range state[:words] {
		binary.LittleEndian.PutUint32(key[i*4:], value)
	}
	return string(key)
}
func phaseValue(state []uint32, reg int) int32 { return int32(state[reg/16] >> uint(2*(reg%16)) & 3) }
func phaseSet(state []uint32, reg int, value int32) {
	if value < 1 || value > 3 {
		panic("saig: invalid ternary value")
	}
	shift := uint(2 * (reg % 16))
	state[reg/16] = state[reg/16]&^(3<<shift) | uint32(value)<<shift
}
func Saig_TsiStateHash(state []uint32, words, bins int32) int32 {
	return aig.Aig_TsiStateHash(state, words, bins)
}
func Saig_TsiStateLookup(p *Saig_Tsim_t, state []uint32, words int32) int32 {
	if _, ok := p.seen[phaseKey(state, words)]; ok {
		return 1
	}
	return 0
}
func Saig_TsiStateInsert(p *Saig_Tsim_t, state []uint32, words int32) {
	key := phaseKey(state, words)
	if _, ok := p.seen[key]; ok {
		panic("saig: duplicate ternary state")
	}
	index := -1
	last := len(p.states) - 1
	if last >= 0 && (len(state) == 0 || len(p.states[last]) > 0 && &p.states[last][0] == &state[0]) {
		index = last
	} else {
		for i, s := range p.states {
			if len(state) > 0 && len(s) > 0 && &s[0] == &state[0] {
				index = i
				break
			}
		}
	}
	if index < 0 {
		panic("saig: ternary state is not owned by manager")
	}
	p.seen[key] = index
}
func Saig_TsiComputePrefix(p *Saig_Tsim_t, state []uint32, words int32) int32 {
	if index, ok := p.seen[phaseKey(state, words)]; ok {
		return int32(index)
	}
	return -1
}
func Saig_TsiStateNew(p *Saig_Tsim_t) []uint32 {
	state := make([]uint32, p.words())
	p.states = append(p.states, state)
	return state
}
func Saig_TsiStateCount(p *Saig_Tsim_t, state []uint32) int32 {
	count := int32(0)
	for i := 0; i < int(aig.Aig_ManRegNum(p.pAig)); i++ {
		if value := phaseValue(state, i); value == SAIG_XVS0 || value == SAIG_XVS1 {
			count++
		}
	}
	return count
}
func Saig_TsiStatePrint(p *Saig_Tsim_t, state []uint32) {
	for i := 0; i < int(aig.Aig_ManRegNum(p.pAig)); i++ {
		Saig_XsimPrint(os.Stdout, phaseValue(state, i))
	}
	fmt.Println()
}
func Saig_TsiStateOrAll(p *Saig_Tsim_t, state []uint32) {
	for _, previous := range p.states {
		for i, value := range previous {
			state[i] |= value
		}
	}
}
func Saig_TsiCountNonXValuedRegisters(p *Saig_Tsim_t, prefix int32) int32 {
	if prefix < 0 || int(prefix) >= len(p.states) {
		panic("saig: invalid ternary prefix")
	}
	p.vNonXRegs = vec.Vec_IntAlloc(16)
	for reg := 0; reg < int(aig.Aig_ManRegNum(p.pAig)); reg++ {
		binary := true
		for _, state := range p.states[prefix:] {
			if phaseValue(state, reg) == SAIG_XVSX {
				binary = false
				break
			}
		}
		if binary {
			vec.Vec_IntPush(p.vNonXRegs, int32(reg))
		}
	}
	return int32(vec.Vec_IntSize(p.vNonXRegs))
}
func Saig_TsiComputeTransient(p *Saig_Tsim_t, prefix int32) *vec.Vec_Int_t {
	if prefix < 0 {
		panic("saig: negative transient length")
	}
	counters := vec.Vec_IntStart(int(prefix))
	for reg := 0; reg < int(aig.Aig_ManRegNum(p.pAig)); reg++ {
		previous, lastStep := int32(-1), -1
		for index, state := range p.states {
			value := phaseValue(state, reg)
			if previous != value {
				previous = value
				lastStep = index
			}
		}
		if previous != SAIG_XVSX && lastStep >= 0 && lastStep < int(prefix) {
			vec.Vec_IntAddToEntry(counters, lastStep, 1)
		}
	}
	return counters
}
func Saig_TsiPrintTraces(p *Saig_Tsim_t, words, prefix, loop int32) {
	fmt.Println("Ternary traces for each register:")
	for reg := 0; reg < int(aig.Aig_ManRegNum(p.pAig)); reg++ {
		fmt.Printf("%5d : ", reg)
		for frame, state := range p.states[:len(p.states)-1] {
			if frame == int(prefix) {
				fmt.Print("  ")
			}
			Saig_XsimPrint(os.Stdout, phaseValue(state, reg))
		}
		fmt.Println()
	}
}
func phaseLit(lit int32, values []int32) int32 {
	value := values[lit>>1]
	if lit&1 != 0 {
		value = Saig_XsimInv(value)
	}
	return value
}
func Saig_ManReachableTernary(graph *aig.Graph, initial *vec.Vec_Int_t, verbose int32) *Saig_Tsim_t {
	p := Saig_TsiStart(graph)
	registers := int(aig.Aig_ManRegNum(graph))
	if initial != nil && vec.Vec_IntSize(initial) != registers {
		panic("saig: invalid initial state dimensions")
	}
	values := make([]int32, p.fast.size())
	values[0] = SAIG_XVS1
	for id := int32(1); id <= p.fast.nPis; id++ {
		values[id] = SAIG_XVSX
	}
	for reg := 0; reg < registers; reg++ {
		value := SAIG_XVS0
		if initial != nil {
			value = Saig_XsimConvertValue(vec.Vec_IntEntry(initial, reg))
			if value < 0 {
				panic("saig: invalid initial ternary value")
			}
		}
		values[int(p.fast.nPis)+1+reg] = value
	}
	for frame := 0; frame < TSIM_MAX_ROUNDS; frame++ {
		state := Saig_TsiStateNew(p)
		for reg := 0; reg < registers; reg++ {
			phaseSet(state, reg, values[int(p.fast.nPis)+1+reg])
		}
		if Saig_TsiStateLookup(p, state, int32(p.words())) != 0 {
			p.nPrefix = Saig_TsiComputePrefix(p, state, int32(p.words()))
			p.nCycle = int32(len(p.states)) - 1 - p.nPrefix
			if verbose != 0 {
				fmt.Printf("Ternary simulation converged after %d frames\n", frame)
			}
			return p
		}
		Saig_TsiStateInsert(p, state, int32(p.words()))
		for i, node := range p.fast.nodes {
			a := phaseLit(node.fanins[0], values)
			value := a
			if node.kind == aig.And {
				value = Saig_XsimAnd(a, phaseLit(node.fanins[1], values))
			} else if node.kind == aig.ExclusiveOr {
				b := phaseLit(node.fanins[1], values)
				value = SAIG_XVSX
				if a != SAIG_XVSX && b != SAIG_XVSX {
					value = SAIG_XVS0
					if a != b {
						value = SAIG_XVS1
					}
				}
			}
			values[int(p.fast.nodeStart())+i] = value
		}
		next := make([]int32, registers)
		for reg := range next {
			next[reg] = phaseLit(p.fast.outputs[int(p.fast.pos())+reg], values)
		}
		for reg, value := range next {
			id := int(p.fast.nPis) + 1 + reg
			if frame < TSIM_ONE_SERIES {
				values[id] = value
			} else if values[id] != value {
				values[id] = SAIG_XVSX
			}
		}
	}
	Saig_TsiStop(p)
	return nil
}
func Saig_ManAnalizeControl(graph *aig.Graph, reg int32) {
	registers := aig.Aig_ManRegNum(graph)
	if reg < 0 || reg >= registers {
		panic("saig: invalid control register")
	}
	inputs := graph.Inputs()
	a, b := inputs[len(inputs)-int(registers)+int(reg)], inputs[len(inputs)-1]
	for id, node := range graph.Nodes() {
		if node.Kind == aig.And && node.Fanins[0].NodeID() == a.NodeID() && node.Fanins[1].NodeID() == b.NodeID() {
			fmt.Printf("Registers %d and %d share AND node %d\n", reg, registers-1, id)
			return
		}
	}
	fmt.Println("Register fanout not found")
}
func Saig_ManFindRegisters(p *Saig_Tsim_t, frames, ignore, verbose int32) int32 {
	if frames < 1 || frames > 256 {
		panic("saig: invalid phase frame count")
	}
	selected := vec.Vec_IntAlloc(16)
	tests := p.nPrefix + 2*p.nCycle
	for _, reg := range vec.Vec_IntArray(p.vNonXRegs) {
		values := make([]int32, frames)
		valid := true
		for frame := int32(0); frame < tests; frame++ {
			index := frame
			if index >= p.nPrefix+p.nCycle {
				index -= p.nCycle
			}
			value := phaseValue(p.states[index], int(reg))
			if value == SAIG_XVSX {
				valid = false
				break
			}
			if frame < frames || ignore != 0 && frame == frames {
				values[frame%frames] = value
			} else if values[frame%frames] != value {
				valid = false
				break
			}
		}
		if !valid {
			continue
		}
		if ignore != 0 {
			constant := true
			for _, value := range values[1:] {
				if value != values[0] {
					constant = false
					break
				}
			}
			if constant {
				continue
			}
		}
		vec.Vec_IntPush(selected, reg)
		if verbose != 0 {
			fmt.Printf("Register %d has phase generator %v\n", reg, values)
		}
	}
	p.vNonXRegs = selected
	return int32(vec.Vec_IntSize(selected))
}
func Saig_ObjFrames(mapping []aig.Edge, frames int32, obj aig.Edge, frame int32) aig.Edge {
	return mapping[int(frames)*obj.NodeID()+int(frame)]
}
func Saig_ObjSetFrames(mapping []aig.Edge, frames int32, obj aig.Edge, frame int32, value aig.Edge) {
	mapping[int(frames)*obj.NodeID()+int(frame)] = value
}
func Saig_ObjChild0Frames(graph *aig.Graph, mapping []aig.Edge, frames int32, obj aig.Edge, frame int32) aig.Edge {
	edge := graph.Node(obj).Fanins[0]
	value := Saig_ObjFrames(mapping, frames, edge, frame)
	if edge.Complemented() && value != aig.InvalidEdge {
		value = value.Not()
	}
	return value
}
func Saig_ObjChild1Frames(graph *aig.Graph, mapping []aig.Edge, frames int32, obj aig.Edge, frame int32) aig.Edge {
	edge := graph.Node(obj).Fanins[1]
	value := Saig_ObjFrames(mapping, frames, edge, frame)
	if edge.Complemented() && value != aig.InvalidEdge {
		value = value.Not()
	}
	return value
}

func Saig_ManPerformAbstraction(p *Saig_Tsim_t, frames, verbose int32) *aig.Graph {
	if frames < 1 || frames > 256 || int(frames) > len(p.states) || vec.Vec_IntSize(p.vNonXRegs) == 0 {
		panic("saig: no phase abstraction candidates")
	}
	source := p.pAig
	registers := int(aig.Aig_ManRegNum(source))
	pis := source.InputCount() - registers
	pos := source.OutputCount() - registers
	graph := aig.NewGraph()
	graph.Name = source.Name
	graph.Spec = source.Spec
	frameInputs := make([][]aig.Edge, frames)
	for frame := range frameInputs {
		frameInputs[frame] = make([]aig.Edge, pis)
		for input := 0; input < pis; input++ {
			frameInputs[frame][input] = graph.NewInput()
		}
	}
	state := make([]aig.Edge, registers)
	for reg := range state {
		state[reg] = graph.NewInput()
	}
	outputs := make([][]aig.Edge, frames)
	for frame := 0; frame < int(frames); frame++ {
		copies := map[int]aig.Edge{0: graph.Const1()}
		for i, input := range source.Inputs() {
			if i < pis {
				copies[input.NodeID()] = frameInputs[frame][i]
			} else {
				copies[input.NodeID()] = state[i-pis]
			}
		}
		for _, reg := range vec.Vec_IntArray(p.vNonXRegs) {
			value := phaseValue(p.states[frame], int(reg))
			edge := graph.Const0()
			if value == SAIG_XVS1 {
				edge = graph.Const1()
			} else if value != SAIG_XVS0 {
				panic("saig: selected register is not binary")
			}
			copies[source.Inputs()[pis+int(reg)].NodeID()] = edge
		}
		var copyEdge func(aig.Edge) aig.Edge
		copyEdge = func(edge aig.Edge) aig.Edge {
			value, ok := copies[edge.NodeID()]
			if !ok {
				node := source.Node(aig.Edge(edge.NodeID()) << 1)
				a := copyEdge(node.Fanins[0])
				switch node.Kind {
				case aig.And:
					value = graph.AND(a, copyEdge(node.Fanins[1]))
				case aig.ExclusiveOr:
					value = graph.NewXor(a, copyEdge(node.Fanins[1]))
				case aig.Buffer, aig.Output:
					value = a
				default:
					panic("saig: invalid phase abstraction node")
				}
				copies[edge.NodeID()] = value
			}
			if edge.Complemented() {
				value = value.Not()
			}
			return value
		}
		outputs[frame] = make([]aig.Edge, pos)
		for i, output := range source.Outputs() {
			value := copyEdge(source.Node(output).Fanins[0])
			if i < pos {
				outputs[frame][i] = value
			} else {
				state[i-pos] = value
			}
		}
	}
	constraints := int(aig.Aig_ManConstrNum(source))
	properties := pos - constraints
	for _, frame := range outputs {
		for _, output := range frame[:properties] {
			graph.AddOutput(output)
		}
	}
	for _, frame := range outputs {
		for _, output := range frame[properties:] {
			graph.AddOutput(output)
		}
	}
	for _, next := range state {
		graph.AddOutput(next)
	}
	aig.Aig_ManSetRegNum(graph, int32(registers))
	aig.Aig_ManSetConstrNum(graph, int32(constraints)*frames)
	aig.Aig_ManSeqCleanup(graph)
	return graph
}
func Saig_ManPhaseFrameNum(graph *aig.Graph, initial *vec.Vec_Int_t) int32 {
	p := Saig_ManReachableTernary(graph, initial, 0)
	if p == nil {
		return 1
	}
	defer Saig_TsiStop(p)
	return p.nCycle
}
func Saig_ManPhasePrefixLength(graph *aig.Graph, verbose, veryVerbose int32, transient **vec.Vec_Int_t) int32 {
	p := Saig_ManReachableTernary(graph, nil, 0)
	if p == nil {
		return 0
	}
	defer Saig_TsiStop(p)
	Saig_TsiCountNonXValuedRegisters(p, p.nPrefix)
	if transient != nil {
		*transient = Saig_TsiComputeTransient(p, p.nPrefix)
	}
	if verbose != 0 {
		fmt.Printf("Lead=%d Loop=%d Binary=%d\n", p.nPrefix, p.nCycle, vec.Vec_IntSize(p.vNonXRegs))
	}
	if veryVerbose != 0 {
		Saig_TsiPrintTraces(p, int32(p.words()), p.nPrefix, p.nCycle)
	}
	return p.nPrefix
}
func phaseAbstractManager(p *Saig_Tsim_t, frames, prefix, ignore, printOnly, verbose int32) *aig.Graph {
	Saig_TsiCountNonXValuedRegisters(p, min(p.nPrefix, prefix))
	if printOnly != 0 {
		Saig_TsiPrintTraces(p, int32(p.words()), p.nPrefix, p.nCycle)
		return nil
	}
	if frames < 2 || frames > 256 || p.nCycle == 1 || p.nCycle%frames != 0 || vec.Vec_IntSize(p.vNonXRegs) == 0 {
		return nil
	}
	if Saig_ManFindRegisters(p, frames, ignore, verbose) == 0 {
		return nil
	}
	return Saig_ManPerformAbstraction(p, frames, verbose)
}
func Saig_ManPhaseAbstract(graph *aig.Graph, initial *vec.Vec_Int_t, frames, prefix, ignore, printOnly, verbose int32) *aig.Graph {
	p := Saig_ManReachableTernary(graph, initial, verbose)
	if p == nil {
		return nil
	}
	defer Saig_TsiStop(p)
	return phaseAbstractManager(p, frames, prefix, ignore, printOnly, verbose)
}
func Saig_ManPhaseAbstractAuto(graph *aig.Graph, verbose int32) *aig.Graph {
	p := Saig_ManReachableTernary(graph, nil, verbose)
	if p == nil {
		return nil
	}
	result := phaseAbstractManager(p, p.nCycle, 0, 0, 0, verbose)
	Saig_TsiStop(p)
	if result == nil || result.InputCount() == int(aig.Aig_ManRegNum(result)) {
		return aig.Aig_ManDupSimple(graph)
	}
	return result
}
func Saig_PhaseTranslateCex(graph *aig.Graph, cex *util.Abc_Cex_t) *util.Abc_Cex_t {
	registers := aig.Aig_ManRegNum(graph)
	pis := int32(graph.InputCount()) - registers
	pos := int32(graph.OutputCount()) - registers
	if cex == nil || pis <= 0 || pos <= 0 || cex.NPis <= 0 || cex.NPis%pis != 0 || cex.IFrame < 0 || cex.IPo < 0 {
		return nil
	}
	expectedBits := int64(cex.NRegs) + (int64(cex.IFrame)+1)*int64(cex.NPis)
	if cex.NRegs < 0 || expectedBits != int64(cex.NBits) || int64(len(cex.PData))*32 < expectedBits {
		return nil
	}
	frames := cex.NPis / pis
	if int64(cex.IPo) >= int64(pos)*int64(frames) {
		return nil
	}
	constraints := aig.Aig_ManConstrNum(graph)
	properties := pos - constraints
	var microframe, output int32
	if int64(cex.IPo) < int64(properties)*int64(frames) {
		microframe = cex.IPo / properties
		output = cex.IPo % properties
	} else {
		index := cex.IPo - properties*frames
		if constraints <= 0 {
			return nil
		}
		microframe = index / constraints
		output = properties + index%constraints
	}
	expandedFrame := int64(cex.IFrame)*int64(frames) + int64(microframe)
	if expandedFrame >= 1<<31-1 || int64(registers)+(expandedFrame+1)*int64(pis) > 1<<31-1 {
		return nil
	}
	frame := int32(expandedFrame)
	result := util.Abc_CexAlloc(registers, pis, frame+1)
	result.IFrame = frame
	result.IPo = output
	for src, dst := cex.NRegs, result.NRegs; dst < result.NBits; src, dst = src+1, dst+1 {
		if src >= cex.NBits || int(src>>5) >= len(cex.PData) {
			return nil
		}
		if cex.PData[src>>5]&(1<<uint(src&31)) != 0 {
			result.PData[dst>>5] |= 1 << uint(dst&31)
		}
	}
	return result
}
