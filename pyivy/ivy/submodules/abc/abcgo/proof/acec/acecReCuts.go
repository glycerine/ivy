package acec

// Three-input cut operations from src/proof/acec/acecRe.c.
import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"sort"
)

func Ree_ManCutMergeOne(a, b, out []int32) int32 {
	copy(out, b[:1+b[0]])
	for _, id := range a[1 : 1+a[0]] {
		found := false
		for _, other := range b[1 : 1+b[0]] {
			if id == other {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if out[0] == 3 {
			return 0
		}
		out[0]++
		out[out[0]] = id
	}
	if out[0] != 2 && out[0] != 3 {
		panic("acec: invalid merged cut size")
	}
	sort.Slice(out[1:1+out[0]], func(i, j int) bool { return out[1+i] < out[1+j] })
	return 1
}
func Ree_ManCutCheckEqual(cuts *Vec_Int_t, cut []int32) int32 {
	list := vec.Vec_IntArray(cuts)
	off := 1
	for i := int32(0); i < list[0]; i++ {
		n := list[off]
		if n == cut[0] {
			same := true
			for k := int32(1); k <= n; k++ {
				if list[off+int(k)] != cut[k] {
					same = false
					break
				}
			}
			if same {
				return 1
			}
		}
		off += int(n) + 2
	}
	return 0
}
func Ree_ManCutFind(id int32, cut []int32) int32 {
	for i := int32(1); i <= 3; i++ {
		if cut[i] == id {
			return i - 1
		}
	}
	panic("acec: cut leaf missing")
}
func Ree_ManCutNotFind(a, b int32, cut []int32) int32 {
	if cut[0] != 3 {
		panic("acec: expected three leaves")
	}
	for i := int32(3); i >= 1; i-- {
		if cut[i] != a && cut[i] != b {
			return 3 - i
		}
	}
	panic("acec: missing third leaf")
}
func Ree_ManCutTruthOne(input, out []int32) int32 {
	t := input[input[0]+1]
	complement := (t >> 7) & 1
	if input[0] == 3 {
		return t
	}
	if complement != 0 {
		t = ^t
	}
	if input[0] == 2 {
		if out[0] == 3 {
			table := [3][8]int32{{0, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77}, {0, 0x05, 0x0a, 0x0f, 0x50, 0x55, 0x5a, 0x5f}, {0, 0x03, 0x0c, 0x0f, 0x30, 0x33, 0x3c, 0x3f}}
			result := table[Ree_ManCutNotFind(input[1], input[2], out)][t&7]
			if complement != 0 {
				result = ^result
			}
			return result & 0xff
		}
		if out[0] != 2 || out[1] != input[1] || out[2] != input[2] {
			panic("acec: incompatible cut")
		}
		return input[input[0]+1]
	}
	if input[0] == 1 {
		table := [3]int32{0x55, 0x33, 0x0f}
		result := table[Ree_ManCutFind(input[1], out)]
		if complement != 0 {
			result = ^result
		}
		return result & 0xff
	}
	panic("acec: invalid input cut")
}
func Ree_ManCutTruth(o *Gia_Obj_t, a, b, out []int32) int32 {
	x, y := Ree_ManCutTruthOne(a, out), Ree_ManCutTruthOne(b, out)
	if gia.Gia_ObjFaninC0(o) != 0 {
		x = ^x
	}
	if gia.Gia_ObjFaninC1(o) != 0 {
		y = ^y
	}
	if gia.Gia_ObjIsXor(o) != 0 {
		return (x ^ y) & 0xff
	}
	return (x & y) & 0xff
}
