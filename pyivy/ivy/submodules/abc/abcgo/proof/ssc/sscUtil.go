package ssc

import (
	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/sat/bsat"
	"github.com/glycerine/abcgo/sat/satoko"
)

// builder owns structural hashing; no GIA-global scratch table is borrowed.
type builder struct {
	graph *gia.Gia_Man_t
	ands  map[[2]int32]int32
}

func newBuilder(inputs, capacity int) *builder {
	b := &builder{gia.Gia_ManStart(int32(capacity)), make(map[[2]int32]int32)}
	for i := 0; i < inputs; i++ {
		gia.Gia_ManAppendCi(b.graph)
	}
	return b
}
func (b *builder) and(a, c int32) int32 {
	if a == 0 || c == 0 || a == c^1 {
		return 0
	}
	if a == 1 {
		return c
	}
	if c == 1 || a == c {
		return a
	}
	if a > c {
		a, c = c, a
	}
	key := [2]int32{a, c}
	if l, ok := b.ands[key]; ok {
		return l
	}
	l := gia.Gia_ManAppendAnd(b.graph, a, c)
	b.ands[key] = l
	return l
}
func (b *builder) or(a, c int32) int32     { return b.and(a^1, c^1) ^ 1 }
func (b *builder) xor(a, c int32) int32    { return b.or(b.and(a, c^1), b.and(a^1, c)) }
func (b *builder) mux(c, t, e int32) int32 { return b.or(b.and(c, t), b.and(c^1, e)) }

// copySelected copies only selected cones, sharing inputs by interface position.
func copySelected(dst *builder, src *gia.Gia_Man_t, outputs, mapping []int32) []int32 {
	cache := make([]int32, src.ObjectCount())
	for i := range cache {
		cache[i] = -1
	}
	cache[0] = 0
	inputs := dst.graph.InputLiterals()
	for i, l := range src.InputLiterals() {
		cache[l>>1] = inputs[i]
	}
	var visit func(int32) int32
	visit = func(l int32) int32 {
		id := l >> 1
		if mapping != nil && mapping[id] != id<<1 {
			return visit(mapping[id]) ^ (l & 1)
		}
		if cache[id] < 0 {
			a, b, c, k := src.Gate(id)
			switch k {
			case 2:
				cache[id] = dst.and(visit(a), visit(b))
			case 3:
				cache[id] = dst.xor(visit(a), visit(b))
			case 4:
				cache[id] = dst.mux(visit(a), visit(b), visit(c))
			case 5, 6:
				cache[id] = visit(a)
			default:
				panic("ssc: invalid selected cone")
			}
		}
		return cache[id] ^ (l & 1)
	}
	out := make([]int32, len(outputs))
	for i, l := range outputs {
		out[i] = visit(l)
	}
	return out
}
func copyInto(dst *builder, src *gia.Gia_Man_t, mapping []int32) []int32 {
	return copySelected(dst, src, src.OutputLiterals(), mapping)
}

// copySubjectIntoFraig creates mappings for subject objects when a caller
// needs SAT queries outside the incremental sweep.
func copySubjectIntoFraig(p *Ssc_Man_t) []int32 {
	literals := make([]int32, 0, p.Original.ObjectCount())
	ids := make([]int32, 0, p.Original.ObjectCount())
	for id := int32(0); int(id) < p.Original.ObjectCount(); id++ {
		_, _, _, kind := p.Original.Gate(id)
		if id == 0 || kind >= 1 && kind <= 6 {
			ids = append(ids, id)
			literals = append(literals, id<<1)
		}
	}
	result := make([]int32, p.Original.ObjectCount())
	for i, lit := range copySelected(p.fraig, p.Original, literals, nil) {
		result[ids[i]] = lit
	}
	return result
}
func copyMapped(src *gia.Gia_Man_t, outputs, mapping []int32) *gia.Gia_Man_t {
	return copyWithRegs(src, outputs, mapping, src.RegisterCount())
}
func copyWithRegs(src *gia.Gia_Man_t, outputs, mapping []int32, regs int) *gia.Gia_Man_t {
	b := newBuilder(len(src.InputLiterals()), src.ObjectCount())
	b.graph.CopyIdentityFrom(src)
	for _, l := range copySelected(b, src, outputs, mapping) {
		gia.Gia_ManAppendCo(b.graph, l)
	}
	if regs > len(outputs) {
		regs = 0
	}
	gia.Gia_ManSetRegNum(b.graph, int32(regs))
	return b.graph
}
func copyOutputs(src *gia.Gia_Man_t, outputs []int32, zero bool) *gia.Gia_Man_t {
	if outputs == nil {
		outputs = src.OutputLiterals()
	}
	if zero {
		outputs = make([]int32, len(outputs))
	}
	return copyMapped(src, outputs, nil)
}
func Gia_ManDropContained(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	m, status := Ssc_ManStart(p, emptyCare(p), nil)
	if status != satoko.SATOKO_SAT {
		return nil
	}
	mapping := copySubjectIntoFraig(m)
	// Only primary outputs participate in containment. Register inputs are
	// transition functions, not members of the output union.
	outputs := p.OutputLiterals()[:int(gia.Gia_ManPoNum(p))]
	keep := make([]bool, len(outputs))
	for i := range keep {
		keep[i] = true
	}
	for i := range outputs {
		var assumptions []bsat.Literal
		for j, l := range outputs {
			if !keep[j] {
				continue
			}
			v := m.encode(0, mapping[l>>1]^(l&1))
			if j != i {
				v ^= 1
			}
			assumptions = append(assumptions, bsat.Literal(v))
		}
		m.sat.SetLimits(0, 0)
		if sscStatus(m.sat.Solve(assumptions...)) == satoko.SATOKO_UNSAT {
			keep[i] = false
		}
	}
	selected := make([]int32, 0, len(outputs))
	for i := range outputs {
		if keep[i] {
			selected = append(selected, int32(i))
		}
	}
	if len(selected) == len(outputs) {
		return gia.Gia_ManDup(p)
	}
	// Following the selected PO cones also retains the transitive register
	// dependencies and their corresponding next-state functions.
	return gia.Gia_ManDupCones(p, selected, int32(len(selected)), 0)
}
func Gia_ManOptimizeRing(p *gia.Gia_Man_t) *gia.Gia_Man_t {
	if gia.Gia_ManConstrNum(p) != 0 || p.RegisterCount() != 0 {
		panic("ssc: ring optimization expects combinational unconstrained outputs")
	}
	q := Gia_ManDropContained(p)
	if q == nil || len(q.OutputLiterals()) < 2 {
		return q
	}
	params := DefaultParams()
	params.Append = true
	n := len(q.OutputLiterals())
	for i := 0; i < n; i++ {
		outputs := q.OutputLiterals()
		outputs[0], outputs[i] = outputs[i], outputs[0]
		q = copyOutputs(q, outputs, false)
		gia.Gia_ManSetConstrNum(q, int32(n-1))
		q = Ssc_PerformSweepingConstr(q, &params)
		outputs = q.OutputLiterals()
		outputs[0], outputs[i] = outputs[i], outputs[0]
		q = copyOutputs(q, outputs, false)
	}
	gia.Gia_ManSetConstrNum(q, 0)
	return q
}
