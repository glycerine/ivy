package saig

import (
	"fmt"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func Saig_SynchNot(w uint32) uint32    { return synchNot(w) }
func Saig_SynchAnd(u, w uint32) uint32 { return synchAnd(u, w) }
func Saig_SynchTernary(v int32) uint32 { return synchValue(uint32(v)) }

// Saig_SynchSequenceRun replays a sequence and stores its resulting binary
// state in the register-output marks. simInfo is retained for API compatibility.
func Saig_SynchSequenceRun(g *aig.Graph, simInfo *vec.Vec_Ptr_t, sequence *vec.Vec_Str_t, ternary int32) int32 {
	if g == nil || sequence == nil {
		panic("saig: missing synchronizing sequence")
	}
	pis := int(Saig_ManPiNum(g))
	data := vec.Vec_StrArray(sequence)
	if pis == 0 {
		if len(data) != 0 {
			panic("saig: invalid synchronizing sequence")
		}
		return Saig_ManRegNum(g)
	}
	if len(data)%pis != 0 {
		panic("saig: invalid synchronizing sequence length")
	}
	s := newSynchState(g, 1)
	s.initRegisters(ternary != 0)
	for i := 0; i < len(data); i += pis {
		s.givenInputs(data[i : i+pis])
		s.simulate()
		s.transfer()
	}
	var unknown int32
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		lo := Saig_ManLo(g, i)
		v := s.words[lo.NodeID()][0] & 3
		if v == 3 {
			unknown++
		}
		if v != 0 {
			aig.Aig_ObjSetMarkA(g, lo)
		} else {
			aig.Aig_ObjClearMarkA(g, lo)
		}
	}
	return unknown
}

// Saig_SynchSequence finds a sequence by selecting, in each frame, the
// candidate that minimizes the number of unknown register values.
func Saig_SynchSequence(g *aig.Graph, nWords int32) *vec.Vec_Str_t {
	if g == nil || Saig_ManRegNum(g) == 0 || Saig_ManPiNum(g) == 0 || nWords <= 0 {
		return nil
	}
	s := newSynchState(g, int(nWords))
	s.initRegisters(true)
	sequence := vec.Vec_StrAlloc(20 * int(Saig_ManRegNum(g)))
	previous := int(Saig_ManRegNum(g))
	for step := 0; step < 100 && previous > 0; step++ {
		pattern, cur := -1, previous
		for attempt := 0; attempt < 100; attempt++ {
			s.randomInputs()
			s.simulate()
			pattern, cur = s.bestPattern()
			if cur < previous {
				break
			}
		}
		if cur >= previous {
			break
		}
		inputs, unknown := s.savePattern(pattern)
		if unknown != cur {
			panic("saig: inconsistent synchronization pattern")
		}
		for _, v := range inputs {
			vec.Vec_StrPush(sequence, v)
		}
		previous = cur
	}
	if previous != 0 {
		return nil
	}
	if Saig_SynchSequenceRun(g, nil, sequence, 1) != 0 {
		panic("saig: synchronization verification failed")
	}
	aig.Aig_ManCleanMarkA(g)
	return sequence
}

// Saig_ManDupInitZero shifts each marked register's state so the duplicate
// starts at zero while preserving the original graph's sequential behavior.
func Saig_ManDupInitZero(g *aig.Graph) *aig.Graph {
	if g == nil {
		return nil
	}
	q := aig.NewGraph()
	q.Name, q.Spec = g.Name, g.Spec
	copies := map[aig.Edge]aig.Edge{g.Const1(): q.Const1()}
	for i := int32(0); i < Saig_ManPiNum(g); i++ {
		copies[aig.Aig_ManCi(g, i)] = q.NewInput()
	}
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		lo := Saig_ManLo(g, i)
		e := q.NewInput()
		if aig.Aig_ObjIsMarkA(g, lo) {
			e = e.Not()
		}
		copies[lo] = e
	}
	var copyEdge func(aig.Edge) aig.Edge
	copyEdge = func(e aig.Edge) aig.Edge {
		reg := aig.Aig_Regular(e)
		if c, ok := copies[reg]; ok {
			return aig.Aig_NotCond(c, int32(e&1))
		}
		n := g.Node(reg)
		var c aig.Edge
		switch n.Kind {
		case aig.And:
			c = q.AND(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
		case aig.ExclusiveOr:
			c = q.NewXor(copyEdge(n.Fanins[0]), copyEdge(n.Fanins[1]))
		case aig.Buffer, aig.Output:
			c = copyEdge(n.Fanins[0])
		default:
			panic("saig: unsupported duplication node")
		}
		copies[reg] = c
		return aig.Aig_NotCond(c, int32(e&1))
	}
	for i := int32(0); i < Saig_ManPoNum(g); i++ {
		q.AddOutput(copyEdge(aig.Aig_ObjChild0(g, aig.Aig_ManCo(g, i))))
	}
	for i := int32(0); i < Saig_ManRegNum(g); i++ {
		lo, li := Saig_ManLo(g, i), Saig_ManLi(g, i)
		driver := copyEdge(aig.Aig_ObjChild0(g, li))
		if aig.Aig_ObjIsMarkA(g, lo) {
			driver = driver.Not()
		}
		q.AddOutput(driver)
	}
	aig.Aig_ManSetRegNum(q, Saig_ManRegNum(g))
	return q
}

func Saig_SynchSequenceApply(g *aig.Graph, nWords, verbose int32) *aig.Graph {
	sequence := Saig_SynchSequence(g, nWords)
	if sequence == nil {
		if verbose != 0 {
			fmt.Println("Synchronizing sequence not found.")
		}
		return nil
	}
	if Saig_SynchSequenceRun(g, nil, sequence, 1) != 0 {
		panic("saig: synchronization verification failed")
	}
	q := Saig_ManDupInitZero(g)
	aig.Aig_ManCleanMarkA(g)
	return q
}

func Saig_Synchronize(a, b *aig.Graph, nWords, verbose int32) *aig.Graph {
	if a == nil || b == nil || Saig_ManPiNum(a) != Saig_ManPiNum(b) {
		return nil
	}
	x, y := Saig_SynchSequence(a, nWords), Saig_SynchSequence(b, nWords)
	if x == nil || y == nil {
		if verbose != 0 {
			fmt.Println("Synchronizing sequence not found.")
		}
		return nil
	}
	if Saig_SynchSequenceRun(a, nil, x, 1) != 0 || Saig_SynchSequenceRun(a, nil, y, 0) != 0 || Saig_SynchSequenceRun(b, nil, y, 1) != 0 {
		panic("saig: synchronization verification failed")
	}
	az, bz := Saig_ManDupInitZero(a), Saig_ManDupInitZero(b)
	aig.Aig_ManCleanMarkA(a)
	aig.Aig_ManCleanMarkA(b)
	m := Saig_ManCreateMiter(az, bz, 0)
	if m != nil {
		aig.Aig_ManCleanup(m)
	}
	return m
}
