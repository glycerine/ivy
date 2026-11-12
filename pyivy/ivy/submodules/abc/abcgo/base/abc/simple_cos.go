package abc

import "fmt"

func Abc_NtkLogicMakeSimpleCosTest(n *Network, _ int32) {
	types := make([]int, len(n.objects))
	var counts [4]int
	var constants, inputs [2]int
	for _, output := range n.cos {
		edge := output.fanins[0]
		phase := int(abcBool(edge.Complement))
		if edge.Object.ID == 0 {
			constants[phase]++
		}
		if edge.Object.IsCI() {
			inputs[phase]++
		}
		types[edge.Object.ID] |= 1 << phase
	}
	for _, kind := range types {
		counts[kind]++
	}
	for i, count := range counts {
		fmt.Printf("%d = %d     ", i, count)
	}
	for i, count := range constants {
		fmt.Printf("c%d = %d     ", i, count)
	}
	for i, count := range inputs {
		fmt.Printf("i%d = %d    ", i, count)
	}
	fmt.Println()
}

// Abc_NtkLogicHasSimpleCos checks edge polarity and output-name sharing.
func Abc_NtkLogicHasSimpleCos(n *Network) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: simple outputs require a logic network")
	}
	names := make(map[*Object]string)
	for _, output := range n.cos {
		edge := output.fanins[0]
		name := Abc_ObjName(output)
		if edge.Complement || (edge.Object.IsCI() && Abc_ObjName(edge.Object) != name) {
			return 0
		}
		if previous, found := names[edge.Object]; found && previous != name {
			return 0
		}
		names[edge.Object] = name
	}
	return 1
}

// Abc_NtkFixCoDriverProblem inserts a buffer or duplicates the driver.
func Abc_NtkFixCoDriverProblem(driver, output *Object, duplicate int32) {
	n := driver.network
	edge := output.fanins[0]
	if edge.Object != driver {
		panic("abc: output driver mismatch")
	}
	var replacement *Object
	if duplicate != 0 && !driver.IsCI() {
		replacement = Abc_NtkDupObj(n, driver, 0)
		for _, fanin := range driver.fanins {
			Abc_ObjAddFanin(replacement, fanin)
		}
		if edge.Complement {
			Abc_NodeComplement(replacement)
		}
	} else if edge.Complement {
		replacement = Abc_NtkCreateNodeInv(n, driver)
	} else {
		replacement = Abc_NtkCreateNodeBuf(n, driver)
	}
	if err := n.PatchFanin(output, 0, replacement.Signal()); err != nil {
		panic(err)
	}
	if len(driver.fanouts) == 0 && driver != n.constant {
		Abc_NtkDeleteObj(driver)
	}
}

// Abc_NtkLogicMakeSimpleCos2 repairs non-simple outputs in traversal order.
func Abc_NtkLogicMakeSimpleCos2(n *Network, duplicate int32) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: simple outputs require a logic network")
	}
	names := make(map[*Object]string)
	var count int32
	for _, output := range n.cos {
		edge := output.fanins[0]
		name := Abc_ObjName(output)
		previous, found := names[edge.Object]
		if edge.Complement || (edge.Object.IsCI() && Abc_ObjName(edge.Object) != name) || (found && previous != name) {
			Abc_NtkFixCoDriverProblem(edge.Object, output, duplicate)
			count++
			continue
		}
		names[edge.Object] = name
	}
	if Abc_NtkLogicHasSimpleCos(n) == 0 {
		panic("abc: failed to simplify outputs")
	}
	return count
}

// Abc_NtkLogicMakeSimpleCos removes complemented and shared output drivers,
// changing an existing driver's polarity when doing so needs no new node.
func Abc_NtkLogicMakeSimpleCos(n *Network, duplicate int32) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: simple outputs require a logic network")
	}
	maxLevel := int(Abc_NtkLevel(n))
	for _, output := range n.cos {
		edge := output.fanins[0]
		driver := edge.Object
		if Abc_NodeIsConst(driver) == 0 {
			continue
		}
		var replacement *Object
		if edge.Complement == (Abc_NodeIsConst0(driver) != 0) {
			replacement = Abc_NtkCreateNodeConst1(n)
		} else {
			replacement = Abc_NtkCreateNodeConst0(n)
		}
		if err := n.PatchFanin(output, 0, replacement.Signal()); err != nil {
			panic(err)
		}
		if len(driver.fanouts) == 0 {
			Abc_NtkDeleteObj(driver)
		}
	}
	var drivers []*Object
	seen := make(map[*Object]bool)
	for _, output := range n.cos {
		edge := output.fanins[0]
		if edge.Complement && !seen[edge.Object] {
			seen[edge.Object] = true
			drivers = append(drivers, edge.Object)
		}
	}
	var total int32
	for _, driver := range drivers {
		var direct, other bool
		for _, fanout := range driver.fanouts {
			if !fanout.IsCO() {
				other = true
			} else if !fanout.fanins[0].Complement {
				direct = true
			}
		}
		if driver.IsCI() || direct || (other && Abc_NtkHasMapping(n) != 0) {
			var replacement *Object
			if duplicate != 0 && driver.Kind == ObjectNode && driver.Level == maxLevel {
				replacement = Abc_NtkDupObj(n, driver, 0)
				for _, fanin := range driver.fanins {
					Abc_ObjAddFanin(replacement, fanin)
				}
				Abc_NodeComplement(replacement)
			} else {
				replacement = Abc_NtkCreateNodeInv(n, driver)
			}
			for _, output := range driver.Fanouts() {
				if output.IsCO() && output.fanins[0].Complement {
					if err := n.PatchFanin(output, 0, replacement.Signal()); err != nil {
						panic(err)
					}
				}
			}
			total++
		} else {
			if driver.Kind != ObjectNode {
				panic("abc: complemented driver is not a node")
			}
			Abc_NodeComplement(driver)
			for _, fanout := range driver.Fanouts() {
				if fanout.IsCO() {
					if err := n.PatchFanin(fanout, 0, driver.Signal()); err != nil {
						panic(err)
					}
				} else if fanout.Kind == ObjectNode {
					Abc_NodeComplementInput(fanout, driver)
				} else {
					panic("abc: invalid complemented driver fanout")
				}
			}
		}
	}
	names := make(map[*Object]string)
	var repair []*Object
	for _, output := range n.cos {
		driver := output.fanins[0].Object
		name := Abc_ObjName(output)
		if driver.IsCI() && Abc_ObjName(driver) != name {
			repair = append(repair, output)
			continue
		}
		if previous, found := names[driver]; found && previous != name {
			repair = append(repair, output)
		} else {
			names[driver] = name
		}
	}
	for _, output := range repair {
		driver := output.fanins[0].Object
		var replacement *Object
		if duplicate != 0 && driver.Kind == ObjectNode && (Abc_NtkHasMapping(n) != 0 || driver.Level == maxLevel) {
			replacement = Abc_NtkDupObj(n, driver, 0)
			for _, fanin := range driver.fanins {
				Abc_ObjAddFanin(replacement, fanin)
			}
		} else {
			replacement = Abc_NtkCreateNodeBuf(n, driver)
			Abc_ObjAssignName(replacement, Abc_ObjName(driver), "_buf")
		}
		if err := n.PatchFanin(output, 0, replacement.Signal()); err != nil {
			panic(err)
		}
		total++
	}
	return total
}
