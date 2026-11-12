package gia

import "math/bits"

// Gia_PtrAre_t preserves the 20-bit item, 11-bit page and mark fields.
type Gia_PtrAre_t struct{ nItem, nPage, fMark uint32 }
type Gia_StaAre_t struct {
	iPrev, iNext Gia_PtrAre_t
	pData        []uint32
}

func Gia_Int2Ptr(n uint32) Gia_PtrAre_t {
	return Gia_PtrAre_t{nItem: n & ((1 << 20) - 1), nPage: (n >> 20) & ((1 << 11) - 1), fMark: n >> 31}
}
func Gia_Ptr2Int(p Gia_PtrAre_t) uint32 { return p.nItem&((1<<20)-1) | (p.nPage&((1<<11)-1))<<20 }
func Gia_StaHasValue0(s *Gia_StaAre_t, reg int32) int32 {
	return giaBool(s.pData[(reg*2)>>5]&(1<<uint((reg*2)&31)) != 0)
}
func Gia_StaHasValue1(s *Gia_StaAre_t, reg int32) int32 {
	bit := reg*2 + 1
	return giaBool(s.pData[bit>>5]&(1<<uint(bit&31)) != 0)
}
func Gia_StaSetValue0(s *Gia_StaAre_t, reg int32) {
	bit := reg * 2
	s.pData[bit>>5] |= 1 << uint(bit&31)
}
func Gia_StaSetValue1(s *Gia_StaAre_t, reg int32) {
	bit := reg*2 + 1
	s.pData[bit>>5] |= 1 << uint(bit&31)
}
func Gia_StaSetUnused(s *Gia_StaAre_t)      { s.iPrev.fMark = 1 }
func Gia_StaIsUnused(s *Gia_StaAre_t) int32 { return int32(s.iPrev.fMark) }
func Gia_StaIsUsed(s *Gia_StaAre_t) int32   { return giaBool(s.iPrev.fMark == 0) }
func Gia_StaAreEqual(a, b *Gia_StaAre_t, nWords int32) int32 {
	for i := int32(0); i < nWords; i++ {
		if a.pData[i] != b.pData[i] {
			return 0
		}
	}
	return 1
}
func Gia_StaAreDisjoint(a, b *Gia_StaAre_t, nWords int32) int32 {
	for i := int32(0); i < nWords; i++ {
		x := a.pData[i] ^ b.pData[i]
		if (x>>1)&x&0x55555555 != 0 {
			return 1
		}
	}
	return 0
}
func Gia_StaAreContain(a, b *Gia_StaAre_t, nWords int32) int32 {
	for i := int32(0); i < nWords; i++ {
		if a.pData[i]|b.pData[i] != b.pData[i] {
			return 0
		}
	}
	return 1
}
func Gia_StaAreDashNum(a, b *Gia_StaAre_t, nWords int32) int32 {
	var n int32
	for i := int32(0); i < nWords; i++ {
		x, y := a.pData[i], b.pData[i]
		n += int32(bits.OnesCount32(^(x ^ (x >> 1)) & (y ^ (y >> 1)) & 0x55555555))
	}
	return n
}
func Gia_StaAreSharpVar(a, b *Gia_StaAre_t, nWords int32) int32 {
	result := int32(-1)
	for i := int32(0); i < nWords; i++ {
		x, y := a.pData[i], b.pData[i]
		word := ^(x ^ (x >> 1)) & (y ^ (y >> 1)) & 0x55555555
		if word == 0 {
			continue
		}
		if bits.OnesCount32(word) != 1 || result >= 0 {
			return -1
		}
		result = 16*i + int32(bits.TrailingZeros32(word))/2
	}
	return result
}
func Gia_StaAreDisjointVar(a, b *Gia_StaAre_t, nWords int32) int32 {
	result := int32(-1)
	for i := int32(0); i < nWords; i++ {
		x := a.pData[i] ^ b.pData[i]
		word := x & (x >> 1) & 0x55555555
		if word == 0 {
			continue
		}
		if bits.OnesCount32(word) != 1 || result >= 0 {
			return -1
		}
		result = 16*i + int32(bits.TrailingZeros32(word))/2
	}
	return result
}
func Gia_ManCountMintermsInCube(cube *Gia_StaAre_t, nVars int32, store []uint32) {
	if nVars < 0 || nVars > 30 {
		panic("gia: unsupported minterm count")
	}
	var mask uint32
	var dashes []int32
	for i := int32(0); i < nVars; i++ {
		if Gia_StaHasValue0(cube, i) != 0 {
			continue
		}
		if Gia_StaHasValue1(cube, i) != 0 {
			mask |= 1 << uint(i)
		} else {
			dashes = append(dashes, i)
		}
	}
	for assignment := uint32(0); assignment < 1<<uint(len(dashes)); assignment++ {
		mint := mask
		for i, v := range dashes {
			if assignment&(1<<uint(i)) != 0 {
				mint |= 1 << uint(v)
			}
		}
		store[mint>>5] |= 1 << uint(mint&31)
	}
}
