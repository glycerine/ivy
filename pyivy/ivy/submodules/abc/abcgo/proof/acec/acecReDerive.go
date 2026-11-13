package acec

// Convert hashed sum/carry cuts into adder records from acecRe.c.
import "github.com/glycerine/abcgo/misc/vec"

func Ree_ManDeriveAdds(h *vec.Hash_IntMan_t, data *Vec_Int_t, verbose int32) *Vec_Int_t {
	n := vec.Hash_IntManEntryNum(h)
	sums, carries := make([][]int32, n+1), make([][]int32, n+1)
	for i := 0; i+2 < vec.Vec_IntSize(data); i += 3 {
		id, key, truth := vec.Vec_IntEntry(data, i), vec.Vec_IntEntry(data, i+1), vec.Vec_IntEntry(data, i+2)
		if truth == 0x66 || truth == 0x99 || truth == 0x69 || truth == 0x96 {
			sums[key] = append(sums[key], id, truth)
		} else {
			carries[key] = append(carries[key], id, truth)
		}
	}
	result := vec.Vec_IntAlloc(1000)
	and := [8]int32{0x88, 0x44, 0x22, 0x11, 0x77, 0xbb, 0xdd, 0xee}
	maj := [8]int32{0xe8, 0xd4, 0xb2, 0x71, 0x8e, 0x4d, 0x2b, 0x17}
	for key := 1; key <= n; key++ {
		x, y := sums[key], carries[key]
		for i := 0; i < len(x); i += 2 {
			for j := 0; j < len(y); j += 2 {
				sign := int32(0)
				if x[i+1] == 0x99 || x[i+1] == 0x69 {
					sign = 8
				}
				idx := -1
				for k, v := range maj {
					if v == y[j+1] {
						idx = k
						break
					}
				}
				if idx < 0 {
					for k, v := range and {
						if v == y[j+1] {
							idx = k
							break
						}
					}
				}
				if idx < 0 {
					panic("acec: unknown carry truth table")
				}
				vec.Vec_IntPushThree(result, vec.Hash_IntObjData0(h, int32(key)), vec.Hash_IntObjData1(h, int32(key)), vec.Hash_IntObjData2(h, int32(key)))
				vec.Vec_IntPushThree(result, x[i], y[j], sign|int32(idx))
			}
		}
	}
	return result
}
