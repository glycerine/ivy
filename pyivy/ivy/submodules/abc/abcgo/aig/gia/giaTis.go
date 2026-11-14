package gia

import "fmt"

// Gia_ManTisDupMuxes replaces eligible AND decompositions with real MUXes.
func Gia_ManTisDupMuxes(src *Gia_Man_t) *Gia_Man_t {
	if src.PMuxes != nil {
		panic("gia: TIS expects an ordinary AIG")
	}
	src.Refs = nil
	Gia_ManCreateRefs(src)
	dst := Gia_ManStart(5000)
	dst.pName, dst.PSpec = src.pName, src.PSpec
	dst.EnableMuxes()
	Gia_ManConst0(src).Value = 0
	for i := int32(0); i < Gia_ManCiNum(src); i++ {
		Gia_ManCi(src, i).Value = uint32(Gia_ManAppendCi(dst))
	}
	Gia_ManHashStart(dst)
	for _, obj := range src.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		var lit int32
		if Gia_ObjIsMuxType(obj) == 0 || Gia_ObjRefNum(src, Gia_ObjFanin0(obj)) > 1 && Gia_ObjRefNum(src, Gia_ObjFanin1(obj)) > 1 {
			lit = Gia_ManHashAnd(dst, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj))
		} else {
			var a, b *Gia_Obj_t
			if Gia_ObjRecognizeExor(obj, &a, &b) != 0 {
				lit = Gia_ManHashXorReal(dst, Gia_ObjLitCopy(src, Gia_ObjToLit(src, a)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, b)))
			} else {
				control := Gia_ObjRecognizeMux(obj, &a, &b)
				lit = Gia_ManHashMuxReal(dst, Gia_ObjLitCopy(src, Gia_ObjToLit(src, control)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, a)), Gia_ObjLitCopy(src, Gia_ObjToLit(src, b)))
			}
		}
		obj.Value = uint32(lit)
	}
	Gia_ManHashStop(dst)
	for i := int32(0); i < Gia_ManCoNum(src); i++ {
		obj := Gia_ManCo(src, i)
		obj.Value = uint32(Gia_ManAppendCo(dst, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(dst, Gia_ManRegNum(src))
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	return result
}
func Gia_ManTisCollectMffc_rec(p *Gia_Man_t, id int32, nodes, leaves *Vec_Int_t) {
	if Gia_ObjIsTravIdCurrentId(p, id) != 0 {
		return
	}
	Gia_ObjSetTravIdCurrentId(p, id)
	obj := Gia_ManObj(p, id)
	if Gia_ObjRefNumId(p, id) > 1 || Gia_ObjIsCi(obj) != 0 {
		Vec_IntPush(leaves, id)
		return
	}
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid TIS cone node")
	}
	Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId0(obj, id), nodes, leaves)
	Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId1(obj, id), nodes, leaves)
	if Gia_ObjIsMuxId(p, id) != 0 {
		Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId2(p, id), nodes, leaves)
	}
	Vec_IntPush(nodes, id)
}
func Gia_ManTisCollectMffc(p *Gia_Man_t, id int32, nodes, leaves *Vec_Int_t) {
	obj := Gia_ManObj(p, id)
	if Gia_ObjIsAnd(obj) == 0 {
		panic("gia: invalid TIS cone root")
	}
	Vec_IntClear(nodes)
	Vec_IntClear(leaves)
	Gia_ManIncrementTravId(p)
	Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId0(obj, id), nodes, leaves)
	Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId1(obj, id), nodes, leaves)
	if Gia_ObjIsMuxId(p, id) != 0 {
		Gia_ManTisCollectMffc_rec(p, Gia_ObjFaninId2(p, id), nodes, leaves)
	}
	Vec_IntPush(nodes, id)
}
func Gia_ManTisPrintMffc(p *Gia_Man_t, id int32, nodes, leaves *Vec_Int_t) {
	fmt.Printf("MFFC %d has %d nodes and %d leaves:\n", id, Vec_IntSize(nodes), Vec_IntSize(leaves))
	for i := Vec_IntSize(nodes) - 1; i >= 0; i-- {
		fmt.Printf("Node %2d : ", Vec_IntSize(nodes)-1-i)
		Gia_ObjPrint(p, Gia_ManObj(p, Vec_IntEntry(nodes, i)))
	}
	for i, id := range Vec_IntArray(leaves) {
		fmt.Printf("Leaf %2d : ", i)
		Gia_ObjPrint(p, Gia_ManObj(p, id))
	}
	fmt.Println()
}
func Gia_ManTisTest(initial *Gia_Man_t) {
	nodes, leaves := Vec_IntAlloc(10), Vec_IntAlloc(10)
	p := Gia_ManTisDupMuxes(initial)
	defer Gia_ManStop(p)
	Gia_ManCreateRefs(p)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 && Gia_ObjRefNumId(p, obj.id) != 1 {
			Gia_ManTisCollectMffc(p, obj.id, nodes, leaves)
			Gia_ManTisPrintMffc(p, obj.id, nodes, leaves)
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		id := Gia_ObjFaninId0p(p, Gia_ManCo(p, i))
		if Gia_ObjRefNumId(p, id) <= 1 {
			Gia_ManTisCollectMffc(p, id, nodes, leaves)
			Gia_ManTisPrintMffc(p, id, nodes, leaves)
		}
	}
}
