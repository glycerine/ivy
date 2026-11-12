package gia

import "github.com/glycerine/abcgo/misc/util"

// Gia_ManTer_t simulates two-bit ternary values over a frontier graph.
type Gia_ManTer_t struct {
	pAig                               *Gia_Man_t
	nIters, nStateWords                int32
	vStates, vFlops                    *Vec_Ptr_t
	vRetired                           *Vec_Int_t
	pRetired                           []byte
	pCount0, pCountX                   []int32
	nBins                              int32
	pBins                              [][][]uint32
	pDataSim, pDataSimCis, pDataSimCos []uint32
}

func Gia_ManTerSimInfoGet(info []uint32, i int32) int32 {
	return int32(info[i>>4] >> uint((i&15)*2) & 3)
}
func Gia_ManTerSimInfoSet(info []uint32, i, value int32) {
	if value < GIA_ZER || value > GIA_UND {
		panic("gia: invalid ternary value")
	}
	shift := uint((i & 15) * 2)
	info[i>>4] = (info[i>>4] &^ (uint32(3) << shift)) | uint32(value)<<shift
}
func giaTerWords(n int32) int32 { return (2*n + 31) / 32 }
func Gia_ManTerCreate(graph *Gia_Man_t) *Gia_ManTer_t {
	front := Gia_ManFront(graph)
	regs := Gia_ManRegNum(graph)
	bins := int32(util.Abc_PrimeCudd(500))
	return &Gia_ManTer_t{pAig: front, nIters: 300, nStateWords: giaTerWords(regs), vStates: Vec_PtrAlloc(1000), vRetired: Vec_IntAlloc(100), pRetired: make([]byte, regs), pCount0: make([]int32, regs), pCountX: make([]int32, regs), nBins: bins, pBins: make([][][]uint32, bins), pDataSim: make([]uint32, giaTerWords(front.nFront)), pDataSimCis: make([]uint32, giaTerWords(Gia_ManCiNum(front))), pDataSimCos: make([]uint32, giaTerWords(Gia_ManCoNum(front)))}
}
func Gia_ManTerStatesFree(states *Vec_Ptr_t) { Vec_PtrFree(states) }
func Gia_ManTerDelete(p *Gia_ManTer_t)       { Gia_ManStop(p.pAig); *p = Gia_ManTer_t{} }
func Gia_ManTerSimulateCi(p *Gia_ManTer_t, obj *Gia_Obj_t, index int32) {
	Gia_ManTerSimInfoSet(p.pDataSim, Gia_ObjValue(obj), Gia_ManTerSimInfoGet(p.pDataSimCis, index))
}
func Gia_ManTerSimulateCo(p *Gia_ManTer_t, index int32, obj *Gia_Obj_t) {
	value := Gia_ManTerSimInfoGet(p.pDataSim, Gia_ObjDiff0(obj))
	Gia_ManTerSimInfoSet(p.pDataSimCos, index, Gia_XsimNotCond(value, Gia_ObjFaninC0(obj)))
}
func Gia_ManTerSimulateNode(p *Gia_ManTer_t, obj *Gia_Obj_t) {
	a, b := Gia_ManTerSimInfoGet(p.pDataSim, Gia_ObjDiff0(obj)), Gia_ManTerSimInfoGet(p.pDataSim, Gia_ObjDiff1(obj))
	Gia_ManTerSimInfoSet(p.pDataSim, Gia_ObjValue(obj), Gia_XsimAndCond(a, Gia_ObjFaninC0(obj), b, Gia_ObjFaninC1(obj)))
}
func Gia_ManTerSimInfoInit(p *Gia_ManTer_t) {
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		value := int32(GIA_ZER)
		if i < Gia_ManPiNum(p.pAig) {
			value = GIA_UND
		}
		Gia_ManTerSimInfoSet(p.pDataSimCis, i, value)
	}
}
func Gia_ManTerSimInfoTransfer(p *Gia_ManTer_t) {
	for i := int32(0); i < Gia_ManCiNum(p.pAig); i++ {
		value := int32(GIA_UND)
		if i >= Gia_ManPiNum(p.pAig) {
			value = Gia_ManTerSimInfoGet(p.pDataSimCos, Gia_ManCoNum(p.pAig)-Gia_ManCiNum(p.pAig)+i)
		}
		Gia_ManTerSimInfoSet(p.pDataSimCis, i, value)
	}
}
func Gia_ManTerSimulateRound(p *Gia_ManTer_t) {
	if p.pAig.nFront <= 0 || Gia_ManConst0(p.pAig).Value != 0 {
		panic("gia: invalid ternary frontier")
	}
	Gia_ManTerSimInfoSet(p.pDataSim, 0, GIA_ZER)
	var ci, co int32
	for _, obj := range p.pAig.pObjs[1:] {
		switch {
		case Gia_ObjIsAndOrConst0(obj) != 0:
			Gia_ManTerSimulateNode(p, obj)
		case Gia_ObjIsCi(obj) != 0:
			Gia_ManTerSimulateCi(p, obj, ci)
			ci++
		default:
			Gia_ManTerSimulateCo(p, co, obj)
			co++
		}
	}
	if ci != Gia_ManCiNum(p.pAig) || co != Gia_ManCoNum(p.pAig) {
		panic("gia: invalid ternary interface")
	}
}
