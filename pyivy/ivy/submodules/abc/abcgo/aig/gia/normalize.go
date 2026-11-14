package gia

import "github.com/glycerine/abcgo/misc/tim"

func Gia_ManIsSeqWithBoxes(p *Gia_Man_t) int32 { return giaBool(p.nRegs > 0 && Gia_ManBoxNum(p) > 0) }
func Gia_ManIsNormalized(p *Gia_Man_t) int32 {
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		if Gia_ObjIsCi(p.pObjs[1+i]) == 0 {
			return 0
		}
	}
	offset := 1 + Gia_ManCiNum(p) + Gia_ManAndNum(p)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		if Gia_ObjIsCo(p.pObjs[offset+i]) == 0 {
			return 0
		}
	}
	return 1
}
func Gia_ManDupNormalize(src *Gia_Man_t, hash int32) *Gia_Man_t {
	Gia_ManFillValue(src)
	dst := Gia_ManStart(src.nObjs)
	dst.CopyIdentityFrom(src)
	// Without hashing, retain simple-mode ANDs with repeated fanins. In
	// particular x & !x must remain an AND, not the buffer C mistakes it for.
	// Hashed copies instead simplify these nodes through Gia_ManHashAnd.
	if hash == 0 {
		dst.fGiaSimple = src.fGiaSimple
	}
	src.pObjs[0].Value = 0
	if src.PMuxes != nil {
		dst.EnableMuxes()
	}
	inputs := append([]int32(nil), Vec_IntArray(src.vCis)...)
	if Gia_ManIsSeqWithBoxes(src) != 0 {
		manager, ok := src.PManTime.(*tim.Tim_Man_t)
		if !ok {
			panic("gia: normalization needs native box timing")
		}
		cis, all := tim.Tim_ManPiNum(manager), tim.Tim_ManCiNum(manager)
		pis := cis - int(src.nRegs)
		if all != len(inputs) || pis <= 0 || cis > all {
			panic("gia: invalid box input partition")
		}
		reordered := append([]int32(nil), inputs[:pis]...)
		reordered = append(reordered, inputs[cis:]...)
		reordered = append(reordered, inputs[pis:cis]...)
		inputs = reordered
	}
	for _, id := range inputs {
		src.pObjs[id].Value = uint32(Gia_ManAppendCi(dst))
	}
	if hash != 0 {
		Gia_ManHashAlloc(dst)
	}
	for _, obj := range src.pObjs {
		a, b, c, kind := src.Gate(obj.id)
		var lit int32
		switch kind {
		case 2:
			if hash != 0 {
				lit = Gia_ManHashAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			} else {
				lit = Gia_ManAppendAnd(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			}
		case 3:
			if hash != 0 {
				lit = Gia_ManHashXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			} else {
				lit = Gia_ManAppendXorReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b))
			}
		case 4:
			if hash != 0 {
				lit = Gia_ManHashMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
			} else {
				lit = Gia_ManAppendMuxReal(dst, Gia_ObjLitCopy(src, a), Gia_ObjLitCopy(src, b), Gia_ObjLitCopy(src, c))
			}
		case 5:
			lit = Gia_ManAppendBuf(dst, Gia_ObjLitCopy(src, a))
		default:
			continue
		}
		obj.Value = uint32(lit)
	}
	if hash != 0 {
		Gia_ManHashStop(dst)
	}
	for _, id := range Vec_IntArray(src.vCos) {
		src.pObjs[id].Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(src.pObjs[id])))
	}
	Gia_ManSetRegNum(dst, src.nRegs)
	dst.nConstrs = src.nConstrs
	if src.PReprs != nil {
		mapping := make([]int32, src.nObjs)
		for id, obj := range src.pObjs {
			mapping[id] = int32(obj.Value) >> 1
		}
		if err := dst.TransferRepresentativesFrom(src, mapping); err != nil {
			panic(err)
		}
	}
	return dst
}

func Gia_ManBoxNum(p *Gia_Man_t) int32 {
	if p.PManTime == nil {
		return 0
	}
	return int32(tim.Tim_ManBoxNum(p.PManTime.(*tim.Tim_Man_t)))
}
func Gia_ManRegBoxNum(p *Gia_Man_t) int32 {
	if p.vRegClasses == nil {
		return 0
	}
	return int32(Vec_IntSize(p.vRegClasses))
}
func Gia_ManNonRegBoxNum(p *Gia_Man_t) int32 { return Gia_ManBoxNum(p) - Gia_ManRegBoxNum(p) }
func Gia_ManBlackBoxNum(p *Gia_Man_t) int32 {
	if p.PManTime == nil {
		return 0
	}
	return int32(tim.Tim_ManBlackBoxNum(p.PManTime.(*tim.Tim_Man_t)))
}
func Gia_ManBoxCiNum(p *Gia_Man_t) int32 {
	if p.PManTime == nil {
		return 0
	}
	return Gia_ManCiNum(p) - int32(tim.Tim_ManPiNum(p.PManTime.(*tim.Tim_Man_t)))
}
func Gia_ManBoxCoNum(p *Gia_Man_t) int32 {
	if p.PManTime == nil {
		return 0
	}
	return Gia_ManCoNum(p) - int32(tim.Tim_ManPoNum(p.PManTime.(*tim.Tim_Man_t)))
}
func Gia_ManClockDomainNum(p *Gia_Man_t) int32 {
	if p.vRegClasses == nil {
		return 0
	}
	domains := make(map[int32]bool)
	for _, class := range Vec_IntArray(p.vRegClasses) {
		if class == 0 {
			panic("gia: register clock domain cannot be zero")
		}
		if class > 0 {
			domains[class] = true
		}
	}
	return int32(len(domains))
}
