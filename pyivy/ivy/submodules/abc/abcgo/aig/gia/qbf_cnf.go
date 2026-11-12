package gia

import (
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/sat/satoko"
)

// qbfEncodeGraph adds a Tseitin encoding and returns its output literal.
// inputLits allows separate copies to share parameters but not internal nodes.
func qbfEncodeGraph(s *satoko.Solver, g *Gia_Man_t, inputLits []int) int {
	if g == nil || g.nRegs != 0 || len(inputLits) != int(Gia_ManCiNum(g)) || Gia_ManCoNum(g) != 1 {
		panic("gia: invalid QBF circuit")
	}
	literals := make([]int, g.nObjs)
	constant := s.AddVariable(false) * 2
	s.AddClause([]int{constant ^ 1})
	literals[0] = constant
	for i, id := range vec.Vec_IntArray(g.vCis) {
		literals[id] = inputLits[i]
	}
	read := func(lit int32) int { return literals[lit>>1] ^ int(lit&1) }
	for id := int32(1); id < g.nObjs; id++ {
		a, b, c, kind := g.Gate(id)
		if kind == 1 {
			continue
		}
		if kind == 5 || kind == 6 {
			literals[id] = read(a)
			continue
		}
		if kind < 2 || kind > 4 {
			panic("gia: unsupported QBF gate")
		}
		fans := []int{read(a), read(b)}
		if kind == 4 {
			fans = append(fans, read(c))
		}
		out := s.AddVariable(false) * 2
		literals[id] = out
		for mask := 0; mask < 1<<uint(len(fans)); mask++ {
			values := [3]bool{}
			clause := make([]int, 0, len(fans)+1)
			for k, lit := range fans {
				values[k] = mask&(1<<uint(k)) != 0
				if values[k] {
					clause = append(clause, lit^1)
				} else {
					clause = append(clause, lit)
				}
			}
			var result bool
			switch kind {
			case 2:
				result = values[0] && values[1]
			case 3:
				result = values[0] != values[1]
			case 4:
				if values[0] {
					result = values[1]
				} else {
					result = values[2]
				}
			}
			if result {
				clause = append(clause, out)
			} else {
				clause = append(clause, out^1)
			}
			s.AddClause(clause)
		}
	}
	return literals[vec.Vec_IntEntry(g.vCos, 0)]
}
