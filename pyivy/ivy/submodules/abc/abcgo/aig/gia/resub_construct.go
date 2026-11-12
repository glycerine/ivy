package gia

import "github.com/glycerine/abcgo/misc/vec"

// Gia_ManConstructFromMap builds a resubstitution gate list using literals in
// used for existing variables and copy for the newly built intermediate gates.
func Gia_ManConstructFromMap(dst *Gia_Man_t, gates *Vec_Int_t, nVars int32, used, copy *Vec_Int_t, hash int32) int32 {
	data := vec.Vec_IntArray(gates)
	if nVars < 0 || vec.Vec_IntSize(used) != int(nVars) || len(data) < 3 || len(data)%2 != 1 || data[len(data)-1]>>1-nVars != int32(len(data)/2-1) {
		panic("gia: invalid resubstitution gate list")
	}
	vec.Vec_IntClear(copy)
	resolve := func(lit int32) int32 {
		id := lit >> 1
		var value int32
		if id < nVars {
			value = vec.Vec_IntEntry(used, int(id))
		} else {
			value = vec.Vec_IntEntry(copy, int(id-nVars))
		}
		if value < 0 {
			panic("gia: unmapped resubstitution input")
		}
		return value ^ (lit & 1)
	}
	for i := 0; i < len(data)-1; i += 2 {
		a, b := data[i], data[i+1]
		v0, v1 := a>>1, b>>1
		x, y := resolve(a), resolve(b)
		var result int32
		if v0 < v1 {
			if hash != 0 {
				result = Gia_ManHashAnd(dst, x, y)
			} else {
				result = Gia_ManAppendAnd(dst, x, y)
			}
		} else if v0 > v1 && a&1 == 0 && b&1 == 0 {
			if hash != 0 {
				result = Gia_ManHashXor(dst, x, y)
			} else {
				result = Gia_ManAppendXor(dst, x, y)
			}
		} else {
			panic("gia: invalid resubstitution gate")
		}
		vec.Vec_IntPush(copy, result)
	}
	return vec.Vec_IntEntryLast(copy)
}
func Gia_ManConstructFromGates(functions *Vec_Wec_t, nVars int32) *Gia_Man_t {
	if nVars < 1 {
		panic("gia: invalid resubstitution variable count")
	}
	dst := Gia_ManStart(100)
	dst.pName = "resub"
	used := vec.Vec_IntStartFull(int(nVars))
	vec.Vec_IntWriteEntry(used, 0, 0)
	copy := vec.Vec_IntAlloc(0)
	for i := 0; i < vec.Vec_WecSize(functions); i++ {
		gates := vec.Vec_IntArray(vec.Vec_WecEntry(functions, i))
		if len(gates) == 0 || len(gates)%2 != 1 {
			panic("gia: invalid resubstitution function")
		}
		for _, lit := range gates {
			id := lit >> 1
			if id > 0 && id < nVars && vec.Vec_IntEntry(used, int(id)) == -1 {
				vec.Vec_IntWriteEntry(used, int(id), Gia_ManAppendCi(dst))
			}
		}
	}
	for i := 0; i < vec.Vec_WecSize(functions); i++ {
		gates := vec.Vec_WecEntry(functions, i)
		top := vec.Vec_IntEntryLast(gates)
		id := top >> 1
		var result int32
		if id < nVars {
			result = vec.Vec_IntEntry(used, int(id))
		} else {
			result = Gia_ManConstructFromMap(dst, gates, nVars, used, copy, 0)
		}
		Gia_ManAppendCo(dst, result^(top&1))
	}
	return dst
}

// Gia_ManInsertOrder returns a dependency order for outputs and replacements.
func Gia_ManInsertOrder(p *Gia_Man_t, objects *Vec_Int_t, functions *Vec_Wec_t) *Vec_Int_t {
	if vec.Vec_IntSize(objects) != vec.Vec_WecSize(functions) {
		panic("gia: mismatched resubstitution replacements")
	}
	replacement := make(map[int32]int)
	for i, id := range vec.Vec_IntArray(objects) {
		if _, ok := replacement[id]; ok || id <= 0 || id >= p.nObjs {
			panic("gia: invalid resubstitution target")
		}
		replacement[id] = i
	}
	state := make([]uint8, p.nObjs)
	result := vec.Vec_IntAlloc(int(p.nObjs))
	var visit func(int32)
	visit = func(id int32) {
		if id == 0 || state[id] == 2 {
			return
		}
		if id < 0 || id >= p.nObjs || state[id] == 1 {
			panic("gia: cyclic resubstitution")
		}
		state[id] = 1
		obj := Gia_ManObj(p, id)
		if index, ok := replacement[id]; ok {
			for _, lit := range vec.Vec_IntArray(vec.Vec_WecEntry(functions, index)) {
				dep := lit >> 1
				if dep < p.nObjs {
					visit(dep)
				}
			}
		} else if Gia_ObjIsCo(obj) != 0 {
			visit(Gia_ObjFaninId0(obj, id))
		} else if Gia_ObjIsAnd(obj) != 0 {
			for _, dep := range giaObjectFanins(p, obj) {
				visit(dep)
			}
		} else if Gia_ObjIsCi(obj) == 0 {
			panic("gia: invalid resubstitution node")
		}
		state[id] = 2
		if Gia_ObjIsCi(obj) == 0 {
			vec.Vec_IntPush(result, id)
		}
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		visit(id)
	}
	return result
}
func Gia_ManInsertFromGates(p *Gia_Man_t, objects *Vec_Int_t, functions *Vec_Wec_t) *Gia_Man_t {
	order := Gia_ManInsertOrder(p, objects, functions)
	replacements := make(map[int32]int)
	for i, id := range vec.Vec_IntArray(objects) {
		replacements[id] = i
	}
	dst := Gia_ManStart(p.nObjs + 1000)
	dst.CopyIdentityFrom(p)
	Gia_ManHashStart(dst)
	used := vec.Vec_IntStartFull(int(p.nObjs))
	vec.Vec_IntWriteEntry(used, 0, 0)
	copy := vec.Vec_IntAlloc(0)
	for _, id := range vec.Vec_IntArray(p.vCis) {
		vec.Vec_IntWriteEntry(used, int(id), Gia_ManAppendCi(dst))
	}
	for _, id := range vec.Vec_IntArray(order) {
		var result int32
		if index, ok := replacements[id]; ok {
			gates := vec.Vec_WecEntry(functions, index)
			top := vec.Vec_IntEntryLast(gates)
			if top>>1 < p.nObjs {
				result = vec.Vec_IntEntry(used, int(top>>1))
			} else {
				result = Gia_ManConstructFromMap(dst, gates, p.nObjs, used, copy, 1)
			}
			result ^= top & 1
		} else {
			a, b, c, kind := p.Gate(id)
			read := func(lit int32) int32 {
				v := vec.Vec_IntEntry(used, int(lit>>1))
				if v < 0 {
					panic("gia: unmapped resubstitution dependency")
				}
				return v ^ (lit & 1)
			}
			switch kind {
			case 2:
				result = Gia_ManHashAnd(dst, read(a), read(b))
			case 3:
				result = Gia_ManHashXor(dst, read(a), read(b))
			case 4:
				result = Gia_ManHashMux(dst, read(a), read(b), read(c))
			case 5, 6:
				result = read(a)
			default:
				panic("gia: invalid resubstitution gate")
			}
		}
		vec.Vec_IntWriteEntry(used, int(id), result)
	}
	for _, id := range vec.Vec_IntArray(p.vCos) {
		Gia_ManAppendCo(dst, vec.Vec_IntEntry(used, int(id)))
	}
	Gia_ManHashStop(dst)
	result := Gia_ManCleanup(dst)
	Gia_ManStop(dst)
	Gia_ManSetRegNum(result, p.nRegs)
	return result
}
