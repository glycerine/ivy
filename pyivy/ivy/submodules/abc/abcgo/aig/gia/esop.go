package gia

import (
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"time"
)

type Eso_Man_t struct {
	pGia                  *Gia_Man_t
	nVars, Cube1          int32
	vEsops, vCubes        *Vec_Wec_t
	cubes                 []*Vec_Int_t
	cubeIDs               map[string]int32
	vCube1, vCube2, vCube *Vec_Int_t
}

func (p *Eso_Man_t) intern(lits []int32) int32 {
	data := make([]byte, 4*len(lits))
	for i, x := range lits {
		binary.LittleEndian.PutUint32(data[4*i:], uint32(x))
	}
	key := string(data)
	if id, ok := p.cubeIDs[key]; ok {
		return id
	}
	id := int32(len(p.cubes))
	v := vec.Vec_IntAlloc(len(lits))
	for _, x := range lits {
		vec.Vec_IntPush(v, x)
	}
	p.cubes = append(p.cubes, v)
	p.cubeIDs[key] = id
	return id
}
func Eso_ManCube(p *Eso_Man_t, id int32) *Vec_Int_t { return p.cubes[id] }
func Eso_ManAlloc(g *Gia_Man_t) *Eso_Man_t {
	n := Gia_ManCiNum(g)
	p := &Eso_Man_t{pGia: g, nVars: n, Cube1: 1000000000, vEsops: vec.Vec_WecStart(int(g.nObjs)), vCubes: vec.Vec_WecStart(int(n + 1)), cubeIDs: make(map[string]int32), vCube1: vec.Vec_IntAlloc(int(n)), vCube2: vec.Vec_IntAlloc(int(n)), vCube: vec.Vec_IntAlloc(int(n))}
	for i := int32(0); i < n; i++ {
		p.intern([]int32{i << 1})
		p.intern([]int32{i<<1 | 1})
		vec.Vec_IntPush(vec.Vec_WecEntry(p.vEsops, int(Gia_ManCi(g, i).id)), i<<1)
	}
	return p
}
func Eso_ManStop(p *Eso_Man_t) {
	if p != nil {
		*p = Eso_Man_t{}
	}
}
func Eso_ManCoverPrint(p *Eso_Man_t, esop *Vec_Int_t) {
	if vec.Vec_IntSize(esop) == 0 {
		fmt.Println("Const 0")
		return
	}
	for _, id := range vec.Vec_IntArray(esop) {
		chars := make([]byte, p.nVars)
		for i := range chars {
			chars[i] = '-'
		}
		if id != p.Cube1 {
			for _, lit := range vec.Vec_IntArray(Eso_ManCube(p, id)) {
				chars[lit>>1] = '1'
				if lit&1 != 0 {
					chars[lit>>1] = '0'
				}
			}
		}
		fmt.Printf("%s 1\n", chars)
	}
	fmt.Println()
}
func Eso_ManCoverDerive(p *Eso_Man_t, cover *Vec_Ptr_t) *Vec_Wec_t {
	out := vec.Vec_WecAlloc(vec.Vec_PtrSize(cover))
	for i := 0; i < vec.Vec_PtrSize(cover); i++ {
		esop := vec.Vec_PtrEntry(cover, i).(*Vec_Int_t)
		for _, id := range vec.Vec_IntArray(esop) {
			level := vec.Vec_WecPushLevel(out)
			if id != p.Cube1 {
				vec.Vec_IntAppend(level, Eso_ManCube(p, id))
			}
			vec.Vec_IntPush(level, -int32(i)-1)
		}
	}
	return out
}
func Eso_ManCoverConvert(p *Eso_Man_t, cover *Vec_Ptr_t) *Gia_Man_t {
	g := p.pGia
	q := Gia_ManStart(g.nObjs)
	q.pName = g.pName
	q.PSpec = g.PSpec
	Gia_ManHashAlloc(q)
	Gia_ObjSetValue(Gia_ManConst0(g), 0)
	for i := int32(0); i < Gia_ManCiNum(g); i++ {
		Gia_ObjSetValue(Gia_ManCi(g, i), Gia_ManAppendCi(q))
	}
	for i := 0; i < vec.Vec_PtrSize(cover); i++ {
		root := int32(0)
		for _, id := range vec.Vec_IntArray(vec.Vec_PtrEntry(cover, i).(*Vec_Int_t)) {
			product := int32(1)
			if id != p.Cube1 {
				for _, lit := range vec.Vec_IntArray(Eso_ManCube(p, id)) {
					product = Gia_ManHashAnd(q, product, lit+2)
				}
			}
			root = Gia_ManHashXor(q, root, product)
		}
		Gia_ManAppendCo(q, root)
	}
	Gia_ManHashStop(q)
	result := Gia_ManCleanup(q)
	Gia_ManStop(q)
	return result
}
func Eso_ManFindDistOneLitEqual(a, b []int32, n int32) int32 {
	diff := int32(-1)
	for i := int32(0); i < n; i++ {
		if a[i] != b[i] {
			if diff != -1 || a[i]>>1 != b[i]>>1 {
				return -1
			}
			diff = i
		}
	}
	return diff
}
func Eso_ManFindDistOneLitNotEqual(short, long []int32, n int32) int32 {
	i, k, diff := int32(0), int32(0), int32(-1)
	for i < n {
		if short[i] != long[k] {
			if diff != -1 {
				return -1
			}
			diff = k
			k++
			continue
		}
		i++
		k++
	}
	if diff < 0 {
		diff = n
	}
	return diff
}
func Eso_ManMinimizeAdd(p *Eso_Man_t, id int32) {
	var cube []int32
	if id != p.Cube1 {
		cube = vec.Vec_IntArray(Eso_ManCube(p, id))
	}
	n := len(cube)
	level := vec.Vec_WecEntry(p.vCubes, n)
	for i, x := range vec.Vec_IntArray(level) {
		if x == id {
			vec.Vec_IntDrop(level, i)
			return
		}
	}
	if id == p.Cube1 {
		vec.Vec_IntPush(level, id)
		return
	}
	if n < int(p.nVars)-1 {
		higher := vec.Vec_WecEntry(p.vCubes, n+1)
		for i, other := range vec.Vec_IntArray(higher) {
			otherCube := vec.Vec_IntArray(Eso_ManCube(p, other))
			index := Eso_ManFindDistOneLitNotEqual(cube, otherCube, int32(n))
			if index < 0 {
				continue
			}
			vec.Vec_IntDrop(higher, i)
			replacement := append([]int32(nil), otherCube...)
			replacement[index] ^= 1
			Eso_ManMinimizeAdd(p, p.intern(replacement))
			return
		}
	}
	for i, other := range vec.Vec_IntArray(level) {
		otherCube := vec.Vec_IntArray(Eso_ManCube(p, other))
		index := Eso_ManFindDistOneLitEqual(otherCube, cube, int32(n))
		if index < 0 {
			continue
		}
		vec.Vec_IntDrop(level, i)
		replacement := make([]int32, 0, n-1)
		replacement = append(replacement, cube[:index]...)
		replacement = append(replacement, cube[index+1:]...)
		if len(replacement) == 0 {
			Eso_ManMinimizeAdd(p, p.Cube1)
		} else {
			Eso_ManMinimizeAdd(p, p.intern(replacement))
		}
		return
	}
	lower := vec.Vec_WecEntry(p.vCubes, n-1)
	if n == 1 && vec.Vec_IntSize(lower) == 1 {
		vec.Vec_IntDrop(lower, 0)
		id ^= 1
	} else {
		for i, other := range vec.Vec_IntArray(lower) {
			otherCube := vec.Vec_IntArray(Eso_ManCube(p, other))
			index := Eso_ManFindDistOneLitNotEqual(otherCube, cube, int32(n-1))
			if index < 0 {
				continue
			}
			vec.Vec_IntDrop(lower, i)
			replacement := append([]int32(nil), cube...)
			replacement[index] ^= 1
			Eso_ManMinimizeAdd(p, p.intern(replacement))
			return
		}
	}
	vec.Vec_IntPush(level, id)
}
func Eso_ManMinimizeCopy(p *Eso_Man_t, esop *Vec_Int_t) {
	vec.Vec_IntClear(esop)
	for i := 0; i < vec.Vec_WecSize(p.vCubes); i++ {
		level := vec.Vec_WecEntry(p.vCubes, i)
		vec.Vec_IntAppend(esop, level)
		vec.Vec_IntClear(level)
	}
}
func Eso_ManComputeAnd(p *Eso_Man_t, a, b, out *Vec_Int_t) int32 {
	x, y := vec.Vec_IntArray(a), vec.Vec_IntArray(b)
	result := make([]int32, 0, len(x)+len(y))
	i, j := 0, 0
	for i < len(x) && j < len(y) {
		if x[i] == y[j] {
			result = append(result, x[i])
			i++
			j++
		} else if x[i]>>1 == y[j]>>1 {
			return -1
		} else if x[i] < y[j] {
			result = append(result, x[i])
			i++
		} else {
			result = append(result, y[j])
			j++
		}
	}
	result = append(result, x[i:]...)
	result = append(result, y[j:]...)
	vec.Vec_IntClear(out)
	for _, lit := range result {
		vec.Vec_IntPush(out, lit)
	}
	return p.intern(result)
}
func Eso_ManComputeOne(p *Eso_Man_t, a, b, out *Vec_Int_t) {
	vec.Vec_IntClear(out)
	for _, x := range vec.Vec_IntArray(a) {
		for _, y := range vec.Vec_IntArray(b) {
			var id int32
			if x == p.Cube1 {
				id = y
			} else if y == p.Cube1 {
				id = x
			} else {
				id = Eso_ManComputeAnd(p, Eso_ManCube(p, x), Eso_ManCube(p, y), p.vCube)
			}
			if id >= 0 {
				Eso_ManMinimizeAdd(p, id)
			}
		}
	}
	Eso_ManMinimizeCopy(p, out)
}
func Eso_ManTransformOne(p *Eso_Man_t, esop *Vec_Int_t, complement int32, out *Vec_Int_t) *Vec_Int_t {
	vec.Vec_IntClear(out)
	entries := vec.Vec_IntArray(esop)
	start := 0
	if complement != 0 {
		if len(entries) == 0 {
			vec.Vec_IntPush(out, p.Cube1)
		} else if entries[0] == p.Cube1 {
			start = 1
		} else if entries[0] < 2*p.nVars {
			vec.Vec_IntPush(out, entries[0]^1)
			start = 1
		} else {
			vec.Vec_IntPush(out, p.Cube1)
		}
	}
	for _, x := range entries[start:] {
		vec.Vec_IntPush(out, x)
	}
	return out
}
func Eso_ManCompute(g *Gia_Man_t, verbose int32, coverOut **Vec_Wec_t) *Gia_Man_t {
	start := time.Now()
	p := Eso_ManAlloc(g)
	var total, used int
	for _, obj := range g.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a := Eso_ManTransformOne(p, vec.Vec_WecEntry(p.vEsops, int(Gia_ObjFaninId0(obj, obj.id))), Gia_ObjFaninC0(obj), p.vCube1)
		b := Eso_ManTransformOne(p, vec.Vec_WecEntry(p.vEsops, int(Gia_ObjFaninId1(obj, obj.id))), Gia_ObjFaninC1(obj), p.vCube2)
		out := vec.Vec_WecEntry(p.vEsops, int(obj.id))
		Eso_ManComputeOne(p, a, b, out)
		total += vec.Vec_IntSize(out)
	}
	cover := vec.Vec_PtrAlloc(int(Gia_ManCoNum(g)))
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		v := Eso_ManTransformOne(p, vec.Vec_WecEntry(p.vEsops, int(Gia_ObjFaninId0(obj, obj.id))), Gia_ObjFaninC0(obj), p.vCube1)
		if verbose != 0 {
			fmt.Printf("Output %3d:  ESOP has %5d cubes\n", i, vec.Vec_IntSize(v))
		}
		used += vec.Vec_IntSize(v)
		vec.Vec_PtrPush(cover, vec.Vec_IntDup(v))
	}
	if verbose != 0 {
		fmt.Printf("Outs = %d.  Cubes = %d.  Used = %d.  Hashed = %d. Time = %.2f sec\n", Gia_ManCoNum(g), total, used, len(p.cubes), time.Since(start).Seconds())
	}
	var result *Gia_Man_t
	if coverOut != nil {
		*coverOut = Eso_ManCoverDerive(p, cover)
	} else {
		result = Eso_ManCoverConvert(p, cover)
	}
	Eso_ManStop(p)
	return result
}
