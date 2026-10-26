package abc

// Tests for timeframe expansion from aig/aig/aigFrames.c.

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"slices"
	"testing"
)

func aigFramesFixture(t testing.TB, seed int) *AigManager {
	t.Helper()
	g := NewAigManager()
	g.Name, g.Spec = "frames-native", "frames-spec"
	ci := make([]AigLit, 4)
	for i := range ci {
		ci[i], _ = g.CreateCI()
	}
	n0, _ := g.And(ci[0].NotIf(seed&1 != 0), ci[2].NotIf(seed&2 != 0))
	n1, _ := g.And(ci[1].NotIf(seed&4 != 0), ci[3].NotIf(seed&8 != 0))
	n2, _ := g.And(n0.NotIf(seed&16 != 0), n1.NotIf(seed&32 != 0))
	g.And(ci[0], ci[1])
	g.CreateCO(n2.NotIf(seed&64 != 0))
	g.CreateCO(n0.NotIf(seed&128 != 0))
	g.CreateCO(n1.NotIf(seed&256 != 0))
	g.CreateCO(n2.NotIf(seed&512 != 0))
	if err := g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	return g
}

func aigFramesObjects(g *AigManager) [][]int {
	result := make([][]int, len(g.objects))
	for i, object := range g.objects {
		row := []int{int(object.Type), -1, -1, int(object.RefCount), int(object.Level), 0}
		if object.Type != AigNone && object.Type != AigConstant1 && object.Type != AigCI {
			row[1] = int(object.Fanin0)
		}
		if object.Type == AigAND || object.Type == AigXOR {
			row[2] = int(object.Fanin1)
		}
		if object.Phase {
			row[5] = 1
		}
		result[i] = row
	}
	return result
}

func TestAigFramesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/aig_frames.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Frames                                              int
		Initialize, Outputs, Registers, Enlarge             int
		Inputs, PrimaryInputs, OutputsCount, PrimaryOutputs int
		RegisterCount                                       int
		Objects                                             [][]int
		Mapping                                             []int64
	}
	if err = json.Unmarshal(data, &rows); err != nil || len(rows) != 48 {
		t.Fatal(len(rows), err)
	}
	for i, row := range rows {
		source := aigFramesFixture(t, row.Frames*257+i%16*19)
		before := source.Clone()
		expansion, err := source.ExpandFrames(AigFramesOptions{
			Frames: row.Frames, Initialize: row.Initialize != 0, Outputs: row.Outputs != 0,
			Registers: row.Registers != 0, Enlarge: row.Enlarge != 0,
		})
		if err != nil {
			t.Fatal(i, err)
		}
		g := expansion.Graph()
		if g.Name != source.Name || g.Spec != source.Spec || g.InputCount() != row.Inputs ||
			g.PrimaryInputCount() != row.PrimaryInputs || g.OutputCount() != row.OutputsCount ||
			g.PrimaryOutputCount() != row.PrimaryOutputs || g.RegisterCount() != row.RegisterCount {
			t.Fatal(i, g.InputCount(), g.PrimaryInputCount(), g.OutputCount(), g.PrimaryOutputCount(), g.RegisterCount())
		}
		if got := aigFramesObjects(g); !reflect.DeepEqual(got, row.Objects) {
			t.Fatalf("case %d objects\ngot  %v\nwant %v", i, got, row.Objects)
		}
		mapping := expansion.Mapping()
		if len(mapping) != len(row.Mapping) {
			t.Fatal(i, len(mapping), len(row.Mapping))
		}
		for j, want := range row.Mapping {
			if want < 0 {
				if mapping[j] != InvalidAigLit {
					t.Fatal(i, j, mapping[j], want)
				}
			} else if mapping[j] != AigLit(want) {
				t.Fatal(i, j, mapping[j], want)
			}
		}
		if !reflect.DeepEqual(source, before) || g.Check() != nil {
			t.Fatal(i, "ownership or invariant")
		}
	}
}

func TestAigFramesAccessErrorsLimitsCleanupAndWide(t *testing.T) {
	source := aigFramesFixture(t, 0)
	expansion, err := source.ExpandFrames(AigFramesOptions{Frames: 2, Outputs: true, Registers: true})
	if err != nil {
		t.Fatal(err)
	}
	if expansion.Frames() != 2 || expansion.Graph() == source {
		t.Fatal("ownership")
	}
	copyMap := expansion.Mapping()
	copyMap[0] = InvalidAigLit
	if again := expansion.Mapping(); again[0] != AigConst1 {
		t.Fatal("mapping aliases")
	}
	for _, query := range []struct {
		id    AigID
		frame int
	}{{0, -1}, {0, 2}, {AigID(source.ObjectCount()), 0}} {
		if got, err := expansion.Object(query.id, query.frame); got != InvalidAigLit || !errors.Is(err, ErrDimension) {
			t.Fatal(query, got, err)
		}
	}
	// The intentionally dangling fourth source AND is removed in every frame.
	if got, err := expansion.Object(8, 0); got != InvalidAigLit || !errors.Is(err, ErrInvalidLiteral) {
		t.Fatal(got, err)
	}
	for _, options := range []AigFramesOptions{
		{}, {Frames: -1}, {Frames: 1, MaxObjects: -1}, {Frames: 1, MaxMappingEntries: -1},
		{Frames: 2, MaxObjects: 1}, {Frames: 2, MaxMappingEntries: 1},
		{Frames: 2, Stop: func() bool { return true }},
	} {
		if got, err := source.ExpandFrames(options); got != nil || !(errors.Is(err, ErrDimension) || errors.Is(err, ErrLimit) || errors.Is(err, ErrCancelled)) {
			t.Fatal(options, got, err)
		}
	}
	for threshold := 2; threshold < 20; threshold++ {
		calls := 0
		got, err := source.ExpandFrames(AigFramesOptions{Frames: 3, Outputs: true, Registers: true, Stop: func() bool {
			calls++
			return calls == threshold
		}})
		if got != nil || !errors.Is(err, ErrCancelled) {
			t.Fatal(threshold, calls, got, err)
		}
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	if got, err := buffered.ExpandFrames(AigFramesOptions{Frames: 1, Outputs: true}); got != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatal(got, err)
	}
	if got, err := (*AigManager)(nil).ExpandFrames(AigFramesOptions{Frames: 1}); got != nil || !errors.Is(err, ErrDimension) {
		t.Fatal(got, err)
	}

	const width = 10000
	wide := NewAigManager()
	root, _ := wide.CreateCI()
	for i := 0; i < width; i++ {
		other, _ := wide.CreateCI()
		root, _ = wide.And(root, other)
	}
	wide.CreateCO(root)
	result, err := wide.ExpandFrames(AigFramesOptions{Frames: 1, Outputs: true, MaxObjects: 30000, MaxMappingEntries: 30000})
	if err != nil || result.Graph().AndCount() != width || result.Graph().Check() != nil {
		t.Fatal(result, err)
	}
}

func FuzzAigFrames(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewAigManager()
		inputs := make([]AigLit, 4)
		for i := range inputs {
			inputs[i], _ = g.CreateCI()
		}
		signals := slices.Clone(inputs)
		for i := 0; i+1 < len(data) && len(signals) < 68; i += 2 {
			left := signals[int(data[i]&0x3f)%len(signals)].NotIf(data[i]&0x80 != 0)
			right := signals[int(data[i+1]&0x3f)%len(signals)].NotIf(data[i+1]&0x80 != 0)
			if literal, err := g.And(left, right); err == nil {
				signals = append(signals, literal)
			}
		}
		for i := 0; i < 3; i++ {
			literal := signals[(i*17+len(data))%len(signals)].NotIf(i < len(data) && data[i]&1 != 0)
			g.CreateCO(literal)
		}
		expansion, err := g.ExpandFrames(AigFramesOptions{Frames: 3, Outputs: true})
		if err != nil {
			t.Fatal(err)
		}
		assignment := make([]bool, 12)
		for i := range assignment {
			if len(data) != 0 {
				assignment[i] = data[i%len(data)]>>uint(i&7)&1 != 0
			}
		}
		got, err := expansion.Graph().Evaluate(assignment)
		if err != nil || len(got) != 9 {
			t.Fatal(len(got), err)
		}
		for frame := 0; frame < 3; frame++ {
			want, err := g.Evaluate(assignment[frame*4 : (frame+1)*4])
			if err != nil || !slices.Equal(got[frame*3:(frame+1)*3], want) {
				t.Fatal(frame, got, want, err)
			}
		}
	})
}
