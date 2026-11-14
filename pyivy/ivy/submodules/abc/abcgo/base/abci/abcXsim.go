package abci

import (
	"fmt"
	"os"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/base/abc"
)

// X-valued simulation uses Object.Temp as the Go representation of ABC's
// pCopy scratch slot, which otherwise stores an integer cast to a pointer.
func Abc_ObjSetXsim(obj *Abc_Obj_t, value int32) { abc.Abc_ObjSetTemp(obj, value) }
func Abc_ObjGetXsim(obj *Abc_Obj_t) int32        { return abc.Abc_ObjTemp(obj) }

func Abc_ObjGetXsimFanin0(obj *Abc_Obj_t) int32 {
	value := Abc_ObjGetXsim(abc.Abc_ObjFanin0(obj))
	if abc.Abc_ObjFaninC0(obj) != 0 {
		return Abc_XsimInv(value)
	}
	return value
}

func Abc_ObjGetXsimFanin1(obj *Abc_Obj_t) int32 {
	value := Abc_ObjGetXsim(abc.Abc_ObjFanin1(obj))
	if abc.Abc_ObjFaninC1(obj) != 0 {
		return Abc_XsimInv(value)
	}
	return value
}

func Abc_XsimPrint(file *os.File, value int32) {
	switch value {
	case XVS0:
		_, _ = fmt.Fprint(file, "0")
	case XVS1:
		_, _ = fmt.Fprint(file, "1")
	case XVSX:
		_, _ = fmt.Fprint(file, "x")
	default:
		panic("abci: invalid X simulation value")
	}
}

func abcXsimLatches(n *Abc_Ntk_t) []*Abc_Obj_t {
	boxes := n.ObjectsOfKind(abc.ObjectLatch)
	return boxes
}

func abcXsimAndNodes(n *Abc_Ntk_t) []*Abc_Obj_t {
	nodes := make([]*Abc_Obj_t, 0, len(n.ObjectsOfKind(abc.ObjectNode)))
	// ABC's Abc_AigForEachAnd walks the object vector in ID order and filters
	// the AIG AND nodes; preserve that order for deterministic simulation.
	for id := int32(0); id < abc.Abc_NtkObjNumMax(n); id++ {
		obj := abc.Abc_NtkObj(n, id)
		if obj != nil && abc.Abc_AigNodeIsAnd(obj) != 0 {
			nodes = append(nodes, obj)
		}
	}
	return nodes
}

func abcXsimLatchInit(latch *Abc_Obj_t) int32 {
	switch {
	case abc.Abc_LatchIsInit0(latch) != 0:
		return XVS0
	case abc.Abc_LatchIsInit1(latch) != 0:
		return XVS1
	case abc.Abc_LatchIsInitDc(latch) != 0:
		return XVSX
	default:
		panic("abci: latch has no supported initial value")
	}
}

func abcXsimInputValue(xInputs int32) int32 {
	if xInputs != 0 {
		return XVSX
	}
	return Abc_XsimRand2()
}

func Abc_NtkXValueSimulate(n *Abc_Ntk_t, frames, xInputs, xState, verbose int32) {
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkXValueSimulate requires a strashed network")
	}
	gia.Gia_ManRandom(1)
	Abc_ObjSetXsim(abc.Abc_AigConst1(n).Object, XVS1)
	for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
		Abc_ObjSetXsim(abc.Abc_NtkPi(n, i), abcXsimInputValue(xInputs))
	}
	for _, latch := range abcXsimLatches(n) {
		value := abcXsimLatchInit(latch)
		if xState != 0 {
			value = XVSX
		}
		Abc_ObjSetXsim(abc.Abc_ObjFanout0(latch), value)
	}

	fmt.Fprintln(os.Stdout, "Frame : Inputs : Latches : Outputs")
	andNodes := abcXsimAndNodes(n)
	for frame := int32(0); frame < frames; frame++ {
		for _, node := range andNodes {
			Abc_ObjSetXsim(node, Abc_XsimAnd(Abc_ObjGetXsimFanin0(node), Abc_ObjGetXsimFanin1(node)))
		}
		for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
			co := abc.Abc_NtkCo(n, i)
			Abc_ObjSetXsim(co, Abc_ObjGetXsimFanin0(co))
		}
		fmt.Fprintf(os.Stdout, "%2d : ", frame)
		for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
			Abc_XsimPrint(os.Stdout, Abc_ObjGetXsim(abc.Abc_NtkPi(n, i)))
		}
		fmt.Fprint(os.Stdout, " : ")
		for _, latch := range abcXsimLatches(n) {
			Abc_XsimPrint(os.Stdout, Abc_ObjGetXsim(abc.Abc_ObjFanout0(latch)))
		}
		fmt.Fprint(os.Stdout, " : ")
		for i := int32(0); i < abc.Abc_NtkPoNum(n); i++ {
			Abc_XsimPrint(os.Stdout, Abc_ObjGetXsim(abc.Abc_NtkPo(n, i)))
		}
		fmt.Fprintln(os.Stdout)

		for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
			Abc_ObjSetXsim(abc.Abc_NtkPi(n, i), abcXsimInputValue(xInputs))
		}
		for _, latch := range abcXsimLatches(n) {
			Abc_ObjSetXsim(abc.Abc_ObjFanout0(latch), Abc_ObjGetXsim(abc.Abc_ObjFanin0(latch)))
		}
	}
}

func Abc_NtkCycleInitState(n *Abc_Ntk_t, frames, useXVal, verbose int32) {
	if abc.Abc_NtkIsStrash(n) == 0 {
		panic("Abc_NtkCycleInitState requires a strashed network")
	}
	gia.Gia_ManRandom(1)
	Abc_ObjSetXsim(abc.Abc_AigConst1(n).Object, XVS1)
	for _, latch := range abcXsimLatches(n) {
		Abc_ObjSetXsim(abc.Abc_ObjFanout0(latch), abcXsimLatchInit(latch))
	}
	andNodes := abcXsimAndNodes(n)
	for frame := int32(0); frame < frames; frame++ {
		for i := int32(0); i < abc.Abc_NtkPiNum(n); i++ {
			Abc_ObjSetXsim(abc.Abc_NtkPi(n, i), abcXsimInputValue(useXVal))
		}
		for _, node := range andNodes {
			Abc_ObjSetXsim(node, Abc_XsimAnd(Abc_ObjGetXsimFanin0(node), Abc_ObjGetXsimFanin1(node)))
		}
		for i := int32(0); i < abc.Abc_NtkCoNum(n); i++ {
			co := abc.Abc_NtkCo(n, i)
			Abc_ObjSetXsim(co, Abc_ObjGetXsimFanin0(co))
		}
		for _, latch := range abcXsimLatches(n) {
			Abc_ObjSetXsim(abc.Abc_ObjFanout0(latch), Abc_ObjGetXsim(abc.Abc_ObjFanin0(latch)))
		}
	}
	for _, latch := range abcXsimLatches(n) {
		switch Abc_ObjGetXsim(abc.Abc_ObjFanout0(latch)) {
		case XVS0:
			abc.Abc_LatchSetInit0(latch)
		case XVS1:
			abc.Abc_LatchSetInit1(latch)
		case XVSX:
			abc.Abc_LatchSetInitDc(latch)
		default:
			panic("abci: invalid simulated latch state")
		}
	}
}
