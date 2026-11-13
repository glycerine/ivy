package gia

import "github.com/glycerine/abcgo/aig/aig"

func Gia_ManReprToAigRepr(pAig *aig.Graph, pGia *Gia_Man_t) {
	if aig.Aig_ManHasReprs(pAig) {
		panic("gia: AIG representative storage already initialized")
	}
	if pGia.pReprs == nil {
		panic("gia: missing GIA representative storage")
	}
	for id := 0; id < int(aig.Aig_ManObjNumMax(pAig)); id++ {
		edge := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(pAig, edge) {
			continue
		}
		copied := aig.Aig_ObjCopy(pAig, edge)
		if id != 0 && copied&1 != 0 {
			panic("gia: complemented AIG/GIA copy")
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
		aig.Aig_ObjCreateRepr(pAig, representative, node)
	}
}
func Gia_ManReprToAigRepr2(pAig *aig.Graph, pGia *Gia_Man_t) {
	if aig.Aig_ManHasReprs(pAig) {
		panic("gia: AIG representative storage already initialized")
	}
	if pGia.pReprs == nil {
		panic("gia: missing GIA representative storage")
	}
	aig.Aig_ManReprStart(pAig, aig.Aig_ManObjNumMax(pAig))
	for id := int32(0); id < pGia.nObjs; id++ {
		repr := Gia_ObjReprObj(pGia, id)
		if repr == nil {
			continue
		}
		representative, node := aig.Aig_Regular(aig.Edge(repr.Value)), aig.Aig_Regular(aig.Edge(pGia.pObjs[id].Value))
		aig.Aig_ObjCreateRepr(pAig, representative, node)
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
func Gia_ManReprFromAigRepr(pAig *aig.Graph, pGia *Gia_Man_t) {
	if !aig.Aig_ManHasReprs(pAig) {
		panic("gia: missing AIG representative storage")
	}
	if pGia.pReprs != nil {
		panic("gia: representative storage already initialized")
	}
	if Gia_ManObjNum(pGia)-Gia_ManCoNum(pGia) != aig.Aig_ManObjNum(pAig)-aig.Aig_ManCoNum(pAig) {
		panic("gia: AIG/GIA object counts differ")
	}
	giaPrepareReprImport(pGia)
	for _, obj := range pGia.pObjs {
		if Gia_ObjIsCo(obj) != 0 {
			continue
		}
		value := Gia_ObjValue(obj)
		if obj.id != 0 && value&1 != 0 {
			panic("gia: complemented GIA/AIG copy")
		}
		aig.Aig_ObjSetCopy(pAig, aig.Edge(uint32(value)>>1)<<1, aig.Edge(obj.id))
	}
	for id := int32(0); id < aig.Aig_ManObjNumMax(pAig); id++ {
		node := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(pAig, node) || aig.Aig_ObjIsCo(pAig, node) {
			continue
		}
		repr := aig.Aig_ObjRepr(pAig, node)
		if repr == aig.InvalidEdge {
			continue
		}
		mapped := aig.Aig_ObjCopy(pAig, node)
		representative := aig.Aig_ObjCopy(pAig, repr)
		if mapped == aig.InvalidEdge || representative == aig.InvalidEdge {
			panic("gia: missing GIA object mapping")
		}
		Gia_ObjSetRepr(pGia, int32(mapped), int32(representative))
	}
	pGia.pNexts = Gia_ManDeriveNexts(pGia)
}
func Gia_ManReprFromAigRepr2(pAig *aig.Graph, pGia *Gia_Man_t) {
	if !aig.Aig_ManHasReprs(pAig) {
		panic("gia: missing AIG representative storage")
	}
	if pGia.pReprs != nil {
		panic("gia: representative storage already initialized")
	}
	if Gia_ManObjNum(pGia)-Gia_ManCoNum(pGia) != aig.Aig_ManObjNum(pAig)-aig.Aig_ManCoNum(pAig) {
		panic("gia: AIG/GIA object counts differ")
	}
	giaPrepareReprImport(pGia)
	for id := int32(0); id < aig.Aig_ManObjNumMax(pAig); id++ {
		node := aig.Edge(id) << 1
		if aig.Aig_ObjIsNone(pAig, node) || aig.Aig_ObjIsCo(pAig, node) {
			continue
		}
		repr := aig.Aig_ObjRepr(pAig, node)
		if repr == aig.InvalidEdge {
			continue
		}
		copied, reprCopy := aig.Aig_ObjCopy(pAig, node), aig.Aig_ObjCopy(pAig, repr)
		if copied == aig.InvalidEdge || reprCopy == aig.InvalidEdge {
			panic("gia: missing AIG/GIA copy mapping")
		}
		Gia_ObjSetRepr(pGia, int32(copied>>1), int32(reprCopy>>1))
	}
	pGia.pNexts = Gia_ManDeriveNexts(pGia)
}
