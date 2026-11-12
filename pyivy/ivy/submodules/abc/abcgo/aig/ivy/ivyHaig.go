// History-AIG lifecycle and three-valued simulation ported from ABC ivyHaig.c.
package ivy

import (
	"fmt"
	"io"
	"os"
)

func Ivy_HaigObjCountClass(root *Ivy_Obj_t) int32 {
	if root == nil || Ivy_IsComplement(root) != 0 || root.nRefs <= 0 {
		panic("ivy: referenced history representative required")
	}
	if root.pEquiv == nil {
		return 1
	}
	if Ivy_IsComplement(root.pEquiv) != 0 {
		panic("ivy: complemented representative link")
	}
	count := int32(1)
	seen := map[*Ivy_Obj_t]bool{root: true}
	for n := root.pEquiv; n != root; n = Ivy_Regular(n.pEquiv) {
		if n == nil || n.owner != root.owner || seen[n] {
			panic("ivy: malformed history choice ring")
		}
		seen[n] = true
		count++
	}
	return count
}
func Ivy_ManHaigStart(p *Ivy_Man_t, verbose int32) {
	if p == nil || p.pHaig != nil || Ivy_ManCheck(p) == 0 {
		panic("ivy: valid manager without history required")
	}
	history := Ivy_ManDup(p)
	var latches []int32
	for _, n := range history.vObjs {
		if n == nil || Ivy_ObjIsLatch(n) == 0 {
			continue
		}
		latches = append(latches, n.Id)
		original := n.Init
		Ivy_TableDelete(history, n)
		n.Init = IVY_INIT_DC
		if existing := Ivy_TableLookup(history, n); existing == nil {
			Ivy_TableInsert(history, n)
		} else {
			// Resetting initial values can merge previously distinct latch keys. Retain
			// the old mapped identity as a buffer rather than leaving a stale hash key.
			n.Init = original
			Ivy_TableInsert(history, n)
			Ivy_ObjReplace(history, n, existing, 1, 0, 0)
		}
	}
	history.pData = latches
	p.pHaig = history
	if verbose != 0 {
		fmt.Print("Starting : ")
		Ivy_ManPrintStats(history)
	}
}

// Ivy_ManHaigTrasfer retains the historical spelling of the C API. Ownership
// moves to destination; source no longer owns a history manager to stop.
func Ivy_ManHaigTrasfer(source, destination *Ivy_Man_t) {
	if source == nil || destination == nil || source == destination || source.pHaig == nil || destination.pHaig != nil || len(source.vPis) != len(destination.vPis) || Ivy_ManCheck(destination) == 0 || Ivy_ManNodeNum(destination) != 0 || Ivy_ManLatchNum(destination) != 0 || Ivy_ManBufNum(destination) != 0 || len(destination.vPos) != 0 {
		panic("ivy: invalid history transfer")
	}
	history := source.pHaig
	destination.pConst1.pEquiv = source.pConst1.pEquiv
	for i, n := range destination.vPis {
		n.pEquiv = source.vPis[i].pEquiv
	}
	destination.pHaig = history
	source.pHaig = nil
}
func Ivy_ManHaigStop(p *Ivy_Man_t) {
	if p == nil || p.pHaig == nil {
		return
	}
	history := p.pHaig
	p.pHaig = nil
	for _, n := range p.vObjs {
		if n != nil {
			n.pEquiv = nil
		}
	}
	Ivy_ManStop(history)
}
func Ivy_ManHaigCountChoices(p *Ivy_Man_t, choices *int32) int32 {
	if p == nil || p.pHaig == nil {
		panic("ivy: history manager required")
	}
	var nodes, total int32
	for _, n := range p.pHaig.vObjs {
		if n == nil || n.Id == 0 || Ivy_ObjIsTerm(n) != 0 || n.nRefs == 0 {
			continue
		}
		size := Ivy_HaigObjCountClass(n)
		if size > 1 {
			nodes++
			total += size - 1
		}
	}
	if choices != nil {
		*choices = total
	}
	return nodes
}
func Ivy_ManHaigPostprocess(p *Ivy_Man_t, verbose int32) {
	if p == nil || p.pHaig == nil {
		panic("ivy: history manager required")
	}
	if verbose != 0 {
		fmt.Print("Final    : ")
		Ivy_ManPrintStats(p)
		fmt.Print("HAIG     : ")
		Ivy_ManPrintStats(p.pHaig)
		var choices int32
		nodes := Ivy_ManHaigCountChoices(p, &choices)
		fmt.Printf("Total choice nodes = %d. Total choices = %d. Skipped classes = %d.\n", nodes, choices, p.pHaig.nClassesSkip)
	}
	if Ivy_ManIsAcyclic(p.pHaig) == 0 {
		fmt.Println("HAIG contains a cycle")
	} else if verbose != 0 {
		fmt.Println("HAIG is acyclic")
	}
}
func Ivy_ManHaigSimulateAnd(a, b Ivy_Init_t) Ivy_Init_t {
	if a < IVY_INIT_0 || a > IVY_INIT_DC || b < IVY_INIT_0 || b > IVY_INIT_DC {
		panic("ivy: invalid history simulation value")
	}
	// This is the original modified rule: unknown dominates even zero.
	if a == IVY_INIT_DC || b == IVY_INIT_DC {
		return IVY_INIT_DC
	}
	if a == IVY_INIT_1 && b == IVY_INIT_1 {
		return IVY_INIT_1
	}
	return IVY_INIT_0
}
func Ivy_ManHaigSimulateChoice(a, b Ivy_Init_t) Ivy_Init_t {
	if a < IVY_INIT_0 || a > IVY_INIT_DC || b < IVY_INIT_0 || b > IVY_INIT_DC {
		panic("ivy: invalid history choice value")
	}
	if a != IVY_INIT_DC && b != IVY_INIT_DC && a != b {
		fmt.Println("Compatibility fails.")
		return IVY_INIT_0
	}
	if a != IVY_INIT_DC {
		return a
	}
	return b
}
func ivyHaigSimulateSteps(working *Ivy_Man_t, rounds int, out io.Writer) []int {
	if working == nil || working.pHaig == nil || rounds < 0 {
		panic("ivy: history manager and nonnegative rounds required")
	}
	if Ivy_ManCheckChoices(working) == 0 || Ivy_ManIsAcyclic(working.pHaig) == 0 {
		panic("ivy: invalid history choices")
	}
	p := working.pHaig
	var latchIDs []int32
	nodes := Ivy_ManDfsSeq(p, &latchIDs)
	values := map[*Ivy_Obj_t]Ivy_Init_t{p.pConst1: IVY_INIT_1}
	for _, n := range p.vPis {
		values[n] = IVY_INIT_0
	}
	for _, id := range latchIDs {
		values[p.vObjs[id]] = IVY_INIT_DC
	}
	originals, ok := p.pData.([]int32)
	if !ok {
		panic("ivy: history latch origins missing")
	}
	for _, id := range originals {
		if id < 0 || int(id) >= len(p.vObjs) || p.vObjs[id] == nil {
			panic("ivy: missing original history latch")
		}
		values[p.vObjs[id]] = IVY_INIT_0
	}
	edgeValue := func(edge *Ivy_Obj_t) Ivy_Init_t {
		value, known := values[Ivy_Regular(edge)]
		if !known {
			panic("ivy: unordered history simulation")
		}
		return Ivy_InitNotCond(value, Ivy_IsComplement(edge))
	}
	counts := make([]int, rounds)
	for round := 0; round < rounds; round++ {
		for _, id := range latchIDs {
			if values[p.vObjs[id]] == IVY_INIT_DC {
				counts[round]++
			}
		}
		if out != nil {
			fmt.Fprintf(out, "Iter %d : Non-determinate = %d\n", round, counts[round])
		}
		for _, id := range nodes {
			n := p.vObjs[id]
			a := edgeValue(n.pFanin0)
			switch n.Type {
			case IVY_BUF:
				values[n] = a
			case IVY_AND:
				values[n] = Ivy_ManHaigSimulateAnd(a, edgeValue(n.pFanin1))
			case IVY_EXOR:
				values[n] = Ivy_InitExor(a, edgeValue(n.pFanin1))
			default:
				panic("ivy: unsupported history simulation node")
			}
			if n.pEquiv != nil && n.nRefs > 0 {
				merged := values[n]
				seen := map[*Ivy_Obj_t]bool{n: true}
				for member := n.pEquiv; member != n; member = Ivy_Regular(member.pEquiv) {
					if member == nil || seen[member] {
						panic("ivy: malformed simulated choice ring")
					}
					seen[member] = true
					value, known := values[member]
					if !known {
						panic("ivy: unordered simulated choice")
					}
					merged = Ivy_ManHaigSimulateChoice(merged, Ivy_InitNotCond(value, Ivy_IsComplement(member.pEquiv)))
				}
				values[n] = merged
			}
		}
		next := make([]Ivy_Init_t, len(latchIDs))
		for i, id := range latchIDs {
			next[i] = edgeValue(p.vObjs[id].pFanin0)
		}
		for i, id := range latchIDs {
			values[p.vObjs[id]] = next[i]
		}
	}
	return counts
}
func Ivy_ManHaigSimulate(p *Ivy_Man_t) { ivyHaigSimulateSteps(p, 10, os.Stdout) }
