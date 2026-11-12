// Experimental DSD2 truth operations. The original file is an unfinished draft;
// its defined operations are retained and its trial loops use the proven DAU
// single/pair extraction engine instead of the source's non-progress TODOs.
package dau

import (
	"github.com/glycerine/abcgo/misc/util"
	"math/bits"
	"strings"
)

const DSD_MAX_VAR = 12
const DSD_MAX_WRD = 1 << (DSD_MAX_VAR - 6)

type Dua_Obj_t struct {
	Type  int32
	nFans int32
	pFans []int32
}
type Dua_Dsd_t struct {
	nSupp, nVars, nWords, nObjs, iRoot int32
	pObjs                              []Dua_Obj_t
	pTruth                             []uint64
	Expression                         string
	state                              *Dau_Dsd_t
	variables                          []int32
}

// The draft's selector is intentionally retained: nonzero takes cofactor zero,
// zero takes cofactor one. The returned 32 bits pack the other five variables.
func Abc_Tt6HalfUnShuffleVars(t uint64, variable, selector int32) uint64 {
	if variable < 0 || variable > 5 {
		panic("dau: half-unshuffle variable")
	}
	result := uint64(0)
	selected := 0
	if selector == 0 {
		selected = 1
	}
	for index := 0; index < 32; index++ {
		lower := index & ((1 << uint(variable)) - 1)
		source := lower | selected<<uint(variable) | (index>>uint(variable))<<uint(variable+1)
		if t>>uint(source)&1 != 0 {
			result |= 1 << uint(index)
		}
	}
	return result
}
func Abc_TtHalfUnShuffleVars(truth []uint64, n, i, j, selector int32) {
	if i == j {
		return
	}
	if i < 0 || i >= j || j >= n || len(truth) < dsWords(n) {
		panic("dau: invalid half-unshuffle range")
	}
	source := append([]uint64(nil), truth[:dsWords(n)]...)
	selected := 0
	if selector == 0 {
		selected = 1
	}
	// Only the destination cofactor with j=0 is specified by the source API.
	for index := 0; index < 64*dsWords(n); index++ {
		if index&(1<<uint(j)) != 0 {
			continue
		}
		packed := (index & ((1 << uint(j)) - 1)) | (index>>uint(j+1))<<uint(j)
		original := (packed & ((1 << uint(i)) - 1)) | selected<<uint(i) | (packed>>uint(i))<<uint(i+1)
		mask := uint64(1) << uint(index&63)
		truth[index>>6] &^= mask
		if dsBit(source, original) {
			truth[index>>6] |= mask
		}
	}
	if j < 6 {
		for w := 0; w < dsWords(n); w++ {
			low := truth[w] & ((uint64(1) << uint(1<<uint(j))) - 1)
			for shift := 1 << uint(j); shift < 64; shift *= 2 {
				low |= low << uint(shift)
			}
			truth[w] = low
		}
	}
}
func Dua_DsdInit(result *Dua_Dsd_t, truth []uint64, n int32) {
	if n < 0 || n > DSD_MAX_VAR || len(truth) < dsWords(n) {
		panic("dau: invalid DSD2 input")
	}
	*result = Dua_Dsd_t{nSupp: n, nVars: n, nWords: int32(dsWords(n)), nObjs: 1, pTruth: append([]uint64(nil), truth[:dsWords(n)]...)}
	root := Dua_Obj_t{Type: 5, nFans: n, pFans: make([]int32, n)}
	for i := range root.pFans {
		root.pFans[i] = int32(2 * i)
	}
	result.pObjs = []Dua_Obj_t{root}
}
func Dua_DsdTryConst(truth []uint64, n int32) int32 {
	if truth[0]&1 == 0 {
		return 0
	}
	dsComplement(truth[:dsWords(n)])
	return 1
}
func Dua_DsdTryVar(truth []uint64, nWords, variable int32) int32 {
	if nWords <= 0 || int(nWords) > len(truth) || variable < 0 || int64(1)<<uint(variable) >= int64(nWords)*64 {
		panic("dau: invalid DSD2 variable trial")
	}
	// Choose among the two input phases and output phases by the first unequal
	// cofactor words, matching the draft's explicitly written ordering tests.
	zero, one := append([]uint64(nil), truth[:nWords]...), append([]uint64(nil), truth[:nWords]...)
	util.Abc_TtCofactor0(zero, nWords, variable)
	util.Abc_TtCofactor1(one, nWords, variable)
	for word := range zero {
		a, b := zero[word], one[word]
		if a == b {
			continue
		}
		flip, complement := false, false
		if b < a && b < ^b {
			flip = true
		} else if ^b < a && ^b < b {
			flip, complement = true, true
		} else if word > 0 && b < a {
			flip = true
		}
		if flip {
			util.Abc_TtFlip(truth, nWords, variable)
		}
		if complement {
			dsComplement(truth[:nWords])
			return 1
		}
		return 0
	}
	return -1 // no dependence; the original draft asserts here
}
func Dua_DsdCheckCof0Const0(truth []uint64, nWords, variable int32) int32 {
	return util.Abc_TtCof0IsConst0(truth, nWords, variable)
}
func Dua_DsdCheckCofsEqualNot(truth []uint64, nWords, variable int32) int32 {
	return util.Abc_TtCofsOpposite(truth, nWords, variable)
}
func Dua_DsdTrySwap(truth []uint64, nWords, variable int32) int32 {
	if variable < 0 || nWords <= 0 || int(nWords) > len(truth) || int64(1)<<uint(variable+1) >= int64(nWords)*64 {
		panic("dau: invalid adjacent-variable swap")
	}
	if nWords&(nWords-1) != 0 {
		panic("dau: truth word count must be a power of two")
	}
	n := int32(6 + bits.Len32(uint32(nWords)) - 1)
	c01 := dsCof(dsCof(truth, n, variable, true), n, variable+1, false)
	c10 := dsCof(dsCof(truth, n, variable, false), n, variable+1, true)
	for word := range c01 {
		if c01[word] == c10[word] {
			continue
		}
		if c01[word] < c10[word] {
			return 0
		}
		util.Abc_TtSwapVars(truth, n, variable, variable+1)
		return 1
	}

	return 2
}
func Dua_DsdCheckDecomp(truth []uint64, nWords, variable int32) int32 {
	equal01, equal03 := true, true
	for index := 0; index < int(nWords)*64; index++ {
		if index&(3<<uint(variable)) != 0 {
			continue
		}
		a := dsBit(truth, index)
		equal01 = equal01 && a == dsBit(truth, index|1<<uint(variable))
		equal03 = equal03 && a == dsBit(truth, index|3<<uint(variable))
		if !equal01 && !equal03 {
			return 0
		}
	}
	if equal01 {
		return 1
	}
	return 2
}

// RangeVars recognizes a common nonconstant binary partition of the bound
// assignments, for every assignment above the range. The return word is its
// local truth table. Mutation commits only after the entire check succeeds.
func Dua_DsdRangeVars(truth []uint64, n, first, last, perform int32) uint64 {
	if first < 0 || last > n || last-first < 3 || last-first > 6 || len(truth) < dsWords(n) {
		panic("dau: invalid bound-variable range")
	}
	free, bound, parts := 1<<uint(first), 1<<uint(last-first), 1<<uint(n-last)
	type cofactors struct{ zero, one []bool }
	all := make([]cofactors, parts)
	partition := uint64(0)
	equal := func(a, b []bool) bool {
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	for part := 0; part < parts; part++ {
		mask := uint64(0)
		for assignment := 0; assignment < bound; assignment++ {
			cofactor := make([]bool, free)
			for low := range cofactor {
				cofactor[low] = dsBit(truth, part<<uint(last)|assignment<<uint(first)|low)
			}
			if all[part].zero == nil {
				all[part].zero = cofactor
			} else if equal(all[part].zero, cofactor) {
				continue
			} else if all[part].one == nil {
				all[part].one = cofactor
				mask |= 1 << uint(assignment)
			} else if equal(all[part].one, cofactor) {
				mask |= 1 << uint(assignment)
			} else {
				return 0
			}
		}
		if all[part].one == nil {
			return 0
		}
		if part == 0 {
			partition = mask
		} else if partition != mask {
			return 0
		}
	}
	if perform != 0 {
		reduced := n - (last - first) + 1
		result := make([]uint64, dsWords(reduced))
		for index := 0; index < 64*len(result); index++ {
			assignment := index & ((1 << uint(reduced)) - 1)
			part := assignment >> uint(first+1)
			low := assignment & (free - 1)
			cofactor := all[part].zero
			if assignment&(1<<uint(first)) != 0 {
				cofactor = all[part].one
			}
			if cofactor[low] {
				dsSet(result, index)
			}
		}
		copy(truth, result)
	}
	return partition
}
func Dua_DsdRangeVars0(truth []uint64, n, bound, perform int32) int32 {
	if bound <= 2 || bound >= n || len(truth) < dsWords(n) {
		panic("dau: invalid low bound set")
	}
	width := 1 << uint(bound)
	parts := 1 << uint(n-bound)
	for part := 1; part < parts; part++ {
		same, opposite := true, true
		for index := 0; index < width; index++ {
			a, b := dsBit(truth, index), dsBit(truth, part*width+index)
			same = same && a == b
			opposite = opposite && a != b
		}
		if !same && !opposite {
			return 0
		}
	}
	return 1
}
func Dua_DsdRangeVars0Derive(truth []uint64, n, bound int32) {
	if Dua_DsdRangeVars0(truth, n, bound, 0) == 0 {
		panic("dau: low bound set is not decomposable")
	}
	// Derive the outer XOR selector: a cofactor differs from the first iff its
	// phase bit is one. The bound function is the original first cofactor.
	variables := n - bound
	result := make([]uint64, dsWords(variables))
	first := dsBit(truth, 0)
	for index := 0; index < 64*len(result); index++ {
		part := index & ((1 << uint(variables)) - 1)
		if dsBit(truth, part<<uint(bound)) != first {
			dsSet(result, index)
		}
	}
	copy(truth, result)
}
func (p *Dua_Dsd_t) prepare() {
	if p.state != nil {
		return
	}
	p.state = &Dau_Dsd_t{fWriteTruth: 1}
	dsInitialize(p.state, p.nSupp)
	p.variables = make([]int32, p.nSupp)
	p.nVars = Dau_DsdMinBase(p.pTruth, p.nSupp, p.variables)
}
func (p *Dua_Dsd_t) record() {
	// Complete the current extraction with its actual residual prime for an
	// inspectable, truth-equivalent trial result; preserve the extraction state.
	copyState := *p.state
	copyState.output = strings.Builder{}
	copyState.output.WriteString(p.state.output.String())
	copyState.fSplitPrime = 0
	if p.nVars > 1 {
		dsPrime(&copyState, p.pTruth, p.variables, p.nVars)
	} else if p.nVars == 1 {
		copyState.writeVar(p.variables[0], p.pTruth[0]&1 != 0)
	} else if p.state.output.Len() == 0 {
		if p.pTruth[0]&1 != 0 {
			copyState.output.WriteByte('1')
		} else {
			copyState.output.WriteByte('0')
		}
	}
	for i := copyState.nConsts - 1; i >= 0; i-- {
		if copyState.uConstMask&(1<<uint(i)) != 0 {
			copyState.output.WriteByte(']')
		} else {
			copyState.output.WriteByte(')')
		}
	}
	p.Expression = dsFlatten(dsParse(copyState.output.String())).text()
	p.nWords = int32(dsWords(p.nVars))
	p.pObjs = nil
	var add func(*dsExpr) int32
	add = func(e *dsExpr) int32 {
		fans := make([]int32, len(e.children))
		for i, c := range e.children {
			fans[i] = add(c)
		}
		kind := int32(5)
		switch e.kind {
		case 'v':
			kind = 1
			fans = []int32{int32(e.variable)}
		case '(':
			kind = 2
		case '[':
			kind = 3
		case '<':
			kind = 4
		}
		id := int32(len(p.pObjs))
		p.pObjs = append(p.pObjs, Dua_Obj_t{Type: kind, nFans: int32(len(fans)), pFans: fans})
		phase := int32(0)
		if e.neg {
			phase = 1
		}
		return 2*id + phase
	}
	p.iRoot = add(dsParse(p.Expression))
	p.nObjs = int32(len(p.pObjs))
}

func Dua_DsdOneVar(p *Dua_Dsd_t) int32 {
	p.prepare()
	if p.nVars > 1 {
		p.nVars = dsSingle(p.state, p.pTruth, p.variables, p.nVars)
	}
	p.record()
	return int32(p.pTruth[0] & 1)
}
func Dua_DsdTwoVars(p *Dua_Dsd_t) int32 {
	p.prepare()
	previous := p.nVars
	if p.nVars > 1 {
		p.nVars = dsDouble(p.state, p.pTruth, p.variables, p.nVars)
	}
	p.record()
	if p.nVars < previous {
		return 1
	}
	return 0
}
func Dua_DsdTest(truth []uint64, n int32) *Dua_Dsd_t {
	p := new(Dua_Dsd_t)
	Dua_DsdInit(p, truth, n)
	Dua_DsdOneVar(p)
	Dua_DsdTwoVars(p)
	return p
}
