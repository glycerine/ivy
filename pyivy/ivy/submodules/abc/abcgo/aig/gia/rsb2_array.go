package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_Rsb2AddNode(result *Vec_Int_t, lit0, lit1, res0, res1 int32) int32 {
	if lit0 == lit1 {
		panic("gia: invalid Rsb2 gate")
	}
	low, high := res0^(lit0&1), res1^(lit1&1)
	if res0 >= res1 {
		low, high = high, low
	}
	if low > high {
		low, high = high, low
	}
	node := int32(vec.Vec_IntSize(result))
	if lit0 < lit1 {
		if low == high {
			return low
		}
		if low == 0 {
			return 0
		}
		if low == 1 {
			return high
		}
		if low == high^1 {
			return 0
		}
		vec.Vec_IntPush(result, low)
		vec.Vec_IntPush(result, high)
	} else {
		if low == high {
			return 0
		}
		if lit0&1 != 0 || lit1&1 != 0 {
			panic("gia: complemented Rsb2 XOR input")
		}
		if low == 0 {
			return high
		}
		if low == 1 {
			return high ^ 1
		}
		if low == high^1 {
			return 1
		}
		vec.Vec_IntPush(result, high)
		vec.Vec_IntPush(result, low)
	}
	return node
}
func Gia_Rsb2ManInsert_rec(result *Vec_Int_t, nPis int32, objects *Vec_Int_t, node int32, resub, divs, copies *Vec_Int_t, id int32) int32 {
	mapped := vec.Vec_IntEntry(copies, int(id))
	if mapped >= 0 {
		return mapped
	}
	if id <= nPis {
		panic("gia: invalid Rsb2 input copy")
	}
	var value int32
	if id == node {
		vars := int32(vec.Vec_IntSize(divs))
		data := vec.Vec_IntArray(resub)
		if len(data) == 0 || len(data)%2 != 1 {
			panic("gia: invalid Rsb2 replacement")
		}
		top := data[len(data)-1]
		resolveOriginal := func(variable int32) int32 {
			object := vec.Vec_IntEntry(divs, int(variable))
			return Gia_Rsb2ManInsert_rec(result, nPis, objects, -1, resub, divs, copies, object)
		}
		if top>>1 == 0 {
			value = 0
		} else if top>>1 < vars {
			value = resolveOriginal(top >> 1)
		} else {
			local := make([]int32, 0, len(data)/2)
			for i := 0; i < len(data)-1; i += 2 {
				a, b := data[i], data[i+1]
				resolve := func(lit int32) int32 {
					v := lit >> 1
					if v < vars {
						return resolveOriginal(v)
					}
					if int(v-vars) >= len(local) {
						panic("gia: invalid Rsb2 replacement dependency")
					}
					return local[v-vars]
				}
				local = append(local, Gia_Rsb2AddNode(result, a, b, resolve(a), resolve(b)))
			}
			value = local[len(local)-1]
		}
		value ^= top & 1
	} else {
		a, b := vec.Vec_IntEntry(objects, int(2*id)), vec.Vec_IntEntry(objects, int(2*id+1))
		x := Gia_Rsb2ManInsert_rec(result, nPis, objects, node, resub, divs, copies, a>>1)
		y := Gia_Rsb2ManInsert_rec(result, nPis, objects, node, resub, divs, copies, b>>1)
		value = Gia_Rsb2AddNode(result, a, b, x, y)
	}
	vec.Vec_IntWriteEntry(copies, int(id), value)
	return value
}
func Gia_Rsb2ManInsert(nPis, nPos int32, objects *Vec_Int_t, node int32, resub, divs, copies *Vec_Int_t) *Vec_Int_t {
	n := int32(vec.Vec_IntSize(objects) / 2)
	first := n - nPos
	if nPis < 0 || nPos < 0 || first <= nPis || vec.Vec_IntSize(objects)%2 != 0 {
		panic("gia: invalid Rsb2 object list")
	}
	result := vec.Vec_IntStart(int(2 * (nPis + 1)))
	vec.Vec_IntFill(copies, int(n), -1)
	for i := int32(0); i <= nPis; i++ {
		vec.Vec_IntWriteEntry(copies, int(i), 2*i)
	}
	for i := first; i < n; i++ {
		id := vec.Vec_IntEntry(objects, int(2*i)) >> 1
		Gia_Rsb2ManInsert_rec(result, nPis, objects, node, resub, divs, copies, id)
	}
	for i := first; i < n; i++ {
		lit := vec.Vec_IntEntry(objects, int(2*i))
		mapped := vec.Vec_IntEntry(copies, int(lit>>1)) ^ (lit & 1)
		vec.Vec_IntPush(result, mapped)
		vec.Vec_IntPush(result, mapped)
	}
	return result
}
func Gia_ManToResub(p *Gia_Man_t) []int32 {
	if Gia_ManIsNormalized(p) == 0 {
		panic("gia: Rsb2 requires normalized graph")
	}
	result := make([]int32, 2*p.nObjs)
	for id := int32(1); id < p.nObjs; id++ {
		a, b, _, kind := p.Gate(id)
		switch kind {
		case 1:
			continue
		case 6:
			result[2*id], result[2*id+1] = a, a
		case 2:
			result[2*id], result[2*id+1] = min(a, b), max(a, b)
		case 3:
			result[2*id], result[2*id+1] = max(a, b), min(a, b)
		default:
			panic("gia: unsupported Rsb2 graph gate")
		}
	}
	return result
}
func Gia_ManFromResub(objects []int32, nObjs, nIns int32) *Gia_Man_t {
	if nObjs < 1 || nIns < 0 || len(objects) != int(2*nObjs) {
		panic("gia: invalid Rsb2 object array")
	}
	g := Gia_ManStart(nObjs)
	copies := make([]int32, nObjs)
	read := func(lit int32, id int32) int32 {
		if lit < 0 || lit>>1 >= id {
			panic("gia: invalid Rsb2 fanin")
		}
		return copies[lit>>1] ^ (lit & 1)
	}
	for id := int32(1); id < nObjs; id++ {
		a, b := objects[2*id], objects[2*id+1]
		if id <= nIns {
			if a != 0 || b != 0 {
				panic("gia: invalid Rsb2 input")
			}
			copies[id] = Gia_ManAppendCi(g)
		} else if a == b {
			copies[id] = Gia_ManAppendCo(g, read(a, id))
		} else if a < b {
			copies[id] = Gia_ManAppendAnd(g, read(a, id), read(b, id))
		} else {
			copies[id] = Gia_ManAppendXor(g, read(a, id), read(b, id))
		}
	}
	return g
}
