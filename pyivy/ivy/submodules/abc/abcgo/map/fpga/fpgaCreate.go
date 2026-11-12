package fpga

import (
	"fmt"
	abcmain "github.com/glycerine/abcgo/base/main"
	"math"
	"os"
)

func Fpga_Regular(n *Fpga_Node_t) *Fpga_Node_t {
	if n != nil && n.regular != nil {
		return n.regular
	}
	return n
}
func Fpga_IsComplement(n *Fpga_Node_t) int32 { return boolInt(n != nil && n.regular != nil) }
func Fpga_Not(n *Fpga_Node_t) *Fpga_Node_t {
	if n == nil {
		return nil
	}
	if n.regular != nil {
		return n.regular
	}
	if n.neg == nil {
		n.neg = &Fpga_Node_t{regular: n}
	}
	return n.neg
}
func Fpga_NotCond(n *Fpga_Node_t, condition int32) *Fpga_Node_t {
	if condition != 0 {
		return Fpga_Not(n)
	}
	return n
}
func Fpga_NodeIsConst(n *Fpga_Node_t) int32 {
	n = Fpga_Regular(n)
	return boolInt(n != nil && n.Num == -1)
}
func Fpga_NodeIsVar(n *Fpga_Node_t) int32 {
	n = Fpga_Regular(n)
	return boolInt(n != nil && n.One == nil && n.Num >= 0)
}
func Fpga_NodeIsAnd(n *Fpga_Node_t) int32 {
	n = Fpga_Regular(n)
	return boolInt(n != nil && n.One != nil)
}
func Fpga_NodeIsSimComplement(n *Fpga_Node_t) bool {
	return Fpga_Regular(n).Phase != (Fpga_IsComplement(n) != 0)
}
func NewManager(inputs, outputs int, lib *Fpga_LutLib_t) (*Fpga_Man_t, error) {
	if inputs < 0 || outputs < 0 {
		return nil, fmt.Errorf("fpga: negative interface size")
	}
	if lib == nil {
		lib = DefaultLibrary(4)
	}
	if err := lib.Validate(); err != nil {
		return nil, err
	}
	if lib.LutMax > FPGA_MAX_LEAVES {
		return nil, fmt.Errorf("fpga: cuts support at most %d LUT inputs", FPGA_MAX_LEAVES)
	}
	p := &Fpga_Man_t{Inputs: make([]*Fpga_Node_t, inputs), Outputs: make([]*Fpga_Node_t, outputs), InputArrivals: make([]float32, inputs), Library: Fpga_LutLibDup(lib), AreaRecovery: true, Epsilon: .001, DelayTarget: -1, CutLimit: FPGA_CUTS_MAX_USE, Output: os.Stdout, table: make(map[nodePair]*Fpga_Node_t)}
	p.Const1 = &Fpga_Node_t{Num: -1, Phase: true, Required: FPGA_FLOAT_LARGE, EstimatedFanouts: -1, manager: p}
	p.Const1.Cuts = &Fpga_Cut_t{Root: p.Const1, Truth: 1}
	p.Const1.Best = p.Const1.Cuts
	for i := range p.Inputs {
		p.Inputs[i] = Fpga_NodeCreate(p, nil, nil)
	}
	p.Ands = Fpga_NodeVecAlloc(0)
	p.Mapping = Fpga_NodeVecAlloc(0)
	return p, nil
}
func Fpga_ManCreate(inputs, outputs, verbose int32) *Fpga_Man_t {
	lib, _ := abcmain.Abc_FrameGetGlobalFrame().Resource("pLibLut").(*Fpga_LutLib_t)
	p, err := NewManager(int(inputs), int(outputs), lib)
	if err != nil {
		panic(err)
	}
	p.Verbose = verbose != 0
	return p
}
func Fpga_ManFree(p *Fpga_Man_t) {
	if p != nil {
		p.Inputs = nil
		p.Outputs = nil
		p.Nodes = nil
		p.table = nil
		p.Ands = nil
		p.Mapping = nil
	}
}
func Fpga_NodeCreate(p *Fpga_Man_t, a, b *Fpga_Node_t) *Fpga_Node_t {
	if (a == nil) != (b == nil) {
		panic("fpga: AND requires two children")
	}
	n := &Fpga_Node_t{Num: int32(len(p.Nodes)), One: a, Two: b, Required: FPGA_FLOAT_LARGE, EstimatedFanouts: -1, manager: p}
	if a != nil {
		if Fpga_Regular(a).manager != p || Fpga_Regular(b).manager != p {
			panic("fpga: foreign fanin")
		}
		n.Level = 1 + max(Fpga_NodeReadLevel(a), Fpga_NodeReadLevel(b))
		n.Phase = Fpga_NodeIsSimComplement(a) && Fpga_NodeIsSimComplement(b)
		Fpga_Regular(a).Refs++
		Fpga_Regular(b).Refs++
		Fpga_NodeAddFaninFanout(Fpga_Regular(a), n)
		Fpga_NodeAddFaninFanout(Fpga_Regular(b), n)
	}
	p.Nodes = append(p.Nodes, n)
	return n
}
func nodeLiteral(n *Fpga_Node_t) int64 {
	return int64(Fpga_Regular(n).Num+1)*2 + int64(Fpga_IsComplement(n))
}
func Fpga_HashKey2(a, b *Fpga_Node_t, size int32) uint32 {
	if size <= 0 {
		panic("fpga: invalid hash size")
	}
	return uint32((uint64(nodeLiteral(a)) + uint64(nodeLiteral(b))*12582917) % uint64(size))
}
func Fpga_TableCreate(p *Fpga_Man_t) {
	p.table = make(map[nodePair]*Fpga_Node_t)
	for _, n := range p.Nodes {
		if n.One != nil {
			p.table[nodePair{n.One, n.Two}] = n
		}
	}
}
func Fpga_TableResize(p *Fpga_Man_t) { Fpga_TableCreate(p) }
func Fpga_TableLookup(p *Fpga_Man_t, a, b *Fpga_Node_t) *Fpga_Node_t {
	if nodeLiteral(a) > nodeLiteral(b) {
		a, b = b, a
	}
	return p.table[nodePair{a, b}]
}
func Fpga_NodeAnd(p *Fpga_Man_t, a, b *Fpga_Node_t) *Fpga_Node_t {
	if a == nil || b == nil || Fpga_Regular(a).manager != p || Fpga_Regular(b).manager != p {
		panic("fpga: invalid AND fanin")
	}
	if a == b {
		return a
	}
	if a == Fpga_Not(b) {
		return Fpga_Not(p.Const1)
	}
	if Fpga_Regular(a) == p.Const1 {
		if Fpga_IsComplement(a) != 0 {
			return a
		}
		return b
	}
	if Fpga_Regular(b) == p.Const1 {
		if Fpga_IsComplement(b) != 0 {
			return b
		}
		return a
	}
	if nodeLiteral(a) > nodeLiteral(b) {
		a, b = b, a
	}
	key := nodePair{a, b}
	if n := p.table[key]; n != nil {
		return n
	}
	n := Fpga_NodeCreate(p, a, b)
	p.table[key] = n
	return n
}
func Fpga_NodeOr(p *Fpga_Man_t, a, b *Fpga_Node_t) *Fpga_Node_t {
	return Fpga_Not(Fpga_NodeAnd(p, Fpga_Not(a), Fpga_Not(b)))
}
func Fpga_NodeExor(p *Fpga_Man_t, a, b *Fpga_Node_t) *Fpga_Node_t {
	return Fpga_NodeOr(p, Fpga_NodeAnd(p, a, Fpga_Not(b)), Fpga_NodeAnd(p, Fpga_Not(a), b))
}
func Fpga_NodeMux(p *Fpga_Man_t, c, t, e *Fpga_Node_t) *Fpga_Node_t {
	return Fpga_NodeOr(p, Fpga_NodeAnd(p, c, t), Fpga_NodeAnd(p, Fpga_Not(c), e))
}
func Fpga_NodeSetChoice(p *Fpga_Man_t, old, new *Fpga_Node_t) {
	old, new = Fpga_Regular(old), Fpga_Regular(new)
	if old == nil || new == nil || old == new || old.manager != p || new.manager != p || new.Repr != nil {
		panic("fpga: invalid choice")
	}
	new.NextE = old.NextE
	old.NextE = new
	new.Repr = old
}
func Fpga_NodeComparePhase(a, b *Fpga_Node_t) int32 {
	return boolInt(Fpga_Regular(a).Phase != Fpga_Regular(b).Phase)
}
func Fpga_ManReadInputNum(p *Fpga_Man_t) int32          { return int32(len(p.Inputs)) }
func Fpga_ManReadOutputNum(p *Fpga_Man_t) int32         { return int32(len(p.Outputs)) }
func Fpga_ManReadInputs(p *Fpga_Man_t) []*Fpga_Node_t   { return p.Inputs }
func Fpga_ManReadOutputs(p *Fpga_Man_t) []*Fpga_Node_t  { return p.Outputs }
func Fpga_ManReadConst1(p *Fpga_Man_t) *Fpga_Node_t     { return p.Const1 }
func Fpga_ManReadInputArrivals(p *Fpga_Man_t) []float32 { return p.InputArrivals }
func Fpga_ManReadVerbose(p *Fpga_Man_t) int32           { return boolInt(p.Verbose) }
func Fpga_ManReadVarMax(p *Fpga_Man_t) int32            { return p.Library.LutMax }
func Fpga_ManReadLutAreas(p *Fpga_Man_t) []float32      { return p.Library.Areas }
func Fpga_ManReadMapping(p *Fpga_Man_t) *Fpga_NodeVec_t { return p.Mapping }
func Fpga_ManSetOutputNames(p *Fpga_Man_t, names []string) {
	p.OutputNames = append([]string(nil), names...)
}
func Fpga_ManSetInputArrivals(p *Fpga_Man_t, arrivals []float32) {
	p.InputArrivals = append([]float32(nil), arrivals...)
}
func Fpga_ManSetAreaRecovery(p *Fpga_Man_t, value int32) { p.AreaRecovery = value != 0 }
func Fpga_ManSetDelayLimit(p *Fpga_Man_t, v float32)     { p.DelayLimit = v }
func Fpga_ManSetAreaLimit(p *Fpga_Man_t, v float32)      { p.AreaLimit = v }
func Fpga_ManSetChoiceNodeNum(p *Fpga_Man_t, v int32)    { p.ChoiceNodes = v }
func Fpga_ManSetChoiceNum(p *Fpga_Man_t, v int32)        { p.Choices = v }
func Fpga_ManSetVerbose(p *Fpga_Man_t, v int32)          { p.Verbose = v != 0 }
func Fpga_ManSetSwitching(p *Fpga_Man_t, v int32)        { p.Switching = v != 0 }
func Fpga_ManSetLatchPaths(p *Fpga_Man_t, v int32)       { p.LatchPaths = v != 0 }
func Fpga_ManSetLatchNum(p *Fpga_Man_t, v int32)         { p.Latches = v }
func Fpga_ManSetDelayTarget(p *Fpga_Man_t, v float32)    { p.DelayTarget = v }
func Fpga_ManSetName(p *Fpga_Man_t, v string)            { p.Name = v }
func Fpga_LibReadLutMax(lib *Fpga_LutLib_t) int32        { return lib.LutMax }
func Fpga_NodeReadData0(n *Fpga_Node_t) any              { return Fpga_Regular(n).Data0 }
func Fpga_NodeReadData1(n *Fpga_Node_t) any              { return Fpga_Regular(n).Data1 }
func Fpga_NodeSetData0(n *Fpga_Node_t, data any)         { Fpga_Regular(n).Data0 = data }
func Fpga_NodeSetData1(n *Fpga_Node_t, data any)         { Fpga_Regular(n).Data1 = data }
func Fpga_NodeReadRefs(n *Fpga_Node_t) int32             { return Fpga_Regular(n).Refs }
func Fpga_NodeReadNum(n *Fpga_Node_t) int32              { return Fpga_Regular(n).Num }
func Fpga_NodeReadLevel(n *Fpga_Node_t) int32            { return Fpga_Regular(n).Level }
func Fpga_NodeReadCuts(n *Fpga_Node_t) *Fpga_Cut_t       { return Fpga_Regular(n).Cuts }
func Fpga_NodeReadCutBest(n *Fpga_Node_t) *Fpga_Cut_t    { return Fpga_Regular(n).Best }
func Fpga_NodeReadOne(n *Fpga_Node_t) *Fpga_Node_t       { return Fpga_Regular(n).One }
func Fpga_NodeReadTwo(n *Fpga_Node_t) *Fpga_Node_t       { return Fpga_Regular(n).Two }
func Fpga_NodeSetNextE(n, next *Fpga_Node_t)             { Fpga_Regular(n).NextE = Fpga_Regular(next) }
func Fpga_NodeSetRepr(n, repr *Fpga_Node_t)              { Fpga_Regular(n).Repr = Fpga_Regular(repr) }
func Fpga_NodeSetSwitching(n *Fpga_Node_t, value float32) {
	if math.IsNaN(float64(value)) || value < 0 || value > 1 {
		panic("fpga: invalid switching probability")
	}
	Fpga_Regular(n).Switching = value
}
func Fpga_CutReadLeavesNum(c *Fpga_Cut_t) int32       { return int32(len(Fpga_CutRegular(c).Leaves)) }
func Fpga_CutReadLeaves(c *Fpga_Cut_t) []*Fpga_Node_t { return Fpga_CutRegular(c).Leaves }
func Fpga_ManPrintTimeStats(p *Fpga_Man_t) {
	fmt.Fprintf(p.Output, "Cuts=%s Match=%s Recovery=%s Total=%s\n", p.Stats.CutTime, p.Stats.MatchTime, p.Stats.RecoveryTime, p.Stats.Total)
}
func Fpga_ManStats(p *Fpga_Man_t) MappingStatistics {
	stats := p.Stats
	stats.Area = p.AreaGlobal
	stats.Delay = Fpga_TimeComputeArrivalMax(p)
	stats.LUTs = len(p.Mapping.Nodes)
	return stats
}
