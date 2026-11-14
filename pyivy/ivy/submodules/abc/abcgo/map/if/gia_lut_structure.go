package if_

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

// Gia_ManFromIfLogicNode implements the requested LUT structure for one cut.
// The packing vector starts with the number of mapped groups.
func Gia_ManFromIfLogicNode(p *If_Man_t, dst *Gia_Man_t, objectID int32, leaves, temp *Vec_Int_t, truth []word, structure string, cover, mapping, records, packing *Vec_Int_t, check75, check44 int32) int32 {
	n := int32(Vec_IntSize(leaves))
	if n < 0 || n > 16 || len(truth) < int(abcutil.Abc_TtWordNum(n)) {
		panic("if: invalid LUT truth table")
	}
	if check75 != 0 {
		structure = "54"
	}
	single := func() int32 { return Gia_ManFromIfLogicCreateLut(dst, truth, leaves, cover, mapping, records) }
	packSingle := func(lit int32) {
		if lit > 1 && gia.Gia_ObjIsCi(Gia_ManObj(dst, lit>>1)) == 0 {
			if packing == nil || Vec_IntSize(packing) == 0 {
				panic("if: missing LUT packing header")
			}
			Vec_IntPush(packing, 1)
			Vec_IntPush(packing, lit>>1)
			vec.Vec_IntAddToEntry(packing, 0, 1)
		}
	}
	if check44 != 0 {
		if n <= 4 {
			lit := single()
			packSingle(lit)
			return lit
		}
		return Gia_ManFromIfLogicCreateLutSpecial(dst, truth, leaves, temp, cover, mapping, records, packing)
	}
	if p.pPars.FLut6Filter != 0 && n == 6 {
		local := vec.Vec_IntAllocArrayCopy([]int32{0}, 1)
		defer Vec_IntFree(local)
		return Gia_ManFromIfLogicCreateLutSpecial(dst, truth, leaves, temp, cover, mapping, records, local)
	}
	if structure == "" {
		return single()
	}
	length := len(structure)
	if length != 2 && length != 3 {
		fmt.Printf("Wrong LUT struct (%s)\n", structure)
		return -1
	}
	var sizes [3]int32
	for i := range structure {
		if structure[i] < '3' || structure[i] > '6' {
			fmt.Printf("The LUT size (%d) should belong to {3,4,5,6}.\n", int(structure[i])-'0')
			return -1
		}
		sizes[i] = int32(structure[i] - '0')
	}
	leaf, root := sizes[0], sizes[length-1]
	var leaf2 int32
	if length == 3 {
		leaf2 = sizes[1]
	}
	capacity := leaf - 1 + root
	if leaf2 != 0 {
		capacity += leaf2 - 1
	}
	if n > capacity {
		fmt.Printf("The node size (%d) is too large for the LUT structure %s.\n", n, structure)
		return -1
	}
	if n <= max(leaf, leaf2, root) {
		lit := single()
		packSingle(lit)
		return lit
	}
	words := abcutil.Abc_TtWordNum(n)
	if abcutil.Abc_TtIsConst0(truth, words) != 0 || abcutil.Abc_TtIsConst1(truth, words) != 0 {
		vec.Vec_IntFillExtra(mapping, 1, 0)
		if Vec_IntEntry(mapping, 0) == 0 {
			vec.Vec_IntWriteEntry(mapping, 0, int32(Vec_IntSize(records)))
			Vec_IntPush(records, 0)
			Vec_IntPush(records, 0)
		}
		if abcutil.Abc_TtIsConst1(truth, words) != 0 {
			return 1
		}
		return 0
	}
	// Look for elementary truth tables without static shared scratch state.
	for variable := int32(0); variable < n; variable++ {
		equal, opposite := true, true
		for x := 0; x < 1<<uint(n); x++ {
			bit := truth[x>>6] >> uint(x&63) & 1
			value := uint64(x >> uint(variable) & 1)
			equal = equal && bit == value
			opposite = opposite && bit != value
		}
		if equal {
			return Vec_IntEntry(leaves, int(variable))
		}
		if opposite {
			return Vec_IntEntry(leaves, int(variable)) ^ 1
		}
	}
	var groups [3][32]byte
	var functions [3]word
	success := int32(0)
	if check75 != 0 {
		for _, pair := range [][2]int32{{4, 4}, {5, 4}, {4, 5}} {
			if n >= 8 && pair == [2]int32{4, 4} {
				continue
			}
			success = If_CluCheckExt(p, truth, n, pair[0], pair[1], groups[0][:], groups[1][:], &functions[0], &functions[1])
			if success != 0 {
				break
			}
		}
	} else if length == 2 {
		success = If_CluCheckExt(p, truth, n, leaf, root, groups[0][:], groups[1][:], &functions[0], &functions[1])
	} else {
		success = If_CluCheckExt3(p, truth, n, leaf, leaf2, root, groups[0][:], groups[1][:], groups[2][:], &functions[0], &functions[1], &functions[2])
	}
	if success == 0 {
		fmt.Printf("Node %d is not decomposable. Deriving LUT structures has failed.\n", objectID)
		return -1
	}
	if packing == nil || Vec_IntSize(packing) == 0 {
		panic("if: missing LUT packing header")
	}
	build := func(group []byte, function word, first, second int32) int32 {
		vec.Vec_IntClear(temp)
		for i := 0; i < int(group[0]); i++ {
			pin := int32(group[2+i])
			switch pin {
			case n:
				Vec_IntPush(temp, first)
			case n + 1:
				Vec_IntPush(temp, second)
			default:
				Vec_IntPush(temp, Vec_IntEntry(leaves, int(pin)))
			}
		}
		return Gia_ManFromIfLogicCreateLut(dst, []word{function}, temp, cover, mapping, records)
	}
	first := build(groups[1][:], functions[1], -1, -1)
	second := int32(-1)
	if length == 3 && groups[2][0] > 0 {
		second = build(groups[2][:], functions[2], first, -1)
		Vec_IntPush(packing, 3)
		Vec_IntPush(packing, first>>1)
		Vec_IntPush(packing, second>>1)
	} else {
		Vec_IntPush(packing, 2)
		Vec_IntPush(packing, first>>1)
	}
	result := build(groups[0][:], functions[0], first, second)
	Vec_IntPush(packing, result>>1)
	vec.Vec_IntAddToEntry(packing, 0, 1)
	return result
}
