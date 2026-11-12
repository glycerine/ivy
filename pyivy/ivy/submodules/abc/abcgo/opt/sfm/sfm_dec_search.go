package sfm

import (
	"github.com/glycerine/abcgo/sat/satoko"
	"slices"
)

// sfmDecSearch ports the SAT-guided decomposition search over window variables.
type sfmDecSearch struct {
	solver        *satoko.Solver
	patterns      sfmDecPatterns
	target, limit int
	conflictLimit int64
	useAndOr      bool
	fallback      []int
	forced        int
	history       []int
}

func (s *sfmDecSearch) record(set int, mask *[2][SFM_SIM_WORDS]uint64) {
	p := &s.patterns
	index := p.counts[set]
	if index >= 64*SFM_SIM_WORDS {
		return
	}
	for d := 0; d < p.divisors; d++ {
		if s.solver.ModelValue(d) == 1 {
			p.sets[set][d*SFM_SIM_WORDS+index/64] |= uint64(1) << uint(index%64)
		}
	}
	p.words[set] = 1 + index/64
	mask[set][index/64] |= uint64(1) << uint(index%64)
	p.counts[set]++
}
func (s *sfmDecSearch) solve(assumptions []int) int {
	return s.solver.SolveAssumptionsLimit(assumptions, s.conflictLimit)
}
func (s *sfmDecSearch) rec(assumptions []int, mask [2][SFM_SIM_WORDS]uint64, used []int, allowFallback bool, extra int) ([]uint64, []int32, bool) {
	p := &s.patterns
	for set := 0; set < 2; set++ {
		empty := true
		for _, v := range mask[set][:p.words[set]] {
			if v != 0 {
				empty = false
				break
			}
		}
		if !empty {
			continue
		}
		status := s.solve(append(append([]int(nil), assumptions...), s.target*2+set))
		if status == satoko.SATOKO_UNDEC {
			return nil, nil, false
		}
		if status == satoko.SATOKO_UNSAT {
			words := 1 << uint(max(0, s.limit-6))
			truth := make([]uint64, words)
			if set != 0 {
				for i := range truth {
					truth[i] = ^uint64(0)
				}
			}
			return truth, nil, true
		}
		s.record(set, &mask)
	}
	variable := s.forced
	s.forced = -1
	if variable < 0 {
		p.implications = [2][]int{}
		for d := 0; d < p.divisors; d++ {
			implications := [2]int{-1, -1}
			for set := 0; set < 2; set++ {
				pattern := p.divisor(set, d)
				ones := sfmMaskedCount(pattern, mask[set][:], false)
				zeros := sfmMaskedCount(pattern, mask[set][:], true)
				if ones != 0 && zeros != 0 {
					continue
				}
				hasOne := 0
				if ones != 0 {
					hasOne = 1
				}
				status := s.solve(append(append(append([]int(nil), assumptions...), s.target*2+set), 2*d+hasOne))
				if status == satoko.SATOKO_UNDEC {
					return nil, nil, false
				}
				if status == satoko.SATOKO_UNSAT {
					implications[set] = 2*d + (hasOne ^ 1)
					p.implications[set] = append(p.implications[set], implications[set])
				} else {
					s.record(set, &mask)
				}
			}
			if implications[0] >= 0 && implications[1] >= 0 {
				if implications[0] == implications[1] {
					p.implications[0] = p.implications[0][:len(p.implications[0])-1]
					p.implications[1] = p.implications[1][:len(p.implications[1])-1]
				} else {
					word := uint64(0xaaaaaaaaaaaaaaaa)
					if implications[0]&1 != 0 {
						word = ^word
					}
					truth := make([]uint64, 1<<uint(max(0, s.limit-6)))
					for i := range truth {
						truth[i] = word
					}
					return truth, []int32{int32(d)}, true
				}
			}
		}
		if extra > s.limit-2 {
			return nil, nil, false
		}
		if s.useAndOr {
			for set := 0; set < 2; set++ {
				candidates := p.implications[1-set]
				if len(candidates) < 2 {
					continue
				}
				as := append(append(append([]int(nil), assumptions...), s.target*2+set), candidates...)
				status := s.solve(as)
				if status == satoko.SATOKO_UNDEC {
					return nil, nil, false
				}
				if status == satoko.SATOKO_UNSAT {
					final := s.solver.FinalConflict()
					var selected []int
					for _, lit := range final {
						if slices.Contains(candidates, lit^1) {
							selected = append(selected, lit^1)
						}
					}
					if len(selected)+extra > 6 {
						continue
					}
					truth := make([]uint64, 1<<uint(max(0, s.limit-6)))
					var support []int32
					for _, lit := range selected {
						support = append(support, int32(lit>>1))
					}
					for a := 0; a < len(truth)*64; a++ {
						value := set != 0
						for i, lit := range selected {
							bit := (a >> uint(i) & 1) != 0
							if set != 0 {
								value = value && (bit != (lit&1 != 0))
							} else {
								value = value || (bit == (lit&1 != 0))
							}
						}
						if value {
							truth[a>>6] |= uint64(1) << uint(a&63)
						}
					}
					return truth, support, true
				}
				s.record(set, &mask)
			}
		}
		masks := [2][]uint64{mask[0][:], mask[1][:]}
		p.decomposition = s.history
		variable = p.bestImplication(masks)
		if variable < 0 && allowFallback {
			for i := len(s.fallback) - 1; i >= 0; i-- {
				if !slices.Contains(s.history, s.fallback[i]) {
					variable = s.fallback[i]
					break
				}
			}
			allowFallback = false
		}
	}
	if variable < 0 || slices.Contains(used, variable) {
		return nil, nil, false
	}
	s.history = append(s.history, variable)
	var truth [2][]uint64
	var support [2][]int32
	for branch := 0; branch < 2; branch++ {
		var next [2][SFM_SIM_WORDS]uint64
		for set := 0; set < 2; set++ {
			pattern := p.divisor(set, variable)
			for i, v := range pattern {
				if branch == 0 {
					v = ^v
				}
				next[set][i] = mask[set][i] & v
			}
		}
		var ok bool
		truth[branch], support[branch], ok = s.rec(append(append([]int(nil), assumptions...), 2*variable+(branch^1)), next, append(append([]int(nil), used...), variable), allowFallback, extra+1+len(support[0]))
		if !ok {
			return nil, nil, false
		}
	}
	result, resultSupport := sfmCombineDecomposition(truth[0], truth[1], support[0], support[1], int32(variable), s.limit)
	return result, resultSupport, result != nil
}
