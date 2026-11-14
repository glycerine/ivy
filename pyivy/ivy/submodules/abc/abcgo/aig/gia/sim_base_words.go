package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManSimPatAssignInputs(p *Gia_Man_t, n int32, sims, inputs *Vec_Wrd_t) {
	if vec.Vec_WrdSize(sims) != int(n*p.nObjs) || vec.Vec_WrdSize(inputs) != int(n*Gia_ManCiNum(p)) {
		panic("gia: invalid simulation dimensions")
	}
	dst, src := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(inputs)
	for i, id := range vec.Vec_IntArray(p.vCis) {
		copy(dst[int(id*n):int((id+1)*n)], src[i*int(n):(i+1)*int(n)])
	}
}
func Gia_ManSimPatSimAnd(p *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims *Vec_Wrd_t) {
	data := vec.Vec_WrdArray(sims)
	a, b := Gia_ObjFaninId0(obj, id), Gia_ObjFaninId1(obj, id)
	for w := int32(0); w < n; w++ {
		x, y := data[a*n+w], data[b*n+w]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		if obj.fCompl1 != 0 {
			y = ^y
		}
		if Gia_ObjIsXor(obj) != 0 {
			data[id*n+w] = x ^ y
		} else {
			data[id*n+w] = x & y
		}
	}
}
func Gia_ManSimPatSimPo(p *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims *Vec_Wrd_t) {
	data := vec.Vec_WrdArray(sims)
	a := Gia_ObjFaninId0(obj, id)
	for w := int32(0); w < n; w++ {
		x := data[a*n+w]
		if obj.fCompl0 != 0 {
			x = ^x
		}
		data[id*n+w] = x
	}
}
func Gia_ManSimPatSimNot(p *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims *Vec_Wrd_t) {
	data := vec.Vec_WrdArray(sims)
	for w := int32(0); w < n; w++ {
		data[id*n+w] = ^data[id*n+w]
	}
}
func giaPatternWords(p *Gia_Man_t) int32 {
	ci := Gia_ManCiNum(p)
	if ci <= 0 || p.VSimsPi == nil || vec.Vec_WrdSize(p.VSimsPi)%int(ci) != 0 {
		panic("gia: invalid input simulation dimensions")
	}
	return int32(vec.Vec_WrdSize(p.VSimsPi) / int(ci))
}
func Gia_ManSimPatSim(p *Gia_Man_t) *Vec_Wrd_t {
	n := giaPatternWords(p)
	sims := vec.Vec_WrdStart(int(p.nObjs * n))
	Gia_ManSimPatAssignInputs(p, n, sims, p.VSimsPi)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManSimPatSimAnd(p, obj.id, obj, n, sims)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManSimPatSimPo(p, id, Gia_ManObj(p, id), n, sims)
	}
	return sims
}
func Gia_ManSimPatResim(p *Gia_Man_t, ids *Vec_Int_t, n int32, sims *Vec_Wrd_t) {
	for i, id := range vec.Vec_IntArray(ids) {
		obj := Gia_ManObj(p, id)
		if i == 0 {
			Gia_ManSimPatSimNot(p, id, obj, n, sims)
		} else if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManSimPatSimAnd(p, id, obj, n, sims)
		} else if Gia_ObjIsCo(obj) == 0 {
			panic("gia: invalid resimulation object")
		}
	}
}
func Gia_ManSimPatWrite(filename string, inputs *Vec_Wrd_t, n int32) {
	vec.Vec_WrdDumpHex(filename, inputs, int(n), 0)
}
func Gia_ManSimPatValuesDerive(p *Gia_Man_t, n int32, sims, values *Vec_Wrd_t) {
	if vec.Vec_WrdSize(sims) != int(n*p.nObjs) || vec.Vec_WrdSize(values) != int(n*Gia_ManCoNum(p)) {
		panic("gia: invalid output simulation dimensions")
	}
	data, out := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(values)
	for i, id := range vec.Vec_IntArray(p.vCos) {
		copy(out[i*int(n):(i+1)*int(n)], data[int(id*n):int((id+1)*n)])
	}
}
func Gia_ManSimPatValues(p *Gia_Man_t) *Vec_Wrd_t {
	n := giaPatternWords(p)
	sims := Gia_ManSimPatSim(p)
	values := vec.Vec_WrdStart(int(n * Gia_ManCoNum(p)))
	Gia_ManSimPatValuesDerive(p, n, sims, values)
	return values
}
