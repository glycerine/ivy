package gia

import "github.com/glycerine/abcgo/misc/vec"

func giaSimBit(words []uint64, bit int32) uint64 { return words[bit>>6] >> uint(bit&63) & 1 }
func giaSimFlip(words []uint64, bit int32)       { words[bit>>6] ^= uint64(1) << uint(bit&63) }

func Gia_ManBuiltInSimPack(p *Gia_Man_t, pattern *Vec_Int_t) int32 {
	lits := vec.Vec_IntArray(pattern)
	if len(lits) == 0 {
		panic("gia: empty simulation pattern")
	}
	for bit := int32(0); bit < p.iPatsPi; bit++ {
		compatible := true
		for _, lit := range lits {
			id := lit >> 1
			if giaSimBit(Gia_ManBuiltInDataPi(p, id), bit) != 0 && giaSimBit(Gia_ManBuiltInData(p, 1+id), bit) == uint64(lit&1) {
				compatible = false
				break
			}
		}
		if compatible {
			return bit
		}
	}
	return -1
}

func giaSimGrowWords(p *Gia_Man_t) {
	old := int(p.NSimWords)
	data := vec.Vec_WrdArray(p.vSims)
	if len(data) != int(p.nObjs)*old {
		panic("gia: inconsistent simulation data")
	}
	expanded := vec.Vec_WrdAlloc(int(p.nObjs) * (old + 1))
	for i, obj := range p.pObjs {
		for _, word := range data[i*old : (i+1)*old] {
			vec.Vec_WrdPush(expanded, word)
		}
		var next uint64
		if Gia_ObjIsCi(obj) != 0 {
			next = Gia_ManRandomW(0)
		} else if Gia_ObjIsAnd(obj) != 0 && obj.fPhase != 0 {
			next = ^uint64(0)
		}
		vec.Vec_WrdPush(expanded, next)
	}
	p.vSims = expanded
	care := vec.Vec_WrdArray(p.VSimsPi)
	expanded = vec.Vec_WrdAlloc(int(Gia_ManCiNum(p)) * (old + 1))
	for i := 0; i < int(Gia_ManCiNum(p)); i++ {
		for _, word := range care[i*old : (i+1)*old] {
			vec.Vec_WrdPush(expanded, word)
		}
		vec.Vec_WrdPush(expanded, 0)
	}
	p.VSimsPi = expanded
	p.NSimWords++
}

func Gia_ManBuiltInSimAddPat(p *Gia_Man_t, pattern *Vec_Int_t) int32 {
	overflow := p.iPatsPi == 64*p.NSimWords && p.NSimWords == p.nSimWordsMax
	bit := Gia_ManBuiltInSimPack(p, pattern)
	if bit < 0 {
		if overflow {
			if p.iPastPiMax&15 == 0 {
				Gia_ManBuiltInSimResimulate(p)
			}
			bit = p.iPastPiMax
			p.iPastPiMax++
			if p.iPastPiMax == 64*p.nSimWordsMax {
				p.iPastPiMax = 0
			}
		} else {
			if p.iPatsPi != 0 && p.iPatsPi&15 == 0 {
				Gia_ManBuiltInSimResimulate(p)
			}
			if p.iPatsPi == 64*p.NSimWords {
				giaSimGrowWords(p)
			}
			bit = p.iPatsPi
			p.iPatsPi++
		}
	}
	if bit < 0 || bit >= p.iPatsPi {
		panic("gia: invalid simulation pattern slot")
	}
	if overflow {
		for _, id := range vec.Vec_IntArray(&p.vSuppVars) {
			care := Gia_ManBuiltInDataPi(p, id)
			if giaSimBit(care, bit) != 0 {
				giaSimFlip(care, bit)
			}
		}
	}
	for _, lit := range vec.Vec_IntArray(pattern) {
		id := lit >> 1
		care := Gia_ManBuiltInDataPi(p, id)
		value := Gia_ManBuiltInData(p, 1+id)
		if !overflow && giaSimBit(care, bit) != 0 {
			if giaSimBit(value, bit) == uint64(lit&1) {
				panic("gia: conflicting packed pattern")
			}
			continue
		}
		if giaSimBit(value, bit) == uint64(lit&1) {
			giaSimFlip(value, bit)
		}
		giaSimFlip(care, bit)
	}
	return 1
}
