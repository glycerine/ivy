package gia

import (
	"github.com/glycerine/abcgo/bool/kit"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/cnf"
)

// Jf_ManGenCnf appends clauses for an output literal equivalent to the
// truth table over leaves. Clauses are represented by offsets into literals.
func Jf_ManGenCnf(truth word, output int32, leaves, literals, clauses, cover *Vec_Int_t) {
	if truth == 0 || ^truth == 0 {
		vec.Vec_IntPush(clauses, int32(vec.Vec_IntSize(literals)))
		if truth == 0 {
			output ^= 1
		}
		vec.Vec_IntPush(literals, output)
		return
	}
	n := vec.Vec_IntSize(leaves)
	if n < 1 || n > 6 {
		panic("JF CNF requires one to six leaves")
	}
	for phase := int32(0); phase < 2; phase++ {
		words := []uint32{uint32(truth), uint32(truth >> 32)}
		var cubes []uint32
		if kit.Kit_TruthIsop(words, int32(n), &cubes, 0) != 0 {
			panic("JF CNF cover derivation failed")
		}
		vec.Vec_IntClear(cover)
		for _, cube := range cubes {
			vec.Vec_IntPush(cover, int32(cube))
			vec.Vec_IntPush(clauses, int32(vec.Vec_IntSize(literals)))
			vec.Vec_IntPush(literals, output^phase)
			for k := 0; k < n; k++ {
				switch (cube >> uint(2*k)) & 3 {
				case 1:
					vec.Vec_IntPush(literals, vec.Vec_IntEntry(leaves, k))
				case 2:
					vec.Vec_IntPush(literals, vec.Vec_IntEntry(leaves, k)^1)
				case 3:
					panic("invalid JF cover cube")
				}
			}
		}
		truth = ^truth
	}
}

// jfClauses consumes the literal vector and uses the recorded offsets to
// reconstruct each clause, including empty clauses.
func jfClauses(literals, offsets *Vec_Int_t) [][]int32 {
	data := vec.Vec_IntReleaseArray(literals)
	starts := vec.Vec_IntArray(offsets)
	clauses := make([][]int32, len(starts))
	for i, start := range starts {
		end := len(data)
		if i+1 < len(starts) {
			end = int(starts[i+1])
		}
		if start < 0 || int(start) > end || end > len(data) {
			panic("invalid JF clause offsets")
		}
		clauses[i] = data[int(start):end]
	}
	return clauses
}

// Jf_ManCreateCnf uses GIA object IDs directly as SAT variable IDs.
func Jf_ManCreateCnf(p *Gia_Man_t, literals, offsets *Vec_Int_t) *cnf.CNF {
	clauses := jfClauses(literals, offsets)
	n := Gia_ManObjNum(p)
	first, count := make([]int32, n), make([]int32, n)
	for i := range first {
		first[i], count[i] = -1, -1
	}
	for i, clause := range clauses {
		if len(clause) == 0 {
			panic("JF mapped clause has no output literal")
		}
		id := clause[0] >> 1
		if id < 0 || id >= n {
			panic("JF clause has invalid output literal")
		}
		if first[id] < 0 {
			first[id], count[id] = int32(i), 1
		} else {
			count[id]++
		}
	}
	return cnf.NewMappedCNF(int(n), clauses, nil, first, count)
}

// Jf_ManCreateCnfRemap compacts used object IDs in reverse object order.
func Jf_ManCreateCnfRemap(p *Gia_Man_t, literals, offsets *Vec_Int_t, addOutputClause int32) *cnf.CNF {
	if addOutputClause != 0 {
		vec.Vec_IntPush(offsets, int32(vec.Vec_IntSize(literals)))
		for i := int32(0); i < Gia_ManPoNum(p); i++ {
			vec.Vec_IntPush(literals, Gia_ObjId(p, Gia_ManPo(p, i))<<1)
		}
	}
	n := Gia_ManObjNum(p)
	mapping := make([]int32, n)
	for i := range mapping {
		mapping[i] = -1
	}
	for _, lit := range vec.Vec_IntArray(literals) {
		id := lit >> 1
		if id < 0 || id >= n {
			panic("JF clause has invalid literal")
		}
		mapping[id] = 0
	}
	vars := int32(0)
	for i := n - 1; i >= 0; i-- {
		if mapping[i] == 0 {
			mapping[i] = vars
			vars++
		}
	}
	data := vec.Vec_IntArray(literals)
	for i, lit := range data {
		data[i] = mapping[lit>>1]<<1 | lit&1
	}
	return cnf.NewMappedCNF(int(vars), jfClauses(literals, offsets), mapping, nil, nil)
}
