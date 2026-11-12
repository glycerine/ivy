package abc

import (
	"fmt"

	"github.com/glycerine/abcgo/misc/vec"
)

// Abc_NtkAddBuffsOne builds or reuses a chain to the requested level.
func Abc_NtkAddBuffsOne(buffers *vec.Vec_Ptr_t, fanin *Object, level, maxLevel int32) *Object {
	if fanin == nil || level <= int32(fanin.Level) || maxLevel <= 0 || level > maxLevel {
		panic("abc: invalid buffer level")
	}
	index := fanin.ID*int(maxLevel) + int(level)
	if index >= vec.Vec_PtrSize(buffers) {
		panic("abc: buffer cache too small")
	}
	if cached := vec.Vec_PtrEntry(buffers, index); cached != nil {
		return cached.(*Object)
	}
	previous := fanin
	if level-1 > int32(fanin.Level) {
		previous = Abc_NtkAddBuffsOne(buffers, fanin, level-1, maxLevel)
	}
	buffer := Abc_NtkCreateNodeBuf(fanin.network, previous)
	vec.Vec_PtrWriteEntry(buffers, index, buffer)
	return buffer
}

// Abc_NtkAddBuffsInt duplicates a network and balances its logic levels.
func Abc_NtkAddBuffsInt(original *Network, reverse, improve, verbose int32) *Network {
	n := Abc_NtkDup(original)
	maxLevel := int(Abc_NtkLevel(n))
	for _, output := range n.cos {
		output.Level = maxLevel + 1
	}
	nodes := Abc_NtkDfs(n, 1)
	if reverse != 0 {
		for i := len(nodes) - 1; i >= 0; i-- {
			node := nodes[i]
			node.Level = 1 << 18
			for _, fanout := range node.fanouts {
				node.Level = min(node.Level, fanout.Level-1)
			}
			if node.Level <= 0 {
				panic("abc: invalid reverse buffer level")
			}
		}
		for _, input := range n.cis {
			input.Level = 0
		}
	}
	for iteration := int32(0); iteration < improve; iteration++ {
		var shifted, gain int
		for j := 0; j < len(nodes); j++ {
			i := j
			if reverse == 0 {
				i = len(nodes) - 1 - j
			}
			node := nodes[i]
			canMove := true
			if reverse != 0 {
				for _, edge := range node.fanins {
					if edge.Object.Level+1 >= node.Level {
						canMove = false
						break
					}
				}
			} else {
				for _, fanout := range node.fanouts {
					if fanout.Level <= node.Level+1 {
						canMove = false
						break
					}
				}
			}
			if !canMove {
				continue
			}
			change := 1
			if reverse != 0 {
				change = -1
			}
			for _, edge := range node.fanins {
				if reverse != 0 {
					change += int(Abc_NtkAddBuffsEval(node, edge.Object))
				} else if Abc_NtkAddBuffsEval2(node, edge.Object) == 0 {
					change--
				}
			}
			if change < 0 {
				continue
			}
			if reverse != 0 {
				node.Level--
			} else {
				node.Level++
			}
			shifted++
			gain += change
		}
		if verbose != 0 {
			direction := "up"
			if reverse != 0 {
				direction = "down"
			}
			fmt.Printf("Shifted %5d nodes %s with total gain %5d.\n", shifted, direction, gain)
		}
		if shifted == 0 {
			break
		}
	}
	objects := n.Objects()
	buffers := vec.Vec_PtrStart(len(n.objects) * (maxLevel + 1))
	for _, node := range objects {
		if node.Kind != ObjectNode && !node.IsCO() {
			continue
		}
		for i, edge := range node.Fanins() {
			level := node.Level - 1
			if level < edge.Object.Level {
				panic("abc: fanin level exceeds buffered node")
			}
			if level == edge.Object.Level {
				continue
			}
			buffer := Abc_NtkAddBuffsOne(buffers, edge.Object, int32(level), int32(maxLevel))
			if err := n.PatchFanin(node, i, Edge{Object: buffer, Complement: edge.Complement}); err != nil {
				panic(err)
			}
		}
	}
	for _, output := range n.cos {
		output.Level = 0
	}
	return n
}

func Abc_NtkAddBuffs(n *Network, direct, reverse, improve, verbose int32) *Network {
	if direct != 0 {
		return Abc_NtkAddBuffsInt(n, 0, improve, verbose)
	}
	if reverse != 0 {
		return Abc_NtkAddBuffsInt(n, 1, improve, verbose)
	}
	directNetwork := Abc_NtkAddBuffsInt(n, 0, improve, verbose)
	reverseNetwork := Abc_NtkAddBuffsInt(n, 1, improve, verbose)
	if len(directNetwork.ObjectsOfKind(ObjectNode)) < len(reverseNetwork.ObjectsOfKind(ObjectNode)) {
		Abc_NtkDelete(reverseNetwork)
		return directNetwork
	}
	Abc_NtkDelete(directNetwork)
	return reverseNetwork
}
