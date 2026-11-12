/***********************************************************************************[SolverTypes.h]
 Glucose -- Copyright (c) 2009, Gilles Audemard, Laurent Simon
                CRIL - Univ. Artois, France
                LRI  - Univ. Paris Sud, France

Glucose sources are based on MiniSat (see below MiniSat copyrights). Permissions and copyrights of
Glucose are exactly the same as Minisat on which it is based on. (see below).

---------------
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

package glucose2

// Literals retain Glucose's compact var*2+sign encoding. Clause references use
// ordinary Go pointers: GC can reclaim deleted clauses without pointer tags.
type Var int32
type Lit int32
type LBool uint8

const (
	VarUndef Var   = -1
	LitUndef Lit   = -2
	LitError Lit   = -1
	LTrue    LBool = 0
	LFalse   LBool = 1
	LUndef   LBool = 2
)

func MkLit(v Var, negative bool) Lit {
	if v < 0 {
		panic("glucose2: negative variable")
	}
	l := Lit(v * 2)
	if negative {
		l ^= 1
	}
	return l
}
func (l Lit) Var() Var   { return Var(l >> 1) }
func (l Lit) Sign() bool { return l&1 != 0 }
func (l Lit) Not() Lit   { return l ^ 1 }
func (l Lit) Xor(negative bool) Lit {
	if negative {
		return l.Not()
	}
	return l
}
func ToLit(value int32) Lit     { return Lit(value) }
func ToLBool(value int32) LBool { return LBool(uint8(value)) }
func (l LBool) IsUndef() bool   { return l&2 != 0 }
func (l LBool) Equal(r LBool) bool {
	return l.IsUndef() && r.IsUndef() || !l.IsUndef() && !r.IsUndef() && l == r
}
func (l LBool) Xor(negative bool) LBool {
	if negative {
		return l ^ 1
	}
	return l
}
func (l LBool) And(r LBool) LBool { return LBool((uint32(0xf7f755f4) >> uint((l<<1)|(r<<3))) & 3) }
func (l LBool) Or(r LBool) LBool  { return LBool((uint32(0xfcfcf400) >> uint((l<<1)|(r<<3))) & 3) }
func (l LBool) Not() LBool        { return l.Xor(true) }

type Clause struct {
	gateReason bool
	gate       Var

	freed                bool
	Lits                 []Lit
	learnt               bool
	mark                 uint8
	extra                bool
	activity             float64
	abstraction          uint32
	lbd                  int
	canBeDel             bool
	sizeWithoutSelectors int
	relocation           *Clause
}
type CRef = *Clause

func NewClause(lits []Lit, learnt, extra bool) *Clause {
	c := &Clause{Lits: append([]Lit(nil), lits...), learnt: learnt, extra: extra || learnt, canBeDel: true, sizeWithoutSelectors: len(lits)}
	if c.extra && !c.learnt {
		c.CalcAbstraction()
	}
	return c
}
func (c *Clause) Size() int          { return len(c.Lits) }
func (c *Clause) Learnt() bool       { return c.learnt }
func (c *Clause) HasExtra() bool     { return c.extra }
func (c *Clause) Mark() uint8        { return c.mark }
func (c *Clause) SetMark(mark uint8) { c.mark = mark & 3 }
func (c *Clause) Last() Lit          { return c.Lits[len(c.Lits)-1] }
func (c *Clause) Shrink(n int) {
	if n < 0 || n > len(c.Lits) {
		panic("glucose2: invalid clause shrink")
	}
	c.Lits = c.Lits[:len(c.Lits)-n]
}
func (c *Clause) Pop() { c.Shrink(1) }
func (c *Clause) CalcAbstraction() {
	var mask uint32
	for _, l := range c.Lits {
		mask |= 1 << uint(l.Var()&31)
	}
	c.abstraction = mask
}
func (c *Clause) Abstraction() uint32           { return c.abstraction }
func (c *Clause) Activity() float64             { return c.activity }
func (c *Clause) SetActivity(value float64)     { c.activity = value }
func (c *Clause) LBD() int                      { return c.lbd }
func (c *Clause) SetLBD(value int)              { c.lbd = value }
func (c *Clause) CanBeDel() bool                { return c.canBeDel }
func (c *Clause) SetCanBeDel(value bool)        { c.canBeDel = value }
func (c *Clause) SizeWithoutSelectors() int     { return c.sizeWithoutSelectors }
func (c *Clause) SetSizeWithoutSelectors(n int) { c.sizeWithoutSelectors = n }
func (c *Clause) Relocated() bool               { return c.relocation != nil }
func (c *Clause) Relocation() *Clause           { return c.relocation }
func (c *Clause) Relocate(to *Clause)           { c.relocation = to }

// Subsumes returns LitUndef for full subsumption, a removable opposite literal
// for self-subsuming resolution, or LitError when neither applies.
func (c *Clause) Subsumes(other *Clause) Lit {
	if c.learnt || other.learnt || !c.extra || !other.extra {
		panic("glucose2: subsumption requires original clauses with abstraction")
	}
	if len(c.Lits) > len(other.Lits) || c.abstraction&^other.abstraction != 0 {
		return LitError
	}
	result := LitUndef
	for _, lit := range c.Lits {
		found := false
		for _, candidate := range other.Lits {
			if lit == candidate {
				found = true
				break
			}
			if result == LitUndef && lit == candidate.Not() {
				result = lit
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
func (c *Clause) Strengthen(remove Lit) {
	for i, l := range c.Lits {
		if l == remove {
			copy(c.Lits[i:], c.Lits[i+1:])
			c.Lits = c.Lits[:len(c.Lits)-1]
			c.CalcAbstraction()
			return
		}
	}
	panic("glucose2: literal not in clause")
}

type ClauseAllocator struct {
	ExtraClauseField bool
	clauses          []*Clause
	wasted           int
}

func (a *ClauseAllocator) Alloc(lits []Lit, learnt bool) *Clause {
	c := NewClause(lits, learnt, a.ExtraClauseField)
	a.clauses = append(a.clauses, c)
	return c
}
func clauseWords(c *Clause) int {
	n := 3 + len(c.Lits)
	if c.extra {
		n++
	}
	return n
}
func (a *ClauseAllocator) Size() int {
	n := 0
	for _, c := range a.clauses {
		n += clauseWords(c)
	}
	return n
}
func (a *ClauseAllocator) Wasted() int { return a.wasted }
func (a *ClauseAllocator) Free(c *Clause) {
	if !c.freed {
		c.freed = true
		c.mark = 1
		a.wasted += clauseWords(c)
	}
}
func (a *ClauseAllocator) MoveTo(to *ClauseAllocator) { *to = *a; *a = ClauseAllocator{} }
func (a *ClauseAllocator) Reloc(c **Clause, to *ClauseAllocator) {
	if *c == nil {
		return
	}
	if (*c).relocation != nil {
		*c = (*c).relocation
		return
	}
	old := *c
	next := to.Alloc(old.Lits, old.learnt)
	next.mark, next.activity, next.lbd, next.canBeDel, next.sizeWithoutSelectors = old.mark, old.activity, old.lbd, old.canBeDel, old.sizeWithoutSelectors
	old.Relocate(next)
	*c = next
}
