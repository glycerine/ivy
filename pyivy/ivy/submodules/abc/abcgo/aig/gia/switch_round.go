package gia

import "math/bits"

func Gia_ManSwiSimulateCi(p *Gia_ManSwi_t, obj *Gia_Obj_t, index int32) {
	copy(Gia_SwiData(p, Gia_ObjValue(obj)), Gia_SwiDataCi(p, index))
}
func Gia_ManSwiSimulateCo(p *Gia_ManSwi_t, index int32, obj *Gia_Obj_t) {
	dst, src := Gia_SwiDataCo(p, index), Gia_SwiData(p, Gia_ObjDiff0(obj))
	for w, v := range src {
		if Gia_ObjFaninC0(obj) != 0 {
			v = ^v
		}
		dst[w] = v
	}
}
func Gia_ManSwiSimulateNode(p *Gia_ManSwi_t, obj *Gia_Obj_t) {
	dst, a, b := Gia_SwiData(p, Gia_ObjValue(obj)), Gia_SwiData(p, Gia_ObjDiff0(obj)), Gia_SwiData(p, Gia_ObjDiff1(obj))
	for w := range dst {
		x, y := a[w], b[w]
		if Gia_ObjFaninC0(obj) != 0 {
			x = ^x
		}
		if Gia_ObjFaninC1(obj) != 0 {
			y = ^y
		}
		dst[w] = x & y
	}
}
func Gia_ManSwiSimInfoInit(p *Gia_ManSwi_t) {
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		if i < Gia_ManPiNum(p.pAig) {
			Gia_ManSwiSimInfoRandom(p, Gia_SwiDataCi(p, i), 0)
		} else {
			Gia_ManSwiSimInfoZero(p, Gia_SwiDataCi(p, i))
		}
	}
}
func Gia_ManSwiSimInfoTransfer(p *Gia_ManSwi_t, factor int32) {
	shift := Gia_ManPoNum(p.pAig) - Gia_ManPiNum(p.pAig)
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		if i < Gia_ManPiNum(p.pAig) {
			Gia_ManSwiSimInfoRandom(p, Gia_SwiDataCi(p, i), factor)
		} else {
			Gia_ManSwiSimInfoCopy(p, Gia_SwiDataCi(p, i), Gia_SwiDataCo(p, shift+i))
		}
	}
}
func Gia_ManSwiSimInfoTransferShift(p *Gia_ManSwi_t, factor int32) {
	shift := Gia_ManPoNum(p.pAig) - Gia_ManPiNum(p.pAig)
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		if i < Gia_ManPiNum(p.pAig) {
			Gia_ManSwiSimInfoRandomShift(p, Gia_SwiDataCi(p, i), factor)
		} else {
			Gia_ManSwiSimInfoCopyShift(p, Gia_SwiDataCi(p, i), Gia_SwiDataCo(p, shift+i))
		}
	}
}
func Gia_ManSwiSimInfoCountOnes(p *Gia_ManSwi_t, slot int32) int32 {
	var count int32
	for _, v := range Gia_SwiData(p, slot) {
		count += int32(bits.OnesCount32(v))
	}
	return count
}
func Gia_ManSwiSimInfoCountTrans(p *Gia_ManSwi_t, slot int32) int32 {
	var count int32
	for _, v := range Gia_SwiData(p, slot) {
		count += 2 * int32(bits.OnesCount32((v^(v>>16))&0xffff))
	}
	return count
}
func Gia_ManSwiSimulateRound(p *Gia_ManSwi_t, count int32) {
	if p.pAig.nFront <= 0 || Gia_ManConst0(p.pAig).Value != 0 {
		panic("gia: invalid switching frontier")
	}
	Gia_ManSwiSimInfoZero(p, Gia_SwiData(p, 0))
	for i, obj := range p.pAig.pObjs[1:] {
		switch {
		case Gia_ObjIsAndOrConst0(obj) != 0:
			Gia_ManSwiSimulateNode(p, obj)
		case Gia_ObjIsCo(obj) != 0:
			Gia_ManSwiSimulateCo(p, Gia_ObjCioId(obj), obj)
		default:
			Gia_ManSwiSimulateCi(p, obj, Gia_ObjCioId(obj))
		}
		if count != 0 && Gia_ObjIsCo(obj) == 0 {
			if p.pPars.fProbTrans != 0 {
				p.pData1[i+1] += Gia_ManSwiSimInfoCountTrans(p, Gia_ObjValue(obj))
			} else {
				p.pData1[i+1] += Gia_ManSwiSimInfoCountOnes(p, Gia_ObjValue(obj))
			}
		}
	}
}
