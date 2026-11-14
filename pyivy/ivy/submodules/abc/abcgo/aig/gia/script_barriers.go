package gia

import "fmt"

func Gia_ManAigPrintPiLevels(p *Gia_Man_t) {
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		fmt.Printf("%d ", Gia_ObjLevel(p, Gia_ManPi(p, i)))
	}
	fmt.Println()
}
func Gia_ManDupFromBarBufs(p *Gia_Man_t) *Gia_Man_t {
	if Gia_ManBufNum(p) == 0 || Gia_ManRegNum(p) != 0 || Gia_ManHasChoices(p) != 0 {
		panic("gia: invalid barrier conversion source")
	}
	q := Gia_ManStart(p.nObjs)
	q.pName, q.PSpec = p.pName, p.PSpec
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(q))
	}
	buf := make([]int32, Gia_ManBufNum(p))
	for i := range buf {
		buf[i] = Gia_ManAppendCi(q)
	}
	k := 0
	for id := int32(1); id < p.nObjs; id++ {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if Gia_ObjIsBuf(obj) != 0 {
			obj.Value = uint32(buf[k])
			buf[k] = Gia_ObjFanin0Copy(obj)
			k++
		} else {
			obj.Value = uint32(Gia_ManAppendAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	if k != len(buf) {
		panic("gia: inconsistent barrier count")
	}
	for _, lit := range buf {
		Gia_ManAppendCo(q, lit)
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManCo(p, i)))
	}
	Gia_ManSetRegNum(q, 0)
	return q
}
func Gia_ManDupToBarBufs(p *Gia_Man_t, count int32) *Gia_Man_t {
	if count < 0 || count > Gia_ManCiNum(p) || count > Gia_ManCoNum(p) || Gia_ManBufNum(p) != 0 || Gia_ManRegNum(p) != 0 || Gia_ManHasChoices(p) == 0 {
		panic("gia: invalid barrier reconstruction source")
	}
	realIn, realOut := Gia_ManCiNum(p)-count, Gia_ManCoNum(p)-count
	q := Gia_ManStart(p.nObjs)
	q.pName, q.PSpec = p.pName, p.PSpec
	q.pSibls = make([]int32, p.nObjs+count+1)
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < realIn; i++ {
		Gia_ManCi(p, i).Value = uint32(Gia_ManAppendCi(q))
	}
	k := int32(0)
	insert := func() {
		for k < count {
			co := Gia_ManCo(p, k)
			if Gia_ObjFanin0(co).Value == ^uint32(0) {
				break
			}
			Gia_ManCi(p, realIn+k).Value = uint32(Gia_ManAppendBuf(q, Gia_ObjFanin0Copy(co)))
			k++
		}
	}
	for id := int32(1); id < p.nObjs; id++ {
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		insert()
		obj.Value = uint32(Gia_ManAppendAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		if sibling := Gia_ObjSibl(p, id); sibling != 0 {
			q.pSibls[obj.Value>>1] = int32(Gia_ManObj(p, sibling).Value >> 1)
		}
	}
	insert()
	if k != count {
		panic("gia: unresolved barrier outputs")
	}
	for i := int32(0); i < realOut; i++ {
		Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManCo(p, count+i)))
	}
	Gia_ManSetRegNum(q, 0)
	return q
}
