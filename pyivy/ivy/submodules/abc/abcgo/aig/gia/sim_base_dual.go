package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManSimPatAssignInputs2(p *Gia_Man_t, n int32, sims, inputs *Vec_Wrd_t) {
	if vec.Vec_WrdSize(sims) != int(2*n*p.nObjs) || vec.Vec_WrdSize(inputs) != int(n*Gia_ManCiNum(p)) {
		panic("gia: invalid dual-polarity dimensions")
	}
	dst, src := vec.Vec_WrdArray(sims), vec.Vec_WrdArray(inputs)
	for i, id := range vec.Vec_IntArray(p.vCis) {
		for w := int32(0); w < n; w++ {
			v := src[int32(i)*n+w]
			dst[2*id*n+w] = v
			dst[(2*id+1)*n+w] = ^v
		}
	}
}
func Gia_ManSimPatSimAnd2(p *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims *Vec_Wrd_t) {
	if Gia_ObjIsXor(obj) != 0 {
		panic("gia: dual-polarity simulator does not support XOR")
	}
	data := vec.Vec_WrdArray(sims)
	a := 2*Gia_ObjFaninId0(obj, id) + Gia_ObjFaninC0(obj)
	b := 2*Gia_ObjFaninId1(obj, id) + Gia_ObjFaninC1(obj)
	for w := int32(0); w < n; w++ {
		v := data[a*n+w] & data[b*n+w]
		data[2*id*n+w] = v
		data[(2*id+1)*n+w] = ^v
	}
}
func Gia_ManSimPatSimPo2(p *Gia_Man_t, id int32, obj *Gia_Obj_t, n int32, sims *Vec_Wrd_t) {
	data := vec.Vec_WrdArray(sims)
	a := 2*Gia_ObjFaninId0(obj, id) + Gia_ObjFaninC0(obj)
	copy(data[id*n:(id+1)*n], data[a*n:(a+1)*n])
}
func Gia_ManSimPatSim2(p *Gia_Man_t) *Vec_Wrd_t {
	n := giaPatternWords(p)
	sims := vec.Vec_WrdStart(int(2 * p.nObjs * n))
	Gia_ManSimPatAssignInputs2(p, n, sims, p.VSimsPi)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			Gia_ManSimPatSimAnd2(p, obj.id, obj, n, sims)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManSimPatSimPo2(p, id, Gia_ManObj(p, id), n, sims)
	}
	return sims
}
func Gia_ManSimCombine(inputs int32, base, add *Vec_Wrd_t, use int32) *Vec_Wrd_t {
	if inputs <= 0 || vec.Vec_WrdSize(base)%int(inputs) != 0 || vec.Vec_WrdSize(add)%int(inputs) != 0 {
		panic("gia: invalid simulation dimensions")
	}
	nBase, nAdd := vec.Vec_WrdSize(base)/int(inputs), vec.Vec_WrdSize(add)/int(inputs)
	if use < 0 || int(use) > nAdd {
		panic("gia: invalid add-on count")
	}
	out := vec.Vec_WrdAlloc(int(inputs) * (nBase + int(use)))
	b, a := vec.Vec_WrdArray(base), vec.Vec_WrdArray(add)
	for i := 0; i < int(inputs); i++ {
		for _, v := range b[i*nBase : (i+1)*nBase] {
			vec.Vec_WrdPush(out, v)
		}
		for _, v := range a[i*nAdd : i*nAdd+int(use)] {
			vec.Vec_WrdPush(out, v)
		}
	}
	return out
}
