package gia

import (
	"encoding/binary"
	"fmt"
	"math/rand"
)

type sweeperCandidate struct {
	id       int32
	polarity bool
}

// Gia_SweeperSweep rewrites probe cones using equivalences proved under the
// current conditions. Random simulation filters candidates; SAT validates each
// replacement, so simulation collisions cannot change the result.
func Gia_SweeperSweep(p *Gia_Man_t, outputs *Vec_Int_t, words, conflicts, verify, verbose int32) *Gia_Man_t {
	sweeper(p)
	if words <= 0 || conflicts < 0 {
		panic("gia: invalid conditional sweep limits")
	}
	subject := Gia_SweeperExtractUserLogic(p, outputs, nil, nil)
	care := Gia_SweeperExtractUserLogic(p, Gia_SweeperCondVector(p), nil, nil)
	count := Gia_ManObjNum(subject)
	numOutputs := Gia_ManPoNum(subject)
	Gia_ManDupAppendShare(subject, care)
	Gia_ManStop(care)
	work := Gia_SweeperStart(subject)
	Gia_SweeperSetConflictLimit(work, conflicts)
	for i := numOutputs; i < Gia_ManPoNum(work); i++ {
		Gia_SweeperCondPush(work, Gia_SweeperProbeCreate(work, Gia_ObjFaninLit0p(work, Gia_ManPo(work, i))))
	}
	// One simulation word per requested batch. Every signature is masked by
	// the same care bits; complements are canonicalized for candidate lookup.
	random := rand.New(rand.NewSource(1))
	values := make([][]uint64, work.nObjs)
	for i := range values {
		values[i] = make([]uint64, words)
	}
	for i := int32(0); i < Gia_ManPiNum(work); i++ {
		for w := range values[Gia_ManPi(work, i).id] {
			values[Gia_ManPi(work, i).id][w] = random.Uint64()
		}
	}
	wordLit := func(lit int32, w int) uint64 {
		v := values[lit>>1][w]
		if lit&1 != 0 {
			return ^v
		}
		return v
	}
	for _, obj := range work.pObjs[1:] {
		if Gia_ObjIsAnd(obj) != 0 {
			for w := 0; w < int(words); w++ {
				values[obj.id][w] = wordLit(Gia_ObjFaninLit0(obj, obj.id), w) & wordLit(Gia_ObjFaninLit1(obj, obj.id), w)
			}
		}
	}
	mask := make([]uint64, words)
	for w := range mask {
		mask[w] = ^uint64(0)
		for i := numOutputs; i < Gia_ManPoNum(work); i++ {
			mask[w] &= ^wordLit(Gia_ObjFaninLit0p(work, Gia_ManPo(work, i)), w)
		}
	}
	groups := make(map[string][]sweeperCandidate)
	reprs := make([]int32, count)
	probes := make([][2]int32, count)
	for i := range probes {
		probes[i] = [2]int32{-1, -1}
	}
	probe := func(id int32, negative bool) int32 {
		phase := 0
		if negative {
			phase = 1
		}
		if probes[id][phase] < 0 {
			probes[id][phase] = Gia_SweeperProbeCreate(work, id<<1|int32(phase))
		}
		return probes[id][phase]
	}
	for id := int32(0); id < count; id++ {
		obj := Gia_ManObj(work, id)
		if id != 0 && Gia_ObjIsCi(obj) == 0 && Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		normal, inverse := make([]byte, 8*int(words)), make([]byte, 8*int(words))
		for w := range mask {
			sig := values[id][w] & mask[w]
			binary.LittleEndian.PutUint64(normal[8*w:], sig)
			binary.LittleEndian.PutUint64(inverse[8*w:], ^sig&mask[w])
		}
		polarity := string(inverse) < string(normal)
		key := string(normal)
		if polarity {
			key = string(inverse)
		}
		reprs[id] = id << 1
		for _, candidate := range groups[key] {
			complement := polarity != candidate.polarity
			if Gia_SweeperCheckEquiv(work, probe(id, false), probe(candidate.id, complement)) == 1 {
				reprs[id] = candidate.id<<1 | giaBool(complement)
				break
			}
		}
		if reprs[id] == id<<1 {
			groups[key] = append(groups[key], sweeperCandidate{id, polarity})
		}
	}
	result := Gia_ManStart(count + 100)
	result.pName, result.pSpec = p.pName, p.pSpec
	mapped := make([]int32, count)
	Gia_ManHashStart(result)
	for i := int32(0); i < Gia_ManPiNum(work); i++ {
		obj := Gia_ManPi(work, i)
		mapped[obj.id] = Gia_ManAppendCi(result)
	}
	for id := int32(1); id < count; id++ {
		obj := Gia_ManObj(work, id)
		if Gia_ObjIsCi(obj) == 0 && Gia_ObjIsAnd(obj) == 0 {
			continue
		}
		if reprs[id] != id<<1 {
			mapped[id] = mapped[reprs[id]>>1] ^ (reprs[id] & 1)
			continue
		}
		if Gia_ObjIsAnd(obj) != 0 {
			a, b := Gia_ObjFaninLit0(obj, id), Gia_ObjFaninLit1(obj, id)
			mapped[id] = Gia_ManHashAnd(result, mapped[a>>1]^(a&1), mapped[b>>1]^(b&1))
		}
	}
	for i := int32(0); i < numOutputs; i++ {
		lit := Gia_ObjFaninLit0p(work, Gia_ManPo(work, i))
		Gia_ManAppendCo(result, mapped[lit>>1]^(lit&1))
	}
	Gia_ManHashStop(result)
	if verify != 0 {
		actual := Gia_SweeperGraft(work, nil, result)
		for i, lit := range Vec_IntArray(actual) {
			want := Gia_ObjFaninLit0p(work, Gia_ManPo(work, int32(i)))
			if Gia_SweeperCheckEquiv(work, Gia_SweeperProbeCreate(work, want), Gia_SweeperProbeCreate(work, lit)) != 1 {
				panic("gia: conditional sweep verification failed")
			}
		}
	}
	if verbose != 0 {
		fmt.Printf("Conditional sweep: %d ANDs -> %d ANDs.\n", Gia_ManAndNum(p), Gia_ManAndNum(result))
	}
	Gia_SweeperStop(work)
	Gia_ManStop(work)
	return result
}
