package abc

// Initial adaptive tied groups from opt/dau/dauCanon.c:
// Abc_TgInitMan, Abc_TgSplitGroup and Abc_TgCreateGroups.

// TruthVariableGroup identifies a contiguous range in a tied input order.
type TruthVariableGroup struct{ Start, Size int }

// TruthTiedGroups records the adaptive canonicalizer's initial signatures.
// Table has only output/input phases applied; input permutation is deferred.
// InputOrder lists source input indices by ascending zero-cofactor population,
// preserving source order within ties. ZeroCofactorCounts follows InputOrder.
// InputPhases is source-indexed. Groups partition InputOrder by signature.
// PhasesDetermined is true when every group's phase is fixed by its signature.
// All slices and Table own their storage.
type TruthTiedGroups struct {
	Table              *TruthTable
	InputOrder         []int
	InputPhases        []bool
	OutputPhase        bool
	PackedOnes         int
	ZeroCofactorCounts []int
	Groups             []TruthVariableGroup
	PhasesDetermined   bool
}

// truthSplitTiedGroup stably sorts one group's labels and signatures, then
// splits it at distinct signatures. coefficients is local to that group and
// is sorted in place. Unlike semicanonicalization's selection sort, ties never
// change their input order. The table itself is not permuted.
func truthSplitTiedGroup(order []int, groups []TruthVariableGroup, index int, coefficients []int) []TruthVariableGroup {
	group := groups[index]
	labels := order[group.Start : group.Start+group.Size]
	for i := 1; i < len(labels); i++ {
		value, label := coefficients[i], labels[i]
		j := i
		for j > 0 && coefficients[j-1] > value {
			coefficients[j], labels[j] = coefficients[j-1], labels[j-1]
			j--
		}
		coefficients[j], labels[j] = value, label
	}
	split := make([]TruthVariableGroup, 0, len(labels))
	for start := 0; start < len(labels); {
		end := start + 1
		for end < len(labels) && coefficients[end] == coefficients[start] {
			end++
		}
		split = append(split, TruthVariableGroup{group.Start + start, end - start})
		start = end
	}
	result := make([]TruthVariableGroup, 0, len(groups)-1+len(split))
	result = append(result, groups[:index]...)
	result = append(result, split...)
	result = append(result, groups[index+1:]...)
	return result
}

// CreateTiedGroups builds population-signature groups for adaptive search.
// Packed counts include small-table repeats. The output has at most half ones;
// each input's zero cofactor has at least as many ones as its one cofactor.
// Sorting group labels does not yet apply the corresponding truth permutation.
// A zero-input table returns an empty partition, extending C's nonempty setup.
func (t *TruthTable) CreateTiedGroups() (TruthTiedGroups, error) {
	pop, e := t.CofactorPopulations()
	if e != nil {
		return TruthTiedGroups{}, e
	}
	r := TruthTiedGroups{Table: t.Clone(), InputOrder: make([]int, t.variables), InputPhases: make([]bool, t.variables), PackedOnes: pop.Ones, ZeroCofactorCounts: pop.Zero}
	if r.PackedOnes > len(t.words)*32 {
		r.OutputPhase = true
		r.PackedOnes = len(t.words)*64 - r.PackedOnes
		for i, w := range r.Table.words {
			r.Table.words[i] = ^w
		}
		for i, c := range r.ZeroCofactorCounts {
			r.ZeroCofactorCounts[i] = len(t.words)*32 - c
		}
	}
	for i, c := range r.ZeroCofactorCounts {
		r.InputOrder[i] = i
		if c < r.PackedOnes-c {
			r.Table.flipVariable(i)
			r.InputPhases[i] = true
			r.ZeroCofactorCounts[i] = r.PackedOnes - c
		}
	}
	if t.variables > 0 {
		r.Groups = truthSplitTiedGroup(r.InputOrder, []TruthVariableGroup{{0, t.variables}}, 0, r.ZeroCofactorCounts)
		r.PhasesDetermined = r.ZeroCofactorCounts[0]*2 != r.PackedOnes
	}
	return r, nil
}
