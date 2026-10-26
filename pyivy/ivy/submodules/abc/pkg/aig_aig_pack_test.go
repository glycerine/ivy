package abc

// Tests for packed simulation and pattern care from aig/aig/aigPack.c.

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigPackFixture(seed int) *AigManager {
	g := NewAigManager()
	ci := make([]AigLit, 8)
	leaf := make([]AigLit, 8)
	for i := range ci {
		ci[i], _ = g.CreateCI()
		leaf[i] = ci[i].NotIf(seed&(1<<uint(i%5)) != 0)
	}
	node := make([]AigLit, 13)
	node[0], _ = g.And(leaf[0], leaf[1])
	node[1], _ = g.And(leaf[2], leaf[3])
	node[2], _ = g.And(node[0], node[1])
	node[3], _ = g.And(node[0].Not(), leaf[4])
	node[4], _ = g.And(node[2], node[3].Not())
	node[5], _ = g.And(node[1].Not(), node[3])
	node[6], _ = g.And(node[4], node[5].Not())
	g.CreateCO(node[6])
	g.CreateCO(node[4].Not())
	node[7], _ = g.And(leaf[0], leaf[1].Not())
	node[8], _ = g.And(node[0], node[7])
	node[9], _ = g.And(node[2], leaf[4])
	node[10], _ = g.And(node[9], leaf[5])
	node[11], _ = g.And(node[10], leaf[6])
	node[12], _ = g.And(node[11], leaf[7])
	return g
}

func aigPackLiterals(t testing.TB, seed, pattern int) []SatLit {
	t.Helper()
	result := make([]SatLit, 3)
	for i := range result {
		variable := (seed + pattern + 2*i) % 8
		negative := seed&(1<<uint((pattern+i)%5)) != 0
		result[i], _ = MakeSatLit(SatVar(variable), negative)
	}
	return result
}

func TestAigPatternPackNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_pack.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Seed                                                      int
		RandomPatterns, RandomSigns                               []uint64
		BasePatterns, BaseSigns                                   []uint64
		Rare, ExperimentCandidates                                []AigID
		TryFirst, TryConflict                                     int
		Patterns, Care, Signs                                     []uint64
		Total, Skipped, Repeated, CareBits                        int
		Histogram                                                 [33]int
		CountZero, FailedZero, CountOne, FailedOne, TotalLiterals int
		ExperimentPatterns, ExperimentCare, ExperimentSigns       []uint64
		ExperimentTotal, ExperimentSkipped, ExperimentRepeated    int
		ExperimentCareBits                                        int
		ExperimentHistogram                                       [33]int
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 32 {
		t.Fatal(len(rows), err)
	}
	for _, row := range rows {
		g := aigPackFixture(row.Seed)
		before := g.Clone()
		pack, err := NewAigPatternPack(g, AigPatternPackOptions{ResetRandom: true})
		if err != nil || !slices.Equal(pack.InputPatterns(), row.RandomPatterns) || !slices.Equal(pack.Signs(), row.RandomSigns) {
			t.Fatal(row.Seed, err, pack.InputPatterns(), row.RandomPatterns, pack.Signs(), row.RandomSigns)
		}
		if err = pack.SetInputPatterns(row.BasePatterns); err != nil {
			t.Fatal(row.Seed, err)
		}
		if err = pack.Simulate(); err != nil || !slices.Equal(pack.Signs(), row.BaseSigns) {
			t.Fatal(row.Seed, err, pack.Signs(), row.BaseSigns)
		}
		rare, err := pack.RareNodes()
		if err != nil || !slices.Equal(rare, row.Rare) {
			t.Fatal(row.Seed, rare, err, row.Rare)
		}
		added, err := pack.TryPattern(1, aigPackLiterals(t, row.Seed, 0))
		if err != nil || added != (row.TryFirst != 0) {
			t.Fatal(row.Seed, added, err, row.TryFirst)
		}
		conflict, _ := MakeSatLit(SatVar(row.Seed%8), row.Seed&1 == 0)
		added, err = pack.TryPattern(1, []SatLit{conflict})
		if err != nil || added != (row.TryConflict != 0) {
			t.Fatal(row.Seed, added, err, row.TryConflict)
		}
		for pattern := 1; pattern <= 8; pattern++ {
			if _, err = pack.AddPattern(aigPackLiterals(t, row.Seed, pattern)); err != nil {
				t.Fatal(row.Seed, pattern, err)
			}
		}
		if err = pack.Simulate(); err != nil {
			t.Fatal(row.Seed, err)
		}
		stats, err := pack.Stats()
		if err != nil || !slices.Equal(pack.InputPatterns(), row.Patterns) || !slices.Equal(pack.InputCare(), row.Care) || !slices.Equal(pack.Signs(), row.Signs) ||
			stats.PatternsTotal != row.Total || stats.PatternsSkipped != row.Skipped || stats.PatternsRepeated != row.Repeated || stats.CareBits != row.CareBits || stats.NodeHistogram != row.Histogram {
			t.Fatal(row.Seed, stats, err, row)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "source mutation")
		}
		experiment, err := g.RunJustificationExperiment(AigJustificationExperimentOptions{PatternPack: AigPatternPackOptions{ResetRandom: true}})
		if err != nil || !slices.Equal(experiment.Candidates, row.ExperimentCandidates) || experiment.CountZero != row.CountZero || experiment.FailedZero != row.FailedZero ||
			experiment.CountOne != row.CountOne || experiment.FailedOne != row.FailedOne || experiment.TotalLiterals != row.TotalLiterals ||
			!slices.Equal(experiment.Patterns, row.ExperimentPatterns) || !slices.Equal(experiment.Care, row.ExperimentCare) || !slices.Equal(experiment.Signs, row.ExperimentSigns) ||
			experiment.PatternStats.PatternsTotal != row.ExperimentTotal || experiment.PatternStats.PatternsSkipped != row.ExperimentSkipped ||
			experiment.PatternStats.PatternsRepeated != row.ExperimentRepeated || experiment.PatternStats.CareBits != row.ExperimentCareBits || experiment.PatternStats.NodeHistogram != row.ExperimentHistogram {
			t.Fatal(row.Seed, experiment, err, row)
		}
		if !reflect.DeepEqual(g, before) {
			t.Fatal(row.Seed, "experiment source mutation")
		}
	}
}

func TestAigPatternPackErrorsLimitsAndOwnership(t *testing.T) {
	g := aigPackFixture(0)
	for _, options := range []AigPatternPackOptions{{MaxObjects: -1}, {MaxPatterns: -1}, {MaxObjects: 1}, {Stop: func() bool { return true }}} {
		pack, err := NewAigPatternPack(g, options)
		if pack != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, pack, err)
		}
	}
	if pack, err := NewAigPatternPack(nil, AigPatternPackOptions{}); pack != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(pack, err)
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	if pack, err := NewAigPatternPack(buffered, AigPatternPackOptions{}); pack != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(pack, err)
	}

	pack, err := NewAigPatternPack(g, AigPatternPackOptions{MaxPatterns: 1})
	if err != nil {
		t.Fatal(err)
	}
	patterns := pack.InputPatterns()
	care := pack.InputCare()
	signs := pack.Signs()
	patterns[0], care[0], signs[0] = 0, 1, 0
	if reflect.DeepEqual(patterns, pack.InputPatterns()) || reflect.DeepEqual(care, pack.InputCare()) || reflect.DeepEqual(signs, pack.Signs()) {
		t.Fatal("getter aliases internal storage")
	}
	if err = pack.SetInputPatterns(nil); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	literalTrue, _ := MakeSatLit(0, false)
	literalFalse, _ := MakeSatLit(0, true)
	if added, err := pack.AddPattern([]SatLit{literalTrue}); err != nil || !added {
		t.Fatal(added, err)
	}
	if added, err := pack.AddPattern([]SatLit{literalFalse}); added || !errors.Is(err, ErrLimit) {
		t.Fatal(added, err)
	}
	for _, bit := range []int{-1, 64} {
		if added, err := pack.TryPattern(bit, nil); added || !errors.Is(err, ErrDimension) {
			t.Fatal(bit, added, err)
		}
	}
	invalid, _ := MakeSatLit(SatVar(g.InputCount()), false)
	if added, err := pack.TryPattern(2, []SatLit{invalid}); added || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(added, err)
	}

	pack, _ = NewAigPatternPack(g, AigPatternPackOptions{})
	if added, err := pack.TryPattern(2, []SatLit{literalTrue, literalFalse}); err != nil || !added || pack.InputPatterns()[0]&(1<<2) != 0 {
		t.Fatal(added, err, pack.InputPatterns()[0])
	}
	pack, _ = NewAigPatternPack(g, AigPatternPackOptions{})
	for bit := 1; bit < 64; bit++ {
		if added, err := pack.TryPattern(bit, []SatLit{literalTrue}); err != nil || !added {
			t.Fatal(bit, added, err)
		}
	}
	if added, err := pack.AddPattern([]SatLit{literalFalse}); err != nil || added {
		t.Fatal(added, err)
	}
	stats, _ := pack.Stats()
	if stats.PatternsTotal != 1 || stats.PatternsSkipped != 1 || stats.CareBits != 63 {
		t.Fatal(stats)
	}

	stale, _ := NewAigPatternPack(g, AigPatternPackOptions{})
	g.CreateCI()
	if err := stale.Simulate(); !errors.Is(err, ErrInvariant) {
		t.Fatal(err)
	}
}

func TestAigJustificationExperimentErrorsAndOwnership(t *testing.T) {
	g := aigPackFixture(4)
	for _, options := range []AigJustificationExperimentOptions{
		{MaxCandidates: -1},
		{MaxCandidates: 1, PatternPack: AigPatternPackOptions{ResetRandom: true}},
		{Justification: AigJustificationOptions{MaxVisits: 1}, PatternPack: AigPatternPackOptions{ResetRandom: true}},
		{PatternPack: AigPatternPackOptions{ResetRandom: true, MaxPatterns: 1}},
	} {
		result, err := g.RunJustificationExperiment(options)
		if result.Candidates != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit)) {
			t.Fatal(options, result, err)
		}
	}
	if result, err := g.RunJustificationExperiment(AigJustificationExperimentOptions{
		Justification: AigJustificationOptions{Stop: func() bool { return true }},
		PatternPack:   AigPatternPackOptions{ResetRandom: true},
	}); result.Candidates != nil || !errors.Is(err, ErrCancelled) {
		t.Fatal(result, err)
	}
	result, err := g.RunJustificationExperiment(AigJustificationExperimentOptions{PatternPack: AigPatternPackOptions{ResetRandom: true}})
	if err != nil || len(result.Candidates) == 0 || len(result.Patterns) == 0 || len(result.Signs) == 0 {
		t.Fatal(result, err)
	}
	again, err := g.RunJustificationExperiment(AigJustificationExperimentOptions{PatternPack: AigPatternPackOptions{ResetRandom: true}})
	if err != nil {
		t.Fatal(err)
	}
	result.Candidates[0], result.Patterns[0], result.Care[0], result.Signs[0] = 0, 0, 0, 0
	if again.Candidates[0] == 0 || again.Patterns[0] == 0 || reflect.DeepEqual(result.Care, again.Care) || again.Signs[0] == 0 {
		t.Fatal("experiment result aliases later result", result, again)
	}
}

func FuzzAigPatternPack(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		signals := make([]AigLit, 4)
		for i := range signals {
			signals[i], _ = g.CreateCI()
		}
		for i := 0; i+1 < len(data) && len(signals) < 28; i += 2 {
			left := signals[int(data[i]&31)%len(signals)].NotIf(data[i]&0x80 != 0)
			right := signals[int(data[i+1]&31)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			literal, err := g.And(left, right)
			if err == nil {
				signals = append(signals, literal)
			}
		}
		root := signals[len(signals)-1]
		g.CreateCO(root)
		pack, err := NewAigPatternPack(g, AigPatternPackOptions{})
		if err != nil {
			t.Fatal(err)
		}
		var words [4]uint64
		for i := range words {
			var bytes [8]byte
			for j := range bytes {
				if len(data) != 0 {
					bytes[j] = data[(8*i+j)%len(data)]
				}
			}
			words[i] = binary.LittleEndian.Uint64(bytes[:])
		}
		if err = pack.SetInputPatterns(words[:]); err != nil {
			t.Fatal(err)
		}
		if err = pack.Simulate(); err != nil {
			t.Fatal(err)
		}
		got := pack.Signs()[g.cos[0]]
		for bit := 0; bit < 64; bit++ {
			inputs := make([]bool, 4)
			for i := range inputs {
				inputs[i] = words[i]&(uint64(1)<<bit) != 0
			}
			outputs, err := g.Evaluate(inputs)
			if err != nil || outputs[0] != (got&(uint64(1)<<bit) != 0) {
				t.Fatal(bit, outputs, err, got)
			}
		}
	})
}
