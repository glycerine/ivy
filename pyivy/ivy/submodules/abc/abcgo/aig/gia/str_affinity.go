package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
)

func Str_Delay2(a, b, size int32) int32 {
	d := max(a>>4, b>>4)
	n := int32(1)
	if d == a>>4 {
		n = a & 15
	}
	if d == b>>4 {
		n += b & 15
	} else {
		n++
	}
	if n > size {
		n = 18
	}
	return d<<4 + n
}
func Str_Delay3(a, b, c, size int32) int32 {
	d := max(a>>4, b>>4, c>>4)
	n := int32(0)
	for _, v := range []int32{a, b, c} {
		if d == v>>4 {
			n += v & 15
		} else {
			n++
		}
	}
	if n > size {
		n = 19
	}
	return d<<4 + n
}
func Str_ObjDelay(p *Gia_Man_t, id, size int32, delays *Vec_Int_t) int32 {
	delay := vec.Vec_IntEntry(delays, int(id))
	if delay != 0 {
		return delay
	}
	obj := Gia_ManObj(p, id)
	a := vec.Vec_IntEntry(delays, int(Gia_ObjFaninId0(obj, id)))
	b := vec.Vec_IntEntry(delays, int(Gia_ObjFaninId1(obj, id)))
	if Gia_ObjIsMuxId(p, id) != 0 {
		c := vec.Vec_IntEntry(delays, int(Gia_ObjFaninId2(p, id)))
		delay = Str_Delay3(a, b, c, size)
	} else {
		delay = Str_Delay2(a, b, size)
	}
	vec.Vec_IntWriteEntry(delays, int(id), delay)
	return delay
}
func transpose64(a []uint64) {
	if len(a) < 64 {
		panic("gia: short transpose matrix")
	}
	mask := uint64(0x00000000ffffffff)
	for j := 32; j != 0; {
		for k := 0; k < 64; k = (k + j + 1) &^ j {
			v := (a[k] ^ (a[k+j] >> uint(j))) & mask
			a[k] ^= v
			a[k+j] ^= v << uint(j)
		}
		j >>= 1
		mask ^= mask << uint(j)
	}
}
func Str_ManNum(p *Gia_Man_t, id int32) int32   { return vec.Vec_IntEntry(&p.vCopies, int(id)) }
func Str_ManSetNum(p *Gia_Man_t, id, num int32) { vec.Vec_IntWriteEntry(&p.vCopies, int(id), num) }
func Str_CountBits(v uint64) int32              { return int32(bits.OnesCount64(v)) }
func Str_ManVectorAffinity(p *Gia_Man_t, super, delays *Vec_Int_t, matrix []uint64, limit int32) int32 {
	size := vec.Vec_IntSize(super)
	if size <= 2 || len(matrix) < 256 || limit < int32(size) || limit > 256 {
		panic("gia: invalid affinity dimensions")
	}
	if size > 64 {
		clear(matrix[:64])
		return 0
	}
	Gia_ManIncrementTravId(p)
	for i, lit := range vec.Vec_IntArray(super) {
		id := lit >> 1
		Gia_ObjSetTravIdCurrentId(p, id)
		Str_ManSetNum(p, id, int32(i))
		matrix[i] = uint64(1) << uint(63-i)
	}
	previous, levels := size, 1
	for i := 0; i < vec.Vec_IntSize(super); i++ {
		id := vec.Vec_IntEntry(super, i) >> 1
		obj := Gia_ManObj(p, id)
		var fans []int32
		if Gia_ObjIsAnd(obj) != 0 {
			fans = append(fans, Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id))
		}
		if Gia_ObjIsMux(p, obj) != 0 {
			fans = append(fans, Gia_ObjFaninId2(p, id))
		}
		stop := false
		for _, fan := range fans {
			if Gia_ObjIsTravIdCurrentId(p, fan) == 0 {
				if vec.Vec_IntSize(super) == int(limit) {
					stop = true
					break
				}
				Gia_ObjSetTravIdCurrentId(p, fan)
				matrix[vec.Vec_IntSize(super)] = 0
				Str_ManSetNum(p, fan, int32(vec.Vec_IntSize(super)))
				vec.Vec_IntPush(super, 2*fan)
			}
			matrix[Str_ManNum(p, fan)] |= matrix[i]
		}
		if stop {
			break
		}
		if previous == i {
			previous = vec.Vec_IntSize(super)
			levels++
		}
		if levels == 8 {
			break
		}
	}
	mask := ^uint64(0) << uint(64-size)
	count := 0
	for i := 0; i < vec.Vec_IntSize(super); i++ {
		v := matrix[i]
		if v == 0 {
			panic("gia: empty affinity vector")
		}
		if v&(v-1) == 0 || v == mask {
			continue
		}
		matrix[count] = v
		count++
		if count == 64 {
			break
		}
	}
	clear(matrix[count:64])
	vec.Vec_IntShrink(super, size)
	if count != 0 {
		transpose64(matrix)
	}
	return int32(count)
}
