package acec

// Cut-set storage and node processing from src/proof/acec/acec2Mult.c.
import (
	"fmt"
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Sdb_CutCountBits(x uint64) int32 { return int32(bits.OnesCount64(x)) }
func Sdb_CutAddUnit(p *Sdb_Sto_t, id int32) {
	v := vec.Vec_WecEntry(p.vCuts, int(id))
	if vec.Vec_IntSize(v) == 0 {
		vec.Vec_IntPush(v, 1)
	} else {
		vec.Vec_IntAddToEntry(v, 0, 1)
	}
	vec.Vec_IntPush(v, 1)
	vec.Vec_IntPush(v, id)
	vec.Vec_IntPush(v, 2)
}
func Sdb_CutAddZero(p *Sdb_Sto_t, id int32) {
	v := vec.Vec_WecEntry(p.vCuts, int(id))
	if vec.Vec_IntSize(v) != 0 {
		panic("acec: constant cut set is not empty")
	}
	vec.Vec_IntPush(v, 1)
	vec.Vec_IntPush(v, 0)
	vec.Vec_IntPush(v, 0)
}
func Sdb_CutTreeLeaves(p *Sdb_Sto_t, c *Sdb_Cut_t) int32 {
	var cost int32
	for _, id := range c.pLeaves[:c.nLeaves] {
		if vec.Vec_IntEntry(p.vRefs, int(id)) == 1 {
			cost++
		}
	}
	return cost
}
func Sdb_StoPrepareSet(p *Sdb_Sto_t, id, index int32) int32 {
	list := vec.Vec_IntArray(vec.Vec_WecEntry(p.vCuts, int(id)))
	n := list[0]
	off := 1
	for i := int32(0); i < n; i++ {
		size := list[off]
		c := &p.pCuts[index][i]
		c.nLeaves = uint32(size)
		copy(c.pLeaves[:], list[off+1:off+1+int(size)])
		c.iFunc = list[off+1+int(size)]
		c.Sign = Sdb_CutGetSign(c)
		c.nTreeLeaves = uint32(Sdb_CutTreeLeaves(p, c))
		off += int(size) + 2
	}
	return n
}
func Sdb_StoInitResult(p *Sdb_Sto_t) {
	for i := range p.ppCuts {
		p.ppCuts[i] = &p.pCuts[2][i]
	}
}
func Sdb_StoStoreResult(p *Sdb_Sto_t, id int32, cuts []*Sdb_Cut_t, n int32) {
	v := vec.Vec_WecEntry(p.vCuts, int(id))
	vec.Vec_IntPush(v, n)
	for _, c := range cuts[:n] {
		vec.Vec_IntPush(v, int32(c.nLeaves))
		for _, leaf := range c.pLeaves[:c.nLeaves] {
			vec.Vec_IntPush(v, leaf)
		}
		vec.Vec_IntPush(v, c.iFunc)
	}
}
func Sdb_CutPrint(p *Sdb_Sto_t, id int32, c *Sdb_Cut_t) {
	if c == nil {
		fmt.Println("No cut.")
		return
	}
	width := len(fmt.Sprint(gia.Gia_ManObjNum(p.pGia)))
	fmt.Printf("%d  {", c.nLeaves)
	for _, leaf := range c.pLeaves[:c.nLeaves] {
		fmt.Printf(" %*d", width, leaf)
	}
	for i := int32(c.nLeaves); i < p.nCutSize; i++ {
		fmt.Printf(" %*s", width, "")
	}
	fmt.Printf("  }  Cost = %3d  CostL = %3d  Tree = %d  \n", c.Cost, c.CostLev, c.nTreeLeaves)
}
func Sdb_StoMergeCuts(p *Sdb_Sto_t, id int32) {
	o := gia.Gia_ManObj(p.pGia, id)
	if gia.Gia_ObjIsBuf(o) != 0 || gia.Gia_ObjIsMux(p.pGia, o) != 0 {
		panic("acec: unsupported cut node")
	}
	x := gia.Gia_ObjIsXor(o)
	a := Sdb_StoPrepareSet(p, gia.Gia_ObjFaninId0(o, id), 0)
	b := Sdb_StoPrepareSet(p, gia.Gia_ObjFaninId1(o, id), 1)
	Sdb_StoInitResult(p)
	cuts := p.ppCuts[:]
	var n int32
	p.CutCount[0] += float64(a * b)
	for i := int32(0); i < a; i++ {
		for j := int32(0); j < b; j++ {
			c0, c1 := &p.pCuts[0][i], &p.pCuts[1][j]
			if int32(c0.nLeaves+c1.nLeaves) > p.nCutSize && Sdb_CutCountBits(c0.Sign|c1.Sign) > p.nCutSize {
				continue
			}
			p.CutCount[1]++
			if Sdb_CutMergeOrder(c0, c1, cuts[n], p.nCutSize) == 0 || Sdb_CutSetLastCutIsContained(cuts, n) != 0 {
				continue
			}
			p.CutCount[2]++
			if p.fCutMin != 0 && Sdb_CutComputeTruth(p, c0, c1, gia.Gia_ObjFaninC0(o), gia.Gia_ObjFaninC1(o), cuts[n], x) != 0 {
				cuts[n].Sign = Sdb_CutGetSign(cuts[n])
			}
			cuts[n].nTreeLeaves = uint32(Sdb_CutTreeLeaves(p, cuts[n]))
			n = Sdb_CutSetAddCut(cuts, n, p.nCutNum)
		}
	}
	p.CutCount[3] += float64(n)
	if n == p.nCutNum-1 {
		p.nCutsOver++
	}
	p.nCutsR = n
	p.Pivot = id
	if n <= 0 || n >= p.nCutNum {
		panic("acec: invalid result count")
	}
	Sdb_CutSetCheckArray(cuts, n)
	Sdb_StoStoreResult(p, id, cuts, n)
	if n > 1 || cuts[0].nLeaves > 1 {
		Sdb_CutAddUnit(p, id)
	}
}
