package msat

/**CFile****************************************************************

  FileName    [msatClauseVec.c]

  PackageName [A C version of SAT solver MINISAT, originally developed
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Procedures working with arrays of SAT clauses.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatClauseVec.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

type Msat_ClauseVec_t struct{ data []*Msat_Clause_t }

func Msat_ClauseVecAlloc(n int) *Msat_ClauseVec_t {
	if n < 0 {
		panic("negative vector capacity")
	}
	if n > 0 && n < 16 {
		n = 16
	}
	return &Msat_ClauseVec_t{data: make([]*Msat_Clause_t, 0, n)}
}
func Msat_ClauseVecFree(p *Msat_ClauseVec_t)                       { p.data = nil }
func Msat_ClauseVecReadArray(p *Msat_ClauseVec_t) []*Msat_Clause_t { return p.data }
func Msat_ClauseVecReadSize(p *Msat_ClauseVec_t) int               { return len(p.data) }
func Msat_ClauseVecGrow(p *Msat_ClauseVec_t, n int) {
	if n <= cap(p.data) {
		return
	}
	a := make([]*Msat_Clause_t, len(p.data), n)
	copy(a, p.data)
	p.data = a
}
func Msat_ClauseVecShrink(p *Msat_ClauseVec_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.data = p.data[:n]
}
func Msat_ClauseVecClear(p *Msat_ClauseVec_t) { p.data = p.data[:0] }
func Msat_ClauseVecPush(p *Msat_ClauseVec_t, value *Msat_Clause_t) {
	if len(p.data) == cap(p.data) {
		n := 2 * cap(p.data)
		if n < 16 {
			n = 16
		}
		Msat_ClauseVecGrow(p, n)
	}
	p.data = append(p.data, value)
}
func Msat_ClauseVecPop(p *Msat_ClauseVec_t) *Msat_Clause_t {
	if len(p.data) == 0 {
		panic("pop empty vector")
	}
	i := len(p.data) - 1
	value := p.data[i]
	p.data = p.data[:i]
	return value
}
func Msat_ClauseVecWriteEntry(p *Msat_ClauseVec_t, i int, value *Msat_Clause_t) { p.data[i] = value }
func Msat_ClauseVecReadEntry(p *Msat_ClauseVec_t, i int) *Msat_Clause_t         { return p.data[i] }
