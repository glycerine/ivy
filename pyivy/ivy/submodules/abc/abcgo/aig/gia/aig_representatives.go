package gia

import "github.com/glycerine/abcgo/aig/aig"

func Gia_ManReprToAigRepr(pAig *aig.Graph, pGia *Gia_Man_t) {
	for id := 0; id < int(aig.Aig_ManObjNumMax(pAig)); id++ {
		edge := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(pAig, edge) {
			continue
		}
		copied := aig.Aig_ObjCopy(pAig, edge)
		if copied == aig.InvalidEdge {
			continue
		}
		pGia.pObjs[copied.NodeID()].Value = uint32(id)
	}
	aig.Aig_ManReprStart(pAig, aig.Aig_ManObjNumMax(pAig))
	for id := int32(0); id < pGia.nObjs; id++ {
		repr := Gia_ObjReprObj(pGia, id)
		if repr == nil {
			continue
		}
		representative, node := aig.Edge(repr.Value)<<1, aig.Edge(pGia.pObjs[id].Value)<<1
		if representative != node {
			aig.Aig_ObjCreateRepr(pAig, representative, node)
		}
	}
}
func Gia_ManReprToAigRepr2(pAig *aig.Graph, pGia *Gia_Man_t) {
	aig.Aig_ManReprStart(pAig, aig.Aig_ManObjNumMax(pAig))
	for id := int32(0); id < pGia.nObjs; id++ {
		repr := Gia_ObjReprObj(pGia, id)
		if repr == nil {
			continue
		}
		representative, node := aig.Aig_Regular(aig.Edge(repr.Value)), aig.Aig_Regular(aig.Edge(pGia.pObjs[id].Value))
		if representative != node {
			aig.Aig_ObjCreateRepr(pAig, representative, node)
		}
	}
}
func giaPrepareReprImport(p *Gia_Man_t) {
	if p.pReprs != nil {
		panic("gia: representative storage already initialized")
	}
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
}
func giaRecordReprImport(p *Gia_Man_t, node, representative int32) {
	giaMergeRepr(p, node, representative)
}

func Gia_ManReprFromAigRepr(pAig *aig.Graph, pGia *Gia_Man_t) {
	giaPrepareReprImport(pGia)
	inverse := make(map[int]int32, pGia.nObjs)
	for _, obj := range pGia.pObjs {
		if Gia_ObjIsCo(obj) == 0 {
			id := aig.Edge(obj.Value).NodeID()
			if first, ok := inverse[id]; ok {
				giaRecordReprImport(pGia, obj.id, first)
			} else {
				inverse[id] = obj.id
			}
		}
	}
	for source, target := range inverse {
		repr := aig.Aig_ObjRepr(pAig, aig.Edge(source)<<1)
		if repr == aig.InvalidEdge {
			continue
		}
		representative, ok := inverse[repr.NodeID()]
		if !ok {
			panic("gia: AIG representative has no GIA copy")
		}
		giaRecordReprImport(pGia, target, representative)
	}
	giaFlattenReprs(pGia)
	pGia.pNexts = Gia_ManDeriveNexts(pGia)
}
func Gia_ManReprFromAigRepr2(pAig *aig.Graph, pGia *Gia_Man_t) {
	giaPrepareReprImport(pGia)
	for id := 0; id < int(aig.Aig_ManObjNumMax(pAig)); id++ {
		edge := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(pAig, edge) {
			continue
		}
		repr := aig.Aig_ObjRepr(pAig, edge)
		if repr == aig.InvalidEdge {
			continue
		}
		copied, reprCopy := aig.Aig_ObjCopy(pAig, edge), aig.Aig_ObjCopy(pAig, repr)
		if copied == aig.InvalidEdge || reprCopy == aig.InvalidEdge {
			panic("gia: missing AIG/GIA copy mapping")
		}
		giaRecordReprImport(pGia, int32(copied>>1), int32(reprCopy>>1))
	}
	giaFlattenReprs(pGia)
	pGia.pNexts = Gia_ManDeriveNexts(pGia)
}
