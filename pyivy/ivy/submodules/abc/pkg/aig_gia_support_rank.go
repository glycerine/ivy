package abc

// Output-support statistics from src/aig/gia/giaDup.c:Gia_ManCoSuppSizeMax,
// Gia_ManCoLargestSupp and Gia_ManSortCoBySuppSize.

// GiaSupportRanking uses C's minimum Maximum of one, even for no outputs.
// Largest is the first largest-support CO, or -1 for no outputs. Order sorts
// by decreasing support size with the pinned C quicksort's nonstable ties.
type GiaSupportRanking struct {
	Maximum, Largest int
	Order            []int
}

// RankOutputSupports computes structural supports and C-compatible statistics.
func (m *GiaManager) RankOutputSupports(maxObjects, maxEntries int) (GiaSupportRanking, error) {
	sets, err := m.OutputSupports(maxObjects, maxEntries)
	if err != nil {
		return GiaSupportRanking{}, err
	}
	result := GiaSupportRanking{Maximum: 1, Largest: -1, Order: make([]int, len(sets))}
	largest := -1
	for i, set := range sets {
		result.Order[i] = i
		if len(set) > largest {
			largest = len(set)
			result.Largest = i
		}
		result.Maximum = max(result.Maximum, len(set))
	}
	// Abc_QuickSort3 actually delegates to QuickSort2 and compares low words.
	// Process the smaller partition immediately to bound auxiliary stack size.
	type span struct{ lo, hi int }
	pending := []span{{0, len(sets) - 1}}
	for len(pending) > 0 {
		s := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for s.lo < s.hi {
			if s.hi-s.lo < 10 {
				for i := s.lo; i < s.hi; i++ {
					best := i
					for j := i + 1; j <= s.hi; j++ {
						if len(sets[result.Order[j]]) > len(sets[result.Order[best]]) {
							best = j
						}
					}
					result.Order[i], result.Order[best] = result.Order[best], result.Order[i]
				}
				break
			}
			pivot := len(sets[result.Order[s.hi]])
			i, j := s.lo-1, s.hi
			for {
				i++
				for len(sets[result.Order[i]]) > pivot {
					i++
				}
				j--
				for pivot > len(sets[result.Order[j]]) {
					if j == s.lo {
						break
					}
					j--
				}
				if i >= j {
					break
				}
				result.Order[i], result.Order[j] = result.Order[j], result.Order[i]
			}
			result.Order[i], result.Order[s.hi] = result.Order[s.hi], result.Order[i]
			left, right := span{s.lo, i - 1}, span{i + 1, s.hi}
			if left.hi-left.lo < right.hi-right.lo {
				pending = append(pending, right)
				s = left
			} else {
				pending = append(pending, left)
				s = right
			}
		}
	}
	return result, nil
}
