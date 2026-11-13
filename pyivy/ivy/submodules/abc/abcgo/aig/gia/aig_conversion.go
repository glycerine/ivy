package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
)

type Aig_Man_t = aig.Graph
type Aig_Obj_t = aig.Edge

func Gia_ObjChild0Copy(p *aig.Graph, obj aig.Edge) int32 {
	return int32(aig.Aig_ObjCopy(p, aig.Aig_ObjFanin0(p, obj))) ^ aig.Aig_ObjFaninC0(p, obj)
}
func Gia_ObjChild1Copy(p *aig.Graph, obj aig.Edge) int32 {
	return int32(aig.Aig_ObjCopy(p, aig.Aig_ObjFanin1(p, obj))) ^ aig.Aig_ObjFaninC1(p, obj)
}
func Gia_ObjChild0Copy2(copies []aig.Edge, obj *Gia_Obj_t, id int32) aig.Edge {
	return copies[Gia_ObjFaninId0(obj, id)] ^ aig.Edge(obj.fCompl0)
}
func Gia_ObjChild1Copy2(copies []aig.Edge, obj *Gia_Obj_t, id int32) aig.Edge {
	return copies[Gia_ObjFaninId1(obj, id)] ^ aig.Edge(obj.fCompl1)
}

const aigCopyInProgress = aig.InvalidEdge - 1

func giaImportAigNode(out *Gia_Man_t, p *aig.Graph, obj aig.Edge, choicesFirst bool) int32 {
	obj = aig.Aig_Regular(obj)
	if copied := aig.Aig_ObjCopy(p, obj); copied != aig.InvalidEdge {
		if copied == aigCopyInProgress {
			panic("gia: cyclic AIG fanin or choice graph")
		}
		return int32(copied)
	}
	aig.Aig_ObjSetCopy(p, obj, aigCopyInProgress)
	node := p.Node(obj)
	edge := func(input aig.Edge) int32 {
		return giaImportAigNode(out, p, aig.Aig_Regular(input), choicesFirst) ^ int32(input&1)
	}
	var result int32
	switch node.Kind {
	case aig.Constant:
		result = 1
	case aig.Input:
		panic("gia: seed AIG inputs before recursive import")
	case aig.Buffer:
		result = edge(node.Fanins[0])
	case aig.And, aig.ExclusiveOr:
		a, b := edge(node.Fanins[0]), edge(node.Fanins[1])
		next := aig.Aig_ObjEquiv(p, obj)
		if choicesFirst && next != aig.InvalidEdge {
			giaImportAigNode(out, p, aig.Aig_Regular(next), true)
		}
		if node.Kind == aig.ExclusiveOr {
			x, y := out.AndLiteral(a, b^1, true), out.AndLiteral(a^1, b, true)
			result = out.AndLiteral(x^1, y^1, true) ^ 1
		} else {
			result = out.AndLiteral(a, b, true)
		}
		aig.Aig_ObjSetCopy(p, obj, aig.Edge(result))
		if !choicesFirst && next != aig.InvalidEdge {
			giaImportAigNode(out, p, aig.Aig_Regular(next), false)
		}
		if next != aig.InvalidEdge {
			target := int32(aig.Aig_ObjCopy(p, aig.Aig_Regular(next))) >> 1
			if target == result>>1 {
				break
			}
			if choicesFirst {
				if len(out.pSibls) < int(out.nObjs) {
					out.pSibls = append(out.pSibls, make([]int32, int(out.nObjs)-len(out.pSibls))...)
				}
				out.pSibls[result>>1] = target
			} else {
				if len(out.pNexts) < int(out.nObjs) {
					out.pNexts = append(out.pNexts, make([]int32, int(out.nObjs)-len(out.pNexts))...)
				}
				out.pNexts[result>>1] = target
			}
		}
	default:
		panic("gia: unsupported AIG object kind")
	}
	aig.Aig_ObjSetCopy(p, obj, aig.Edge(result))
	return result
}
func Gia_ManFromAig_rec(out *Gia_Man_t, p *aig.Graph, obj aig.Edge) {
	giaImportAigNode(out, p, obj, false)
}
func Gia_ManFromAigChoices_rec(out *Gia_Man_t, p *aig.Graph, obj aig.Edge) {
	if obj != aig.InvalidEdge {
		giaImportAigNode(out, p, obj, true)
	}
}
func giaImportAig(p *aig.Graph, choicesFirst, allNodes, switchOutputs bool) *Gia_Man_t {
	out := Gia_ManStart(aig.Aig_ManObjNumMax(p) + 1)
	out.pName = p.Name
	out.pSpec = p.Spec
	out.nConstrs = aig.Aig_ManConstrNum(p)
	aig.Aig_ManCleanData(p)
	aig.Aig_ObjSetCopy(p, p.Const1(), 1)
	for _, input := range p.Inputs() {
		aig.Aig_ObjSetCopy(p, input, aig.Edge(Gia_ManAppendCi(out)))
	}
	nodes := p.Nodes()
	if allNodes || switchOutputs {
		for id, node := range nodes {
			object := aig.Edge(id) << 1
			if aig.Aig_ObjIsNone(p, object) || (node.Kind != aig.And && node.Kind != aig.ExclusiveOr && node.Kind != aig.Buffer) {
				continue
			}
			if allNodes || node.Refs == 0 {
				lit := giaImportAigNode(out, p, object, choicesFirst)
				if switchOutputs && node.Refs == 0 {
					Gia_ManAppendCo(out, lit)
				}
			}
		}
	}
	for _, output := range p.Outputs() {
		driver := nodes[output.NodeID()].Fanins[0]
		giaImportAigNode(out, p, aig.Aig_Regular(driver), choicesFirst)
	}
	for _, output := range p.Outputs() {
		driver := nodes[output.NodeID()].Fanins[0]
		lit := int32(aig.Aig_ObjCopy(p, aig.Aig_Regular(driver))) ^ int32(driver&1)
		id := Gia_ManAppendCo(out, lit)
		aig.Aig_ObjSetCopy(p, output, aig.Edge(id))
	}
	out.nRegs = aig.Aig_ManRegNum(p)
	if len(out.pNexts) > 0 {
		if len(out.pNexts) < int(out.nObjs) {
			out.pNexts = append(out.pNexts, make([]int32, int(out.nObjs)-len(out.pNexts))...)
		}
		Gia_ManDeriveReprs(out)
	}
	return out
}
func Gia_ManFromAig(p *aig.Graph) *Gia_Man_t        { return giaImportAig(p, false, false, false) }
func Gia_ManFromAigChoices(p *aig.Graph) *Gia_Man_t { return giaImportAig(p, true, false, false) }
func Gia_ManFromAigSimple(p *aig.Graph) *Gia_Man_t  { return giaImportAig(p, false, true, false) }
func Gia_ManFromAigSwitch(p *aig.Graph) *Gia_Man_t  { return giaImportAig(p, false, false, true) }

func giaExportAigNode(out *aig.Graph, copies []aig.Edge, p *Gia_Man_t, obj *Gia_Obj_t, choices, record bool) aig.Edge {
	id := Gia_ObjId(p, obj)
	if copied := copies[id]; copied != aig.InvalidEdge {
		if copied == aigCopyInProgress {
			panic("gia: cyclic GIA choices")
		}
		return copied
	}
	copies[id] = aigCopyInProgress
	edge := func(diff, compl uint32) aig.Edge {
		return giaExportAigNode(out, copies, p, p.pObjs[id-int32(diff)], choices, record) ^ aig.Edge(compl)
	}
	var result aig.Edge
	switch {
	case id == 0:
		result = out.Const0()
	case Gia_ObjIsCi(obj) != 0:
		result = out.NewInput()
	case Gia_ObjIsAnd(obj) != 0:
		a, b := edge(obj.iDiff0, obj.fCompl0), edge(obj.iDiff1, obj.fCompl1)
		if int(id) < len(p.pMuxes) && p.pMuxes[id] != 0 {
			control := p.pMuxes[id]
			c := giaExportAigNode(out, copies, p, p.pObjs[control>>1], choices, record) ^ aig.Edge(control&1)
			result = out.Mux(c, b, a)
		} else if Gia_ObjIsXor(obj) != 0 {
			result = out.NewXor(a, b)
		} else {
			result = out.AND(a, b)
		}
	default:
		panic("gia: invalid GIA conversion root")
	}
	copies[id] = result
	if record {
		obj.Value = uint32(result)
	}
	if choices && int(id) < len(p.pNexts) && p.pNexts[id] != 0 {
		next := giaExportAigNode(out, copies, p, p.pObjs[p.pNexts[id]], true, record)
		aig.Aig_ObjSetEquiv(out, aig.Aig_Regular(result), aig.Aig_Regular(next))
	}
	return result
}
func Gia_ManToAig_rec(out *aig.Graph, copies []aig.Edge, p *Gia_Man_t, obj *Gia_Obj_t) {
	giaExportAigNode(out, copies, p, obj, p.pNexts != nil, false)
}
func giaExportAig(p *Gia_Man_t, choices, allNodes, levels bool, skip int32) *aig.Graph {
	if skip <= 0 {
		panic("gia: invalid output stride")
	}
	if choices && (p.pNexts == nil || p.pReprs == nil) {
		panic("gia: choice export requires representatives and next links")
	}
	out := aig.Aig_ManStart(p.nObjs)
	out.Name = p.pName
	out.Spec = p.pSpec
	copies := make([]aig.Edge, p.nObjs)
	for i := range copies {
		copies[i] = aig.InvalidEdge
	}
	copies[0] = out.Const0()
	if allNodes {
		p.pObjs[0].Value = uint32(copies[0])
	}
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		obj := Gia_ManCi(p, i)
		copies[obj.id] = out.NewInput()
		if allNodes {
			obj.Value = uint32(copies[obj.id])
		}
		if levels && p.vLevels != nil {
			aig.Aig_ObjSetLevel(out, copies[obj.id], Vec_IntEntry(p.vLevels, int(obj.id)))
		}
	}
	if allNodes {
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				giaExportAigNode(out, copies, p, obj, choices, allNodes)
			}
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		driver := giaExportAigNode(out, copies, p, Gia_ObjFanin0(obj), choices, allNodes) ^ aig.Edge(obj.fCompl0)
		if i%skip != 0 {
			continue
		}
		copies[obj.id] = out.AddOutput(driver)
		if allNodes {
			obj.Value = uint32(copies[obj.id])
		}
	}
	aig.Aig_ManSetRegNum(out, p.nRegs)
	aig.Aig_ManSetConstrNum(out, p.nConstrs)
	return out
}
func Gia_ManToAig(p *Gia_Man_t, choices int32) *aig.Graph {
	return giaExportAig(p, choices != 0, false, true, 1)
}
func Gia_ManToAigSkip(p *Gia_Man_t, stride int32) *aig.Graph {
	if p.pNexts != nil || p.pReprs != nil || stride <= 0 || Gia_ManCoNum(p)%stride != 0 {
		panic("gia: invalid strided AIG export")
	}
	return giaExportAig(p, false, false, false, stride)
}
func Gia_ManToAigSimple(p *Gia_Man_t) *aig.Graph { return giaExportAig(p, false, true, false, 1) }

func Gia_ManCheckChoices_rec(p *Gia_Man_t, obj *Gia_Obj_t) {
	if obj == nil {
		return
	}
	stack := []int32{obj.id}
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		node := p.pObjs[id]
		if Gia_ObjIsAnd(node) == 0 || node.fPhase != 0 {
			continue
		}
		node.fPhase = 1
		stack = append(stack, giaObjectFanins(p, node)...)
		if int(id) < len(p.pSibls) && p.pSibls[id] != 0 {
			stack = append(stack, p.pSibls[id])
		}
	}
}
func Gia_ManCheckChoices(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		obj.fPhase = 0
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManCheckChoices_rec(p, Gia_ObjFanin0(Gia_ManCo(p, i)))
	}
	found := false
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && obj.fPhase == 0 {
			fmt.Printf("Object %d is dangling.\n", obj.id)
			found = true
		}
	}
	if !found {
		fmt.Println("There are no dangling objects.")
	}
	for _, obj := range p.pObjs {
		obj.fPhase = 0
	}
}
