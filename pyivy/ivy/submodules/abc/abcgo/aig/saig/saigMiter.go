package saig

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/proof/cec/core"
	"time"
)

// Sec_MiterStatus classifies outputs by constants, primary inputs, and the
// all-zero input phase; it is a quick structural test rather than a SAT proof.
func Sec_MiterStatus(p *aig.Graph) Sec_MtrStatus_t {
	s := Sec_MtrStatus_t{nInputs: Saig_ManPiNum(p), nNodes: aig.Aig_ManNodeNum(p), nOutputs: Saig_ManPoNum(p), iOut: -1}
	for i, o := range p.Outputs()[:s.nOutputs] {
		e := aig.Aig_ObjChild0(p, o)
		switch {
		case e == p.Const0():
			s.nUnsat++
		case e == p.Const1() || Saig_ObjIsPi(p, e) || p.Node(e).Phase != e.Complemented():
			s.nSat++
			if s.iOut < 0 {
				s.iOut = int32(i)
			}
		default:
			s.nUndec++
		}
	}
	return s
}

func miterMap(src *aig.Graph, dst *aig.Graph) []aig.Edge {
	copies := make([]aig.Edge, aig.Aig_ManObjNumMax(src))
	for i := range copies {
		copies[i] = aig.InvalidEdge
	}
	copies[0] = dst.Const1()
	return copies
}
func miterMapped(copies []aig.Edge, e aig.Edge) aig.Edge {
	c := copies[e.NodeID()]
	if c == aig.InvalidEdge {
		panic("saig: unmapped miter fanin")
	}
	if e.Complemented() {
		c = c.Not()
	}
	return c
}

// Missing inputs deliberately leave their entire dependent cone unmapped when
// extracting one independent register half.
func miterCopyLogic(src, dst *aig.Graph, copies []aig.Edge) {
	for _, e := range src.TopologicalObjects() {
		n := src.Node(e)
		if n.Kind == aig.Constant || n.Kind == aig.Input {
			continue
		}
		a := n.Fanins[0]
		if copies[a.NodeID()] == aig.InvalidEdge {
			continue
		}
		x := miterMapped(copies, a)
		switch n.Kind {
		case aig.Buffer, aig.Output:
			copies[e.NodeID()] = x
		case aig.And, aig.ExclusiveOr:
			b := n.Fanins[1]
			if copies[b.NodeID()] == aig.InvalidEdge {
				continue
			}
			y := miterMapped(copies, b)
			if n.Kind == aig.And {
				copies[e.NodeID()] = dst.AND(x, y)
			} else {
				copies[e.NodeID()] = dst.XOR(x, y)
			}
		}
	}
}
func miterOperation(dst *aig.Graph, a, b aig.Edge, oper int32) aig.Edge {
	switch oper {
	case 0:
		return dst.XOR(a, b)
	case 1:
		return dst.AND(a, b.Not())
	default:
		panic("saig: miter operation must be XOR or implication")
	}
}
func miterCompatible(a, b *aig.Graph, sequential bool) {
	if sequential {
		if Saig_ManPiNum(a) != Saig_ManPiNum(b) || Saig_ManPoNum(a) != Saig_ManPoNum(b) || Saig_ManRegNum(a)+Saig_ManRegNum(b) == 0 {
			panic("saig: incompatible sequential interfaces")
		}
	} else if a.InputCount() != b.InputCount() || len(a.Outputs()) != len(b.Outputs()) {
		panic("saig: incompatible combinational interfaces")
	}
}
func Saig_ManCreateMiter(p0, p1 *aig.Graph, oper int32) *aig.Graph {
	result, _, _ := Saig_ManCreateMiterWithCopies(p0, p1, oper)
	return result
}

// Saig_ManCreateMiterWithCopies additionally returns the literal mapping of
// each source object to its copy in the miter. A mapped literal retains phase;
// an unavailable object is aig.InvalidEdge.
func Saig_ManCreateMiterWithCopies(p0, p1 *aig.Graph, oper int32) (*aig.Graph, []aig.Edge, []aig.Edge) {
	miterCompatible(p0, p1, true)
	dst := aig.NewGraph()
	dst.Name = "miter"
	c0, c1 := miterMap(p0, dst), miterMap(p1, dst)
	in0, in1 := p0.Inputs(), p1.Inputs()
	pi := int(Saig_ManPiNum(p0))
	for i := 0; i < pi; i++ {
		e := dst.NewInput()
		c0[in0[i].NodeID()] = e
		c1[in1[i].NodeID()] = e
	}
	for _, e := range in0[pi:] {
		c0[e.NodeID()] = dst.NewInput()
	}
	for _, e := range in1[pi:] {
		c1[e.NodeID()] = dst.NewInput()
	}
	miterCopyLogic(p0, dst, c0)
	miterCopyLogic(p1, dst, c1)
	out0, out1 := p0.Outputs(), p1.Outputs()
	po := int(Saig_ManPoNum(p0))
	for i := 0; i < po; i++ {
		dst.AddOutput(miterOperation(dst, miterMapped(c0, out0[i]), miterMapped(c1, out1[i]), oper))
	}
	for _, e := range out0[po:] {
		dst.AddOutput(miterMapped(c0, e))
	}
	for _, e := range out1[po:] {
		dst.AddOutput(miterMapped(c1, e))
	}
	aig.Aig_ManSetRegNum(dst, Saig_ManRegNum(p0)+Saig_ManRegNum(p1))
	return dst, c0, c1
}
func Saig_ManCreateMiterComb(p0, p1 *aig.Graph, oper int32) *aig.Graph {
	miterCompatible(p0, p1, false)
	dst := aig.NewGraph()
	dst.Name = "miter"
	c0, c1 := miterMap(p0, dst), miterMap(p1, dst)
	in1 := p1.Inputs()
	for i, e := range p0.Inputs() {
		x := dst.NewInput()
		c0[e.NodeID()] = x
		c1[in1[i].NodeID()] = x
	}
	miterCopyLogic(p0, dst, c0)
	miterCopyLogic(p1, dst, c1)
	out1 := p1.Outputs()
	for i, e := range p0.Outputs() {
		dst.AddOutput(miterOperation(dst, miterMapped(c0, e), miterMapped(c1, out1[i]), oper))
	}
	aig.Aig_ManCleanup(dst)
	return dst
}

// Saig_AndDualRail takes explicit false/true rail storage instead of borrowing
// the source objects' pData/pNext scratch fields.
func Saig_AndDualRail(dst, src *aig.Graph, obj aig.Edge, rails [][2]aig.Edge) (aig.Edge, aig.Edge) {
	n := src.Node(obj)
	a, b := rails[n.Fanins[0].NodeID()], rails[n.Fanins[1].NodeID()]
	if n.Fanins[0].Complemented() {
		a[0], a[1] = a[1], a[0]
	}
	if n.Fanins[1].Complemented() {
		b[0], b[1] = b[1], b[0]
	}
	return dst.OR(dst.AND(a[0], a[1].Not()), dst.AND(b[0], b[1].Not())), dst.AND(dst.AND(a[0].Not(), a[1]), dst.AND(b[0].Not(), b[1]))
}
func Saig_ManDualRail(src *aig.Graph, miter int32) *aig.Graph {
	dst := aig.NewGraph()
	dst.Name, dst.Spec = src.Name, src.Spec
	rails := make([][2]aig.Edge, aig.Aig_ManObjNumMax(src))
	rails[0] = [2]aig.Edge{dst.Const0(), dst.Const1()}
	for _, e := range src.Inputs() {
		rails[e.NodeID()] = [2]aig.Edge{dst.NewInput(), dst.NewInput()}
	}
	rail := func(e aig.Edge) [2]aig.Edge {
		r := rails[e.NodeID()]
		if e.Complemented() {
			r[0], r[1] = r[1], r[0]
		}
		return r
	}
	for _, e := range src.TopologicalObjects() {
		n := src.Node(e)
		switch n.Kind {
		case aig.And:
			rails[e.NodeID()][0], rails[e.NodeID()][1] = Saig_AndDualRail(dst, src, e, rails)
		case aig.ExclusiveOr:
			a, b := rail(n.Fanins[0]), rail(n.Fanins[1])
			a0, a1 := dst.AND(a[0], a[1].Not()), dst.AND(a[1], a[0].Not())
			b0, b1 := dst.AND(b[0], b[1].Not()), dst.AND(b[1], b[0].Not())
			rails[e.NodeID()] = [2]aig.Edge{dst.OR(dst.AND(a0, b0), dst.AND(a1, b1)), dst.OR(dst.AND(a0, b1), dst.AND(a1, b0))}
		case aig.Buffer, aig.Output:
			rails[e.NodeID()] = rail(n.Fanins[0])
		}
	}
	outs := src.Outputs()
	if miter != 0 {
		r := dst.Const1()
		for _, e := range src.Inputs()[Saig_ManPiNum(src):] {
			x := rail(e)
			r = dst.AND(r, dst.OR(x[0], x[1]))
		}
		dst.AddOutput(r)
		outs = outs[Saig_ManPoNum(src):]
	}
	for _, e := range outs {
		r := rail(e)
		dst.AddOutput(r[0])
		dst.AddOutput(r[1])
	}
	aig.Aig_ManSetRegNum(dst, 2*Saig_ManRegNum(src))
	aig.Aig_ManCleanup(dst)
	return dst
}

func Saig_ManUnrollTwo(bot, top *aig.Graph, nFrames int32) *aig.Graph {
	miterCompatible(bot, top, true)
	if Saig_ManRegNum(bot) != Saig_ManRegNum(top) || nFrames < 1 {
		panic("saig: unrolling needs matching registers and positive frame count")
	}
	dst := aig.NewGraph()
	dst.Name = "frames"
	state := make([]aig.Edge, Saig_ManRegNum(bot))
	for i := range state {
		state[i] = dst.NewInput()
	}
	src := bot
	for f := int32(0); f < nFrames; f++ {
		copies := miterMap(src, dst)
		pi, po := int(Saig_ManPiNum(src)), int(Saig_ManPoNum(src))
		for i, e := range src.Inputs() {
			if i < pi {
				copies[e.NodeID()] = dst.NewInput()
			} else {
				copies[e.NodeID()] = state[i-pi]
			}
		}
		miterCopyLogic(src, dst, copies)
		outs := src.Outputs()
		for i, e := range outs {
			v := miterMapped(copies, e)
			if i < po || f == nFrames-1 {
				dst.AddOutput(v)
			}
			if i >= po {
				state[i-po] = v
			}
		}
		src = top
	}
	aig.Aig_ManCleanup(dst)
	return dst
}

func miterDupSelected(src *aig.Graph, selected []aig.Edge, part int) *aig.Graph {
	dst := aig.NewGraph()
	dst.Name = src.Name
	copies := miterMap(src, dst)
	pi, po := int(Saig_ManPiNum(src)), int(Saig_ManPoNum(src))
	regs := int(Saig_ManRegNum(src))
	first, last := 0, regs
	if part >= 0 {
		if regs%2 != 0 {
			panic("saig: register halves need an even count")
		}
		if part == 0 {
			last = regs / 2
		} else {
			first = regs / 2
		}
	}
	for i, e := range src.Inputs() {
		if i < pi || i-pi >= first && i-pi < last {
			copies[e.NodeID()] = dst.NewInput()
		}
	}
	miterCopyLogic(src, dst, copies)
	for _, e := range selected {
		dst.AddOutput(miterMapped(copies, e))
	}
	outs := src.Outputs()
	for _, e := range outs[po+first : po+last] {
		dst.AddOutput(miterMapped(copies, e))
	}
	aig.Aig_ManSetRegNum(dst, int32(last-first))
	if part < 0 {
		aig.Aig_ManSeqCleanup(dst)
	} else {
		aig.Aig_ManCleanup(dst)
	}
	return aig.Aig_ManDupSimpleDfs(dst)
}
func Aig_ManDupNodesAll(src *aig.Graph, selected []aig.Edge) *aig.Graph {
	return miterDupSelected(src, selected, -1)
}
func Aig_ManDupNodesHalf(src *aig.Graph, selected []aig.Edge, part int32) *aig.Graph {
	if part != 0 && part != 1 {
		panic("saig: half must be zero or one")
	}
	return miterDupSelected(src, selected, int(part))
}

func miterSplitXor(src *aig.Graph, o aig.Edge) (aig.Edge, aig.Edge, bool) {
	driver := aig.Aig_ObjChild0(src, o)
	a, b, ok := aig.Aig_ObjRecognizeExor(src, driver&^1)
	if ok && driver.Complemented() {
		a = a.Not()
	}
	return a, b, ok
}
func miterOrdered(a, b aig.Edge) (aig.Edge, aig.Edge) {
	if a.NodeID() < b.NodeID() {
		return a, b
	}
	return b, a
}
func miterDemiterSimple(src *aig.Graph, left, right **aig.Graph, all, skipBad bool) int32 {
	if !all && Saig_ManRegNum(src)%2 != 0 {
		panic("saig: demiter requires even register count")
	}
	var set0, set1 []aig.Edge
	for i, o := range src.Outputs()[:Saig_ManPoNum(src)] {
		driver := aig.Aig_ObjChild0(src, o)
		if driver.NodeID() == 0 {
			if driver == src.Const1() {
				fmt.Printf("The output number %d of the miter is constant 1.\n", i)
			}
			continue
		}
		a, b, ok := miterSplitXor(src, o)
		if !ok {
			if skipBad {
				fmt.Printf("The output number %d cannot be demitered.\n", i)
				continue
			}
			fmt.Println("The miter cannot be demitered.")
			return 0
		}
		a, b = miterOrdered(a, b)
		set0 = append(set0, a)
		set1 = append(set1, b)
	}
	if left != nil {
		if all {
			*left = Aig_ManDupNodesAll(src, set0)
		} else {
			*left = Aig_ManDupNodesHalf(src, set0, 0)
		}
		(*left).Name = "part0"
	}
	if right != nil {
		if all {
			*right = Aig_ManDupNodesAll(src, set1)
		} else {
			*right = Aig_ManDupNodesHalf(src, set1, 1)
		}
		(*right).Name = "part1"
	}
	return 1
}
func Saig_ManDemiterSimple(src *aig.Graph, left, right **aig.Graph) int32 {
	return miterDemiterSimple(src, left, right, false, false)
}
func Saig_ManDemiterSimpleDiff_old(src *aig.Graph, left, right **aig.Graph) int32 {
	return miterDemiterSimple(src, left, right, true, true)
}

func Saig_ManDemiterMarkPos(src *aig.Graph) {
	aig.Aig_ManCleanMarkAB(src)
	for i, e := range src.Inputs()[Saig_ManPiNum(src):] {
		if i < int(Saig_ManRegNum(src)/2) {
			aig.Aig_ObjSetMarkA(src, e)
		} else {
			aig.Aig_ObjSetMarkB(src, e)
		}
	}
	for _, e := range src.TopologicalObjects() {
		n := src.Node(e)
		if n.Kind != aig.And && n.Kind != aig.ExclusiveOr && n.Kind != aig.Buffer {
			continue
		}
		a := src.Node(n.Fanins[0])
		ma, mb := a.MarkA, a.MarkB
		if n.Kind != aig.Buffer {
			b := src.Node(n.Fanins[1])
			ma = ma || b.MarkA
			mb = mb || b.MarkB
		}
		if ma {
			aig.Aig_ObjSetMarkA(src, e)
		}
		if mb {
			aig.Aig_ObjSetMarkB(src, e)
		}
	}
}
func Saig_ManDemiterCheckPo(src *aig.Graph, o aig.Edge, left, right *aig.Edge) int32 {
	if !Saig_ObjIsPo(src, o) {
		panic("saig: demiter expects primary output")
	}
	d := aig.Aig_ObjChild0(src, o)
	if d.NodeID() == 0 {
		if d == src.Const1() {
			return 0
		}
		*left, *right = src.Const0(), src.Const0()
		return 1
	}
	a, b, ok := miterSplitXor(src, o)
	if !ok {
		return 0
	}
	x, y := src.Node(a), src.Node(b)
	if x.MarkA && x.MarkB || y.MarkA && y.MarkB || x.MarkA && y.MarkA || x.MarkB && y.MarkB {
		return 0
	}
	// A support-free side can belong to either half; a labeled side must
	// still remain in its own register half, regardless of object ID.
	if y.MarkA || x.MarkB {
		a, b = b, a
	} else if !x.MarkA && !y.MarkB {
		a, b = miterOrdered(a, b)
	}
	*left, *right = a, b
	return 1
}
func Saig_ManDemiterSimpleDiff(src *aig.Graph, left, right **aig.Graph) int32 {
	if Saig_ManRegNum(src) == 0 || Saig_ManRegNum(src)%2 != 0 {
		return 0
	}
	Saig_ManDemiterMarkPos(src)
	defer aig.Aig_ManCleanMarkAB(src)
	var set0, set1 []aig.Edge
	for _, o := range src.Outputs()[:Saig_ManPoNum(src)] {
		var a, b aig.Edge
		if Saig_ManDemiterCheckPo(src, o, &a, &b) == 0 {
			return 0
		}
		set0 = append(set0, a)
		set1 = append(set1, b)
	}
	a, b := Aig_ManDupNodesHalf(src, set0, 0), Aig_ManDupNodesHalf(src, set1, 1)
	a.Name, b.Name = "part0", "part1"
	if left != nil {
		*left = a
	}
	if right != nil {
		*right = b
	}
	return 1
}
func Saig_ManDemiterDual(src *aig.Graph, left, right **aig.Graph) int32 {
	var sets [2][]aig.Edge
	for i, e := range src.Outputs()[:Saig_ManPoNum(src)] {
		sets[i&1] = append(sets[i&1], aig.Aig_ObjChild0(src, e))
	}
	a, b := Aig_ManDupNodesAll(src, sets[0]), Aig_ManDupNodesAll(src, sets[1])
	if left != nil {
		*left = a
	}
	if right != nil {
		*right = b
	}
	return 1
}

func Saig_ManDemiterLabel_rec(src *aig.Graph, e aig.Edge, value int32) {
	e &^= 1
	if aig.Aig_ObjIsTravIdCurrent(src, e) {
		return
	}
	aig.Aig_ObjSetTravIdCurrent(src, e)
	if value != 0 {
		aig.Aig_ObjSetMarkB(src, e)
	} else {
		aig.Aig_ObjSetMarkA(src, e)
	}
	n := src.Node(e)
	if n.Kind == aig.Constant || Saig_ObjIsPi(src, e) {
		return
	}
	if Saig_ObjIsLo(src, e) {
		Saig_ManDemiterLabel_rec(src, aig.Aig_ObjChild0(src, Saig_ObjLoToLi(src, e)), value)
		return
	}
	Saig_ManDemiterLabel_rec(src, n.Fanins[0], value)
	if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
		Saig_ManDemiterLabel_rec(src, n.Fanins[1], value)
	}
}
func Saig_ManGetLabeledRegister_rec(src *aig.Graph, e aig.Edge) aig.Edge {
	e &^= 1
	if aig.Aig_ObjIsTravIdCurrent(src, e) {
		return aig.InvalidEdge
	}
	aig.Aig_ObjSetTravIdCurrent(src, e)
	n := src.Node(e)
	if n.Kind == aig.Constant || Saig_ObjIsPi(src, e) {
		return aig.InvalidEdge
	}
	if Saig_ObjIsLo(src, e) {
		if n.MarkA || n.MarkB {
			return e
		}
		return Saig_ManGetLabeledRegister_rec(src, aig.Aig_ObjChild0(src, Saig_ObjLoToLi(src, e)))
	}
	r := Saig_ManGetLabeledRegister_rec(src, n.Fanins[0])
	if r != aig.InvalidEdge {
		return r
	}
	if n.Kind == aig.And || n.Kind == aig.ExclusiveOr {
		return Saig_ManGetLabeledRegister_rec(src, n.Fanins[1])
	}
	return aig.InvalidEdge
}

// The original general demiter is unfinished and asserts before producing either
// requested graph. Preserve that explicit failure, while retaining its labeling
// analysis for callers requesting diagnostics only.
func Saig_ManDemiter(src *aig.Graph, left, right **aig.Graph) int32 {
	if Saig_ManRegNum(src) == 0 {
		panic("saig: general demiter needs registers")
	}
	var pairs [][2]aig.Edge
	for _, o := range src.Outputs()[:Saig_ManPoNum(src)] {
		a, b, ok := aig.Aig_ObjRecognizeExor(src, aig.Aig_ObjFanin0(src, o))
		if !ok {
			return 0
		}
		pairs = append(pairs, [2]aig.Edge{a, b})
	}
	if len(pairs) == 0 {
		return 0
	}
	for i, pair := range pairs {
		for j, e := range pair {
			aig.Aig_ManIncrementTravId(src)
			value := int32(j)
			if i > 0 {
				reg := Saig_ManGetLabeledRegister_rec(src, e)
				value = 0
				if reg != aig.InvalidEdge && src.Node(reg).MarkB {
					value = 1
				}
				// The search visited this cone; labeling needs a fresh traversal.
				aig.Aig_ManIncrementTravId(src)
			}
			Saig_ManDemiterLabel_rec(src, e, value)
		}
	}
	shared := 0
	for _, e := range src.Inputs()[Saig_ManPiNum(src):] {
		n := src.Node(e)
		if n.MarkA && n.MarkB {
			shared++
		}
	}
	if shared > 0 {
		fmt.Printf("The miters contains %d flops reachable from both AIGs.\n", shared)
	}
	if left != nil || right != nil {
		panic("saig: original general demiter graph construction is unfinished")
	}
	return 1
}
func Saig_ManCreateMiterTwo(old, new *aig.Graph, nFrames int32) *aig.Graph {
	return Saig_ManCreateMiterComb(Saig_ManUnrollTwo(old, old, nFrames), Saig_ManUnrollTwo(new, old, nFrames), 0)
}
func Ssw_SecCexResimulate(src *aig.Graph, model []int32, nOutputs *int32) int32 {
	if len(model) != src.InputCount() {
		panic("saig: counterexample input count differs")
	}
	in := make([]bool, len(model))
	for i, v := range model {
		in[i] = v != 0
	}
	out, err := src.Evaluate(in)
	if err != nil {
		panic(err)
	}
	*nOutputs = 0
	first := int32(-1)
	for i, v := range out {
		if v {
			if first < 0 {
				first = int32(i)
			}
			*nOutputs++
		}
	}
	aig.Aig_ManCleanMarkA(src)
	return first
}

// The native SAT backend proves the same combinational unrolling obligation.
// It retains the source's 100000-conflict limit, without importing the legacy
// FRA package (which depends on SAIG).
func Ssw_SecSpecial(a, b *aig.Graph, nFrames, verbose int32) int32 {
	start := time.Now()
	m := Saig_ManCreateMiterTwo(a, b, nFrames)
	result, err := core.VerifyMiter(miterNetwork{m}, core.Options{ConflictLimit: 100000})
	if err != nil {
		panic(err)
	}
	status := int32(result.Status)
	switch status {
	case 1:
		fmt.Print("Networks are equivalent.  ")
	case 0:
		fmt.Print("Networks are NOT EQUIVALENT.  ")
	default:
		fmt.Print("Networks are UNDECIDED.   ")
	}
	fmt.Printf("Time = %.2f sec\n", time.Since(start).Seconds())
	if result.Counterexample != nil {
		model := make([]int32, len(result.Counterexample.Inputs))
		for i, v := range result.Counterexample.Inputs {
			if v {
				model[i] = 1
			}
		}
		var count int32
		first := Ssw_SecCexResimulate(m, model, &count)
		if first < 0 {
			panic("saig: SAT counterexample failed resimulation")
		}
		po := Saig_ManPoNum(a)
		if first < po*nFrames {
			fmt.Printf("Primary output %d has failed in frame %d.\n", first%po, first/po)
		} else {
			fmt.Printf("Flop input %d has failed in the last frame.\n", first-po*nFrames)
		}
		fmt.Printf("The counter-example detected %d incorrect POs or flop inputs.\n", count)
	}
	return status
}
func Ssw_SecSpecialMiter(a, b *aig.Graph, nFrames, verbose int32) int32 {
	var p0, p1 *aig.Graph
	if b == nil {
		if Saig_ManDemiterSimpleDiff(a, &p0, &p1) == 0 {
			fmt.Println("Demitering has failed.")
			return -1
		}
		if Saig_ManRegNum(p0) != Saig_ManRegNum(p1) {
			fmt.Println("After demitering AIGs have different number of flops. Quitting.")
			return -1
		}
	} else {
		p0, p1 = aig.Aig_ManDupSimple(a), aig.Aig_ManDupSimple(b)
	}
	miterCompatible(p0, p1, true)
	if Saig_ManRegNum(p0) != Saig_ManRegNum(p1) {
		panic("saig: sequential verification requires matching registers")
	}
	r := Ssw_SecSpecial(p0, p1, nFrames, verbose)
	if r != 1 && aig.Aig_ManNodeNum(p0) >= aig.Aig_ManNodeNum(p1) {
		r = Ssw_SecSpecial(p1, p0, nFrames, verbose)
	}
	return r
}

// miterNetwork adapts the AIG constant-one convention to the proof core's
// constant-zero convention without rebuilding the graph.
type miterNetwork struct{ g *aig.Graph }

func miterLit(e aig.Edge) int32 {
	if e.NodeID() == 0 {
		return int32(e) ^ 1
	}
	return int32(e)
}
func (n miterNetwork) ObjectCount() int { return int(aig.Aig_ManObjNumMax(n.g)) }
func (n miterNetwork) InputLiterals() []int32 {
	out := make([]int32, n.g.InputCount())
	for i, e := range n.g.Inputs() {
		out[i] = miterLit(e)
	}
	return out
}
func (n miterNetwork) OutputLiterals() []int32 {
	edges := n.g.Outputs()
	out := make([]int32, len(edges))
	for i, e := range edges {
		out[i] = miterLit(e)
	}
	return out
}
func (n miterNetwork) Gate(id int32) (int32, int32, int32, int) {
	e := aig.Edge(id) << 1
	if !n.g.ValidEdge(e) {
		return 0, 0, 0, core.Absent
	}
	x := n.g.Node(e)
	switch x.Kind {
	case aig.Constant:
		return 0, 0, 0, core.Constant
	case aig.Input:
		return 0, 0, 0, core.Input
	case aig.Output:
		return miterLit(x.Fanins[0]), 0, 0, core.Output
	case aig.Buffer:
		return miterLit(x.Fanins[0]), 0, 0, core.Buffer
	case aig.And:
		return miterLit(x.Fanins[0]), miterLit(x.Fanins[1]), 0, core.And
	case aig.ExclusiveOr:
		return miterLit(x.Fanins[0]), miterLit(x.Fanins[1]), 0, core.Xor
	}
	panic("saig: unknown gate")
}

func Saig_ManDemiterNew(src *aig.Graph) int32 {
	for i, o := range src.Outputs()[:Saig_ManPoNum(src)-Saig_ManConstrNum(src)] {
		fmt.Printf("Output %3d : ", i)
		d := aig.Aig_ObjChild0(src, o)
		if d.NodeID() == 0 {
			if d == src.Const1() {
				fmt.Println("Const1")
			} else {
				fmt.Println("Const0")
			}
			continue
		}
		if !aig.Aig_ObjIsNode(src, d) {
			fmt.Println("Terminal")
			continue
		}
		a, b, ok := aig.Aig_ObjRecognizeExor(src, d&^1)
		if !d.Complemented() {
			fmt.Printf("AND   %s\n", miterYesNo(ok))
			continue
		}
		var super []aig.Edge
		aig.Aig_ObjCollectSuper(src, d&^1, &super)
		fmt.Printf("OR with %d inputs    ", len(super))
		if len(super) == 2 {
			fmt.Print(miterYesNo(ok))
			if ok {
				fmt.Println()
				for _, e := range []aig.Edge{a, b} {
					for _, s := range aig.Aig_Support(src, aig.Aig_Regular(e)) {
						if Saig_ObjIsLo(src, s) {
							fmt.Printf(" %d", aig.Aig_ObjCioId(src, s))
						}
					}
					fmt.Println()
				}
			}
		}
		fmt.Println()
	}
	return 1
}
func miterYesNo(ok bool) string {
	if ok {
		return "Yes"
	}
	return "No"
}
