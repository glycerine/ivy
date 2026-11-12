package gia

import "math/bits"

// Gia_ManOneHot produces violation outputs: all are false exactly when one
// of the non-skipped inputs is asserted.
func Gia_ManOneHot(skips, vars int32) *Gia_Man_t {
	if skips < 0 || vars < 0 || vars > 1<<28 {
		panic("gia: invalid one-hot size")
	}
	width := 1
	if vars > 1 {
		width = 1 << bits.Len32(uint32(vars-1))
	}
	temp := make([]int32, width)
	out := Gia_ManStart(max(1, skips+4*vars+1))
	out.pName = "onehot"
	for i := int32(0); i < skips; i++ {
		Gia_ManAppendCi(out)
	}
	for i := int32(0); i < vars; i++ {
		temp[i] = Gia_ManAppendCi(out)
	}
	Gia_ManHashStart(out)
	for shift := 1; shift < width; shift *= 2 {
		for i := 0; i < width; i += 2 * shift {
			lit := Gia_ManHashAnd(out, temp[i], temp[i+shift])
			if lit != 0 {
				Gia_ManAppendCo(out, lit)
			}
			temp[i] = Gia_ManHashOr(out, temp[i], temp[i+shift])
		}
	}
	Gia_ManHashStop(out)
	Gia_ManAppendCo(out, temp[0]^1)
	return out
}

func Gia_ManDupOneHot(src *Gia_Man_t) *Gia_Man_t {
	skips, variables := int32(0), Gia_ManCiNum(src)
	if Gia_ManRegNum(src) > 0 {
		skips, variables = Gia_ManPiNum(src), Gia_ManRegNum(src)
	}
	constraint := Gia_ManOneHot(skips, variables)
	defer Gia_ManStop(constraint)
	out := giaQuantStart(src)
	giaComposeNodes(out, src, true)
	for i := int32(0); i < Gia_ManPoNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManPo(src, i))
	}
	Gia_ManDupAppendShare(out, constraint)
	for i := int32(0); i < Gia_ManRegNum(src); i++ {
		giaDupMappedOutput(out, src, Gia_ManRi(src, i))
	}
	out.nConstrs = src.nConstrs + Gia_ManPoNum(constraint)
	Gia_ManSetRegNum(out, Gia_ManRegNum(src))
	Gia_ManHashStop(out)
	return out
}
