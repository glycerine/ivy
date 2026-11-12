package gia

import "github.com/glycerine/abcgo/misc/vec"

func Spl_ManToWecMapping(p *Gia_Man_t) *Vec_Wec_t {
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: missing LUT mapping")
	}
	result := vec.Vec_WecStart(int(p.nObjs))
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) != 0 {
			for _, fan := range Gia_ObjLutFanins(p, id) {
				vec.Vec_WecPush(result, int(id), fan)
			}
		}
	}
	return result
}
func Spl_ManFromWecMapping(p *Gia_Man_t, mapping *Vec_Wec_t) *Vec_Int_t {
	if Gia_ManHasMapping2(p) == 0 {
		panic("gia: missing internal LUT mapping")
	}
	result := vec.Vec_IntStart(int(p.nObjs))
	for id := 0; id < vec.Vec_WecSize(mapping); id++ {
		fans := vec.Vec_IntArray(vec.Vec_WecEntry(mapping, id))
		if len(fans) == 0 {
			continue
		}
		vec.Vec_IntWriteEntry(result, id, int32(vec.Vec_IntSize(result)))
		vec.Vec_IntPush(result, int32(len(fans)))
		for _, fan := range fans {
			vec.Vec_IntPush(result, fan)
		}
		vec.Vec_IntPush(result, int32(id))
	}
	return result
}
