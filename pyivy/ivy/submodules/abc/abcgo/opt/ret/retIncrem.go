package ret

import (
	"fmt"
	"github.com/glycerine/abcgo/base/abc"
	"reflect"
)

type latchTerminals struct{ in, out *abc.Object }
type LatchBindings struct {
	old map[*abc.Object]latchTerminals
}

func Abc_NtkRetimePrepareLatches(n *abc.Network) *LatchBindings {
	bindings := &LatchBindings{old: make(map[*abc.Object]latchTerminals)}
	for _, latch := range retLatches(n) {
		in, out := retFan0(latch), latch.Fanouts()[0]
		bindings.old[latch] = latchTerminals{in, out}
		driver := in.Fanins()[0]
		retTransfer(in, driver.Object)
		if driver.Complement {
			for i, e := range latch.Fanins() {
				retPatch(latch, i, e.Not())
			}
		}
		if err := n.DeleteFanin(in, 0); err != nil {
			panic(err)
		}
		retTransfer(out, latch)
		if err := n.DeleteFanin(out, 0); err != nil {
			panic(err)
		}
	}
	return bindings
}
func Abc_NtkRetimeFinalizeLatches(n *abc.Network, bindings *LatchBindings, idStart, useOldNames int32) int32 {
	for _, latch := range retLatches(n) {
		term, found := bindings.old[latch]
		if !found {
			term = latchTerminals{abc.Abc_NtkCreateBi(n), abc.Abc_NtkCreateBo(n)}
			base := latch.Name()
			if useOldNames == 0 {
				base = retFan0(latch).Name()
			}
			term.in.SetName(base + "_in")
			term.out.SetName(base + "_out")
		}
		driver := latch.Fanins()[0]
		retAdd(term.in, driver)
		retPatch(latch, 0, term.in.Signal())
		retTransfer(latch, term.out)
		retAdd(term.out, latch.Signal())
	}
	for _, term := range bindings.old {
		for _, o := range []*abc.Object{term.in, term.out} {
			if o.Network() == n && len(o.Fanins()) == 0 && len(o.Fanouts()) == 0 {
				retDelete(o)
			}
		}
	}
	abc.Abc_NtkOrderCisCos(n)
	return 1
}
func Abc_NtkRetimeIncremental(n *abc.Network, delay, forward, minDelay, oneStep, useOldNames, verbose int32) int32 {
	if len(retNodes(n)) == 0 {
		return 0
	}
	before := int32(len(retLatches(n)))
	idStart := abc.Abc_NtkObjNumMax(n)
	abc.Abc_NtkOrderCisCos(n)
	var copy *abc.Network
	iterations := int32(-1)
	if minDelay != 0 {
		iterations = 2 * abc.Abc_NtkLevel(n)
		if oneStep != 0 {
			iterations = 1
		}
		copy = abc.Abc_NtkDup(n)
		Abc_NtkRetimePrepareLatches(copy)
	}
	bindings := Abc_NtkRetimePrepareLatches(n)
	Abc_NtkRetimeShareLatches(n, 0)
	if minDelay != 0 {
		Abc_NtkRetimeMinDelay(n, copy, delay, iterations, forward, verbose)
	} else {
		Abc_NtkRetimeOneWay(n, forward, verbose)
	}
	Abc_NtkRetimeShareLatches(n, 0)
	if Abc_NtkRetimeFinalizeLatches(n, bindings, idStart, useOldNames) == 0 {
		return 0
	}
	return before - int32(len(retLatches(n)))
}
func Abc_NtkRetimeOneWay(n *abc.Network, forward, verbose int32) int32 {
	var cone *abc.Network
	var values []int32
	if forward != 0 {
		Abc_NtkRetimeTranferToCopy(n)
	} else {
		values = Abc_NtkRetimeCollectLatchValues(n)
		cone = Abc_NtkRetimeBackwardInitialStart(n)
	}
	moves := 0
	for moves < 10000 {
		changed := false
		for _, o := range n.Objects() {
			if o.Kind == abc.ObjectNode && Abc_NtkRetimeNodeIsEnabled(o, forward) != 0 {
				Abc_NtkRetimeNode(o, forward, 1)
				changed = true
				moves++
				if moves == 10000 {
					fmt.Println("Stopped after 10000 latch moves.")
					break
				}
			}
		}
		if !changed {
			break
		}
	}
	if forward != 0 {
		Abc_NtkRetimeTranferFromCopy(n)
	} else {
		Abc_NtkRetimeBackwardInitialFinish(n, cone, values, verbose)
	}
	return 0
}
func Abc_NtkRetimeNodeIsEnabled(o *abc.Object, forward int32) int32 {
	if o.Kind != abc.ObjectNode {
		panic("ret: expected logic node")
	}
	if forward != 0 {
		for _, e := range o.Fanins() {
			if e.Object.Kind != abc.ObjectLatch {
				return 0
			}
		}
	} else {
		for _, next := range o.Fanouts() {
			if next.Kind != abc.ObjectLatch {
				return 0
			}
		}
	}
	return 1
}
func Abc_NtkRetimeNode(o *abc.Object, forward, initial int32) {
	n := o.Network()
	if forward != 0 {
		if initial != 0 {
			o.Temp = int(Abc_ObjSopSimulate(o))
		}
		for i, e := range o.Fanins() {
			latch := e.Object
			if latch.Kind != abc.ObjectLatch {
				panic("ret: enabled node has non-latch fanin")
			}
			driver := latch.Fanins()[0]
			if e.Complement {
				driver = driver.Not()
			}
			retPatch(o, i, driver)
			if len(latch.Fanouts()) == 0 {
				retDelete(latch)
			}
		}
		latch := abc.Abc_NtkCreateLatch(n)
		retTransfer(o, latch)
		retAdd(latch, o.Signal())
		if initial != 0 {
			latch.Temp = o.Temp
		}
	} else {
		fanouts := o.Fanouts()
		if len(fanouts) == 0 {
			panic("ret: backward node has no latch fanout")
		}
		var cone *abc.Network
		if initial != 0 {
			if fanouts[0].Copy == nil {
				panic("ret: missing backward initial-state mapping")
			}
			cone = fanouts[0].Copy.Network()
			o.Copy = retSopNode(cone, o)
			for _, next := range fanouts {
				retAdd(next.Copy, o.Copy.Signal())
			}
		}
		for _, latch := range fanouts {
			if latch.Kind != abc.ObjectLatch {
				panic("ret: enabled node has non-latch fanout")
			}
			retTransfer(latch, o)
			retDelete(latch)
		}
		for i, e := range o.Fanins() {
			latch := abc.Abc_NtkCreateLatch(n)
			retPatch(o, i, latch.Signal())
			retAdd(latch, e)
			if initial != 0 {
				latch.Copy = abc.Abc_NtkCreateNodeBuf(cone, nil)
				latch.Copy.SetName(e.Object.Name() + "_buf")
				retAdd(o.Copy, latch.Copy.Signal())
			}
		}
	}
}
func Abc_NtkRetimeCheckCompatibleLatchFanouts(o *abc.Object) int32 {
	var first *abc.Object
	count := int32(0)
	for _, next := range o.Fanouts() {
		if next.Kind != abc.ObjectLatch {
			continue
		}
		if first == nil {
			first = next
		}
		if reflect.DeepEqual(next.Data, first.Data) && next.Fanins()[0].Complement == first.Fanins()[0].Complement {
			count++
		}
	}
	return count
}
func Abc_NtkRetimeShareLatches(n *abc.Network, initial int32) {
	for _, o := range n.Objects() {
		if Abc_NtkRetimeCheckCompatibleLatchFanouts(o) <= 1 {
			continue
		}
		var top *abc.Object
		for _, latch := range o.Fanouts() {
			if latch.Kind != abc.ObjectLatch {
				continue
			}
			if top == nil {
				top = latch
				continue
			}
			if !reflect.DeepEqual(latch.Data, top.Data) || latch.Fanins()[0].Complement != top.Fanins()[0].Complement {
				continue
			}
			if initial != 0 {
				retAdd(latch.Copy, top.Copy.Signal())
			}
			retTransfer(latch, top)
			retDelete(latch)
		}
	}
}
