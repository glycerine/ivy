package acec

// Half-adder detection from src/proof/acec/acecFadds.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManDetectHalfAdders(p *Gia_Man_t, verbose int32) *Vec_Int_t {
	if !acecCanDetectArithmetic(p) {
		panic("acec: half-adder detection requires AND/XOR nodes without native MUXes or buffers")
	}
	result := vec.Vec_IntAlloc(1000)
	var counts [5]int
	gia.Gia_ManHashStart(p)
	real := gia.Gia_ManXorNum(p) != 0
	if !real {
		p.Refs = nil
		gia.Gia_ManCreateRefs(p)
	}
	for id := int32(1); id < gia.Gia_ManObjNum(p); id++ {
		o := gia.Gia_ManObj(p, id)
		if gia.Gia_ObjIsAnd(o) == 0 {
			continue
		}
		count := 0
		var a, b int32
		var choices [][2]int32
		if real {
			if gia.Gia_ObjIsXor(o) == 0 {
				continue
			}
			a, b = gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)
			choices = [][2]int32{{0, 0}, {1, 1}, {0, 1}, {1, 0}}
		} else {
			var x, y *Gia_Obj_t
			if gia.Gia_ObjRecognizeExor(o, &x, &y) == 0 {
				continue
			}
			for _, f := range []int32{gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id)} {
				if gia.Gia_ObjRefNumId(p, f) > 1 {
					vec.Vec_IntPushTwo(result, id, f)
					count++
				}
			}
			a, b = gia.Gia_ObjId(p, gia.Gia_Regular(x)), gia.Gia_ObjId(p, gia.Gia_Regular(y))
			left, right := gia.Gia_ObjFanin0(o), gia.Gia_ObjFanin1(o)
			diff := gia.Gia_ObjFaninC0(left) ^ gia.Gia_ObjFaninC1(left)
			if diff != gia.Gia_ObjFaninC0(right)^gia.Gia_ObjFaninC1(right) {
				panic("acec: invalid XOR fanin polarity")
			}
			if diff != 0 {
				choices = [][2]int32{{0, 0}, {1, 1}}
			} else {
				choices = [][2]int32{{0, 1}, {1, 0}}
			}
		}
		for _, c := range choices {
			lit := gia.Gia_ManHashLookupInt(p, a<<1|c[0], b<<1|c[1])
			if lit != 0 {
				vec.Vec_IntPushTwo(result, id, lit>>1)
				count++
			}
		}
		counts[count]++
	}
	if !real {
		p.Refs = nil
	}
	gia.Gia_ManHashStop(p)
	if verbose != 0 {
		fmt.Printf("Found %d half-adders with XOR gates: ", vec.Vec_IntSize(result)/2)
		for i, n := range counts {
			fmt.Printf("%d=%d ", i, n)
		}
		fmt.Println()
		for i := 0; i < vec.Vec_IntSize(result); i += 2 {
			x, y := vec.Vec_IntEntry(result, i), vec.Vec_IntEntry(result, i+1)
			o := gia.Gia_ManObj(p, x)
			fmt.Printf("%3d : %5d %5d -> %5d %5d\n", i, gia.Gia_ObjFaninId0(o, x), gia.Gia_ObjFaninId1(o, x), x, y)
		}
	}
	return result
}
