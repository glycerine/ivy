package fra

// Simulation management and FRAIG refinement, ported from fraSim.c.
import (
	"fmt"
	"math/bits"
	"os"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

func fraBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Fra_SmlNodeHash(s *Fra_Sml_t, obj aig.Edge, size int32) int32 {
	return int32(s.Hash(obj.NodeID()) % uint64(size))
}
func Fra_SmlNodeIsConst(s *Fra_Sml_t, obj aig.Edge) int32 { return fraBool(s.IsConstant(obj.NodeID())) }
func Fra_SmlNodesAreEqual(s *Fra_Sml_t, a, b aig.Edge) int32 {
	return fraBool(s.Equal(a.NodeID(), b.NodeID()))
}
func Fra_SmlNodeNotEquWeight(s *Fra_Sml_t, a, b int32) int32 {
	return int32(s.DifferenceWeight(int(a), int(b)))
}
func Fra_SmlNodeIsZero(s *Fra_Sml_t, obj aig.Edge) int32    { return fraBool(s.IsConstant(obj.NodeID())) }
func Fra_SmlNodeCountOnes(s *Fra_Sml_t, obj aig.Edge) int32 { return int32(s.CountOnes(obj.NodeID())) }
func Fra_SmlSavePattern0(p *Fra_Man_t, _ int32)             { clear(p.pPatWords) }
func Fra_SmlSavePattern1(p *Fra_Man_t, init int32) {
	for i := range p.pPatWords {
		p.pPatWords[i] = ^uint32(0)
	}
	if init != 0 {
		regs := int(aig.Aig_ManRegNum(p.pManAig))
		pis := len(p.pManAig.Inputs()) - regs
		for i := 0; i < regs; i++ {
			bit := pis*int(p.nFramesAll) + i
			p.pPatWords[bit/32] &^= uint32(1) << uint(bit%32)
		}
	}
}
func Fra_SmlSavePattern(p *Fra_Man_t) {
	clear(p.pPatWords)
	for i, input := range p.pManFraig.Inputs() {
		value, _ := p.pSat.Value(Fra_ObjSatNum(p, input))
		if value {
			util.Abc_InfoSetBit(p.pPatWords, int32(i))
		}
	}
	if p.vCex != nil {
		vec.Vec_IntClear(p.vCex)
		pis := len(p.pManAig.Inputs()) - int(aig.Aig_ManRegNum(p.pManAig))
		push := func(i int) { vec.Vec_IntPush(p.vCex, int32((p.pPatWords[i/32]>>uint(i%32))&1)) }
		for i := 0; i < pis; i++ {
			push(i)
		}
		inputs := len(p.pManFraig.Inputs())
		regs := int(aig.Aig_ManRegNum(p.pManFraig))
		for i := inputs - regs; i < inputs; i++ {
			push(i)
		}
	}
}
func Fra_SmlCheckOutputSavePattern(p *Fra_Man_t, output aig.Edge) {
	s := p.pSml
	fanin := p.pManAig.Node(output).Fanins[0]
	words := s.NodeWords(fanin.NodeID())
	word := -1
	for i, value := range words {
		if value != 0 {
			word = i
			break
		}
	}
	if word < 0 {
		panic("fra: no failing simulation pattern")
	}
	bit := 32*word + bits.TrailingZeros32(words[word])
	inputs := p.pManAig.Inputs()
	model := make([]int32, len(p.pManFraig.Inputs())+1)
	if len(model) <= len(inputs) {
		panic("fra: model has too few inputs")
	}
	for i, input := range inputs {
		model[i] = int32((s.NodeWords(input.NodeID())[bit/32] >> uint(bit%32)) & 1)
	}
	model[len(inputs)] = int32(output.NodeID())
	if p.pManFraig.Data != nil {
		panic("fra: output model already exists")
	}
	p.pManFraig.Data = model
}
func Fra_SmlCheckOutput(p *Fra_Man_t) int32 {
	for _, output := range p.pManAig.Outputs() {
		fanin := p.pManAig.Node(output).Fanins[0]
		if !p.pSml.IsConstant(fanin.NodeID()) {
			Fra_SmlCheckOutputSavePattern(p, output)
			return 1
		}
	}
	return 0
}
func Fra_SmlAssignRandom(s *Fra_Sml_t, obj aig.Edge) {
	for i := range s.NodeWords(obj.NodeID()) {
		s.NodeWords(obj.NodeID())[i] = aig.Aig_ManRandom(0)
	}
}
func Fra_SmlAssignConst(s *Fra_Sml_t, obj aig.Edge, value, frame int32) {
	s.AssignFrame(obj.NodeID(), int(frame), value != 0)
}
func Fra_SmlInitialize(s *Fra_Sml_t, init int32)                  { s.Initialize(init != 0) }
func Fra_SmlAssignDist1(s *Fra_Sml_t, pattern []uint32)           { s.AssignDistance1(pattern) }
func Fra_SmlNodeSimulate(s *Fra_Sml_t, obj aig.Edge, frame int32) { s.SimulateNode(obj, int(frame)) }
func Fra_SmlNodesCompareInFrame(s *Fra_Sml_t, a, b aig.Edge, frameA, frameB int32) int32 {
	x, y := s.FrameWords(a.NodeID(), int(frameA)), s.FrameWords(b.NodeID(), int(frameB))
	for i := range x {
		if x[i] != y[i] {
			return 0
		}
	}
	return 1
}
func Fra_SmlNodeCopyFanin(s *Fra_Sml_t, obj aig.Edge, frame int32) {
	nodes := s.Graph.Nodes()
	child := nodes[obj.NodeID()].Fanins[0]
	src, dst := s.FrameWords(child.NodeID(), int(frame)), s.FrameWords(obj.NodeID(), int(frame))
	flip := nodes[child.NodeID()].Phase != child.Complemented()
	for i := range dst {
		dst[i] = src[i]
		if flip {
			dst[i] = ^dst[i]
		}
	}
}
func Fra_SmlNodeTransferNext(s *Fra_Sml_t, out, in aig.Edge, frame int32) {
	copy(s.FrameWords(in.NodeID(), int(frame)+1), s.FrameWords(out.NodeID(), int(frame)))
}
func Fra_SmlCheckNonConstOutputs(s *Fra_Sml_t) int32 { return fraBool(s.HasNonconstantOutput()) }
func Fra_SmlSimulateOne(s *Fra_Sml_t)                { s.SimulateOne() }
func Fra_SmlResimulate(p *Fra_Man_t) {
	p.pCla.simulation = p.pSml
	p.pSml.AssignDistance1(p.pPatWords)
	p.pSml.SimulateOne()
	if p.pPars.FProve != 0 && Fra_SmlCheckOutput(p) != 0 {
		return
	}
	start := time.Now()
	changes := Fra_ClassesRefine(p.pCla) + Fra_ClassesRefine1(p.pCla, 1, nil)
	if p.pCla.vImps != nil {
		changes += Fra_ImpRefineUsingCex(p, p.pCla.vImps)
	}
	if p.vOneHots != nil {
		changes += Fra_OneHotRefineUsingCex(p, p.vOneHots)
	}
	p.timeRef += time.Since(start)
	if p.pPars.NFramesK == 0 && changes < 1 {
		fmt.Println("Error: A counter-example did not refine classes!")
	}
}
func Fra_SmlSimulate(p *Fra_Man_t, init int32) {
	p.pCla.simulation = p.pSml
	run := func() bool { p.pSml.SimulateOne(); return p.pPars.FProve != 0 && Fra_SmlCheckOutput(p) != 0 }
	refine := func() int32 {
		start := time.Now()
		n := Fra_ClassesRefine(p.pCla) + Fra_ClassesRefine1(p.pCla, 1, nil)
		p.timeRef += time.Since(start)
		return n
	}
	p.pSml.Initialize(init != 0)
	if run() {
		return
	}
	Fra_ClassesPrepare(p.pCla, p.pPars.FLatchCorr, 0)
	for _, one := range []bool{false, true} {
		if one {
			Fra_SmlSavePattern1(p, init)
		} else {
			Fra_SmlSavePattern0(p, init)
		}
		p.pSml.AssignDistance1(p.pPatWords)
		if run() {
			return
		}
		refine()
	}
	for {
		p.pSml.Initialize(init != 0)
		if run() {
			return
		}
		count := len(p.pCla.classes)
		changes := refine()
		if count == 0 || float64(changes)/float64(count) <= p.pPars.DSimSatur {
			return
		}
	}
}
func Fra_SmlStart(g *aig.Graph, prefix, frames, words int32) *Fra_Sml_t {
	return newFraSimulation(g, int(prefix), int(frames), int(words))
}
func Fra_SmlStop(_ *Fra_Sml_t) {}
func Fra_SmlSimulateComb(g *aig.Graph, words, check int32) *Fra_Sml_t {
	s := Fra_SmlStart(g, 0, 1, words)
	s.Initialize(false)
	s.SimulateOne()
	if check != 0 {
		s.NonconstantOutput = s.HasNonconstantOutput()
	}
	return s
}
func Fra_SmlSimulateReadFile(name string) *vec.Vec_Str_t {
	data, err := os.ReadFile(name)
	if err != nil {
		fmt.Printf("Cannot open file %q with simulation patterns.\n", name)
		return nil
	}
	result := vec.Vec_StrAlloc(len(data))
	for _, c := range data {
		switch c {
		case '0', '1':
			vec.Vec_StrPush(result, c-'0')
		case ' ', '\r', '\n', '\t':
		default:
			fmt.Printf("File %q contains symbol (%c) other than '0' or '1'.\n", name, c)
			return nil
		}
	}
	return result
}
func Fra_SmlInitializeGiven(s *Fra_Sml_t, values *vec.Vec_Str_t) {
	s.InitializeGiven(vec.Vec_StrArray(values))
}
func Fra_SmlPrintOutputs(s *Fra_Sml_t, patterns int32) {
	for i := 0; i < int(patterns); i++ {
		for _, out := range s.Graph.Outputs() {
			fmt.Printf("%d", (s.NodeWords(out.NodeID())[i/32]>>uint(i%32))&1)
		}
		fmt.Println()
	}
}
func Fra_SmlSimulateCombGiven(g *aig.Graph, name string, check, verbose int32) *Fra_Sml_t {
	if aig.Aig_ManRegNum(g) != 0 {
		panic("fra: expected combinational AIG")
	}
	values := Fra_SmlSimulateReadFile(name)
	if values == nil {
		return nil
	}
	n := vec.Vec_StrSize(values)
	inputs := len(g.Inputs())
	if inputs == 0 || n == 0 || n%inputs != 0 {
		fmt.Printf("File %q: invalid number of input digits (%d).\n", name, n)
		return nil
	}
	patterns := n / inputs
	s := Fra_SmlStart(g, 0, 1, int32((patterns+31)/32))
	Fra_SmlInitializeGiven(s, values)
	s.SimulateOne()
	if check != 0 {
		s.NonconstantOutput = s.HasNonconstantOutput()
	}
	if verbose != 0 {
		Fra_SmlPrintOutputs(s, int32(patterns))
	}
	return s
}
func Fra_SmlSimulateSeq(g *aig.Graph, prefix, frames, words, check int32) *Fra_Sml_t {
	s := Fra_SmlStart(g, prefix, frames, words)
	s.Initialize(true)
	s.SimulateOne()
	if check != 0 {
		s.NonconstantOutput = s.HasNonconstantOutput()
	}
	return s
}
func Fra_SmlGetCounterExample(s *Fra_Sml_t) *util.Abc_Cex_t { return s.CounterExample() }
func Fra_SmlCopyCounterExample(g, frames *aig.Graph, model []int32) *util.Abc_Cex_t {
	return copyFrameCounterExample(g, frames, model)
}
