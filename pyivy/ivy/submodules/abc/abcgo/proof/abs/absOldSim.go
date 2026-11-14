package abs

import (
	"fmt"
	"time"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// evalOldSimNode computes a node's value for a frame. The second encoding
// tracks visited binary values; the first supports an unknown value.
func evalOldSimNode(graph *aig.Graph, object aig.Edge, frame int32, info *vec.Vec_Ptr_t, visited bool) int32 {
	node := graph.Node(object)
	get := Saig_ManSimInfoGet
	set := Saig_ManSimInfoSet
	not := Saig_ManSimInfoNot
	and := Saig_ManSimInfoAnd
	if visited {
		get, set = Saig_ManSimInfo2Get, Saig_ManSimInfo2Set
		not, and = Saig_ManSimInfo2Not, Saig_ManSimInfo2And
	}
	fanin := aig.Aig_Regular(node.Fanins[0])
	a := get(info, &fanin, frame)
	if node.Fanins[0].Complemented() {
		a = not(a)
	}
	value := a
	if node.Kind != aig.Output && node.Kind != aig.Buffer {
		if node.Kind != aig.And {
			panic("abstraction: ternary simulation expects AND nodes")
		}
		fanin = aig.Aig_Regular(node.Fanins[1])
		b := get(info, &fanin, frame)
		if node.Fanins[1].Complemented() {
			b = not(b)
		}
		value = and(a, b)
	}
	set(info, &object, frame, value)
	return value
}

func Saig_ManExtendOneEval(graph *aig.Graph, info *vec.Vec_Ptr_t, object aig.Edge, frame int32) int32 {
	return evalOldSimNode(graph, object, frame, info, false)
}
func Saig_ManExtendOneEval2(graph *aig.Graph, info *vec.Vec_Ptr_t, object aig.Edge, frame int32) int32 {
	return evalOldSimNode(graph, object, frame, info, true)
}

// simulateOldCex initializes a frame-by-frame trace from a concrete CEX. In
// ternary mode, entries named by unknownInputs are replaced with unknowns.
func simulateOldCex(graph *aig.Graph, cex *util.Abc_Cex_t, info *vec.Vec_Ptr_t, unknownInputs *vec.Vec_Int_t, visited bool) int32 {
	set := Saig_ManSimInfoSet
	zero, one := SAIG_ZER, SAIG_ONE
	if visited {
		set, zero, one = Saig_ManSimInfo2Set, SAIG_ZER_NEW, SAIG_ONE_NEW
	}
	bit := int32(0)
	bitValue := func() int32 {
		v := zero
		if util.Abc_InfoHasBit(cex.PData, bit) {
			v = one
		}
		bit++
		return v
	}
	inputs, outputs := graph.Inputs(), graph.Outputs()
	piCount := saig.Saig_ManPiNum(graph)
	regCount := saig.Saig_ManRegNum(graph)
	for i := int32(0); i < regCount; i++ {
		object := saig.Saig_ManLo(graph, i)
		set(info, &object, 0, bitValue())
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		constant := graph.Const1()
		set(info, &constant, frame, one)
		for i := int32(0); i < piCount; i++ {
			object := inputs[i]
			set(info, &object, frame, bitValue())
		}
		if !visited && unknownInputs != nil {
			for _, input := range vec.Vec_IntArray(unknownInputs) {
				object := inputs[input]
				set(info, &object, frame, SAIG_UND)
			}
		}
		// ABC traverses internal nodes before combinational outputs.
		for _, object := range graph.LiveEdges() {
			node := graph.Node(object)
			if node.Kind == aig.And || node.Kind == aig.ExclusiveOr || node.Kind == aig.Buffer {
				evalOldSimNode(graph, object, frame, info, visited)
			}
		}
		for _, object := range outputs {
			evalOldSimNode(graph, object, frame, info, visited)
		}
		if frame == cex.IFrame {
			break
		}
		for i := int32(0); i < regCount; i++ {
			li, lo := saig.Saig_ManLi(graph, i), saig.Saig_ManLo(graph, i)
			get := Saig_ManSimInfoGet
			if visited {
				get = Saig_ManSimInfo2Get
			}
			set(info, &lo, frame+1, get(info, &li, frame))
		}
	}
	output := outputs[cex.IPo]
	if visited {
		return Saig_ManSimInfo2Get(info, &output, cex.IFrame)
	}
	return Saig_ManSimInfoGet(info, &output, cex.IFrame)
}

func Saig_ManSimDataInit(graph *aig.Graph, cex *util.Abc_Cex_t, info *vec.Vec_Ptr_t, unknownInputs *vec.Vec_Int_t) int32 {
	return simulateOldCex(graph, cex, info, unknownInputs, false)
}
func Saig_ManSimDataInit2(graph *aig.Graph, cex *util.Abc_Cex_t, info *vec.Vec_Ptr_t) int32 {
	return simulateOldCex(graph, cex, info, nil, true)
}

// Saig_ManSetAndDriveImplications_rec marks a value and propagates the
// consequences to fanouts and across latch boundaries.
func Saig_ManSetAndDriveImplications_rec(graph *aig.Graph, object aig.Edge, frame, last int32, info *vec.Vec_Ptr_t) {
	value := Saig_ManSimInfo2Get(info, &object, frame)
	if Saig_ManSimInfo2IsOld(value) != 0 {
		panic("abstraction: simulation value already visited")
	}
	Saig_ManSimInfo2Set(info, &object, frame, Saig_ManSimInfo2SetOld(value))
	if (aig.Aig_ObjIsCo(graph, object) && frame == last) || saig.Saig_ObjIsPo(graph, object) {
		return
	}
	if saig.Saig_ObjIsLi(graph, object) {
		if frame >= last {
			panic("abstraction: latch input beyond final frame")
		}
		fanout := saig.Saig_ObjLiToLo(graph, object)
		if Saig_ManSimInfo2IsOld(Saig_ManSimInfo2Get(info, &fanout, frame+1)) == 0 {
			Saig_ManSetAndDriveImplications_rec(graph, fanout, frame+1, last, info)
		}
		return
	}
	for i, fan := int32(0), aig.Aig_ObjFanout0Int(graph, int32(object.NodeID())); i < aig.Aig_ObjRefs(graph, object); i, fan = i+1, aig.Aig_ObjFanoutNext(graph, fan) {
		fanout := aig.Edge(uint32(fan>>1) << 1)
		value = Saig_ManSimInfo2Get(info, &fanout, frame)
		if Saig_ManSimInfo2IsOld(value) != 0 {
			continue
		}
		if aig.Aig_ObjIsCo(graph, fanout) {
			Saig_ManSetAndDriveImplications_rec(graph, fanout, frame, last, info)
			continue
		}
		node := graph.Node(fanout)
		if node.Kind != aig.And {
			panic("abstraction: sensitization expects AND nodes")
		}
		a, b := aig.Aig_Regular(node.Fanins[0]), aig.Aig_Regular(node.Fanins[1])
		av, bv := Saig_ManSimInfo2Get(info, &a, frame), Saig_ManSimInfo2Get(info, &b, frame)
		if node.Fanins[0].Complemented() {
			av = Saig_ManSimInfo2Not(av)
		}
		if node.Fanins[1].Complemented() {
			bv = Saig_ManSimInfo2Not(bv)
		}
		if av == SAIG_ZER_OLD || bv == SAIG_ZER_OLD || (av == SAIG_ONE_OLD && bv == SAIG_ONE_OLD) {
			Saig_ManSetAndDriveImplications_rec(graph, fanout, frame, last, info)
		}
	}
}

func Saig_ManExplorePaths_rec(graph *aig.Graph, object aig.Edge, frame, last int32, info *vec.Vec_Ptr_t) {
	value := Saig_ManSimInfo2Get(info, &object, frame)
	if Saig_ManSimInfo2IsOld(value) != 0 {
		return
	}
	Saig_ManSetAndDriveImplications_rec(graph, object, frame, last, info)
	if aig.Aig_ObjIsConst1(graph, object) {
		panic("abstraction: unexpected constant in justification")
	}
	if saig.Saig_ObjIsLo(graph, object) && frame == 0 {
		return
	}
	if saig.Saig_ObjIsPi(graph, object) {
		index := aig.Aig_ObjCioId(graph, object)
		for f := last; f >= 0; f-- {
			if f != frame {
				Saig_ManSetAndDriveImplications_rec(graph, graph.Inputs()[index], f, last, info)
			}
		}
		return
	}
	if saig.Saig_ObjIsLo(graph, object) {
		Saig_ManExplorePaths_rec(graph, saig.Saig_ObjLoToLi(graph, object), frame-1, last, info)
		return
	}
	if aig.Aig_ObjIsCo(graph, object) {
		Saig_ManExplorePaths_rec(graph, aig.Aig_ObjFanin0(graph, object), frame, last, info)
		return
	}
	if !aig.Aig_ObjIsNode(graph, object) {
		panic("abstraction: unexpected justification node")
	}
	Saig_ManExplorePaths_rec(graph, aig.Aig_ObjFanin0(graph, object), frame, last, info)
	// Keep the source's visited-value comparison and resulting traversal.
	if value != SAIG_ZER_OLD {
		Saig_ManExplorePaths_rec(graph, aig.Aig_ObjFanin1(graph, object), frame, last, info)
	}
}

func Saig_ManProcessCex(graph *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, info *vec.Vec_Ptr_t, verbose int32) *vec.Vec_Int_t {
	if Saig_ManSimDataInit2(graph, cex, info) != SAIG_ONE_NEW {
		panic("abstraction: counterexample output is not asserted")
	}
	for i := int32(0); i < saig.Saig_ManRegNum(graph); i++ {
		Saig_ManSetAndDriveImplications_rec(graph, saig.Saig_ManLo(graph, i), 0, cex.IFrame, info)
	}
	inputs := graph.Inputs()
	for frame := cex.IFrame; frame >= 0; frame-- {
		Saig_ManSetAndDriveImplications_rec(graph, graph.Const1(), frame, cex.IFrame, info)
		for i := int32(0); i < firstFlopPI; i++ {
			Saig_ManSetAndDriveImplications_rec(graph, inputs[i], frame, cex.IFrame, info)
		}
	}
	Saig_ManExplorePaths_rec(graph, graph.Outputs()[cex.IPo], cex.IFrame, cex.IFrame, info)
	result, omitted := vec.Vec_IntAlloc(0), vec.Vec_IntAlloc(0)
	for i := firstFlopPI; i < saig.Saig_ManPiNum(graph); i++ {
		found := false
		object := inputs[i]
		for frame := cex.IFrame; frame >= 0; frame-- {
			if Saig_ManSimInfo2IsOld(Saig_ManSimInfo2Get(info, &object, frame)) != 0 {
				found = true
				break
			}
		}
		if found {
			vec.Vec_IntPush(result, i)
		} else {
			vec.Vec_IntPush(omitted, i)
		}
	}
	if Saig_ManSimDataInit(graph, cex, info, omitted) != SAIG_ONE {
		panic("abstraction: selected counterexample inputs are insufficient")
	}
	return result
}

func Saig_ManExtendCounterExampleTest2(graph *aig.Graph, firstFlopPI int32, cex *util.Abc_Cex_t, verbose int32) *vec.Vec_Int_t {
	if saig.Saig_ManPiNum(graph) != cex.NPis {
		fmt.Printf("Saig_ManExtendCounterExampleTest2(): The PI count of AIG (%d) does not match that of cex (%d).\n", aig.Aig_ManCiNum(graph), cex.NPis)
		return nil
	}
	aig.Aig_ManFanoutStart(graph)
	defer aig.Aig_ManFanoutStop(graph)
	info := vec.Vec_PtrAllocSimInfo(len(graph.Nodes()), int((2*(cex.IFrame+1)+31)/32))
	started := time.Now()
	result := Saig_ManProcessCex(graph, firstFlopPI, cex, info, verbose)
	if verbose != 0 {
		fmt.Printf("Total new PIs = %3d. Non-removable PIs = %3d. Time = %9.2f sec\n", saig.Saig_ManPiNum(graph)-firstFlopPI, vec.Vec_IntSize(result), time.Since(started).Seconds())
	}
	return result
}
