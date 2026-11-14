package gia

import "github.com/glycerine/abcgo/misc/vec"

// Seg_ManCountIntEdges enumerates mapped LUT edges between internal nodes.
func Seg_ManCountIntEdges(p *Gia_Man_t, polars, skip *Vec_Int_t, maxFanouts int32) *Vec_Int_t {
	if Gia_ManHasMapping(p) == 0 {
		panic("gia: SAT edge assignment requires mapping")
	}
	edges := vec.Vec_IntAlloc(1000)
	vec.Vec_IntClear(polars)
	vec.Vec_IntClear(skip)
	if maxFanouts != 0 {
		Gia_ManSetLutRefs(p)
		defer func() { p.pLutRefs = nil }()
	}
	for id := int32(1); id < p.nObjs; id++ {
		if Gia_ObjIsLut(p, id) == 0 {
			continue
		}
		for _, fan := range Gia_ObjLutFanins(p, id) {
			if Gia_ObjIsAnd(Gia_ManObj(p, fan)) == 0 {
				continue
			}
			index := int32(vec.Vec_IntSize(edges) / 2)
			if p.VEdge1 != nil && Gia_ObjCheckEdge(p, fan, id) != 0 {
				vec.Vec_IntPush(polars, index)
			}
			if maxFanouts != 0 && Gia_ObjLutRefNumId(p, fan) >= maxFanouts {
				vec.Vec_IntPush(skip, index)
			}
			vec.Vec_IntPushTwo(edges, fan, id)
		}
	}
	return edges
}
func Seg_ManCollectObjEdges(edges *Vec_Int_t, nObjs int32) *Vec_Wec_t {
	data := vec.Vec_IntArray(edges)
	if nObjs < 0 || len(data)%2 != 0 {
		panic("gia: invalid SAT edge list")
	}
	result := vec.Vec_WecStart(int(nObjs))
	for i := 0; i < len(data); i += 2 {
		if data[i] < 0 || data[i] >= nObjs || data[i+1] < 0 || data[i+1] >= nObjs {
			panic("gia: invalid SAT edge endpoint")
		}
		vec.Vec_WecPush(result, int(data[i]), int32(i/2))
		vec.Vec_WecPush(result, int(data[i+1]), int32(i/2))
	}
	return result
}
