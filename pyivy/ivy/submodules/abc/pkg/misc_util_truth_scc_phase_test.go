package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestTruthSCCPhaseNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/truth_scc_phase.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Mode, High, Kind, PhaseInputs, CandidateLog2, PermutationCost int
		Source, Masks                                                            []uint64
		PhaseGroups                                                              []int
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 540 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		source, e := TruthTableFromWords(row.Variables, row.Source, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := source.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseMode(row.Kind), HighSymmetry: row.High != 0})
		if e != nil || r.PhaseInputs != row.PhaseInputs || r.CandidateLog2 != row.CandidateLog2 || r.PermutationCost != row.PermutationCost || !slices.Equal(r.Masks, row.Masks) || !slices.Equal(r.PhaseGroups, row.PhaseGroups) {
			t.Fatal(index, row, r, e)
		}
		if !slices.Equal(source.Words(), row.Source) {
			t.Fatal("source changed")
		}
	}
}
func TestTruthSCCPhaseLimitsAndOwnership(t *testing.T) {
	source, _ := NewVariableTruthTable(18, 17, 0)
	before := source.Words()
	for _, tc := range []struct {
		options TruthSCCPhaseOptions
		want    error
	}{
		{TruthSCCPhaseOptions{Mode: 3}, ErrDimension}, {TruthSCCPhaseOptions{MaxConfigurations: -1}, ErrDimension},
		{TruthSCCPhaseOptions{MaxConfigurations: 1}, ErrLimit}, {TruthSCCPhaseOptions{Stop: func() bool { return true }}, ErrCancelled},
	} {
		r, e := source.FilterAdaptivePhases(tc.options)
		if r.Masks != nil || !errors.Is(e, tc.want) {
			t.Fatal(r, e)
		}
	}
	if r, e := (*TruthTable)(nil).FilterAdaptivePhases(TruthSCCPhaseOptions{}); r.Masks != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(r, e)
	}
	calls := 0
	r, e := source.FilterAdaptivePhases(TruthSCCPhaseOptions{Stop: func() bool { calls++; return calls == 4 }})
	if r.Masks != nil || !errors.Is(e, ErrCancelled) {
		t.Fatal(r, e)
	}
	r, e = source.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseSignature, HighSymmetry: true})
	if e != nil || r.Visited != 2 || r.PhaseInputs != 1 || r.PhaseGroups[0] != 17 {
		t.Fatal(r, e)
	}
	r.PhaseGroups[0] = 99
	r.Masks[0] = 99
	r.Coefficients[0] = 99
	fresh, e := source.FilterAdaptivePhases(TruthSCCPhaseOptions{Mode: TruthSCCPhaseSignature, HighSymmetry: true})
	if e != nil || fresh.PhaseGroups[0] != 17 || fresh.Masks[0] == 99 || fresh.Coefficients[0] == 99 || !slices.Equal(source.Words(), before) {
		t.Fatal("ownership", e)
	}
	for _, value := range []bool{false, true} {
		source, _ := NewConstantTruthTable(0, value, 0)
		r, e := source.FilterAdaptivePhases(TruthSCCPhaseOptions{})
		if e != nil || !slices.Equal(r.Masks, []uint64{0}) || r.Total != 0 || r.Visited != 1 {
			t.Fatal("zero input", r, e)
		}
	}
	if cost := truthSCCPermutationCost([]TruthVariableGroup{{0, 18}}, make([]uint64, 18)); cost != 5251 {
		t.Fatal("dynamic factorial cost", cost)
	}
}

// Build every phase assignment independently from a Boolean model; enumerate
// masks numerically and compare signatures without Gray traversal or state
// mutation. Native fixtures separately verify the retained masks' visit order.
func checkTruthSCCPhaseModel(t *testing.T, source *TruthTable, options TruthSCCPhaseOptions) {
	t.Helper()
	reduction, e := source.ReduceInputSymmetry(options.HighSymmetry)
	if e != nil {
		t.Fatal(e)
	}
	actual, e := source.FilterAdaptivePhases(options)
	if e != nil {
		t.Fatal(e)
	}
	n := len(actual.PhaseGroups)
	if actual.Visited != 1<<uint(n) {
		t.Fatal("visited", actual)
	}
	width := max(6, source.variables)
	minimum := []uint64(nil)
	masks := []uint64{}
	var bestTotal uint64
	var bestCoefficients []uint64
	bestCost := 0
	for mask := 0; mask < 1<<uint(n); mask++ {
		flips := 0
		for i, head := range actual.PhaseGroups {
			if mask&(1<<uint(i)) != 0 {
				for v := head; v >= 0; v = reduction.Links[v] {
					if reduction.PhaseWeights[v] != 0 {
						flips ^= 1 << uint(v)
					}
				}
			}
		}
		order := reduction.ExpandedOrder
		total := uint64(0)
		zero := make([]uint64, source.variables)
		for x := 0; x < 1<<uint(width); x++ {
			old := flips
			for target, v := range order {
				if x&(1<<uint(target)) != 0 {
					old ^= 1 << uint(v)
				}
			}
			value, _ := reduction.Table.Get(uint64(old))
			if !value {
				continue
			}
			weight := uint64(1)
			for bit := 0; bit < width; bit++ {
				if x&(1<<uint(bit)) != 0 {
					weight *= 2
				}
			}
			total += weight
			for v := range zero {
				if x&(1<<uint(v)) == 0 {
					zero[v] += weight
				}
			}
		}
		coefficients := []uint64{}
		position := 0
		for _, head := range reduction.Representatives {
			coefficients = append(coefficients, zero[position])
			for v := head; v >= 0; v = reduction.Links[v] {
				position++
			}
		}
		for _, group := range reduction.Groups {
			slices.Sort(coefficients[group.Start : group.Start+group.Size])
		}
		cost := truthSCCPermutationCost(reduction.Groups, coefficients)
		signature := []uint64{}
		if options.Mode == TruthSCCPhaseCost {
			signature = append(signature, uint64(cost))
		}
		signature = append(signature, total)
		if options.Mode != TruthSCCPhaseTotal {
			signature = append(signature, coefficients...)
		}
		compare := -1
		if minimum != nil {
			compare = slices.Compare(signature, minimum)
		}
		if compare < 0 {
			minimum = signature
			masks = nil
			bestTotal = total
			bestCoefficients = slices.Clone(coefficients)
			bestCost = cost
		}
		if compare <= 0 {
			masks = append(masks, uint64(mask))
		}
	}
	if reduction.PhasesDetermined {
		masks = nil
	}
	got := slices.Clone(actual.Masks)
	slices.Sort(got)
	if !slices.Equal(got, masks) || actual.Total != bestTotal {
		t.Fatal("minimum phases", options.Mode, actual, masks, bestTotal)
	}
	if options.Mode != TruthSCCPhaseTotal {
		if !slices.Equal(actual.Coefficients, bestCoefficients) || actual.PermutationCost != bestCost {
			t.Fatal("signature", actual, bestCoefficients, bestCost)
		}
	}
}
func FuzzTruthSCCPhase(f *testing.F) {
	f.Add(uint8(4), uint64(0x96254718abcdef31))
	f.Add(uint8(128), uint64(0))
	f.Fuzz(func(t *testing.T, flags uint8, word uint64) {
		variables := int(flags % 6)
		width := 1 << uint(variables)
		word &= (uint64(1) << uint(width)) - 1
		for shift := width; shift < 64; shift *= 2 {
			word |= word << uint(shift)
		}
		source, e := TruthTableFromWords(variables, []uint64{word}, 0)
		if e != nil {
			t.Fatal(e)
		}
		options := TruthSCCPhaseOptions{Mode: TruthSCCPhaseMode((flags / 6) % 3), HighSymmetry: flags&128 != 0}
		checkTruthSCCPhaseModel(t, source, options)
		if source.words[0] != word {
			t.Fatal("source changed")
		}
	})
}
