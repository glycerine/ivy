package abc

import "fmt"

// Abc_NtkToLogic converts a flat netlist or a strashed AIG to logic.
func Abc_NtkToLogic(src *Network) *Network {
	if Abc_NtkIsStrash(src) != 0 {
		return Abc_NtkAigToLogicSop(src)
	}
	if Abc_NtkIsNetlist(src) == 0 || Abc_NtkWhiteboxNum(src) != 0 || Abc_NtkBlackboxNum(src) != 0 {
		panic("abc: logic conversion requires a flat netlist")
	}
	dst := Abc_NtkStartFrom(src, ABC_NTK_LOGIC, Abc_NtkFunc_t(src.Function))
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		copy := Abc_NtkDupObj(dst, node, 0)
		Abc_ObjAssignName(copy, Abc_ObjName(node.fanouts[0]), "")
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		for _, edge := range node.fanins {
			net := edge.Object
			driver := net.fanins[0]
			copy := ntkCopiedEdge(driver)
			copy.Complement = copy.Complement != edge.Complement
			Abc_ObjAddFanin(node.Copy, copy)
		}
	}
	Abc_NtkFinalize(src, dst)
	Abc_NtkLogicMakeSimpleCos(dst, 0)
	if src.Exdc != nil {
		dst.Exdc = Abc_NtkToLogic(src.Exdc)
	}
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkToLogic(): Network check has failed.")
	}
	return dst
}

// Abc_NtkToNetlist converts logic or strashed AIG networks to netlists.
func Abc_NtkToNetlist(src *Network) *Network {
	if Abc_NtkIsStrash(src) != 0 {
		tmp := Abc_NtkAigToLogicSop(src)
		dst := Abc_NtkLogicToNetlist(tmp)
		Abc_NtkDelete(tmp)
		return dst
	}
	return Abc_NtkLogicToNetlist(src)
}

func Abc_NtkLogicToNetlist(src *Network) *Network {
	if Abc_NtkIsLogic(src) == 0 {
		panic("abc: netlist conversion requires a logic network")
	}
	Abc_NtkCleanup(src, 0)
	Abc_NtkCheckUniqueCiNames(src)
	Abc_NtkCheckUniqueCoNames(src)
	Abc_NtkCheckUniqueCioNames(src)
	if Abc_NtkLogicHasSimpleCos(src) == 0 {
		Abc_NtkLogicMakeSimpleCos(src, 0)
	}
	dst := Abc_NtkStartFrom(src, ABC_NTK_NETLIST, Abc_NtkFunc_t(src.Function))
	nets := make(map[*Object]*Object)
	for _, input := range src.cis {
		net := Abc_NtkFindOrCreateNet(dst, Abc_ObjName(input))
		Abc_ObjAddFanin(net, getCopy(input))
		nets[input] = net
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		Abc_NtkDupObj(dst, node, 0)
	}
	for _, output := range src.cos {
		driver := output.fanins[0].Object
		net := nets[driver]
		if net == nil {
			if driver.Kind != ObjectNode {
				panic("abc: invalid netlist output driver")
			}
			net = Abc_NtkFindOrCreateNet(dst, Abc_ObjName(output))
			Abc_ObjAddFanin(net, driver.Copy)
			nets[driver] = net
		} else if Abc_ObjName(net) != Abc_ObjName(output) {
			panic("abc: output driver has a different net name")
		}
		Abc_ObjAddFanin(output.Copy, net)
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		if nets[node] == nil {
			net := Abc_NtkFindOrCreateNet(dst, "new_"+Abc_ObjName(node)+"_")
			Abc_ObjAddFanin(net, node.Copy)
			nets[node] = net
		}
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		for _, edge := range node.fanins {
			Abc_ObjAddFanin(node.Copy, Edge{Object: nets[edge.Object], Complement: edge.Complement})
		}
	}
	if src.Exdc != nil {
		dst.Exdc = Abc_NtkToNetlist(src.Exdc)
	}
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkLogicToNetlist(): Network check has failed.")
	}
	return dst
}

// Abc_NtkAigToLogicSop converts AIG ANDs and choice chains to SOP nodes.
func Abc_NtkAigToLogicSop(src *Network) *Network {
	if Abc_NtkIsStrash(src) == 0 {
		panic("abc: SOP conversion requires a strashed network")
	}
	dst := Abc_NtkStartFrom(src, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	if len(src.constant.fanouts) != 0 {
		setCopy(src.constant, Abc_NtkCreateNodeConst1(dst).Signal())
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		copy := Abc_NtkDupObj(dst, node, 0)
		copy.Data = Abc_SopCreateAnd2(nil, abcBool(node.fanins[0].Complement), abcBool(node.fanins[1].Complement))
	}
	choices := make(map[*Object]*Object)
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		if Abc_AigNodeIsChoice(node) == 0 {
			continue
		}
		choice := dst.NewObject(ObjectNode)
		var phases []int32
		for alt := node; alt != nil; alt = alt.Choice {
			phases = append(phases, abcBool(node.Phase != alt.Phase))
			Abc_ObjAddFanin(choice, alt.Copy)
		}
		choice.Data = Abc_SopCreateOrMultiCube(nil, int32(len(phases)), phases)
		choices[node] = choice
	}
	resolve := func(o *Object) *Object {
		if c := choices[o]; c != nil {
			return c
		}
		return getCopy(o).Object
	}
	for _, node := range src.ObjectsOfKind(ObjectNode) {
		for _, edge := range node.fanins {
			Abc_ObjAddFanin(node.Copy, resolve(edge.Object))
		}
	}
	for _, output := range src.cos {
		edge := output.fanins[0]
		Abc_ObjAddFanin(output.Copy, Edge{Object: resolve(edge.Object), Complement: edge.Complement})
	}
	if len(choices) != 0 {
		Abc_NtkLogicMakeSimpleCos2(dst, 0)
	} else {
		Abc_NtkLogicMakeSimpleCos(dst, 0)
	}
	if src.Exdc != nil {
		if Abc_NtkIsStrash(src.Exdc) != 0 {
			dst.Exdc = Abc_NtkAigToLogicSop(src.Exdc)
		} else {
			dst.Exdc = Abc_NtkDup(src.Exdc)
		}
	}
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkAigToLogicSop(): Network check has failed.")
	}
	return dst
}

// Abc_NtkToNetlistBench converts an AIG to a netlist without choice nodes.
func Abc_NtkToNetlistBench(src *Network) *Network {
	tmp := Abc_NtkAigToLogicSopBench(src)
	dst := Abc_NtkLogicToNetlist(tmp)
	Abc_NtkDelete(tmp)
	return dst
}

func Abc_NtkAigToLogicSopBench(src *Network) *Network {
	return aigToLogicBench(src, false)
}

func Abc_NtkAigToLogicSopNand(src *Network) *Network {
	return aigToLogicBench(src, true)
}

// NAND mode reverses each internal node's polarity. Interpret edges with the
// corresponding inversion rather than temporarily changing the source graph.
func aigToLogicBench(src *Network, nand bool) *Network {
	if Abc_NtkIsStrash(src) == 0 {
		panic("abc: bench conversion requires a strashed network")
	}
	if Abc_NtkGetChoiceNum(src) != 0 {
		fmt.Println("Warning: Choice nodes are skipped.")
	}
	dst := Abc_NtkStartFrom(src, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	nodes := Abc_NtkDfs(src, 0)
	if len(src.constant.fanouts) != 0 {
		setCopy(src.constant, Abc_NtkCreateNodeConst1(dst).Signal())
	}
	for _, node := range nodes {
		copy := Abc_NtkDupObj(dst, node, 0)
		if nand {
			copy.Data = Abc_SopCreateNand(nil, 2)
		} else {
			copy.Data = Abc_SopCreateAnd(nil, 2, nil)
		}
	}
	inverses := make(map[*Object]*Object)
	resolve := func(edge Edge) *Object {
		copy := getCopy(edge.Object).Object
		if nand && edge.Object.Kind == ObjectNode {
			edge.Complement = !edge.Complement
		}
		if !edge.Complement {
			return copy
		}
		if inverses[edge.Object] == nil {
			inverses[edge.Object] = Abc_NtkCreateNodeInv(dst, copy)
		}
		return inverses[edge.Object]
	}
	for _, node := range nodes {
		for _, edge := range node.fanins {
			Abc_ObjAddFanin(node.Copy, resolve(edge))
		}
	}
	for _, output := range src.cos {
		Abc_ObjAddFanin(output.Copy, resolve(output.fanins[0]))
	}
	Abc_NtkLogicMakeSimpleCos(dst, 0)
	if src.Exdc != nil {
		fmt.Println("Warning: The EXDc network is skipped.")
	}
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkAigToLogicSopBench(): Network check has failed.")
	}
	return dst
}
