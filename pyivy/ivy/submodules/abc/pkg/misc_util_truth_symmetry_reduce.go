package abc

// Adaptive symmetry reduction from opt/dau/dauCanon.c: Abc_TgGroupSymmetry,
// Abc_TgPurgeSymmetry, Abc_TgExpendSymmetry and symmetry-phase helpers.

// TruthSymmetryReduction contains phase-normalized truth and a deferred input
// order. Representatives are partitioned by Groups; each representative heads
// a chain in Links. Links has one extra entry, at Variables(), heading unused
// inputs in C's reverse encounter order. ExpandedOrder expands representative
// chains then appends unused inputs. SupportedInputs counts expanded active
// inputs. PhaseWeights preserves C's accumulated symPhase values, including
// zero for phase-redundant inputs. InputPhases is source-indexed; the truth
// permutation remains deferred. All storage is owned by this result.
type TruthSymmetryReduction struct {
	Table                              *TruthTable
	Representatives                    []int
	Groups                             []TruthVariableGroup
	Links, PhaseWeights, ExpandedOrder []int
	InputPhases                        []bool
	OutputPhase, PhasesDetermined      bool
	SupportedInputs                    int
}

func (r *TruthSymmetryReduction) chain(head int) []int {
	result := []int{}
	for v := head; v >= 0; v = r.Links[v] {
		result = append(result, v)
	}
	return result
}
func (r *TruthSymmetryReduction) flipChain(head int) {
	for v := head; v >= 0; v = r.Links[v] {
		if r.PhaseWeights[v] != 0 {
			r.Table.flipVariable(v)
			r.InputPhases[v] = !r.InputPhases[v]
		}
	}
}
func (r *TruthSymmetryReduction) reduceGroup(index int, high bool) int {
	group := r.Groups[index]
	labels := r.Representatives[group.Start : group.Start+group.Size]
	done := make([]bool, len(labels))
	sizes := make([]int, len(labels))
	kinds := make([]TruthSymmetry, len(labels))
	for i := range sizes {
		sizes[i] = 1
	}
	includePhase := index == 0 && !r.PhasesDetermined
	removed := 0
	for {
		modified := false
		for i := 0; i < len(labels)-1; i++ {
			a := labels[i]
			if a < 0 || done[i] {
				continue
			}
			for j := i + 1; j < len(labels); j++ {
				b := labels[j]
				kinds[j] = 0
				if b < 0 || sizes[j] != sizes[i] {
					continue
				}
				if sizes[j] == 1 {
					kinds[j], _ = r.Table.InputSymmetry(a, b, includePhase)
				} else {
					phases := make([]bool, r.Table.variables)
					for v, w := range r.PhaseWeights {
						phases[v] = w != 0
					}
					kinds[j], _ = r.Table.GroupSymmetry(r.chain(a), r.chain(b), phases, includePhase)
				}
			}
			done[i] = true
			// C computes all comparisons before merging; retain that decision order.
			for j := i + 1; j < len(labels); j++ {
				b := labels[j]
				switch kinds[j] {
				case TruthSymmetryPhased:
					r.flipChain(b)
					r.PhaseWeights[a] += r.PhaseWeights[b]
				case TruthSymmetryDirect:
					r.PhaseWeights[a] += r.PhaseWeights[b]
				case TruthSymmetryBoth:
					for v := b; v >= 0; v = r.Links[v] {
						r.PhaseWeights[v] = 0
					}
				default:
					continue
				}
				tail := a
				for r.Links[tail] >= 0 {
					tail = r.Links[tail]
				}
				r.Links[tail] = b
				labels[j] = -1
				sizes[i] += sizes[j]
				modified = true
				done[i] = false
				removed++
			}
		}
		if !high || !modified {
			break
		}
	}
	return removed
}

// ReduceInputSymmetry creates adaptive signature groups then merges symmetric
// input chains. high enables repeated equal-size group comparisons, in addition
// to individual-input symmetry. Unsupported inputs are kept in a separate
// chain. This ports the reduction stage, not adaptive canonicalization/search.
func (t *TruthTable) ReduceInputSymmetry(high bool) (TruthSymmetryReduction, error) {
	initial, e := t.CreateTiedGroups()
	if e != nil {
		return TruthSymmetryReduction{}, e
	}
	r := TruthSymmetryReduction{Table: initial.Table, Representatives: initial.InputOrder, Groups: initial.Groups, Links: make([]int, t.variables+1), PhaseWeights: make([]int, t.variables), InputPhases: initial.InputPhases, OutputPhase: initial.OutputPhase, PhasesDetermined: initial.PhasesDetermined}
	for i := range r.Links {
		r.Links[i] = -1
	}
	for i := range r.PhaseWeights {
		r.PhaseWeights[i] = 1
	}
	removed := make([]int, len(r.Groups))
	if len(r.Groups) > 0 && !r.PhasesDetermined {
		for j := 0; j < r.Groups[0].Size; j++ {
			v := r.Representatives[j]
			used, _ := r.Table.HasVariable(v)
			if !used {
				r.Links[v] = r.Links[t.variables]
				r.Links[t.variables] = v
				r.Representatives[j] = -1
				removed[0]++
			}
		}
	}
	for index := range r.Groups {
		removed[index] += r.reduceGroup(index, high)
	}
	labels := make([]int, 0, t.variables)
	for _, v := range r.Representatives {
		if v >= 0 {
			labels = append(labels, v)
		}
	}
	r.Representatives = labels
	sum := 0
	for i := range r.Groups {
		r.Groups[i].Size -= removed[i]
		r.Groups[i].Start -= sum
		sum += removed[i]
	}
	if len(r.Groups) > 0 && r.Groups[0].Size == 0 {
		r.Groups = r.Groups[1:]
	}
	r.ExpandedOrder = make([]int, 0, t.variables)
	for _, v := range r.Representatives {
		r.ExpandedOrder = append(r.ExpandedOrder, r.chain(v)...)
	}
	r.SupportedInputs = len(r.ExpandedOrder)
	r.ExpandedOrder = append(r.ExpandedOrder, r.chain(r.Links[t.variables])...)
	return r, nil
}
