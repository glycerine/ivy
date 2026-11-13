package acec

// Full-adder chain discovery from src/proof/acec/acecFadds.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_ManCreateMap(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Int_t {
	m := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; 5*i < vec.Vec_IntSize(adds); i++ {
		vec.Vec_IntWriteEntry(m, int(vec.Vec_IntEntry(adds, 5*i+4)), int32(i))
	}
	return m
}
func Gia_ManFindChains_rec(p *Gia_Man_t, carry int32, adds, m, length *Vec_Int_t) int32 {
	box := vec.Vec_IntEntry(m, int(carry))
	if box < 0 {
		panic("acec: missing full-adder carry")
	}
	if v := vec.Vec_IntEntry(length, int(carry)); v >= 0 {
		return v
	}
	if gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, carry)) == 0 {
		panic("acec: carry is not AND")
	}
	var lens [3]int32
	off := int(5 * box)
	for k := 0; k < 3; k++ {
		id := vec.Vec_IntEntry(adds, off+k)
		if vec.Vec_IntEntry(m, int(id)) != -1 {
			lens[k] = Gia_ManFindChains_rec(p, id, adds, m, length)
		}
	}
	best := 0
	if lens[1] > lens[best] {
		best = 1
	}
	if lens[2] > lens[best] {
		best = 2
	}
	if best != 0 {
		a := vec.Vec_IntEntry(adds, off)
		vec.Vec_IntWriteEntry(adds, off, vec.Vec_IntEntry(adds, off+best))
		vec.Vec_IntWriteEntry(adds, off+best, a)
	}
	vec.Vec_IntWriteEntry(length, int(carry), lens[best]+1)
	return lens[best] + 1
}
func Gia_ManFindChains(p *Gia_Man_t, adds, m *Vec_Int_t) {
	length := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; 5*i < vec.Vec_IntSize(adds); i++ {
		Gia_ManFindChains_rec(p, vec.Vec_IntEntry(adds, 5*i+4), adds, m, length)
	}
}
func Gia_ManCollectOneChain(p *Gia_Man_t, adds *Vec_Int_t, top int32, m, chain *Vec_Int_t) {
	vec.Vec_IntClear(chain)
	for box := top; box >= 0; box = vec.Vec_IntEntry(m, int(vec.Vec_IntEntry(adds, int(5*box)))) {
		if gia.Gia_ObjIsTravIdCurrentId(p, vec.Vec_IntEntry(adds, int(5*box+3))) != 0 || gia.Gia_ObjIsTravIdCurrentId(p, vec.Vec_IntEntry(adds, int(5*box+4))) != 0 {
			break
		}
		vec.Vec_IntPush(chain, box)
	}
	vec.Vec_IntReverseOrder(chain)
}
func Gia_ManMarkWithTravId_rec(p *Gia_Man_t, id int32) {
	if gia.Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	gia.Gia_ObjSetTravIdCurrentId(p, id)
	o := gia.Gia_ManObj(p, id)
	if gia.Gia_ObjIsAnd(o) != 0 {
		Gia_ManMarkWithTravId_rec(p, gia.Gia_ObjFaninId0(o, id))
		Gia_ManMarkWithTravId_rec(p, gia.Gia_ObjFaninId1(o, id))
	}
}
func Gia_ManCollectTopmost(p *Gia_Man_t, adds, m *Vec_Int_t, min int32) *Vec_Wec_t {
	chain := vec.Vec_IntAlloc(100)
	chains := vec.Vec_WecAlloc(vec.Vec_IntSize(adds) / 5)
	marked := vec.Vec_BitStart(vec.Vec_IntSize(adds) / 5)
	for i := 0; 5*i < vec.Vec_IntSize(adds); i++ {
		id := vec.Vec_IntEntry(m, int(vec.Vec_IntEntry(adds, 5*i)))
		if id >= 0 {
			vec.Vec_BitWriteEntry(marked, int(id), 1)
		}
	}
	gia.Gia_ManIncrementTravId(p)
	for i := 0; 5*i < vec.Vec_IntSize(adds); i++ {
		if vec.Vec_BitEntry(marked, i) != 0 {
			continue
		}
		Gia_ManCollectOneChain(p, adds, int32(i), m, chain)
		if int32(vec.Vec_IntSize(chain)) < min {
			continue
		}
		vec.Vec_IntAppend(vec.Vec_WecPushLevel(chains), chain)
		for _, box := range vec.Vec_IntArray(chain) {
			sum := vec.Vec_IntEntry(adds, int(5*box+3))
			carry := vec.Vec_IntEntry(adds, int(5*box+4))
			if gia.Gia_ObjIsTravIdCurrentId(p, sum) != 0 || gia.Gia_ObjIsTravIdCurrentId(p, carry) != 0 {
				panic("acec: overlapping adder chains")
			}
			Gia_ManMarkWithTravId_rec(p, sum)
			Gia_ManMarkWithTravId_rec(p, carry)
		}
	}
	return chains
}
func Gia_ManPrintChains(p *Gia_Man_t, adds, m *Vec_Int_t, chains *Vec_Wec_t) {
	total := 0
	for i := 0; i < vec.Vec_WecSize(chains); i++ {
		chain := vec.Vec_WecEntry(chains, i)
		total += vec.Vec_IntSize(chain)
		if i < 10 {
			fmt.Printf("Chain %4d : %4d    ", i, vec.Vec_IntSize(chain))
			for k, box := range vec.Vec_IntArray(chain) {
				fmt.Printf("%d(%d) ", box, vec.Vec_IntEntry(adds, int(5*box+4)))
				if k != vec.Vec_IntSize(chain)-1 {
					fmt.Print("-> ")
				}
				if k > 6 {
					fmt.Print("...")
					break
				}
			}
			fmt.Println()
		} else if i == 10 {
			fmt.Println("...")
		}
	}
	fmt.Printf("Total chains = %d. Total full-adders = %d.\n", vec.Vec_WecSize(chains), total)
}
func Gia_ManFindMapping(p *Gia_Man_t, adds, m *Vec_Int_t, chains *Vec_Wec_t) *Vec_Int_t {
	result := vec.Vec_IntStartFull(int(gia.Gia_ManObjNum(p)))
	for i := 0; i < vec.Vec_WecSize(chains); i++ {
		chain := vec.Vec_WecEntry(chains, i)
		if vec.Vec_IntSize(chain) == 0 {
			panic("acec: empty adder chain")
		}
		var box int32
		for _, box = range vec.Vec_IntArray(chain) {
			id := vec.Vec_IntEntry(adds, int(5*box+3))
			if vec.Vec_IntEntry(result, int(id)) != -1 {
				panic("acec: repeated sum")
			}
			vec.Vec_IntWriteEntry(result, int(id), int32(i))
		}
		id := vec.Vec_IntEntry(adds, int(5*box+4))
		if vec.Vec_IntEntry(result, int(id)) != -1 {
			panic("acec: repeated carry")
		}
		vec.Vec_IntWriteEntry(result, int(id), int32(i))
	}
	return result
}
func Gia_ManCollectTruthTables(p *Gia_Man_t, adds *Vec_Int_t) *Vec_Int_t {
	truths := vec.Vec_IntAlloc(2 * vec.Vec_IntSize(adds) / 5)
	gia.Gia_ManCleanValue(p)
	for i := 0; 5*i < vec.Vec_IntSize(adds); i++ {
		cut := []int32{3, vec.Vec_IntEntry(adds, 5*i), vec.Vec_IntEntry(adds, 5*i+1), vec.Vec_IntEntry(adds, 5*i+2)}
		for k := 3; k < 5; k++ {
			var truth int32
			typ := Dtc_ObjComputeTruth(p, vec.Vec_IntEntry(adds, 5*i+k), cut, &truth)
			if typ != int32(k-2) {
				panic("acec: unexpected full-adder truth")
			}
			vec.Vec_IntPush(truths, truth)
		}
	}
	return truths
}
