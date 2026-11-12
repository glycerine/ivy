package bsat2

/***********************************************************************************[SolverTypes.h]
Copyright (c) 2003-2006, Niklas Een, Niklas Sorensson
Copyright (c) 2007-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

type Var int32
type Lit int32

const VarUndef Var = -1
const LitUndef Lit = -2
const LitError Lit = -1

func MakeLit(v Var, negative bool) Lit {
	if v < 0 || v > 0x3fffffff {
		panic("invalid SAT variable")
	}
	l := Lit(v) << 1
	if negative {
		l |= 1
	}
	return l
}
func (l Lit) Not() Lit   { return l ^ 1 }
func (l Lit) Var() Var   { return Var(l >> 1) }
func (l Lit) Sign() bool { return l&1 != 0 }

type LBool uint8

const (
	True LBool = iota
	False
	Undef
)

func (v LBool) Not() LBool {
	if v >= Undef {
		return Undef
	}
	return v ^ 1
}
func (v LBool) Xor(negative bool) LBool {
	if negative {
		return v.Not()
	}
	return v
}
func (v LBool) And(w LBool) LBool {
	return normalizeBool(LBool((uint32(0xF7F755F4) >> ((uint(v) << 1) | (uint(w) << 3))) & 3))
}
func (v LBool) Or(w LBool) LBool {
	return normalizeBool(LBool((uint32(0xFCFCF400) >> ((uint(v) << 1) | (uint(w) << 3))) & 3))
}

type CRef = Ref

const CRefUndef = RefUndef

type Clause struct {
	Lits             []Lit
	Learnt, HasExtra bool
	Mark             uint8
	Activity         float32
	Abstraction      uint32
	relocated        bool
	relocation       CRef
}

func (c *Clause) CalcAbstraction() {
	if !c.HasExtra {
		panic("clause has no abstraction")
	}
	c.Abstraction = 0
	for _, l := range c.Lits {
		c.Abstraction |= 1 << uint(l.Var()&31)
	}
}
func (c *Clause) Shrink(count int) {
	if count < 0 || count > len(c.Lits) {
		panic("invalid clause shrink")
	}
	clear(c.Lits[len(c.Lits)-count:])
	c.Lits = c.Lits[:len(c.Lits)-count]
}
func (c *Clause) Subsumes(other *Clause) Lit {
	if c.Learnt || other.Learnt || !c.HasExtra || !other.HasExtra {
		panic("subsumption requires original clause abstractions")
	}
	if len(other.Lits) < len(c.Lits) || c.Abstraction&^other.Abstraction != 0 {
		return LitError
	}
	result := LitUndef
	for _, literal := range c.Lits {
		found := false
		for _, candidate := range other.Lits {
			if literal == candidate {
				found = true
				break
			}
			if result == LitUndef && literal == candidate.Not() {
				result = literal
				found = true
				break
			}
		}
		if !found {
			return LitError
		}
	}
	return result
}
func (c *Clause) Strengthen(literal Lit) {
	for i, l := range c.Lits {
		if l == literal {
			copy(c.Lits[i:], c.Lits[i+1:])
			c.Shrink(1)
			c.CalcAbstraction()
			return
		}
	}
	panic("literal missing from strengthened clause")
}

type ClauseAllocator struct {
	ExtraClauseField bool
	clauses          map[CRef]*Clause
	size, wasted     uint32
}

func NewClauseAllocator() *ClauseAllocator { return &ClauseAllocator{clauses: make(map[CRef]*Clause)} }
func clauseWords(c *Clause) uint32 {
	extra := 0
	if c.HasExtra {
		extra = 1
	}
	return uint32(1 + len(c.Lits) + extra)
}
func (a *ClauseAllocator) Alloc(literals []Lit, learnt bool) CRef {
	c := &Clause{Lits: append([]Lit(nil), literals...), Learnt: learnt, HasExtra: learnt || a.ExtraClauseField}
	words := clauseWords(c)
	if uint64(a.size)+uint64(words) >= uint64(CRefUndef) {
		panic("clause allocator overflow")
	}
	h := CRef(a.size)
	a.size += words
	if a.clauses == nil {
		a.clauses = make(map[CRef]*Clause)
	}
	if c.HasExtra && !learnt {
		c.CalcAbstraction()
	}
	a.clauses[h] = c
	return h
}
func (a *ClauseAllocator) At(h CRef) *Clause {
	c, ok := a.clauses[h]
	if !ok {
		panic("invalid clause reference")
	}
	return c
}
func (a *ClauseAllocator) Size() uint32   { return a.size }
func (a *ClauseAllocator) Wasted() uint32 { return a.wasted }
func (a *ClauseAllocator) Free(h CRef)    { c := a.At(h); a.wasted += clauseWords(c) }
func (a *ClauseAllocator) Relocate(handle *CRef, to *ClauseAllocator) {
	c := a.At(*handle)
	if c.relocated {
		*handle = c.relocation
		return
	}
	h := to.Alloc(c.Lits, c.Learnt)
	dest := to.At(h)
	dest.Mark = c.Mark
	dest.Activity = c.Activity
	c.relocated, c.relocation = true, h
	*handle = h
}
func (a *ClauseAllocator) MoveTo(to *ClauseAllocator) {
	if a == to {
		return
	}
	*to = *a
	*a = ClauseAllocator{}
}

type OccLists[T any] struct {
	Lists   [][]T
	dirty   []bool
	dirties []int
	deleted func(T) bool
}

func NewOccLists[T any](deleted func(T) bool) *OccLists[T] { return &OccLists[T]{deleted: deleted} }
func (o *OccLists[T]) Init(index int) {
	for len(o.Lists) <= index {
		o.Lists = append(o.Lists, nil)
		o.dirty = append(o.dirty, false)
	}
}
func (o *OccLists[T]) Smudge(index int) {
	if !o.dirty[index] {
		o.dirty[index] = true
		o.dirties = append(o.dirties, index)
	}
}
func (o *OccLists[T]) Lookup(index int) []T {
	if o.dirty[index] {
		o.Clean(index)
	}
	return o.Lists[index]
}
func (o *OccLists[T]) Clean(index int) {
	list := o.Lists[index]
	n := 0
	for _, entry := range list {
		if !o.deleted(entry) {
			list[n] = entry
			n++
		}
	}
	clear(list[n:])
	o.Lists[index] = list[:n]
	o.dirty[index] = false
}
func (o *OccLists[T]) CleanAll() {
	for _, index := range o.dirties {
		if o.dirty[index] {
			o.Clean(index)
		}
	}
	o.dirties = o.dirties[:0]
}
func (o *OccLists[T]) Clear() { o.Lists = nil; o.dirty = nil; o.dirties = nil }

func normalizeBool(v LBool) LBool {
	if v >= Undef {
		return Undef
	}
	return v
}
