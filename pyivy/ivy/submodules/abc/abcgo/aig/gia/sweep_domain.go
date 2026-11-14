package gia

// SweepCorrespondence provides sequential equivalence computation without
// requiring the GIA package to import the higher-level CEC package.
type SweepCorrespondence interface{ SweepCorrespondence(*Gia_Man_t) }

func Gia_ManSweepComputeOneDomainEquivs(p *Gia_Man_t, classes *Vec_Int_t, domain int32, parameters any, constant, equivalent, verbose int32) {
	if domain < 1 || domain > Vec_IntFindMax(classes) || p.PManTime != nil || Gia_ManRegNum(p) <= 0 || Vec_IntSize(classes) != int(Gia_ManRegNum(p)) {
		panic("gia: invalid sweep domain")
	}
	perm := Vec_IntAlloc(int(Gia_ManRegNum(p)))
	for i, class := range Vec_IntArray(classes) {
		if class != domain {
			Vec_IntPush(perm, int32(i))
		}
	}
	other := Vec_IntSize(perm)
	for i, class := range Vec_IntArray(classes) {
		if class == domain {
			Vec_IntPush(perm, int32(i))
		}
	}
	count := int32(Vec_IntSize(perm) - other)
	copy := Gia_ManDupPermFlop(p, perm)
	if Gia_ManObjNum(copy) != Gia_ManObjNum(p) {
		panic("gia: flop permutation changed object count")
	}
	copy.nRegs = count
	if parameters != nil {
		engine, ok := parameters.(SweepCorrespondence)
		if !ok {
			panic("gia: invalid sequential correspondence engine")
		}
		engine.SweepCorrespondence(copy)
	} else {
		Gia_ManSeqCleanupClasses(copy, constant, equivalent, verbose)
	}
	copy.nRegs = Gia_ManRegNum(p)
	for _, obj := range p.pObjs {
		if obj.Value&1 != 0 {
			panic("gia: complemented flop permutation copy")
		}
		Gia_ManObj(copy, int32(obj.Value>>1)).Value = uint32(obj.id << 1)
	}
	Gia_ManDupRemapEquiv(p, copy)
	Gia_ManStop(copy)
}
