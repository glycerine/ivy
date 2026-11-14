package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManBuiltInDataPi(p *Gia_Man_t, id int32) []uint64 {
	start := int(p.NSimWords * id)
	return vec.Vec_WrdArray(p.VSimsPi)[start : start+int(p.NSimWords)]
}
func Gia_ManBuiltInData(p *Gia_Man_t, id int32) []uint64 {
	start := int(p.NSimWords * id)
	return vec.Vec_WrdArray(p.vSims)[start : start+int(p.NSimWords)]
}
func Gia_ManBuiltInDataPrint(_ *Gia_Man_t, _ int32) {}
func Gia_ManBuiltInSimStart(p *Gia_Man_t, words, objects int32) {
	if p.fBuiltInSim != 0 || Gia_ManAndNum(p) != 0 || words <= 0 || objects < p.nObjs {
		panic("gia: invalid built-in simulation start")
	}
	p.fBuiltInSim = 1
	p.iPatsPi = 0
	p.iPastPiMax = 0
	p.NSimWords = words
	p.nSimWordsMax = 8
	Gia_ManRandomW(1)
	p.VSimsPi = vec.Vec_WrdStart(int(words * Gia_ManCiNum(p)))
	p.vSims = vec.Vec_WrdAlloc(int(words * objects))
	vec.Vec_WrdFill(p.vSims, int(words), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		for k := int32(0); k < words; k++ {
			vec.Vec_WrdPush(p.vSims, Gia_ManRandomW(0))
		}
	}
}
func Gia_ManBuiltInSimResimulateCone_rec(p *Gia_Man_t, id int32) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: simulation cone expects AND")
	}
	Gia_ManBuiltInSimResimulateCone_rec(p, Gia_ObjFaninId0(obj, id))
	Gia_ManBuiltInSimResimulateCone_rec(p, Gia_ObjFaninId1(obj, id))
	Gia_ManBuiltInSimPerformInt(p, id)
}
func Gia_ManBuiltInSimResimulateCone(p *Gia_Man_t, a, b int32) {
	Gia_ManIncrementTravId(p)
	Gia_ManBuiltInSimResimulateCone_rec(p, a>>1)
	Gia_ManBuiltInSimResimulateCone_rec(p, b>>1)
}
func Gia_ManBuiltInSimResimulate(p *Gia_Man_t) {
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManBuiltInSimPerformInt(p, obj.id)
		}
	}
}
func Gia_ManBuiltInSimCheckOver(p *Gia_Man_t, a, b int32) int32 {
	if p.fBuiltInSim == 0 && p.fIncrSim == 0 {
		panic("gia: simulation is not active")
	}
	x, y := Gia_ManBuiltInData(p, a>>1), Gia_ManBuiltInData(p, b>>1)
	for i := range x {
		v, w := x[i], y[i]
		if a&1 != 0 {
			v = ^v
		}
		if b&1 != 0 {
			w = ^w
		}
		if v&w != 0 {
			return 1
		}
	}
	return 0
}
func Gia_ManBuiltInSimCheckEqual(p *Gia_Man_t, a, b int32) int32 {
	if p.fBuiltInSim == 0 && p.fIncrSim == 0 {
		panic("gia: simulation is not active")
	}
	x, y := Gia_ManBuiltInData(p, a>>1), Gia_ManBuiltInData(p, b>>1)
	for i := range x {
		v, w := x[i], y[i]
		if a&1 != 0 {
			v = ^v
		}
		if b&1 != 0 {
			w = ^w
		}
		if v != w {
			return 0
		}
	}
	return 1
}
