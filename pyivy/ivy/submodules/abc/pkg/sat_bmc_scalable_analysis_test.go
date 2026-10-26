package abc

// Go-only analysis tests for src/sat/bmc/bmcBmc3.c.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestAigBMC3TernaryPoliciesAndCone(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	register, _ := graph.CreateCI()
	graph.CreateCO(register)
	graph.CreateCO(input)
	graph.SetRegisterCount(1)
	before := slices.Clone(graph.objects)
	stable, err := graph.AnalyzeBMC3Ternary(AigBMC3TernaryOptions{Mode: AigBMC3StableRegisters})
	if err != nil {
		t.Fatal(err)
	}
	if !stable.Complete || len(stable.Values) != 2 || !slices.Equal(stable.RegisterBinaryCounts, []int{0, 0}) || !slices.Equal(stable.OutputBinaryCounts, []int{1, 0}) {
		t.Fatal(stable.Complete, len(stable.Values), stable.RegisterBinaryCounts, stable.OutputBinaryCounts)
	}
	unknown, err := graph.AnalyzeBMC3Ternary(AigBMC3TernaryOptions{Mode: AigBMC3FirstUnknownOutput, MaxFrames: 4})
	if err != nil || !unknown.Complete || unknown.FirstUnknownOutputFrame != 1 || len(unknown.Values) != 2 {
		t.Fatal(unknown.Complete, unknown.FirstUnknownOutputFrame, len(unknown.Values), err)
	}
	counts, err := unknown.UnknownConeCounts(1, 0)
	if err != nil || !slices.Equal(counts, []int{2, 1}) {
		t.Fatal(counts, err)
	}
	unknown.Values[0][0] = AigTernaryZero
	again, err := graph.AnalyzeBMC3Ternary(AigBMC3TernaryOptions{Mode: AigBMC3FirstUnknownOutput, MaxFrames: 4})
	if err != nil || again.Values[0][0] != AigTernaryOne {
		t.Fatal(err)
	}
	if !slices.Equal(graph.objects, before) {
		t.Fatal("ternary analysis mutated graph")
	}
	if _, err = unknown.UnknownConeCounts(2, 0); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestAigBMC3TernaryLimits(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0xc, 0, 0xa)
	for _, options := range []AigBMC3TernaryOptions{
		{MaxFrames: -1},
		{MaxObjects: 1},
		{MaxEntries: graph.ObjectCount()},
		{Stop: func() bool { return true }},
	} {
		result, err := graph.AnalyzeBMC3Ternary(options)
		if err == nil || result.Values != nil {
			t.Fatal(options, result, err)
		}
	}
	buffered := NewAigManager()
	a, _ := buffered.CreateCI()
	b, _ := buffered.CreateBuffer(a)
	buffered.CreateCO(b)
	buffered.CreateCO(a)
	buffered.SetRegisterCount(1)
	if _, err := buffered.AnalyzeBMC3Ternary(AigBMC3TernaryOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
}

func TestAigBMC3Sections(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	register, _ := graph.CreateCI()
	graph.CreateCO(register)
	graph.CreateCO(input)
	graph.SetRegisterCount(1)
	sections, err := graph.BMC3Sections(AigBMC3SectionOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(sections) != 2 || !slices.Equal(sections[0], []AigID{graph.cis[1]}) || !slices.Equal(sections[1], []AigID{graph.cis[0]}) {
		t.Fatal(sections)
	}
	sections[0][0] = 0
	again, _ := graph.BMC3Sections(AigBMC3SectionOptions{})
	if again[0][0] == 0 {
		t.Fatal("sections alias prior result")
	}
	for _, options := range []AigBMC3SectionOptions{{MaxSections: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		if _, err = graph.BMC3Sections(options); err == nil {
			t.Fatal(options)
		}
	}
}

func TestAigBMC3Supergate(t *testing.T) {
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	and, _ := graph.And(a.Not(), b.Not())
	graph.CreateCO(and.Not())
	graph.CreateCO(a)
	supergate, err := graph.BMC3Supergate(0, 0)
	if err != nil || !slices.Equal(supergate.Leaves, []AigID{a.ID(), b.ID()}) || supergate.Referenced != 1 {
		t.Fatal(supergate, err)
	}
	single, err := graph.BMC3Supergate(1, 0)
	if err != nil || !slices.Equal(single.Leaves, []AigID{a.ID()}) || single.Referenced != 1 {
		t.Fatal(single, err)
	}
	if _, err = graph.BMC3Supergate(0, 1); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	order, err := graph.BMC3ReferenceOrder([]AigID{and.ID(), b.ID(), a.ID()})
	if err != nil || !slices.Equal(order, []AigID{b.ID(), and.ID(), a.ID()}) {
		t.Fatal(order, err)
	}
}

func TestBMC3TruthReduction(t *testing.T) {
	truth := uint16(0xca00)
	reduced, literals, err := BMC3ReduceTruth(truth, [4]int{0, 2, 3, 4})
	if err != nil || literals[0] != -1 {
		t.Fatal(reduced, literals, err)
	}
	for assignment := 0; assignment < 16; assignment++ {
		if assignment&1 != 0 {
			continue
		}
		if truth>>assignment&1 != reduced>>assignment&1 {
			t.Fatal(assignment, truth, reduced)
		}
	}
	zero, _ := BMC3TruthCofactor(0xaaaa, 0, false)
	one, _ := BMC3TruthCofactor(0xaaaa, 0, true)
	if zero != 0 || one != 0xffff {
		t.Fatal(zero, one)
	}
	if _, err = BMC3TruthCofactor(0, 4, false); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestAigBMC3MappingAndBLIF(t *testing.T) {
	graph := mappedCNFFixture(t, 0, 0)
	before := slices.Clone(graph.objects)
	mapping, err := graph.BMC3Mapping(DARCutOptions{}, CNFMappingArrayOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(mapping.Packed) < graph.ObjectCount() || len(mapping.Cuts) == 0 || len(mapping.References) != graph.ObjectCount() {
		t.Fatal(len(mapping.Packed), len(mapping.Cuts), len(mapping.References))
	}
	manual := make([]uint32, graph.ObjectCount())
	for _, id := range graph.cos {
		manual[graph.objects[id].Fanin0.ID()]++
	}
	for _, cut := range mapping.Cuts {
		for _, leaf := range cut.Leaves {
			manual[leaf]++
		}
	}
	if !slices.Equal(mapping.References, manual) || !slices.Equal(graph.objects, before) {
		t.Fatal("mapping references or source ownership")
	}
	text, err := mapping.FormatBLIF(graph)
	if err != nil || !strings.HasPrefix(text, ".model test\n.inputs") || !strings.Contains(text, ".outputs") || !strings.HasSuffix(text, ".end\n") {
		t.Fatal(text, err)
	}
	path := filepath.Join(t.TempDir(), "mapping.blif")
	if err = mapping.WriteBLIFFile(graph, path); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != text {
		t.Fatal(err)
	}
}

func TestAigBMC3AnalysisDeep(t *testing.T) {
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	root := a
	for index := 0; index < 10000; index++ {
		var err error
		root, err = graph.And(root, b.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.CreateCO(root.Not())
	supergate, err := graph.BMC3Supergate(0, 20000)
	if err != nil || len(supergate.Leaves) == 0 {
		t.Fatal(len(supergate.Leaves), err)
	}
	sections, err := graph.BMC3Sections(AigBMC3SectionOptions{MaxVisits: 20000})
	if err != nil || len(sections) != 1 || len(sections[0]) != 10002 {
		t.Fatal(len(sections), err)
	}
}

func FuzzBMC3TruthReduction(f *testing.F) {
	f.Add(uint16(0xaaaa), int8(0), int8(2), int8(3), int8(-1))
	f.Add(uint16(0x6996), int8(1), int8(2), int8(4), int8(6))
	f.Fuzz(func(t *testing.T, truth uint16, a, b, c, d int8) {
		toLiteral := func(value int8) int {
			switch value & 3 {
			case 0:
				return 0
			case 1:
				return 1
			case 2:
				return -1
			default:
				return int(uint8(value)) + 2
			}
		}
		original := [4]int{toLiteral(a), toLiteral(b), toLiteral(c), toLiteral(d)}
		reduced, literals, err := BMC3ReduceTruth(truth, original)
		if errors.Is(err, ErrInvariant) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 16; assignment++ {
			compatible := true
			for variable, literal := range original {
				if literal == 0 && assignment>>variable&1 != 0 || literal == 1 && assignment>>variable&1 == 0 {
					compatible = false
				}
			}
			if compatible && truth>>assignment&1 != reduced>>assignment&1 {
				t.Fatalf("truth=%x reduced=%x assignment=%d", truth, reduced, assignment)
			}
		}
		for variable, literal := range literals {
			if literal == -1 && !bmc3TruthIndependent(reduced, variable) {
				t.Fatal("removed dependent variable", variable)
			}
		}
	})
}

func FuzzAigBMC3TernaryPolicy(f *testing.F) {
	f.Add(uint8(0xc), uint8(0), uint8(0xa), uint8(4))
	f.Add(uint8(0x8), uint8(0x6), uint8(0xc), uint8(3))
	f.Fuzz(func(t *testing.T, output0, output1, next, frameByte uint8) {
		frames := int(frameByte%5) + 1
		graph := aigIntervalTruthGraph(t, output0&15, output1&15, next&15)
		got, err := graph.AnalyzeBMC3Ternary(AigBMC3TernaryOptions{Mode: AigBMC3FirstUnknownOutput, MaxFrames: frames})
		if err != nil {
			t.Fatal(err)
		}
		state := []uint8{1}
		wantFirst := -1
		for frame := 0; frame < frames; frame++ {
			values := make([]uint8, graph.ObjectCount())
			values[0] = 2
			values[graph.cis[0]] = 3
			values[graph.cis[1]] = state[0]
			read := func(literal AigLit) uint8 {
				value := values[literal.ID()]
				if !literal.Complemented() {
					return value
				}
				return (value&1)<<1 | (value&2)>>1
			}
			for id, object := range graph.objects {
				switch object.Type {
				case AigAND:
					a, b := read(object.Fanin0), read(object.Fanin1)
					var possible uint8
					for av := 0; av < 2; av++ {
						for bv := 0; bv < 2; bv++ {
							if a>>av&1 != 0 && b>>bv&1 != 0 {
								possible |= 1 << uint(av&bv)
							}
						}
					}
					values[id] = possible
				case AigCO:
					values[id] = read(object.Fanin0)
				}
			}
			if frame >= len(got.Values) {
				t.Fatal("analysis stopped before independent model", frame)
			}
			for id, value := range values {
				if uint8(got.Values[frame][id]) != value {
					t.Fatalf("frame=%d id=%d got=%d want=%d", frame, id, got.Values[frame][id], value)
				}
			}
			unknown := false
			for _, id := range graph.cos[:graph.PrimaryOutputCount()] {
				unknown = unknown || values[id] == 3
			}
			state[0] = values[graph.cos[graph.PrimaryOutputCount()]]
			if unknown {
				wantFirst = frame
				break
			}
		}
		if got.FirstUnknownOutputFrame != wantFirst {
			t.Fatal(got.FirstUnknownOutputFrame, wantFirst)
		}
	})
}
