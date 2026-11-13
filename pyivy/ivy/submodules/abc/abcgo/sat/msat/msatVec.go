package msat

/**CFile****************************************************************

  FileName    [msatVec.c]

  PackageName [A C version of SAT solver MINISAT, originally developed
  in C++ by Niklas Een and Niklas Sorensson, Chalmers University of
  Technology, Sweden: http://www.cs.chalmers.se/~een/Satzoo.]

  Synopsis    [Integer vector borrowed from Extra.]

  Author      [Alan Mishchenko <alanmi@eecs.berkeley.edu>]

  Affiliation [UC Berkeley]

  Date        [Ver. 1.0. Started - January 1, 2004.]

  Revision    [$Id: msatVec.c,v 1.0 2004/01/01 1:00:00 alanmi Exp $]

***********************************************************************/

import "sort"

type Msat_IntVec_t struct{ data []int }

func Msat_IntVecAlloc(n int) *Msat_IntVec_t {
	if n < 0 {
		panic("negative vector capacity")
	}
	if n > 0 && n < 16 {
		n = 16
	}
	return &Msat_IntVec_t{data: make([]int, 0, n)}
}
func Msat_IntVecFree(p *Msat_IntVec_t)            { p.data = nil }
func Msat_IntVecReadArray(p *Msat_IntVec_t) []int { return p.data }
func Msat_IntVecReadSize(p *Msat_IntVec_t) int    { return len(p.data) }
func Msat_IntVecGrow(p *Msat_IntVec_t, n int) {
	if n <= cap(p.data) {
		return
	}
	a := make([]int, len(p.data), n)
	copy(a, p.data)
	p.data = a
}
func Msat_IntVecShrink(p *Msat_IntVec_t, n int) {
	if n < 0 || n > len(p.data) {
		panic("invalid vector shrink")
	}
	p.data = p.data[:n]
}
func Msat_IntVecClear(p *Msat_IntVec_t) { p.data = p.data[:0] }
func Msat_IntVecPush(p *Msat_IntVec_t, value int) {
	if len(p.data) == cap(p.data) {
		n := 2 * cap(p.data)
		if n < 16 {
			n = 16
		}
		Msat_IntVecGrow(p, n)
	}
	p.data = append(p.data, value)
}
func Msat_IntVecPop(p *Msat_IntVec_t) int {
	if len(p.data) == 0 {
		panic("pop empty vector")
	}
	i := len(p.data) - 1
	value := p.data[i]
	p.data = p.data[:i]
	return value
}
func Msat_IntVecWriteEntry(p *Msat_IntVec_t, i int, value int) { p.data[i] = value }
func Msat_IntVecReadEntry(p *Msat_IntVec_t, i int) int         { return p.data[i] }

func Msat_IntVecAllocArray(a []int, n int) *Msat_IntVec_t { return &Msat_IntVec_t{data: a[:n]} }
func Msat_IntVecAllocArrayCopy(a []int, n int) *Msat_IntVec_t {
	return &Msat_IntVec_t{data: append([]int(nil), a[:n]...)}
}
func Msat_IntVecDup(p *Msat_IntVec_t) *Msat_IntVec_t {
	q := Msat_IntVecAlloc(cap(p.data))
	q.data = append(q.data, p.data...)
	return q
}
func Msat_IntVecDupArray(p *Msat_IntVec_t) *Msat_IntVec_t {
	q := &Msat_IntVec_t{data: p.data}
	p.data = nil
	return q
}
func Msat_IntVecFill(p *Msat_IntVec_t, n, value int) {
	Msat_IntVecGrow(p, n)
	p.data = p.data[:n]
	for i := range p.data {
		p.data[i] = value
	}
}
func Msat_IntVecReleaseArray(p *Msat_IntVec_t) []int { a := p.data; p.data = nil; return a }
func Msat_IntVecReadEntryLast(p *Msat_IntVec_t) int  { return p.data[len(p.data)-1] }
func Msat_IntVecPushUnique(p *Msat_IntVec_t, value int) int {
	for _, v := range p.data {
		if v == value {
			return 1
		}
	}
	Msat_IntVecPush(p, value)
	return 0
}
func Msat_IntVecPushUniqueOrder(p *Msat_IntVec_t, value, increase int) {
	Msat_IntVecPushUnique(p, value)
	for i := len(p.data) - 1; i > 0; i-- {
		if increase != 0 && p.data[i] >= p.data[i-1] || increase == 0 && p.data[i] <= p.data[i-1] {
			break
		}
		p.data[i], p.data[i-1] = p.data[i-1], p.data[i]
	}
}
func Msat_IntVecSort(p *Msat_IntVec_t, reverse int) {
	sort.Ints(p.data)
	if reverse != 0 {
		for i, j := 0, len(p.data)-1; i < j; i, j = i+1, j-1 {
			p.data[i], p.data[j] = p.data[j], p.data[i]
		}
	}
}
