package if_

import (
	"fmt"
	"slices"
	"strings"

	abcutil "github.com/glycerine/abcgo/misc/util"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

// The canonical DSD is a tree of shared, unlabeled structures. A variable
// occurrence consumes the next entry of the caller's signed input permutation.
func If_DsdObjAlloc(p *If_DsdMan_t, kind, fans int32) *If_DsdObj_t {
	o := &If_DsdObj_t{Id: int32(Vec_PtrSize(&p.vObjs)), Type: If_DsdType_t(kind), nFans: fans, fMark: p.fNewAsUseless, pFans: make([]int32, fans)}
	Vec_PtrPush(&p.vObjs, o)
	Vec_IntPush(&p.vNexts, 0)
	Vec_IntPush(&p.vTruths, -1)
	return o
}
func If_DsdObjTruthId(p *If_DsdMan_t, o *If_DsdObj_t) int32 {
	if o.Type == IF_DSD_PRIME && o.nFans > 2 {
		return Vec_IntEntry(&p.vTruths, int(o.Id))
	}
	return -1
}
func If_DsdObjTruth(p *If_DsdMan_t, o *If_DsdObj_t) []word {
	return vec.Vec_MemReadEntry(p.vTtMem[o.nFans], int(If_DsdObjTruthId(p, o)))
}
func If_DsdObjCompare(p *If_DsdMan_t, objects *Vec_Ptr_t, a, b int32) int32 {
	x, y := If_DsdVecObj(objects, a>>1), If_DsdVecObj(objects, b>>1)
	if x.Type < y.Type {
		return -1
	}
	if x.Type > y.Type {
		return 1
	}
	if x.Type < IF_DSD_AND {
		return 0
	}
	if x.nFans < y.nFans {
		return -1
	}
	if x.nFans > y.nFans {
		return 1
	}
	if x.Type == IF_DSD_PRIME {
		i, j := If_DsdObjTruthId(p, x), If_DsdObjTruthId(p, y)
		if i < j {
			return -1
		}
		if i > j {
			return 1
		}
	}
	for i := range x.pFans {
		if c := If_DsdObjCompare(p, objects, x.pFans[i], y.pFans[i]); c != 0 {
			return c
		}
	}
	if a&1 > b&1 {
		return -1
	}
	if a&1 < b&1 {
		return 1
	}
	assert(a == b, "canonical DSD structures are interned")
	return 0
}
func If_DsdObjSort(p *If_DsdMan_t, objects *Vec_Ptr_t, lits []int32, n int32, perm []int32) {
	for i := int32(0); i < n-1; i++ {
		best := i
		for j := i + 1; j < n; j++ {
			if If_DsdObjCompare(p, objects, lits[best], lits[j]) == 1 {
				best = j
			}
		}
		if i != best {
			lits[i], lits[best] = lits[best], lits[i]
			if perm != nil {
				perm[i], perm[best] = perm[best], perm[i]
			}
		}
	}
}
func ifDsdHashKey(p *If_DsdMan_t, kind int32, lits []int32, truthID int32) int {
	primes := [...]uint32{1049, 1297, 1559, 1823, 2089, 2371, 2663, 2909, 3221, 3517, 3779, 4073, 4363, 4663, 4973, 5281}
	hash := uint32(kind)*7873 + uint32(len(lits))*8147
	for i, lit := range lits {
		hash += uint32(lit) * primes[i&15]
	}
	if If_DsdType_t(kind) == IF_DSD_PRIME {
		hash += uint32(truthID) * primes[len(lits)&15]
	}
	return int(hash % uint32(p.nBins))
}
func If_DsdObjHashLookup(p *If_DsdMan_t, kind int32, lits []int32, n int32, truthID int32) *int32 {
	spot := &p.pBins[ifDsdHashKey(p, kind, lits[:n], truthID)]
	for *spot != 0 {
		o := If_DsdVecObj(&p.vObjs, *spot)
		if int32(o.Type) == kind && o.nFans == n && slices.Equal(o.pFans, lits[:n]) && If_DsdObjTruthId(p, o) == truthID {
			p.nUniqueHits++
			return spot
		}
		spot = &Vec_IntArray(&p.vNexts)[o.Id]
	}
	p.nUniqueMisses++
	return spot
}
func ifDsdHashResize(p *If_DsdMan_t) {
	misses := p.nUniqueMisses
	p.nBins = abcutil.Abc_PrimeCudd(uint32(2 * p.nBins))
	p.pBins = make([]int32, p.nBins)
	vec.Vec_IntFill(&p.vNexts, Vec_PtrSize(&p.vObjs), 0)
	for i := 2; i < Vec_PtrSize(&p.vObjs); i++ {
		o := If_DsdVecObj(&p.vObjs, int32(i))
		spot := If_DsdObjHashLookup(p, int32(o.Type), o.pFans, o.nFans, If_DsdObjTruthId(p, o))
		assert(*spot == 0, "resized canonical DSD entry is new")
		*spot = o.Id
	}
	assert(p.nUniqueMisses-misses == int32(Vec_PtrSize(&p.vObjs)-2), "all DSD entries rehashed")
	p.nUniqueMisses = misses
}
func If_DsdObjCreate(p *If_DsdMan_t, kind int32, lits []int32, n, truthID int32) int32 {
	assert(If_DsdType_t(kind) != IF_DSD_MUX || n == 3, "DSD mux has three fanins")
	o := If_DsdObjAlloc(p, kind, n)
	if o.Type == IF_DSD_PRIME {
		vec.Vec_IntWriteEntry(&p.vTruths, int(o.Id), truthID)
	}
	copy(o.pFans, lits[:n])
	for _, lit := range o.pFans {
		o.nSupp += If_DsdVecLitSuppSize(&p.vObjs, lit)
	}
	if p.LutSize != 0 && If_DsdManCheckXY(p, o.Id<<1, p.LutSize, 0, 0, 0, 0) == 0 {
		o.fMark = 1
	}
	return o.Id
}
func If_DsdObjFindOrAdd(p *If_DsdMan_t, kind int32, lits []int32, n int32, truth []word) int32 {
	truthID, previous := int32(-1), -1
	if If_DsdType_t(kind) == IF_DSD_PRIME {
		previous = vec.Vec_MemEntryNum(p.vTtMem[n])
		truthID = int32(vec.Vec_MemHashInsert(p.vTtMem[n], truth))
	}
	spot := If_DsdObjHashLookup(p, kind, lits, n, truthID)
	if *spot != 0 {
		return *spot
	}
	if p.LutSize != 0 && truthID >= 0 && int(truthID) == Vec_PtrSize(p.vTtDecs[n]) {
		var schedules [16][]int32
		copy(schedules[:], p.pSched[:])
		sets := dau.Dau_DecFindSets_int(truth, n, schedules)
		Vec_PtrPush(p.vTtDecs[n], sets)
	}
	if p.vIsops[n] != nil && truthID >= 0 && previous != vec.Vec_MemEntryNum(p.vTtMem[n]) {
		level := vec.Vec_WecPushLevel(p.vIsops[n])
		cover, complement := ifDsdPrimeCover(truth, n)
		if len(cover) <= 8 && complement >= 0 {
			for _, cube := range cover {
				Vec_IntPush(level, cube)
			}
			p.vIsopComplement[n] = append(p.vIsopComplement[n], complement != 0)
		} else {
			p.vIsopComplement[n] = append(p.vIsopComplement[n], false)
		}
	}
	*spot = int32(Vec_PtrSize(&p.vObjs))
	id := If_DsdObjCreate(p, kind, lits, n, truthID)
	if Vec_PtrSize(&p.vObjs) > int(p.nBins) {
		ifDsdHashResize(p)
	}
	return id
}
func If_DsdManComputeFirstArray(p *If_DsdMan_t, lits []int32, n int32, first []int32) int32 {
	var size int32
	for i := int32(0); i < n; i++ {
		first[i] = size
		size += If_DsdVecLitSuppSize(&p.vObjs, lits[i])
	}
	return size
}
func If_DsdManComputeFirst(p *If_DsdMan_t, o *If_DsdObj_t, first []int32) int32 {
	return If_DsdManComputeFirstArray(p, o.pFans, o.nFans, first)
}
func If_DsdManCheckInv_rec(p *If_DsdMan_t, lit int32) int32 {
	o := dsdObject(p, lit)
	switch o.Type {
	case IF_DSD_VAR:
		return 1
	case IF_DSD_AND, IF_DSD_PRIME:
		return 0
	case IF_DSD_XOR:
		for _, fan := range o.pFans {
			if If_DsdManCheckInv_rec(p, fan) != 0 {
				return 1
			}
		}
		return 0
	case IF_DSD_MUX:
		if If_DsdManCheckInv_rec(p, o.pFans[1]) != 0 && If_DsdManCheckInv_rec(p, o.pFans[2]) != 0 {
			return 1
		}
		return 0
	}
	panic("if: invalid DSD inversion node")
}
func If_DsdManPushInv_rec(p *If_DsdMan_t, lit int32, perm []byte) int32 {
	o := dsdObject(p, lit)
	switch o.Type {
	case IF_DSD_VAR:
		perm[0] ^= 1
	case IF_DSD_XOR:
		for _, fan := range o.pFans {
			if If_DsdManCheckInv_rec(p, fan) != 0 {
				If_DsdManPushInv_rec(p, fan, perm)
				break
			}
			perm = perm[If_DsdVecLitSuppSize(&p.vObjs, fan):]
		}
	case IF_DSD_MUX:
		assert(If_DsdManCheckInv_rec(p, o.pFans[1]) != 0 && If_DsdManCheckInv_rec(p, o.pFans[2]) != 0, "invertible DSD mux branches")
		perm = perm[If_DsdVecLitSuppSize(&p.vObjs, o.pFans[0]):]
		If_DsdManPushInv_rec(p, o.pFans[1], perm)
		perm = perm[If_DsdVecLitSuppSize(&p.vObjs, o.pFans[1]):]
		If_DsdManPushInv_rec(p, o.pFans[2], perm)
	default:
		panic("if: invalid DSD inversion push")
	}
	return 1
}
func If_DsdManPushInv(p *If_DsdMan_t, lit int32, perm []byte) int32 {
	if lit&1 != 0 && If_DsdManCheckInv_rec(p, lit) != 0 {
		return If_DsdManPushInv_rec(p, lit, perm)
	}
	return 0
}

func If_DsdManOperation(p *If_DsdMan_t, kind int32, lits []int32, n int32, perm []byte, truth []word) int32 {
	var children, intervals [16]int32
	var reordered [16]byte
	count, total, offset, complement := int32(0), int32(0), int32(0), int32(0)
	switch If_DsdType_t(kind) {
	case IF_DSD_AND, IF_DSD_XOR:
		for k := int32(0); k < n; k++ {
			if If_DsdType_t(kind) == IF_DSD_XOR && lits[k]&1 != 0 {
				lits[k] ^= 1
				complement ^= 1
			}
			o := dsdObject(p, lits[k])
			if o.Type == If_DsdType_t(kind) && (o.Type == IF_DSD_XOR || lits[k]&1 == 0) {
				for _, fan := range o.pFans {
					size := If_DsdVecLitSuppSize(&p.vObjs, fan)
					children[count], intervals[count] = fan, total<<16|(total+size)
					count++
					total += size
				}
			} else {
				children[count] = lits[k] ^ If_DsdManPushInv(p, lits[k], perm[offset:])
				intervals[count] = total<<16 | (total + o.nSupp)
				count++
				total += o.nSupp
			}
			offset += o.nSupp
		}
		If_DsdObjSort(p, &p.vObjs, children[:], count, intervals[:])
		j := 0
		for i := int32(0); i < count; i++ {
			for k := intervals[i] >> 16; k < intervals[i]&255; k++ {
				reordered[j] = perm[k]
				j++
			}
		}
		assert(int32(j) == total, "DSD reordered support")
		copy(perm[:total], reordered[:total])
	case IF_DSD_MUX:
		assert(n == 3, "DSD mux arity")
		for k := int32(0); k < n; k++ {
			o := dsdObject(p, lits[k])
			lits[k] ^= If_DsdManPushInv(p, lits[k], perm[offset:])
			offset += o.nSupp
		}
		compared := If_DsdObjCompare(p, &p.vObjs, lits[1], lits[2])
		if compared == 1 || compared == 0 && lits[0]&1 != 0 {
			a, b, c := If_DsdVecLitSuppSize(&p.vObjs, lits[0]), If_DsdVecLitSuppSize(&p.vObjs, lits[1]), If_DsdVecLitSuppSize(&p.vObjs, lits[2])
			lits[0] ^= 1
			lits[1], lits[2] = lits[2], lits[1]
			copy(reordered[:a], perm[:a])
			copy(reordered[a:a+c], perm[a+b:a+b+c])
			copy(reordered[a+c:a+c+b], perm[a:a+b])
			copy(perm[:a+b+c], reordered[:a+b+c])
		}
		if lits[1]&1 != 0 {
			lits[1] ^= 1
			lits[2] ^= 1
			complement ^= 1
		}
		offset = 0
		for k := int32(0); k < n; k++ {
			o := dsdObject(p, lits[k])
			children[count] = lits[k] ^ If_DsdManPushInv(p, lits[k], perm[offset:])
			count++
			offset += o.nSupp
		}
	case IF_DSD_PRIME:
		var canonical [16]byte
		var first [16]int32
		phase := dau.Abc_TtCanonicize(truth, n, canonical[:])
		complement = int32(phase >> uint(n) & 1)
		total = If_DsdManComputeFirstArray(p, lits, n, first[:])
		j := 0
		for i := int32(0); i < n; i++ {
			old := int32(canonical[i])
			lit := lits[old] ^ int32(phase>>uint(i)&1)
			o := dsdObject(p, lit)
			input := perm[first[old]:]
			children[count] = lit ^ If_DsdManPushInv(p, lit, input)
			count++
			j += copy(reordered[j:], input[:o.nSupp])
		}
		assert(int32(j) == total, "canonical prime support")
		copy(perm[:total], reordered[:total])
		abcutil.Abc_TtStretch6(truth, n, p.nVars)
	default:
		panic("if: invalid DSD operation")
	}
	id := If_DsdObjFindOrAdd(p, kind, children[:], count, truth)
	return id<<1 | complement
}

func If_DsdManAddDsd(p *If_DsdMan_t, text string, truth []word, perm []byte, support *int32) int32 {
	text = strings.TrimRight(text, "\x00")
	complement := int32(0)
	if strings.HasPrefix(text, "!") {
		text = text[1:]
		complement = 1
	}
	if text == "0" {
		return complement
	}
	if text == "1" {
		return complement ^ 1
	}
	if dau.Dau_DsdIsVar(text) != 0 {
		perm[*support] = byte(dau.Dau_DsdReadVar(text))
		*support++
		return 2 ^ complement
	}
	at := 0
	result := ifDsdAddRec(p, text, &at, truth, perm, support)
	assert(at == len(text)-1, "DSD parser consumed expression")
	return result ^ complement
}
func ifDsdAddRec(p *If_DsdMan_t, text string, at *int, truth []word, perm []byte, support *int32) int32 {
	start := *support
	complement := int32(0)
	if text[*at] == '!' {
		complement = 1
		*at++
	}
	c := text[*at]
	if c >= 'a' && c <= 'z' {
		perm[*support] = byte(2*(c-'a')) | byte(complement)
		*support++
		return 2
	}
	kind := int32(0)
	local := truth
	if c >= 'A' && c <= 'F' || c >= '0' && c <= '9' {
		local = make([]word, dsdWordCount(p.nVars))
		variables := abcutil.Abc_TtReadHex(local, text[*at:])
		digits := 1 << uint(max(variables-2, 0))
		*at += int(digits)
		assert(text[*at] == '{', "prime DSD opening brace")
		kind = int32(IF_DSD_PRIME)
	} else {
		switch c {
		case '(':
			kind = int32(IF_DSD_AND)
		case '[':
			kind = int32(IF_DSD_XOR)
		case '<':
			kind = int32(IF_DSD_MUX)
		case '{':
			kind = int32(IF_DSD_PRIME)
		default:
			panic(fmt.Sprintf("if: invalid DSD token %c", c))
		}
	}
	open := text[*at]
	close := byte(')')
	if open == '[' {
		close = ']'
	}
	if open == '<' {
		close = '>'
	}
	if open == '{' {
		close = '}'
	}
	*at++
	var lits [16]int32
	n := int32(0)
	for text[*at] != close {
		lits[n] = ifDsdAddRec(p, text, at, truth, perm, support)
		n++
		*at++
	}
	return If_DsdManOperation(p, kind, lits[:], n, perm[start:*support], local) ^ complement
}
