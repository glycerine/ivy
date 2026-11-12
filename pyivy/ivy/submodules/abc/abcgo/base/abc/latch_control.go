package abc

import "github.com/glycerine/abcgo/misc/vec"

type latchControlCounts struct{ notMux, other, opposite, enabled int }

// abcClassifyLatchControls removes supported feedback muxes and records their
// shared enable signals for the retiming pass.
func abcClassifyLatchControls(n *Network) (*vec.Vec_Ptr_t, *vec.Vec_Int_t, latchControlCounts) {
	controls := vec.Vec_PtrAlloc(0)
	classes := vec.Vec_IntAlloc(0)
	var counts latchControlCounts
	for _, latch := range n.ObjectsOfKind(ObjectLatch) {
		output := Abc_ObjFanout0(latch)
		input := Abc_ObjFanin0(latch)
		driver := Abc_ObjFanin0(input)
		if Abc_ObjFaninNum(driver) != 3 {
			vec.Vec_IntPush(classes, -1)
			counts.notMux++
			continue
		}
		if Abc_ObjFanin(driver, 1) != output && Abc_ObjFanin(driver, 2) != output {
			vec.Vec_IntPush(classes, -1)
			counts.other++
			continue
		}
		if Abc_ObjFanin(driver, 1) == output {
			vec.Vec_IntPush(classes, -1)
			counts.opposite++
			continue
		}
		counts.enabled++
		control := Abc_ObjFanin0(driver)
		vec.Vec_PtrPushUnique(controls, control)
		vec.Vec_IntPush(classes, int32(vec.Vec_PtrFind(controls, control)))
		Abc_ObjPatchFanin(input, driver, Abc_ObjFanin(driver, 1))
	}
	return controls, classes, counts
}

func abcExposeLatchControls(n *Network, controls *vec.Vec_Ptr_t, verbose int32) {
	for i, item := range vec.Vec_PtrArray(controls) {
		signal, ok := item.(*Object)
		if !ok || signal.network != n {
			panic("abc: invalid latch control")
		}
		output := Abc_NtkCreatePo(n)
		Abc_ObjAddFanin(output, signal)
		Abc_ObjAssignName(output, Abc_ObjName(output), "")
		vec.Vec_PtrWriteEntry(controls, i, output)
	}
	Abc_NtkOrderCisCos(n)
	Abc_NtkCleanup(n, verbose)
}
