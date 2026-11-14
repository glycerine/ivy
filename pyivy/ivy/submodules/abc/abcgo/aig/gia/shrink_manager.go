package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/bool/bdc"
	"github.com/glycerine/abcgo/bool/rsb"
	"github.com/glycerine/abcgo/misc/vec"
)

type Shr_Fan_t struct{ iFan, Next int32 }
type Shr_Man_t struct {
	pGia, pNew                       *Gia_Man_t
	nDivMax, nNewSize                int32
	fanMem                           []Shr_Fan_t
	vObj2Fan                         *Vec_Int_t
	fanIter                          int32
	vDivs, vPrio, vDivResub, vLeaves *Vec_Int_t
	vTruths, vDivTruths              *Vec_Wrd_t
	pManRsb                          *rsb.Rsb_Man_t
	pManDec                          *bdc.Bdc_Man_t
	Pars                             bdc.Bdc_Par_t
}

func Shr_ManAlloc(g *Gia_Man_t) *Shr_Man_t {
	p := &Shr_Man_t{pGia: g, nDivMax: 64, nNewSize: 2 * g.nObjs, fanMem: []Shr_Fan_t{{}}}
	p.vObj2Fan = vec.Vec_IntStart(int(p.nNewSize))
	p.vDivs = vec.Vec_IntAlloc(1000)
	p.vPrio = vec.Vec_IntAlloc(1000)
	p.vTruths = vec.Vec_WrdStart(int(p.nNewSize))
	p.vDivTruths = vec.Vec_WrdAlloc(100)
	p.vDivResub = vec.Vec_IntAlloc(6)
	p.vLeaves = vec.Vec_IntAlloc(6)
	p.pNew = Gia_ManStart(p.nNewSize)
	p.pNew.pName, p.pNew.PSpec = g.pName, g.PSpec
	Gia_ManHashStart(p.pNew)
	Gia_ManCleanLevels(p.pNew, p.nNewSize)
	Gia_ManIncrementTravId(p.pNew)
	p.Pars.NVarsMax = 6
	p.pManDec = bdc.Bdc_ManAlloc(&p.Pars)
	p.pManRsb = rsb.Rsb_ManAlloc(6, p.nDivMax, 4, 1)
	return p
}
func Shr_ManFree(p *Shr_Man_t) *Gia_Man_t {
	q := p.pNew
	Gia_ManHashStop(q)
	q.vLevels = nil
	if Gia_ManHasDangling(q) != 0 {
		old := q
		q = Gia_ManCleanup(q)
		if Gia_ManAndNum(q) != Gia_ManAndNum(old) {
			fmt.Printf("Node reduction after sweep %6d -> %6d.\n", Gia_ManAndNum(old), Gia_ManAndNum(q))
		}
	}
	Gia_ManSetRegNum(q, Gia_ManRegNum(p.pGia))
	rsb.Rsb_ManFree(p.pManRsb)
	bdc.Bdc_ManFree(p.pManDec)
	p.pNew = nil
	return q
}
func Shr_ManAddFanout(p *Shr_Man_t, fanin, fanout int32) {
	if fanin < 0 || int(fanin) >= vec.Vec_IntSize(p.vObj2Fan) || fanout <= 0 {
		panic("gia: invalid shrink fanout")
	}
	p.fanMem = append(p.fanMem, Shr_Fan_t{fanout, vec.Vec_IntEntry(p.vObj2Fan, int(fanin))})
	vec.Vec_IntWriteEntry(p.vObj2Fan, int(fanin), int32(len(p.fanMem)-1))
}
func Shr_ManFanIterStart(p *Shr_Man_t, id int32) int32 {
	p.fanIter = vec.Vec_IntEntry(p.vObj2Fan, int(id))
	if p.fanIter == 0 {
		return 0
	}
	return p.fanMem[p.fanIter].iFan
}
func Shr_ManFanIterNext(p *Shr_Man_t) int32 {
	if p.fanIter == 0 {
		return 0
	}
	p.fanIter = p.fanMem[p.fanIter].Next
	if p.fanIter == 0 {
		return 0
	}
	return p.fanMem[p.fanIter].iFan
}
func Shr_ManDivPushOrderByLevel(p *Shr_Man_t, id int32) int32 {
	vec.Vec_IntPush(p.vPrio, id)
	data := vec.Vec_IntArray(p.vPrio)
	i := len(data) - 1
	for i > 0 && Gia_ObjLevelId(p.pNew, data[i-1]) > Gia_ObjLevelId(p.pNew, data[i]) {
		data[i-1], data[i] = data[i], data[i-1]
		i--
	}
	return int32(i)
}
func Shr_ManCollectDivisors(p *Shr_Man_t, leaves *Vec_Int_t, limit, maxFanout int32) int32 {
	if limit <= 6 {
		panic("gia: invalid shrink divisor limit")
	}
	vec.Vec_IntClear(p.vDivs)
	vec.Vec_IntClear(p.vPrio)
	Gia_ManIncrementTravId(p.pNew)
	for _, id := range vec.Vec_IntArray(leaves) {
		vec.Vec_IntPush(p.vDivs, id)
		Shr_ManDivPushOrderByLevel(p, id)
		Gia_ObjSetTravIdCurrentId(p.pNew, id)
	}
	for i := 0; i < vec.Vec_IntSize(p.vPrio); i++ {
		id := vec.Vec_IntEntry(p.vPrio, i)
		count := int32(0)
		for fan := Shr_ManFanIterStart(p, id); fan != 0; fan = Shr_ManFanIterNext(p) {
			if count == maxFanout {
				break
			}
			count++
			if Gia_ObjIsTravIdCurrentId(p.pNew, fan) != 0 {
				continue
			}
			obj := Gia_ManObj(p.pNew, fan)
			if Gia_ObjIsAnd(obj) == 0 {
				panic("gia: invalid shrink fanout node")
			}
			if Gia_ObjIsTravIdCurrentId(p.pNew, Gia_ObjFaninId0(obj, fan)) == 0 || Gia_ObjIsTravIdCurrentId(p.pNew, Gia_ObjFaninId1(obj, fan)) == 0 {
				continue
			}
			vec.Vec_IntPush(p.vDivs, fan)
			Gia_ObjSetTravIdCurrentId(p.pNew, fan)
			if int(Shr_ManDivPushOrderByLevel(p, fan)) <= i {
				panic("gia: shrink divisor order")
			}
			if int32(vec.Vec_IntSize(p.vDivs)) == limit {
				return limit
			}
		}
	}
	return int32(vec.Vec_IntSize(p.vDivs))
}
