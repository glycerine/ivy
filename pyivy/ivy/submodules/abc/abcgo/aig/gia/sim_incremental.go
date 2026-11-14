package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManIncrSimUpdate(p *Gia_Man_t) {
	if p.fIncrSim == 0 || vec.Vec_IntSize(p.vTimeStamps) > int(p.nObjs) || vec.Vec_WrdSize(p.vSims) > int(p.nObjs*p.NSimWords) || p.iNextPi > Gia_ManCiNum(p) {
		panic("gia: invalid incremental simulation state")
	}
	vec.Vec_IntFillExtra(p.vTimeStamps, int(p.nObjs), 0)
	vec.Vec_WrdFillExtra(p.vSims, int(p.nObjs*p.NSimWords), 0)
	for i := p.iNextPi; i < Gia_ManCiNum(p); i++ {
		words := Gia_ManBuiltInData(p, Gia_ManCi(p, i).id)
		for k := range words {
			words[k] = Gia_ManRandomW(0)
		}
	}
	p.iNextPi = Gia_ManCiNum(p)
}
func Gia_ManIncrSimStart(p *Gia_Man_t, words, objects int32) {
	if p.fIncrSim != 0 || words <= 0 || objects < p.nObjs {
		panic("gia: invalid incremental simulation start")
	}
	p.fIncrSim = 1
	p.iPatsPi = 0
	p.NSimWords = words
	p.iTimeStamp = 1
	p.iNextPi = 0
	p.vTimeStamps = vec.Vec_IntAlloc(int(words))
	p.vSims = vec.Vec_WrdAlloc(int(words * objects))
	Gia_ManRandomW(1)
}
func Gia_ManIncrSimStop(p *Gia_Man_t) {
	if p.fIncrSim == 0 {
		panic("gia: incremental simulation is not active")
	}
	p.fIncrSim = 0
	p.iPatsPi = 0
	p.NSimWords = 0
	p.iTimeStamp = 1
	p.vTimeStamps = nil
	p.vSims = nil
}
func Gia_ManIncrSimSet(p *Gia_Man_t, lits *Vec_Int_t) {
	if vec.Vec_IntSize(lits) == 0 {
		panic("gia: empty incremental assignment")
	}
	p.iTimeStamp++
	for _, lit := range vec.Vec_IntArray(lits) {
		id := lit >> 1
		obj := Gia_ManObj(p, id)
		if Gia_ObjIsAnd(obj) != 0 {
			continue
		}
		words := Gia_ManBuiltInData(p, id)
		vec.Vec_IntWriteEntry(p.vTimeStamps, int(id), p.iTimeStamp)
		if giaSimBit(words, p.iPatsPi) == uint64(lit&1) {
			giaSimFlip(words, p.iPatsPi)
		}
	}
	p.iPatsPi++
	if p.iPatsPi == 64*p.NSimWords {
		p.iPatsPi = 0
	}
}
func Gia_ManIncrSimCone_rec(p *Gia_Man_t, id int32) {
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsCi(obj) != 0 {
		return
	}
	stamp := vec.Vec_IntEntry(p.vTimeStamps, int(id))
	if stamp == p.iTimeStamp {
		return
	}
	if stamp > p.iTimeStamp || Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid incremental simulation cone")
	}
	vec.Vec_IntWriteEntry(p.vTimeStamps, int(id), p.iTimeStamp)
	Gia_ManIncrSimCone_rec(p, Gia_ObjFaninId0(obj, id))
	Gia_ManIncrSimCone_rec(p, Gia_ObjFaninId1(obj, id))
	Gia_ManBuiltInSimPerformInt(p, id)
}
func Gia_ManIncrSimCheckOver(p *Gia_Man_t, a, b int32) int32 {
	if a <= 1 || b <= 1 {
		panic("gia: invalid simulation literals")
	}
	Gia_ManIncrSimUpdate(p)
	Gia_ManIncrSimCone_rec(p, a>>1)
	Gia_ManIncrSimCone_rec(p, b>>1)
	return Gia_ManBuiltInSimCheckOver(p, a, b)
}
func Gia_ManIncrSimCheckEqual(p *Gia_Man_t, a, b int32) int32 {
	if a <= 1 || b <= 1 {
		panic("gia: invalid simulation literals")
	}
	Gia_ManIncrSimUpdate(p)
	Gia_ManIncrSimCone_rec(p, a>>1)
	Gia_ManIncrSimCone_rec(p, b>>1)
	return Gia_ManBuiltInSimCheckEqual(p, a, b)
}
