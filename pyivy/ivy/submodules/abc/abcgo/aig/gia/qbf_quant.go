package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_QbfQuantifyOne(p *Gia_Man_t, variable, andAll, orAll int32) *Gia_Man_t {
	if p == nil || p.nRegs != 0 || variable < 0 || variable >= Gia_ManPiNum(p) {
		panic("gia: invalid QBF quantification")
	}
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	Gia_ManHashAlloc(q)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		Gia_ManPi(p, i).Value = uint32(Gia_ManAppendCi(q))
	}
	outputs := make([][]int32, 2)
	for value := int32(0); value < 2; value++ {
		Gia_ManPi(p, variable).Value = uint32(value)
		for _, obj := range p.pObjs {
			if Gia_ObjIsAnd(obj) != 0 {
				obj.Value = uint32(Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
			}
		}
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			outputs[value] = append(outputs[value], Gia_ObjFanin0Copy(Gia_ManPo(p, i)))
		}
	}
	if andAll != 0 || orAll != 0 {
		for i := range outputs[0] {
			lit := int32(0)
			if andAll != 0 {
				lit = Gia_ManHashAnd(q, outputs[0][i], outputs[1][i])
			} else {
				lit = Gia_ManHashOr(q, outputs[0][i], outputs[1][i])
			}
			Gia_ManAppendCo(q, lit)
		}
	} else {
		for _, phase := range outputs {
			for _, lit := range phase {
				Gia_ManAppendCo(q, lit)
			}
		}
	}
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Gia_QbfQuantifyAll(p *Gia_Man_t, parameters, andAll, orAll int32) *Gia_Man_t {
	if parameters < 0 || parameters > Gia_ManPiNum(p) {
		panic("gia: invalid QBF parameter count")
	}
	result := Gia_ManDup(p)
	for i := Gia_ManPiNum(p) - 1; i >= parameters; i-- {
		old := result
		result = Gia_QbfQuantifyOne(old, i, andAll, orAll)
		Gia_ManStop(old)
	}
	return result
}
func Gia_QbfCofactor(p *Gia_Man_t, parameters int32, values, parameterMap *Vec_Int_t) *Gia_Man_t {
	if p == nil || p.nRegs != 0 || parameters < 0 || Gia_ManPiNum(p) != parameters+int32(vec.Vec_IntSize(values)) || int32(vec.Vec_IntSize(parameterMap)) < parameters {
		panic("gia: invalid QBF cofactor")
	}
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	Gia_ManHashAlloc(q)
	Gia_ManConst0(p).Value = 0
	for i := int32(0); i < Gia_ManPiNum(p); i++ {
		var lit int32
		if i < parameters {
			lit = Gia_ManAppendCi(q)
			if replacement := vec.Vec_IntEntry(parameterMap, int(i)); replacement != -1 {
				lit = replacement
			}
		} else {
			lit = vec.Vec_IntEntry(values, int(i-parameters))
		}
		Gia_ManPi(p, i).Value = uint32(lit)
	}
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) != 0 {
			obj.Value = uint32(Gia_ManHashAnd(q, Gia_ObjFanin0Copy(obj), Gia_ObjFanin1Copy(obj)))
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManAppendCo(q, Gia_ObjFanin0Copy(Gia_ManObj(p, id)))
	}
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
