package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math"
	"math/bits"
)

const (
	KF_LEAF_MAX  = 16
	KF_CUT_MAX   = 32
	KF_PROC_MAX  = 32
	KF_WORD_MAX  = 1 << (KF_LEAF_MAX - 6)
	KF_LOG_TABLE = 8
	KF_ADD_ON1   = 2
	KF_ADD_ON2   = 4
)

type Kf_Cut_t struct {
	Sign                  word
	Polar, Delay          int32
	Area                  float32
	iFunc, iNext, nLeaves int32
	pLeaves               [KF_LEAF_MAX]int32
}

func Kf_SetLoadCuts(cuts []Kf_Cut_t, packed []int32) int32 {
	n := packed[0]
	offset := KF_ADD_ON1
	for i := int32(0); i < n; i++ {
		size := packed[offset]
		if size < 0 || size > KF_LEAF_MAX || offset+int(size)+KF_ADD_ON2 > len(packed) {
			panic("KF malformed cut")
		}
		cut := &cuts[i]
		cut.Sign = 0
		cut.Polar = 0
		cut.nLeaves = size
		cut.iFunc = packed[offset+int(size)+1]
		cut.Delay = packed[offset+int(size)+2]
		cut.Area = math.Float32frombits(uint32(packed[offset+int(size)+3]))
		for k := int32(0); k < size; k++ {
			lit := packed[offset+int(k)+1]
			cut.pLeaves[k] = lit >> 1
			cut.Sign |= word(1) << uint(cut.pLeaves[k]&63)
			cut.Polar |= (lit & 1) << uint(k)
		}
		offset += int(size) + KF_ADD_ON2
	}
	return n
}
func Kf_ManStoreStart(v *Vec_Int_t, n int32) {
	vec.Vec_IntClear(v)
	vec.Vec_IntPush(v, n)
	vec.Vec_IntPush(v, -1)
}
func Kf_ManStoreAddUnit(v *Vec_Int_t, id, time int32, area float32) {
	vec.Vec_IntWriteEntry(v, 0, vec.Vec_IntEntry(v, 0)+1)
	for _, x := range []int32{1, id << 1, 2, time, int32(math.Float32bits(area))} {
		vec.Vec_IntPush(v, x)
	}
}
func Kf_ManSaveResults(cuts []*Kf_Cut_t, n int32, best *Kf_Cut_t, v *Vec_Int_t) {
	if n <= 0 || n >= KF_CUT_MAX {
		panic("KF invalid cut count")
	}
	Kf_ManStoreStart(v, n)
	for _, cut := range cuts[:n] {
		if cut == best {
			vec.Vec_IntWriteEntry(v, 1, int32(vec.Vec_IntSize(v)))
		}
		vec.Vec_IntPush(v, cut.nLeaves)
		for _, leaf := range cut.pLeaves[:cut.nLeaves] {
			vec.Vec_IntPush(v, leaf<<1)
		}
		vec.Vec_IntPush(v, cut.iFunc)
		vec.Vec_IntPush(v, cut.Delay)
		vec.Vec_IntPush(v, int32(math.Float32bits(cut.Area)))
	}
	if vec.Vec_IntEntry(v, 1) <= 0 {
		panic("KF best cut missing")
	}
}
func Kf_CutCompare(a, b *Kf_Cut_t, area int32) int32 {
	cmp := func(x, y float32) int32 {
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
		return 0
	}
	intCmp := func(x, y int32) int32 {
		if x < y {
			return -1
		}
		if x > y {
			return 1
		}
		return 0
	}
	if area != 0 {
		if n := cmp(a.Area, b.Area); n != 0 {
			return n
		}
		if n := intCmp(a.Delay, b.Delay); n != 0 {
			return n
		}
		return intCmp(a.nLeaves, b.nLeaves)
	}
	if n := intCmp(a.Delay, b.Delay); n != 0 {
		return n
	}
	if n := intCmp(a.nLeaves, b.nLeaves); n != 0 {
		return n
	}
	return cmp(a.Area, b.Area)
}
func Kf_SetCutGetSign(c *Kf_Cut_t) word {
	var sign word
	for _, id := range c.pLeaves[:c.nLeaves] {
		sign |= word(1) << uint(id&63)
	}
	return sign
}
func Kf_SetCountBits(x word) int32 { return int32(bits.OnesCount64(x)) }
func Kf_SetCutIsContainedSimple(base, cut *Kf_Cut_t) bool {
	for _, v := range cut.pLeaves[:cut.nLeaves] {
		found := false
		for _, b := range base.pLeaves[:base.nLeaves] {
			if b == v {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func Kf_SetCutIsContainedOrder(base, cut *Kf_Cut_t) bool {
	if base.nLeaves < cut.nLeaves {
		return false
	}
	k := int32(0)
	for _, v := range base.pLeaves[:base.nLeaves] {
		if k == cut.nLeaves {
			return true
		}
		if v > cut.pLeaves[k] {
			return false
		}
		if v == cut.pLeaves[k] {
			k++
		}
	}
	return k == cut.nLeaves
}
func Kf_SetMergeSimpleOne(a, b, out *Kf_Cut_t, limit int32) int32 {
	n := a.nLeaves
	for _, v := range b.pLeaves[:b.nLeaves] {
		found := false
		for _, x := range a.pLeaves[:a.nLeaves] {
			if x == v {
				found = true
				break
			}
		}
		if !found {
			if n == limit {
				return 0
			}
			out.pLeaves[n] = v
			n++
		}
	}
	copy(out.pLeaves[:], a.pLeaves[:a.nLeaves])
	out.nLeaves = n
	return 1
}
func Kf_SetMergeOrderOne(a, b, out *Kf_Cut_t, limit int32) int32 {
	i, j, n := int32(0), int32(0), int32(0)
	for i < a.nLeaves || j < b.nLeaves {
		if n == limit {
			return 0
		}
		var v int32
		if j >= b.nLeaves || (i < a.nLeaves && a.pLeaves[i] < b.pLeaves[j]) {
			v = a.pLeaves[i]
			i++
		} else if i >= a.nLeaves || b.pLeaves[j] < a.pLeaves[i] {
			v = b.pLeaves[j]
			j++
		} else {
			v = a.pLeaves[i]
			i++
			j++
		}
		out.pLeaves[n] = v
		n++
	}
	out.nLeaves = n
	return 1
}
