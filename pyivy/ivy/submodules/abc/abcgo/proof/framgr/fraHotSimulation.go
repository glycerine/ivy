package framgr

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"time"
)

func Fra_RegToLit(n, complement int32) int32 {
	if complement != 0 {
		return -n - 1
	}
	return n + 1
}
func Fra_LitReg(n int32) int32 {
	if n > 0 {
		return n - 1
	}
	return -n - 1
}
func Fra_LitSign(n int32) int32 {
	if n < 0 {
		return 1
	}
	return 0
}
func Fra_OneHotNodeIsConst(s *Fra_Sml_t, node aig.Edge) int32 {
	return fraBool(s.IsConstant(node.NodeID()))
}
func Fra_OneHotNodesAreEqual(s *Fra_Sml_t, a, b aig.Edge) int32 {
	return fraBool(s.Equal(a.NodeID(), b.NodeID()))
}

func oneHotViolation(a, b uint32, negA, negB bool) uint32 {
	switch {
	case negA && negB:
		return a & b
	case negA:
		return a &^ b
	case negB:
		return ^a & b
	default:
		panic("fra: one-hot clause has no negative literal")
	}
}
func Fra_OneHotNodesAreClause(s *Fra_Sml_t, a, b aig.Edge, negA, negB int32) int32 {
	left, right := s.nodeWords(a.NodeID()), s.nodeWords(b.NodeID())
	for i := s.prefixWords; i < s.wordsTotal; i++ {
		if oneHotViolation(left[i], right[i], negA != 0, negB != 0) != 0 {
			return 0
		}
	}
	return 1
}
func Fra_OneHotRefineUsingCex(p *Fra_Man_t, clauses *vec.Vec_Int_t) int32 {
	if p.pSml.graph != p.pManAig {
		panic("fra: wrong one-hot simulation graph")
	}
	inputs := p.pManAig.Inputs()
	first := len(inputs) - int(p.pManAig.RegisterCount())
	var changed int32
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		x, y := vec.Vec_IntEntry(clauses, i), vec.Vec_IntEntry(clauses, i+1)
		if x == 0 && y == 0 {
			continue
		}
		if Fra_OneHotNodesAreClause(p.pSml, inputs[first+int(Fra_LitReg(x))], inputs[first+int(Fra_LitReg(y))], Fra_LitSign(x), Fra_LitSign(y)) == 0 {
			vec.Vec_IntWriteEntry(clauses, i, 0)
			vec.Vec_IntWriteEntry(clauses, i+1, 0)
			changed = 1
		}
	}
	return changed
}
func Fra_OneHotCount(_ *Fra_Man_t, clauses *vec.Vec_Int_t) int32 {
	var count int32
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		if vec.Vec_IntEntry(clauses, i) != 0 || vec.Vec_IntEntry(clauses, i+1) != 0 {
			count++
		}
	}
	return count
}
func Fra_OneHotEstimateCoverage(p *Fra_Man_t, clauses *vec.Vec_Int_t) {
	started := time.Now()
	const words = 1 << 14
	registers := make([][]uint32, p.pManAig.RegisterCount())
	aig.Aig_ManRandom(1)
	for i := range registers {
		registers[i] = make([]uint32, words)
		for j := range registers[i] {
			registers[i][j] = Fra_ObjRandomSim()
		}
	}
	violations := make([]uint32, words)
	for i := 0; i < vec.Vec_IntSize(clauses); i += 2 {
		x, y := vec.Vec_IntEntry(clauses, i), vec.Vec_IntEntry(clauses, i+1)
		if x == 0 && y == 0 {
			continue
		}
		a, b := registers[Fra_LitReg(x)], registers[Fra_LitReg(y)]
		for j := range violations {
			violations[j] |= oneHotViolation(a[j], b[j], x < 0, y < 0)
		}
	}
	covered := 0
	for _, word := range violations {
		covered += bits.OnesCount32(word)
	}
	fmt.Printf("Care states ratio = %f. (%d out of %d patterns)  Time = %.2f sec\n", float64(words*32-covered)/float64(words*32), words*32-covered, words*32, time.Since(started).Seconds())
}
