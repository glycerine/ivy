// Go port of pdrUtil.c by Alan Mishchenko, UC Berkeley (2010).
package pdr

import (
	"fmt"
	"io"
	"os"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"github.com/glycerine/abcgo/misc/hash"
	"github.com/glycerine/abcgo/misc/vec"
)

func pdrSetPattern(p *Pdr_Set_t, nRegs int32, counts *vec.Vec_Int_t) []byte {
	if nRegs < 0 {
		panic("pdr: negative register count")
	}
	pattern := make([]byte, nRegs)
	for i := range pattern {
		pattern[i] = '-'
	}
	for _, lit := range p.Lits[:p.nLits] {
		if lit == -1 {
			continue
		}
		if lit < 0 || lit>>1 >= nRegs {
			panic("pdr: cube literal outside register range")
		}
		pattern[lit>>1] = '1' - byte(lit&1)
	}
	if counts == nil {
		return pattern
	}
	selected := make([]byte, 0, len(pattern))
	for i, count := range vec.Vec_IntArray(counts) {
		if count != 0 {
			if i >= len(pattern) {
				panic("pdr: register filter out of range")
			}
			selected = append(selected, pattern[i])
		}
	}
	return selected
}

func Pdr_SetPrint(w io.Writer, p *Pdr_Set_t, nRegs int32, counts *vec.Vec_Int_t) {
	_, _ = w.Write(pdrSetPattern(p, nRegs, counts))
}
func ZPdr_SetPrint(p *Pdr_Set_t) {
	for _, lit := range p.Lits[:p.nLits] {
		fmt.Fprintf(os.Stdout, "%d ", lit)
	}
	fmt.Fprintln(os.Stdout)
}
func Pdr_SetPrintStr(out *vec.Vec_Str_t, p *Pdr_Set_t, nRegs int32, counts *vec.Vec_Int_t) {
	pattern := pdrSetPattern(p, nRegs, counts)
	vec.Vec_StrPushBuffer(out, pattern, len(pattern))
	vec.Vec_StrPushBuffer(out, []byte(" 1\n"), 3)
}

func ZPdr_SetIntersection(a, b *Pdr_Set_t, keep *hash.Hash_Int_t) *Pdr_Set_t {
	common := vec.Vec_IntAlloc(int(min(a.nLits, b.nLits)))
	i, j := 0, 0
	for i < int(a.nLits) || j < int(b.nLits) {
		if i < int(a.nLits) && j < int(b.nLits) && a.Lits[i] == b.Lits[j] {
			vec.Vec_IntPush(common, a.Lits[i])
			i++
			j++
			continue
		}
		var dropped int32
		if j == int(b.nLits) || i < int(a.nLits) && a.Lits[i] < b.Lits[j] {
			dropped = a.Lits[i]
			i++
		} else {
			dropped = b.Lits[j]
			j++
		}
		if keep != nil && hash.Hash_IntExists(keep, dropped) != 0 {
			return nil
		}
	}
	return Pdr_SetCreate(common, vec.Vec_IntAlloc(0))
}

// Pdr_OblStart transfers ownership of state and next to the new obligation.
func Pdr_OblStart(k, prio int32, state *Pdr_Set_t, next *Pdr_Obl_t) *Pdr_Obl_t {
	return &Pdr_Obl_t{iFrame: k, prio: prio, nRefs: 1, pState: state, pNext: next}
}
func Pdr_OblRef(p *Pdr_Obl_t) *Pdr_Obl_t { p.nRefs++; return p }
func Pdr_OblDeref(p *Pdr_Obl_t) {
	if p == nil || p.nRefs <= 0 {
		panic("pdr: invalid obligation reference count")
	}
	p.nRefs--
	if p.nRefs == 0 {
		if p.pNext != nil {
			Pdr_OblDeref(p.pNext)
			p.pNext = nil
		}
		if p.pState != nil {
			Pdr_SetDeref(p.pState)
			p.pState = nil
		}
		p.pLink = nil
	}
}
func Pdr_QueueIsEmpty(p *Pdr_Man_t) int32 {
	if p.pQueue == nil {
		return 1
	}
	return 0
}
func Pdr_QueueHead(p *Pdr_Man_t) *Pdr_Obl_t { return p.pQueue }
func Pdr_QueuePop(p *Pdr_Man_t) *Pdr_Obl_t {
	head := p.pQueue
	if head == nil {
		return nil
	}
	p.pQueue = head.pLink
	head.pLink = nil
	Pdr_OblDeref(head)
	p.nQueCur--
	return head
}
func Pdr_QueueClean(p *Pdr_Man_t) {
	for entry := Pdr_QueuePop(p); entry != nil; entry = Pdr_QueuePop(p) {
		Pdr_OblDeref(entry)
	}
}
func Pdr_QueuePush(p *Pdr_Man_t, entry *Pdr_Obl_t) {
	p.nObligs++
	p.nQueCur++
	if p.nQueCur > p.nQueMax {
		p.nQueMax = p.nQueCur
	}
	Pdr_OblRef(entry)
	link := &p.pQueue
	for *link != nil && ((*link).iFrame < entry.iFrame || ((*link).iFrame == entry.iFrame && (*link).prio <= entry.prio)) {
		link = &(*link).pLink
	}
	entry.pLink = *link
	*link = entry
}
func Pdr_QueuePrint(p *Pdr_Man_t) {
	for entry := p.pQueue; entry != nil; entry = entry.pLink {
		fmt.Printf("Frame = %2d.  Prio = %8d.\n", entry.iFrame, entry.prio)
	}
}
func Pdr_QueueStop(p *Pdr_Man_t) { Pdr_QueueClean(p); p.nQueCur = 0 }

const (
	PDR_VAL0 int32 = 1
	PDR_VAL1 int32 = 2
	PDR_VALX int32 = 3
)

func Pdr_ObjSatValue(graph *aig.Graph, node aig.Edge, complemented int32) int32 {
	if !aig.Aig_ObjIsTravIdCurrent(graph, node) {
		return PDR_VALX
	}
	value := aig.Aig_ObjIsMarkA(graph, node) != (complemented != 0)
	if value {
		return PDR_VAL1
	}
	return PDR_VAL0
}

func Pdr_NtkFindSatAssign_rec(graph *aig.Graph, node aig.Edge, value int32, cube *Pdr_Set_t, heuristic int32) int32 {
	if aig.Aig_ObjIsConst1(graph, node) {
		return 1
	}
	if aig.Aig_ObjIsTravIdCurrent(graph, node) {
		if aig.Aig_ObjIsMarkA(graph, node) == (value != 0) {
			return 1
		}
		return 0
	}
	aig.Aig_ObjSetTravIdCurrent(graph, node)
	if value != 0 {
		aig.Aig_ObjSetMarkA(graph, node)
	} else {
		aig.Aig_ObjClearMarkA(graph, node)
	}
	if aig.Aig_ObjIsCi(graph, node) {
		if saig.Saig_ObjIsLo(graph, node) {
			lit := 2*(aig.Aig_ObjCioId(graph, node)-saig.Saig_ManPiNum(graph)) + value
			if int(cube.nLits) == len(cube.Lits) {
				cube.Lits = append(cube.Lits, lit)
			} else {
				cube.Lits[cube.nLits] = lit
			}
			cube.nLits++
			cube.Sign |= uint64(1) << uint(lit%63)
		}
		return 1
	}
	if !aig.Aig_ObjIsNode(graph, node) {
		panic("pdr: expected AND node")
	}
	a, b := aig.Aig_ObjFanin0(graph, node), aig.Aig_ObjFanin1(graph, node)
	ca, cb := aig.Aig_ObjFaninC0(graph, node), aig.Aig_ObjFaninC1(graph, node)
	if value != 0 {
		if Pdr_NtkFindSatAssign_rec(graph, a, ca^1, cube, heuristic) == 0 {
			return 0
		}
		return Pdr_NtkFindSatAssign_rec(graph, b, cb^1, cube, heuristic)
	}
	va, vb := Pdr_ObjSatValue(graph, a, ca), Pdr_ObjSatValue(graph, b, cb)
	if va == PDR_VAL0 || vb == PDR_VAL0 {
		return 1
	}
	if va == PDR_VAL1 && vb == PDR_VAL1 {
		return 0
	}
	if va == PDR_VAL1 {
		return Pdr_NtkFindSatAssign_rec(graph, b, cb, cube, heuristic)
	}
	if vb == PDR_VAL1 {
		return Pdr_NtkFindSatAssign_rec(graph, a, ca, cube, heuristic)
	}
	if int32(node.NodeID())%4 == heuristic {
		return Pdr_NtkFindSatAssign_rec(graph, b, cb, cube, heuristic)
	}
	return Pdr_NtkFindSatAssign_rec(graph, a, ca, cube, heuristic)
}
