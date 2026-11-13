package acec

// Polynomial node ordering from src/proof/acec/acecOrder.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
)

func Gia_PolynFindOrder(p *Gia_Man_t, fadds, hadds *Vec_Int_t, verbose, veryVerbose int32) *Vec_Int_t {
	record := vec.Vec_IntAlloc(100)
	frontier := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	for i := int32(0); i < gia.Gia_ManCoNum(p); i++ {
		vec.Vec_IntWriteEntry(frontier, int(gia.Gia_ObjFaninId0p(p, gia.Gia_ManCo(p, i))), 1)
	}
	xors := Gia_PolynCollectLastXor(p, verbose)
	p.SetPolynXors(xors)
	fmt.Printf("Collected %d topmost XORs\n", vec.Vec_IntSize(xors))
	for _, id := range vec.Vec_IntArray(xors) {
		o := gia.Gia_ManObj(p, id)
		if vec.Vec_IntEntry(frontier, int(id)) == 0 {
			panic("acec: top XOR absent from frontier")
		}
		vec.Vec_IntWriteEntry(frontier, int(id), 0)
		vec.Vec_IntWriteEntry(frontier, int(gia.Gia_ObjFaninId0(o, id)), 1)
		vec.Vec_IntWriteEntry(frontier, int(gia.Gia_ObjFaninId1(o, id)), 1)
		vec.Vec_IntPush(record, id<<2|3)
		if veryVerbose != 0 {
			fmt.Printf("Recognizing %d => XXXOR(%d %d)\n", id, gia.Gia_ObjFaninId0(o, id), gia.Gia_ObjFaninId1(o, id))
		}
	}
	for iteration := 0; ; iteration++ {
		if veryVerbose != 0 {
			fmt.Printf("Iteration %d\n", iteration)
		}
		foundAll := false
		for {
			found := false
			for i := vec.Vec_IntSize(fadds)/5 - 1; i >= 0; i-- {
				sum, carry := vec.Vec_IntEntry(fadds, 5*i+3), vec.Vec_IntEntry(fadds, 5*i+4)
				if vec.Vec_IntEntry(frontier, int(sum)) == 0 || vec.Vec_IntEntry(frontier, int(carry)) == 0 {
					continue
				}
				vec.Vec_IntWriteEntry(frontier, int(sum), 0)
				vec.Vec_IntWriteEntry(frontier, int(carry), 0)
				for k := 0; k < 3; k++ {
					vec.Vec_IntWriteEntry(frontier, int(vec.Vec_IntEntry(fadds, 5*i+k)), 1)
				}
				vec.Vec_IntPush(record, int32(i)<<2|2)
				found, foundAll = true, true
				if veryVerbose != 0 {
					fmt.Printf("Recognizing (%d %d) => FA(%d %d %d)\n", sum, carry, vec.Vec_IntEntry(fadds, 5*i), vec.Vec_IntEntry(fadds, 5*i+1), vec.Vec_IntEntry(fadds, 5*i+2))
				}
			}
			if !found {
				break
			}
		}
		for i := vec.Vec_IntSize(hadds)/2 - 1; i >= 0; i-- {
			sum, carry := vec.Vec_IntEntry(hadds, 2*i), vec.Vec_IntEntry(hadds, 2*i+1)
			if vec.Vec_IntEntry(frontier, int(sum)) == 0 || vec.Vec_IntEntry(frontier, int(carry)) == 0 {
				continue
			}
			o := gia.Gia_ManObj(p, carry)
			vec.Vec_IntWriteEntry(frontier, int(sum), 0)
			vec.Vec_IntWriteEntry(frontier, int(carry), 0)
			vec.Vec_IntWriteEntry(frontier, int(gia.Gia_ObjFaninId0(o, carry)), 1)
			vec.Vec_IntWriteEntry(frontier, int(gia.Gia_ObjFaninId1(o, carry)), 1)
			vec.Vec_IntPush(record, int32(i)<<2|1)
			foundAll = true
			if veryVerbose != 0 {
				fmt.Printf("Recognizing (%d %d) => HA(%d %d)\n", sum, carry, gia.Gia_ObjFaninId0(o, carry), gia.Gia_ObjFaninId1(o, carry))
			}
		}
		if !foundAll {
			break
		}
	}
	left := vec.Vec_IntAlloc(100)
	for id := int32(0); id < gia.Gia_ManObjNum(p); id++ {
		if vec.Vec_IntEntry(frontier, int(id)) != 0 && gia.Gia_ObjIsAnd(gia.Gia_ManObj(p, id)) != 0 {
			vec.Vec_IntPush(left, id)
		}
	}
	result := vec.Vec_IntAlloc(100)
	gia.Gia_ManIncrementTravId(p)
	gia.Gia_ManCollectAnds(p, vec.Vec_IntArray(left), int32(vec.Vec_IntSize(left)), result, nil)
	for i := 0; i < vec.Vec_IntSize(result); i++ {
		vec.Vec_IntWriteEntry(result, i, vec.Vec_IntEntry(result, i)<<2)
	}
	vec.Vec_IntReverseOrder(record)
	vec.Vec_IntAppend(result, record)
	return result
}
func Gia_PolynReorder(p *Gia_Man_t, verbose, veryVerbose int32) *Vec_Int_t {
	fadds := Gia_ManDetectFullAdders(p, veryVerbose, nil)
	hadds := Gia_ManDetectHalfAdders(p, veryVerbose)
	record := Gia_PolynFindOrder(p, fadds, hadds, verbose, veryVerbose)
	order := vec.Vec_IntAlloc(int(gia.Gia_ManAndNum(p)))
	result := vec.Vec_IntStart(int(gia.Gia_ManObjNum(p)))
	gia.Gia_ManIncrementTravId(p)
	for _, entry := range vec.Vec_IntArray(record) {
		node, attr := entry>>2, entry&3
		if attr == 2 {
			for k := 3; k < 5; k++ {
				gia.Gia_ManCollectAnds_rec(p, vec.Vec_IntEntry(fadds, int(5*node)+k), order)
			}
		} else if attr == 1 {
			for k := 0; k < 2; k++ {
				gia.Gia_ManCollectAnds_rec(p, vec.Vec_IntEntry(hadds, int(2*node)+k), order)
			}
		} else {
			gia.Gia_ManCollectAnds_rec(p, node, order)
		}
	}
	for i := int32(0); i < gia.Gia_ManCiNum(p); i++ {
		vec.Vec_IntWriteEntry(result, int(gia.Gia_ObjId(p, gia.Gia_ManCi(p, i))), 1+i)
	}
	for i, id := range vec.Vec_IntArray(order) {
		vec.Vec_IntWriteEntry(result, int(id), 1+gia.Gia_ManCiNum(p)+int32(i))
	}
	return result
}
