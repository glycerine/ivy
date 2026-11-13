package gia

import "github.com/glycerine/abcgo/misc/vec"

func Gia_ManOrigIdsInit(p *Gia_Man_t)  { p.vIdsOrig = vec.Vec_IntStartNatural(int(p.nObjs)) }
func Gia_ManOrigIdsStart(p *Gia_Man_t) { p.vIdsOrig = vec.Vec_IntStartFull(int(p.nObjs)) }

// Gia_ManOrigIds returns the current object's original-ID mapping.
func Gia_ManOrigIds(p *Gia_Man_t) *Vec_Int_t { return p.vIdsOrig }

// Gia_ManOrigEquivPairs returns the original-ID pairs proved by SAT sweeping.
func Gia_ManOrigEquivPairs(p *Gia_Man_t) *Vec_Int_t { return p.vIdsEquiv }
func Gia_ManOrigEquivPairsReset(p *Gia_Man_t) {
	p.vIdsEquiv = vec.Vec_IntAlloc(1000)
}

// Gia_ManOrigIdsSet installs a mapping; -1 denotes a synthetic node.
func Gia_ManOrigIdsSet(p *Gia_Man_t, ids []int32) {
	if len(ids) != int(p.nObjs) {
		panic("gia: original ID mapping size mismatch")
	}
	p.vIdsOrig = vec.Vec_IntAlloc(len(ids))
	for _, id := range ids {
		vec.Vec_IntPush(p.vIdsOrig, id)
	}
}
func Gia_ManOrigIdsRemap(p, q *Gia_Man_t) {
	if p.vIdsOrig == nil {
		return
	}
	Gia_ManOrigIdsStart(q)
	vec.Vec_IntWriteEntry(q.vIdsOrig, 0, 0)
	for _, obj := range p.pObjs[1:] {
		lit := Gia_ObjValue(obj)
		if lit < 0 || lit>>1 == 0 {
			continue
		}
		old := vec.Vec_IntEntry(p.vIdsOrig, int(obj.id))
		if old != -1 && vec.Vec_IntEntry(q.vIdsOrig, int(lit>>1)) == -1 {
			vec.Vec_IntWriteEntry(q.vIdsOrig, int(lit>>1), old)
		}
	}
}
func Gia_ManOrigIdsRemapPairsInsert(mapping *Vec_Int_t, a, b int32) {
	for a != b {
		small, big := min(a, b), max(a, b)
		old := vec.Vec_IntEntry(mapping, int(big))
		if old < 0 {
			vec.Vec_IntWriteEntry(mapping, int(big), small)
			return
		}
		a, b = small, old
	}
}
func Gia_ManOrigIdsRemapPairsExtract(mapping *Vec_Int_t, id int32) int32 {
	for {
		next := vec.Vec_IntEntry(mapping, int(id))
		if next < 0 {
			return id
		}
		id = next
	}
}
func Gia_ManOrigIdsRemapPairs(pairs *Vec_Int_t, nObjs int32) *Vec_Int_t {
	entries := vec.Vec_IntArray(pairs)
	if len(entries)%2 != 0 {
		panic("gia: incomplete equivalence pair")
	}
	links := vec.Vec_IntStartFull(int(nObjs))
	for i := 0; i < len(entries); i += 2 {
		Gia_ManOrigIdsRemapPairsInsert(links, entries[i], entries[i+1])
	}
	result := vec.Vec_IntStartFull(int(nObjs))
	for id, parent := range vec.Vec_IntArray(links) {
		if parent >= 0 {
			vec.Vec_IntWriteEntry(result, id, Gia_ManOrigIdsRemapPairsExtract(links, parent))
		}
	}
	return result
}
func Gia_ManOrigIdsReduce(p *Gia_Man_t, pairs *Vec_Int_t) *Gia_Man_t {
	mapping := Gia_ManOrigIdsRemapPairs(pairs, p.nObjs)
	Gia_ManSetPhase(p)
	q := Gia_ManStart(p.nObjs)
	q.pName = p.pName
	q.pSpec = p.pSpec
	Gia_ManFillValue(p)
	Gia_ObjSetValue(Gia_ManConst0(p), 0)
	for i := int32(0); i < Gia_ManCiNum(p); i++ {
		Gia_ObjSetValue(Gia_ManCi(p, i), Gia_ManAppendCi(q))
	}
	Gia_ManHashAlloc(q)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		repr := vec.Vec_IntEntry(mapping, int(obj.id))
		if repr < 0 {
			a, b, c, kind := p.Gate(obj.id)
			var lit int32
			switch kind {
			case 2:
				lit = Gia_ManHashAnd(q, Gia_ObjLitCopy(p, a), Gia_ObjLitCopy(p, b))
			case 3:
				lit = Gia_ManHashXor(q, Gia_ObjLitCopy(p, a), Gia_ObjLitCopy(p, b))
			case 4:
				lit = Gia_ManHashMux(q, Gia_ObjLitCopy(p, a), Gia_ObjLitCopy(p, b), Gia_ObjLitCopy(p, c))
			case 5:
				lit = Gia_ObjLitCopy(p, a)
			default:
				panic("gia: unsupported original-ID reduction node")
			}
			Gia_ObjSetValue(obj, lit)
		} else {
			other := Gia_ManObj(p, repr)
			Gia_ObjSetValue(obj, Gia_ObjValue(other)^int32(other.fPhase^obj.fPhase))
		}
	}
	Gia_ManHashStop(q)
	for i := int32(0); i < Gia_ManCoNum(p); i++ {
		obj := Gia_ManCo(p, i)
		Gia_ObjSetValue(obj, Gia_ManAppendCo(q, Gia_ObjFanin0Copy(obj)))
	}
	Gia_ManSetRegNum(q, Gia_ManRegNum(p))
	Gia_ManSetConstrNum(q, Gia_ManConstrNum(p))
	if p.pReprs != nil || p.pNexts != nil {
		panic("gia: existing equivalence classes")
	}
	p.pReprs = make([]Gia_Rpr_t, p.nObjs)
	for i := range p.pReprs {
		p.pReprs[i].iRepr = GIA_VOID
	}
	Gia_ManFillValue(q)
	for _, obj := range p.pObjs {
		if Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		repr := Gia_ObjValue(obj) >> 1
		if repr == 0 {
			Gia_ObjSetRepr(p, obj.id, 0)
			continue
		}
		other := Gia_ManObj(q, repr)
		if Gia_ObjValue(other) == -1 {
			Gia_ObjSetValue(other, obj.id)
		} else {
			Gia_ObjSetRepr(p, obj.id, Gia_ObjValue(other))
		}
	}
	p.pNexts = Gia_ManDeriveNexts(p)
	return q
}
