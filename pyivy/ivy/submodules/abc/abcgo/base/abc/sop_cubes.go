package abc

import (
	"fmt"
	"strings"
)

// Abc_NtkComputeDelay estimates logic delay with ABC's normalized fanin model.
func Abc_NtkComputeDelay(n *Network) float32 {
	delays := [...]float64{1, 1, 2, 2.58, 3, 3.32, 3.58, 3.81, 4, 4.17, 4.32, 4.46, 4.58, 4.70, 4.81}
	maximum := int(Abc_NtkGetFaninMax(n))
	if maximum <= 1 || maximum >= len(delays) {
		panic("abc: unsupported fanin count for delay estimate")
	}
	for _, input := range n.cis {
		input.DTemp = 0
	}
	for _, node := range Abc_NtkDfs(n, 1) {
		node.DTemp = 0
		for _, edge := range node.fanins {
			node.DTemp = max(node.DTemp, edge.Object.DTemp)
		}
		node.DTemp += float64(float32(delays[len(node.fanins)] / delays[maximum]))
	}
	var result float64
	for _, output := range n.cos {
		if len(output.fanins) != 1 {
			panic("abc: output has no driver")
		}
		result = max(result, output.fanins[0].Object.DTemp)
	}
	return float32(result)
}

// Abc_NodeSopToCubes expands a local SOP into cube nodes and an OR/XOR root.
func Abc_NodeSopToCubes(old *Object, dst *Network, xor int32) {
	sop, ok := old.Data.(string)
	if !ok {
		panic("abc: node has no SOP")
	}
	if Abc_SopGetCubeNum(sop) < 2 {
		copy := Abc_NtkDupObj(dst, old, 0)
		for _, edge := range old.fanins {
			Abc_ObjAddFanin(copy, ntkCopiedEdge(edge))
		}
		return
	}
	root := dst.NewObject(ObjectNode)
	cubes := Abc_SopGetCubeNum(sop)
	var function string
	if xor != 0 {
		function = Abc_SopCreateXorSpecial(nil, cubes)
	} else {
		function = Abc_SopCreateOr(nil, cubes, nil)
	}
	for _, row := range strings.Split(strings.TrimSuffix(sop, "\n"), "\n") {
		if len(row) < len(old.fanins)+2 {
			panic("abc: malformed SOP cube")
		}
		var inputs []Edge
		var phases []int32
		for i, value := range row[:len(old.fanins)] {
			if value != '0' && value != '1' {
				continue
			}
			inputs = append(inputs, ntkCopiedEdge(old.fanins[i]))
			phases = append(phases, abcBool(value == '0'))
		}
		var cube *Object
		if len(inputs) == 0 {
			cube = Abc_NtkCreateNodeConst1(dst)
		} else {
			cube = dst.NewObject(ObjectNode)
			cube.Data = Abc_SopCreateAnd(nil, int32(len(inputs)), phases)
			for _, input := range inputs {
				Abc_ObjAddFanin(cube, input)
			}
		}
		Abc_ObjAddFanin(root, cube)
	}
	if Abc_SopIsComplement(sop) != 0 {
		function = Abc_SopComplement(function)
	}
	root.Data = function
	setCopy(old, root.Signal())
}

func Abc_NtkSopToCubes(n *Network, xor int32) *Network {
	if Abc_NtkIsSopLogic(n) == 0 {
		panic("abc: cube conversion requires SOP logic")
	}
	dst := Abc_NtkStartFrom(n, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	for _, node := range Abc_NtkDfs(n, 0) {
		Abc_NodeSopToCubes(node, dst, xor)
	}
	Abc_NtkFinalize(n, dst)
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkSopToCubes: The network check has failed.")
		Abc_NtkDelete(dst)
		return nil
	}
	return dst
}

// Abc_NtkSplitSop splits large covers into smaller covers feeding an OR node.
func Abc_NtkSplitSop(n *Network, cubeLimit, verbose int32) *Network {
	if n == nil {
		return nil
	}
	if Abc_NtkIsSopLogic(n) == 0 || cubeLimit <= 0 {
		panic("abc: SOP splitting requires SOP logic and a positive limit")
	}
	dst := Abc_NtkStartFrom(n, ABC_NTK_LOGIC, ABC_FUNC_SOP)
	for _, node := range Abc_NtkDfs(n, 0) {
		sop, ok := node.Data.(string)
		if !ok {
			panic("abc: node has no SOP")
		}
		root := Abc_NtkDupObj(dst, node, 0)
		count := int(Abc_SopGetCubeNum(sop))
		if count <= int(cubeLimit) {
			for _, edge := range node.fanins {
				Abc_ObjAddFanin(root, ntkCopiedEdge(edge))
			}
			continue
		}
		phase := Abc_SopIsComplement(sop) != 0
		if phase {
			sop = Abc_SopComplement(sop)
		}
		rows := strings.SplitAfter(sop, "\n")
		if rows[len(rows)-1] == "" {
			rows = rows[:len(rows)-1]
		}
		splits := (len(rows) + int(cubeLimit) - 1) / int(cubeLimit)
		root.Data = Abc_SopCreateOr(nil, int32(splits), nil)
		if phase {
			root.Data = Abc_SopComplement(root.Data.(string))
		}
		for start := 0; start < len(rows); start += int(cubeLimit) {
			end := min(start+int(cubeLimit), len(rows))
			part := dst.NewObject(ObjectNode)
			part.Data = strings.Join(rows[start:end], "")
			for _, edge := range node.fanins {
				Abc_ObjAddFanin(part, ntkCopiedEdge(edge))
			}
			Abc_ObjAddFanin(root, part)
		}
		setCopy(node, root.Signal())
	}
	Abc_NtkFinalize(n, dst)
	if Abc_NtkCheck(dst) == 0 {
		fmt.Println("Abc_NtkDup(): Network check has failed.")
	}
	n.Copy = dst
	return dst
}

func Abc_NtkIsTopo(n *Network) int32 {
	Abc_NtkIncrementTravId(n)
	for _, input := range n.cis {
		Abc_NodeSetTravIdCurrent(input)
	}
	var count int
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		for _, edge := range node.fanins {
			if Abc_NodeIsTravIdCurrent(edge.Object) == 0 {
				if count == 0 {
					fmt.Printf("Node %d is out of topo order.\n", node.ID)
				}
				count++
				break
			}
		}
		Abc_NodeSetTravIdCurrent(node)
	}
	if count != 0 {
		fmt.Printf("Topological order does not hold for %d internal nodes.\n", count)
	}
	return abcBool(count == 0)
}

func Abc_NtkTransferPhases(dst, src *Network) {
	if src.Phases == nil || len(src.Phases) != len(src.objects) || dst.Phases != nil {
		panic("abc: invalid phase transfer")
	}
	dst.Phases = make([]int32, len(dst.objects))
	for _, object := range src.Objects() {
		copy := getCopy(object).Object
		if copy != nil && dst.contains(copy) {
			dst.Phases[copy.ID] = src.Phases[object.ID]
		}
	}
}
