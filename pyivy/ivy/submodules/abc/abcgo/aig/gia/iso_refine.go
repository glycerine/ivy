package gia

import (
	"fmt"
	"sort"
	"time"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
)

const ISO_MASK = 0xff

var s_256Primes = [ISO_MASK + 1]uint32{
	0x984b6ad9, 0x18a6eed3, 0x950353e2, 0x6222f6eb, 0xdfbedd47, 0xef0f9023, 0xac932a26, 0x590eaf55,
	0x97d0a034, 0xdc36cd2e, 0x22736b37, 0xdc9066b0, 0x2eb2f98b, 0x5d9c7baf, 0x85747c9e, 0x8aca1055,
	0x50d66b74, 0x2f01ae9e, 0xa1a80123, 0x3e1ce2dc, 0xebedbc57, 0x4e68bc34, 0x855ee0cf, 0x17275120,
	0x2ae7f2df, 0xf71039eb, 0x7c283eec, 0x70cd1137, 0x7cf651f3, 0xa87bfa7a, 0x14d87f02, 0xe82e197d,
	0x8d8a5ebe, 0x1e6a15dc, 0x197d49db, 0x5bab9c89, 0x4b55dea7, 0x55dede49, 0x9a6a8080, 0xe5e51035,
	0xe148d658, 0x8a17eb3b, 0xe22e4b38, 0xe5be2a9a, 0xbe938cbb, 0x3b981069, 0x7f9c0c8e, 0xf756df10,
	0x8fa783f7, 0x252062ce, 0x3dc46b4b, 0xf70f6432, 0x3f378276, 0x44b137a1, 0x2bf74b77, 0x04892ed6,
	0xfd318de1, 0xd58c235e, 0x94c6d25b, 0x7aa5f218, 0x35c9e921, 0x5732fbbb, 0x06026481, 0xf584a44f,
	0x946e1b5f, 0x8463d5b2, 0x4ebca7b2, 0x54887b15, 0x08d1e804, 0x5b22067d, 0x794580f6, 0xb351ea43,
	0xbce555b9, 0x19ae2194, 0xd32f1396, 0x6fc1a7f1, 0x1fd8a867, 0x3a89fdb0, 0xea49c61c, 0x25f8a879,
	0xde1e6437, 0x7c74afca, 0x8ba63e50, 0xb1572074, 0xe4655092, 0xdb6f8b1c, 0xc2955f3c, 0x327f85ba,
	0x60a17021, 0x95bd261d, 0xdea94f28, 0x04528b65, 0xbe0109cc, 0x26dd5688, 0x6ab2729d, 0xc4f029ce,
	0xacf7a0be, 0x4c912f55, 0x34c06e65, 0x4fbb938e, 0x1533fb5f, 0x03da06bd, 0x48262889, 0xc2523d7d,
	0x28a71d57, 0x89f9713a, 0xf574c551, 0x7a99deb5, 0x52834d91, 0x5a6f4484, 0xc67ba946, 0x13ae698f,
	0x3e390f34, 0x34fc9593, 0x894c7932, 0x6cf414a3, 0xdb7928ab, 0x13a3b8a3, 0x4b381c1d, 0xa10b54cb,
	0x55359d9d, 0x35a3422a, 0x58d1b551, 0x0fd4de20, 0x199eb3f4, 0x167e09e2, 0x3ee6a956, 0x5371a7fa,
	0xd424efda, 0x74f521c5, 0xcb899ff6, 0x4a42e4f4, 0x747917b6, 0x4b08df0b, 0x090c7a39, 0x11e909e4,
	0x258e2e32, 0xd9fad92d, 0x48fe5f69, 0x0545cde6, 0x55937b37, 0x9b4ae4e4, 0x1332b40e, 0xc3792351,
	0xaff982ef, 0x4dba132a, 0x38b81ef1, 0x28e641bf, 0x227208c1, 0xec4bbe37, 0xc4e1821c, 0x512c9d09,
	0xdaef1257, 0xb63e7784, 0x043e04d7, 0x9c2cea47, 0x45a0e59a, 0x281315ca, 0x849f0aac, 0xa4071ed3,
	0x0ef707b3, 0xfe8dac02, 0x12173864, 0x471f6d46, 0x24a53c0a, 0x35ab9265, 0xbbf77406, 0xa2144e79,
	0xb39a884a, 0x0baf5b6d, 0xcccee3dd, 0x12c77584, 0x2907325b, 0xfd1adcd2, 0xd16ee972, 0x345ad6c1,
	0x315ebe66, 0xc7ad2b8d, 0x99e82c8d, 0xe52da8c8, 0xba50f1d3, 0x66689cd8, 0x2e8e9138, 0x43e15e74,
	0xf1ced14d, 0x188ec52a, 0xe0ef3cbb, 0xa958aedc, 0x4107a1bc, 0x5a9e7a3e, 0x3bde939f, 0xb5b28d5a,
	0x596fe848, 0xe85ad00c, 0x0b6b3aae, 0x44503086, 0x25b5695c, 0xc0c31dcd, 0x5ee617f0, 0x74d40c3a,
	0xd2cb2b9f, 0x1e19f5fa, 0x81e24faf, 0xa01ed68f, 0xcee172fc, 0x7fdf2e4d, 0x002f4774, 0x664f82dd,
	0xc569c39a, 0xa2d4dcbe, 0xaadea306, 0xa4c947bf, 0xa413e4e3, 0x81fb5486, 0x8a404970, 0x752c980c,
	0x98d1d881, 0x5c932c1e, 0xeee65dfb, 0x37592cdd, 0x0fd4e65b, 0xad1d383f, 0x62a1452f, 0x8872f68d,
	0xb58c919b, 0x345c8ee3, 0xb583a6d6, 0x43d72cb3, 0x77aaa0aa, 0xeb508242, 0xf2db64f8, 0x86294328,
	0x82211731, 0x1239a9d5, 0x673ba5de, 0xaf4af007, 0x44203b19, 0x2399d955, 0xa175cd12, 0x595928a7,
	0x6918928b, 0xde3126bb, 0x6c99835c, 0x63ba1fa2, 0xdebbdff0, 0x3d02e541, 0xd6f7aac6, 0xe80b4cd0,
	0xd0fa29f1, 0x804cac5e, 0x2c226798, 0x462f624c, 0xad05b377, 0x22924fcd, 0xfbea205c, 0x1b47586d,
}

type Gia_IsoMan_t struct {
	pGia                                                           *Gia_Man_t
	nObjs, nUniques, nSingles, nEntries                            int32
	pLevels, pUniques                                              []int32
	pStoreW                                                        []word
	vLevCounts, vClasses, vClasses2                                *Vec_Int_t
	timeStart, timeSim, timeRefine, timeSort, timeOther, timeTotal int64
}

func Gia_IsoGetValue(p *Gia_IsoMan_t, i int32) uint32 { return uint32(p.pStoreW[i]) }
func Gia_IsoGetItem(p *Gia_IsoMan_t, i int32) uint32  { return uint32(p.pStoreW[i] >> 32) }
func Gia_IsoSetValue(p *Gia_IsoMan_t, i int32, value uint32) {
	p.pStoreW[i] = p.pStoreW[i]&0xffffffff00000000 | word(value)
}
func Gia_IsoSetItem(p *Gia_IsoMan_t, i int32, value uint32) {
	p.pStoreW[i] = p.pStoreW[i]&0xffffffff | word(value)<<32
}
func Gia_IsoManStart(graph *Gia_Man_t) *Gia_IsoMan_t {
	n := Gia_ManObjNum(graph)
	return &Gia_IsoMan_t{pGia: graph, nObjs: n, nUniques: 1, nEntries: n, pLevels: make([]int32, n), pUniques: make([]int32, n), pStoreW: make([]word, n), vClasses: vec.Vec_IntAlloc(int(n / 4)), vClasses2: vec.Vec_IntAlloc(int(n / 4))}
}
func Gia_IsoManStop(p *Gia_IsoMan_t) {
	if p != nil {
		*p = Gia_IsoMan_t{}
	}
}
func Gia_IsoManTransferUnique(p *Gia_IsoMan_t) {
	for i, id := range p.pUniques {
		Gia_ManObj(p.pGia, int32(i)).Value = uint32(id)
	}
}
func Gia_IsoPrintClasses(p *Gia_IsoMan_t) {
	classes := vec.Vec_IntArray(p.vClasses)
	fmt.Printf("The total of %d classes:\n", len(classes)/2)
	for i := 0; i+1 < len(classes); i += 2 {
		fmt.Printf("%5d : (%3d,%3d)  \n", i/2, classes[i], classes[i+1])
	}
}
func Gia_IsoPrint(p *Gia_IsoMan_t, iter int32, elapsed int64) {
	fmt.Printf("Iter %4d :  Entries =%8d.  Uniques =%8d.  Singles =%8d.  %9.2f sec\n", iter, p.nEntries, p.nUniques, p.nSingles, float64(elapsed)/float64(time.Second))
}

// Gia_IsoPrepare creates initial equivalence classes based on graph levels.
func Gia_IsoPrepare(p *Gia_IsoMan_t) {
	var maxLevel int32
	for id := int32(1); id < p.nObjs; id++ {
		obj := Gia_ManObj(p.pGia, id)
		if Gia_ObjIsAnd(obj) != 0 {
			p.pLevels[id] = 1 + max(p.pLevels[Gia_ObjFaninId0(obj, id)], p.pLevels[Gia_ObjFaninId1(obj, id)])
		}
	}
	for i := int32(0); i < Gia_ManCoNum(p.pGia); i++ {
		obj := Gia_ManCo(p.pGia, i)
		id := Gia_ObjId(p.pGia, obj)
		p.pLevels[id] = 1 + p.pLevels[Gia_ObjFaninId0(obj, id)]
		maxLevel = max(maxLevel, p.pLevels[id])
	}
	sizes := make([]int32, maxLevel+1)
	for id := int32(1); id < p.nObjs; id++ {
		sizes[p.pLevels[id]]++
	}
	vec.Vec_IntClear(p.vClasses)
	vec.Vec_IntPush(p.vClasses, 0)
	vec.Vec_IntPush(p.vClasses, 1)
	begins := make([]int32, maxLevel+2)
	begins[0] = 1
	for level := int32(0); level <= maxLevel; level++ {
		if sizes[level] == 0 {
			panic("gia: isomorphism requires nonempty graph levels")
		}
		vec.Vec_IntPush(p.vClasses, begins[level])
		vec.Vec_IntPush(p.vClasses, sizes[level])
		begins[level+1] = begins[level] + sizes[level]
	}
	if begins[maxLevel+1] != p.nObjs {
		panic("gia: inconsistent isomorphism levels")
	}
	for id := int32(1); id < p.nObjs; id++ {
		level := p.pLevels[id]
		Gia_IsoSetItem(p, begins[level], uint32(id))
		begins[level]++
	}
}
func Gia_IsoAssignUnique(p *Gia_IsoMan_t) {
	p.nSingles = 0
	vec.Vec_IntClear(p.vClasses2)
	classes := vec.Vec_IntArray(p.vClasses)
	for i := 0; i+1 < len(classes); i += 2 {
		begin, size := classes[i], classes[i+1]
		if size == 1 {
			id := Gia_IsoGetItem(p, begin)
			if p.pUniques[id] != 0 {
				panic("gia: duplicate unique assignment")
			}
			p.pUniques[id] = p.nUniques
			p.nUniques++
			p.nSingles++
		} else {
			vec.Vec_IntPush(p.vClasses2, begin)
			vec.Vec_IntPush(p.vClasses2, size)
		}
	}
	p.vClasses, p.vClasses2 = p.vClasses2, p.vClasses
	p.nEntries -= p.nSingles
}

// Gia_IsoSort partitions each nonsingleton class by its current signature.
func Gia_IsoSort(p *Gia_IsoMan_t) int32 {
	p.nSingles = 0
	vec.Vec_IntClear(p.vClasses2)
	classes := vec.Vec_IntArray(p.vClasses)
	var refined int32
	add := func(begin, size int32) {
		if size == 1 {
			id := Gia_IsoGetItem(p, begin)
			if p.pUniques[id] != 0 {
				panic("gia: duplicate unique assignment")
			}
			p.pUniques[id] = p.nUniques
			p.nUniques++
			p.nSingles++
		} else {
			vec.Vec_IntPush(p.vClasses2, begin)
			vec.Vec_IntPush(p.vClasses2, size)
		}
	}
	for i := 0; i+1 < len(classes); i += 2 {
		begin, size := classes[i], classes[i+1]
		if size <= 1 {
			panic("gia: unexpected singleton class")
		}
		first := Gia_ManObj(p.pGia, int32(Gia_IsoGetItem(p, begin))).Value
		same := true
		for k := int32(0); k < size; k++ {
			value := Gia_ManObj(p.pGia, int32(Gia_IsoGetItem(p, begin+k))).Value
			Gia_IsoSetValue(p, begin+k, value)
			if value != first {
				same = false
			}
		}
		if same {
			add(begin, size)
			continue
		}
		refined = 1
		start := time.Now()
		abcutil.Abc_QuickSort3(p.pStoreW[begin:begin+size], size, 0)
		p.timeSort += time.Since(start).Nanoseconds()
		firstBegin := begin
		for k := int32(1); k < size; k++ {
			if Gia_IsoGetValue(p, begin+k) == Gia_IsoGetValue(p, begin+k-1) {
				continue
			}
			add(firstBegin, begin+k-firstBegin)
			firstBegin = begin + k
		}
		add(firstBegin, begin+size-firstBegin)
	}
	p.vClasses, p.vClasses2 = p.vClasses2, p.vClasses
	p.nEntries -= p.nSingles
	return refined
}
func Gia_IsoUpdateValue(value, complement int32) uint32 {
	return uint32(value+1) * s_256Primes[(uint32(value)*2+uint32(complement))&ISO_MASK]
}
func Gia_IsoUpdate(p *Gia_IsoMan_t, iteration, id, complement int32) uint32 {
	if iteration == 0 {
		return Gia_IsoUpdateValue(p.pLevels[id], complement)
	}
	if p.pUniques[id] > 0 {
		return Gia_IsoUpdateValue(p.pUniques[id], complement)
	}
	return 0
}

// Gia_IsoSimulate propagates structural signatures forward, including registers.
func Gia_IsoSimulate(p *Gia_IsoMan_t, iteration int32) {
	g := p.pGia
	Gia_ManConst0(g).Value += s_256Primes[ISO_MASK]
	for i := int32(0); i < Gia_ManPiNum(g); i++ {
		Gia_ManCi(g, i).Value += s_256Primes[ISO_MASK-1]
	}
	if iteration == 0 {
		for i := int32(0); i < Gia_ManRegNum(g); i++ {
			Gia_ManRo(g, i).Value += s_256Primes[ISO_MASK-2]
		}
	}
	for id := int32(1); id < Gia_ManObjNum(g); id++ {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		a, b := Gia_ObjFanin0(obj), Gia_ObjFanin1(obj)
		obj.Value += a.Value + Gia_IsoUpdate(p, iteration, Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj))
		obj.Value += b.Value + Gia_IsoUpdate(p, iteration, Gia_ObjFaninId1(obj, id), Gia_ObjFaninC1(obj))
	}
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		id := Gia_ObjId(g, obj)
		obj.Value += Gia_ObjFanin0(obj).Value + Gia_IsoUpdate(p, iteration, Gia_ObjFaninId0(obj, id), Gia_ObjFaninC0(obj))
	}
	for i := int32(0); i < Gia_ManRegNum(g); i++ {
		Gia_ManRo(g, i).Value += Gia_ManRi(g, i).Value
	}
}
func Gia_IsoSimulateBack(p *Gia_IsoMan_t, iteration int32) {
	g := p.pGia
	for i := int32(0); i < Gia_ManCoNum(g); i++ {
		obj := Gia_ManCo(g, i)
		id := Gia_ObjId(g, obj)
		Gia_ObjFanin0(obj).Value += obj.Value + Gia_IsoUpdate(p, iteration, id, Gia_ObjFaninC0(obj))
	}
	for id := Gia_ManObjNum(g) - 1; id > 0; id-- {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		Gia_ObjFanin0(obj).Value += obj.Value + Gia_IsoUpdate(p, iteration, id, Gia_ObjFaninC0(obj))
		Gia_ObjFanin1(obj).Value += obj.Value + Gia_IsoUpdate(p, iteration, id, Gia_ObjFaninC1(obj))
	}
	for i := int32(0); i < Gia_ManRegNum(g); i++ {
		Gia_ManRi(g, i).Value += Gia_ManRo(g, i).Value
	}
}

func Gia_IsoAssignOneClass(p *Gia_IsoMan_t, verbose int32) {
	classes := vec.Vec_IntArray(p.vClasses)
	if len(classes) == 0 {
		panic("gia: no isomorphism class to assign")
	}
	last := classes[len(classes)-2]
	level := p.pLevels[Gia_IsoGetItem(p, last)]
	i := len(classes) - 2
	for i >= 0 && p.pLevels[Gia_IsoGetItem(p, classes[i])] == level {
		i -= 2
	}
	start := i + 2
	for i = start; i < len(classes); i += 2 {
		begin, size := classes[i], classes[i+1]
		for k := int32(0); k < size; k++ {
			id := Gia_IsoGetItem(p, begin+k)
			if p.pUniques[id] != 0 {
				panic("gia: duplicate unique assignment")
			}
			p.pUniques[id] = p.nUniques
			p.nUniques++
			p.nSingles++
			p.nEntries--
		}
		if verbose != 0 {
			fmt.Printf("Broke ties in class of size %d at level %d.\n", size, level)
		}
	}
	vec.Vec_IntShrink(p.vClasses, start)
}
func Gia_IsoCollectCosClasses(p *Gia_IsoMan_t, verbose int32) *Vec_Ptr_t {
	var groups [][]int32
	for i := int32(0); i < Gia_ManPoNum(p.pGia); i++ {
		obj := Gia_ManPo(p.pGia, i)
		if p.pUniques[Gia_ObjId(p.pGia, obj)] > 0 {
			groups = append(groups, []int32{i})
		}
	}
	classes := vec.Vec_IntArray(p.vClasses)
	for i := 0; i+1 < len(classes); i += 2 {
		begin, size := classes[i], classes[i+1]
		var group []int32
		for k := int32(0); k < size; k++ {
			obj := Gia_ManObj(p.pGia, int32(Gia_IsoGetItem(p, begin+k)))
			if Gia_ObjIsPo(p.pGia, obj) != 0 {
				group = append(group, Gia_ObjCioId(obj))
			}
		}
		if len(group) > 0 {
			sort.Slice(group, func(i, j int) bool { return group[i] < group[j] })
			groups = append(groups, group)
		}
	}
	sort.Slice(groups, func(i, j int) bool { return groups[i][0] < groups[j][0] })
	result := vec.Vec_PtrAlloc(len(groups))
	for _, group := range groups {
		vec.Vec_PtrPush(result, vec.Vec_IntAllocArrayCopy(group, len(group)))
	}
	return result
}
func Gia_IsoRecognizeMuxes(g *Gia_Man_t) {
	for id := int32(1); id < Gia_ManObjNum(g); id++ {
		obj := Gia_ManObj(g, id)
		if Gia_ObjIsAnd(obj) == 0 || Gia_ObjIsMuxType(obj) == 0 {
			continue
		}
		var then, otherwise *Gia_Obj_t
		control := Gia_ObjRecognizeMux(obj, &then, &otherwise)
		if Gia_Regular(then) == Gia_Regular(otherwise) {
			obj.Value += s_256Primes[233]
			Gia_Regular(control).Value += s_256Primes[234]
			Gia_Regular(otherwise).Value += s_256Primes[234]
		} else {
			obj.Value += s_256Primes[235]
			Gia_Regular(control).Value += s_256Primes[236]
			Gia_Regular(otherwise).Value += s_256Primes[237]
			Gia_Regular(then).Value += s_256Primes[237]
		}
	}
}

// Gia_IsoDeriveEquivPos refines structural signatures and groups primary outputs.
func Gia_IsoDeriveEquivPos(g *Gia_Man_t, forward, verbose int32) *Vec_Ptr_t {
	if Gia_ManCiNum(g) == 0 || Gia_ManPoNum(g) == 0 {
		panic("gia: isomorphism needs inputs and outputs")
	}
	total := time.Now()
	Gia_ManCleanValue(g)
	p := Gia_IsoManStart(g)
	defer Gia_IsoManStop(p)
	Gia_IsoPrepare(p)
	Gia_IsoAssignUnique(p)
	p.timeStart = time.Since(total).Nanoseconds()
	if verbose != 0 {
		Gia_IsoPrint(p, 0, time.Since(total).Nanoseconds())
	}
	const maxIter = 10000
	iteration := int32(0)
	run := func(back bool, fixed int32) bool {
		var unchanged int32
		any := false
		for iteration < maxIter && unchanged < fixed {
			start := time.Now()
			if back {
				Gia_IsoSimulateBack(p, iteration)
			} else {
				Gia_IsoSimulate(p, iteration)
			}
			p.timeSim += time.Since(start).Nanoseconds()
			start = time.Now()
			refined := Gia_IsoSort(p)
			p.timeRefine += time.Since(start).Nanoseconds()
			iteration++
			if verbose != 0 {
				Gia_IsoPrint(p, iteration, time.Since(total).Nanoseconds())
			}
			if refined != 0 {
				unchanged = 0
				any = true
			} else {
				unchanged++
			}
		}
		return any
	}
	if forward != 0 {
		run(false, 2)
	} else {
		for Vec_IntSize(p.vClasses) > 0 {
			a, b := run(false, 2), run(true, 2)
			if !a && !b {
				break
			}
		}
		for Vec_IntSize(p.vClasses) > 0 {
			Gia_IsoAssignOneClass(p, verbose)
			for iteration < maxIter {
				a, b := run(true, 1), run(false, 1)
				if !a && !b {
					break
				}
			}
		}
	}
	var result *Vec_Ptr_t
	if Gia_ManPoNum(g) > 1 {
		result = Gia_IsoCollectCosClasses(p, verbose)
	}
	Gia_IsoManTransferUnique(p)
	return result
}

func isoAssignPartialClass(p *Gia_IsoMan_t, index int) (int32, int32) {
	classes := vec.Vec_IntArray(p.vClasses)
	begin, size := classes[index], classes[index+1]
	if size <= 1 {
		panic("gia: expected nonsingleton isomorphism class")
	}
	count := int32(1)
	if size == 2 {
		count = 2
	}
	for k := int32(0); k < count; k++ {
		id := Gia_IsoGetItem(p, begin+k)
		if p.pUniques[id] != 0 {
			panic("gia: duplicate unique assignment")
		}
		p.pUniques[id] = p.nUniques
		p.nUniques++
		p.nSingles++
		p.nEntries--
	}
	copy(classes[index:], classes[index+2:])
	vec.Vec_IntShrink(p.vClasses, len(classes)-2)
	return begin, size
}
func Gia_IsoAssignOneClass2(p *Gia_IsoMan_t) {
	classes := vec.Vec_IntArray(p.vClasses)
	if len(classes) == 0 {
		panic("gia: no isomorphism class")
	}
	index := len(classes) - 2
	for i := 0; i < len(classes); i += 2 {
		if classes[i+1] == 2 {
			index = i
			break
		}
	}
	begin, size := isoAssignPartialClass(p, index)
	fmt.Printf("Broke ties in class %d of size %d at level %d.\n", index/2, size, p.pLevels[Gia_IsoGetItem(p, begin)])
}
func Gia_IsoAssignOneClass3(p *Gia_IsoMan_t) {
	n := vec.Vec_IntSize(p.vClasses)
	if n == 0 {
		panic("gia: no isomorphism class")
	}
	begin, size := isoAssignPartialClass(p, n-2)
	fmt.Printf("Broke ties in last class of size %d at level %d.\n", size, p.pLevels[Gia_IsoGetItem(p, begin)])
}
func Gia_IsoReportTopmost(p *Gia_IsoMan_t) {
	g := p.pGia
	marked := make([]bool, Gia_ManObjNum(g))
	classes := vec.Vec_IntArray(p.vClasses)
	for i := 0; i+1 < len(classes); i += 2 {
		for k := int32(0); k < classes[i+1]; k++ {
			obj := Gia_ManObj(g, int32(Gia_IsoGetItem(p, classes[i]+k)))
			if Gia_ObjIsAnd(obj) != 0 {
				marked[Gia_ObjId(g, Gia_ObjFanin0(obj))] = true
				marked[Gia_ObjId(g, Gia_ObjFanin1(obj))] = true
			} else if Gia_ObjIsRo(g, obj) != 0 {
				marked[Gia_ObjId(g, Gia_ObjFanin0(Gia_ObjRoToRi(g, obj)))] = true
			}
		}
	}
	count := 0
	for i := 0; i+1 < len(classes); i += 2 {
		for k := int32(0); k < classes[i+1]; k++ {
			id := Gia_IsoGetItem(p, classes[i]+k)
			if !marked[id] {
				count++
				fmt.Printf("%5d : Obj %6d : Level = %4d.  iBegin = %4d.  Size = %4d.\n", count, id, p.pLevels[id], classes[i], classes[i+1])
				break
			}
		}
	}
}
