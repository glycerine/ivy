package abs

import "github.com/glycerine/abcgo/sat/bsat"

// Ga2_ManCnfAddDynamic emits the minimum four-input SOP covers for a cut.
func Ga2_ManCnfAddDynamic(p *Ga2_Man_t, truth int32, lits []int32, out, proofID int32) {
	if len(p.pSops) != 1<<16 {
		panic("abs: invalid dynamic CNF input")
	}
	table := uint32(truth) & 0xffff
	for polarity := 0; polarity < 2; polarity++ {
		if polarity == 1 {
			table ^= 0xffff
		}
		for _, encoded := range p.pSops[table] {
			output := out
			if polarity == 1 {
				output ^= 1
			}
			clause := []int32{output}
			cube := int(encoded)
			for b := 3; b >= 0; b-- {
				digit := cube % 3
				cube /= 3
				if digit == 2 {
					continue
				}
				if b >= len(lits) || lits[b] <= 1 {
					panic("abs: invalid cut literal")
				}
				lit := lits[b]
				if digit == 1 {
					lit ^= 1
				}
				clause = append(clause, lit)
			}
			bsat.Sat_solver2_addclause(p.pSat, clause, proofID)
		}
	}
}
