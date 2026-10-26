package abc

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestAigerConstantsAndLatchInit(t *testing.T) {
	for _, test := range []struct {
		data    string
		initial []LatchInit
		ci      []bool
		want    []bool
	}{
		{"aig 0 0 0 0 0\n", nil, nil, []bool{}},
		{"aig 0 0 0 2 0\n0\n1\n", nil, nil, []bool{false, true}},
		{"aig 2 1 1 1 0\n2\n4\n", []LatchInit{LatchInitZero}, []bool{true, false}, []bool{false, true}},
		{"aig 2 1 1 1 0\n2 1\n4\n", []LatchInit{LatchInitOne}, []bool{false, true}, []bool{true, false}},
		{"aig 2 1 1 1 0\n2 4\n4\n", []LatchInit{LatchInitUnknown}, []bool{false, true}, []bool{true, false}},
	} {
		model, err := ReadAiger([]byte(test.data), AigerLimits{})
		if err != nil {
			t.Fatal(test.data, err)
		}
		if !slices.Equal(model.InitialValues(), test.initial) {
			t.Fatal("initial", model.InitialValues())
		}
		got, err := model.Graph().Evaluate(test.ci)
		if err != nil || !slices.Equal(got, test.want) {
			t.Fatal(got, test.want, err)
		}
		encoded, err := model.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		again, err := ReadAiger(encoded, AigerLimits{})
		if err != nil {
			t.Fatal(err)
		}
		got, err = again.Graph().Evaluate(test.ci)
		if err != nil || !slices.Equal(got, test.want) || !slices.Equal(again.InitialValues(), test.initial) {
			t.Fatal("round trip", got, err)
		}
	}
}
func TestAigerAndAndSymbols(t *testing.T) {
	// AND variable 3 = input 1 AND NOT input 2: 6-5=1, 5-2=3.
	data := append([]byte("aig 3 2 0 1 1\n6\n"), 1, 3)
	data = append(data, []byte("i1 second input\ni0 first input\no0 conjunction\nc\nraw\x00comment\n")...)
	model, err := ReadAiger(data, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	wantNames := []AigerSymbolName{{AigerSymbol{'i', 1}, "second input"}, {AigerSymbol{'i', 0}, "first input"}, {AigerSymbol{'o', 0}, "conjunction"}}
	if !slices.Equal(model.Symbols(), wantNames) {
		t.Fatal(model.Symbols())
	}
	for assignment := 0; assignment < 4; assignment++ {
		got, err := model.Graph().Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
		if err != nil || got[0] != (assignment&1 != 0 && assignment&2 == 0) {
			t.Fatal(assignment, got, err)
		}
	}
	if err := model.SetSymbol(AigerSymbol{'i', 1}, "renamed"); err != nil {
		t.Fatal(err)
	}
	wantNames[0].Name = "renamed"
	if !slices.Equal(model.Symbols(), wantNames) {
		t.Fatal("update changed order")
	}
	if err := model.SetSymbol(AigerSymbol{'i', 2}, "invalid"); err == nil {
		t.Fatal("symbol bounds")
	}
	names := model.Symbols()
	names[0].Name = "aliased"
	if model.Symbols()[0].Name != "renamed" {
		t.Fatal("symbol alias")
	}
	comment := model.Comment()
	comment[0] = 'X'
	if model.Comment()[0] != 'r' {
		t.Fatal("comment alias")
	}
	encoded, err := model.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadAiger(encoded, AigerLimits{})
	if err != nil || !slices.Equal(again.Symbols(), wantNames) || !bytes.Equal(again.Comment(), model.Comment()) {
		t.Fatal("symbol round trip", err)
	}
}
func TestAigerExtendedSafetySections(t *testing.T) {
	data := []byte("aig 1 1 0 1 0 1 1 0 0\n2\n3\n2\nc0 assume positive\nb0 bad negative\no0 output\n")
	m, err := ReadAiger(data, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if m.OrdinaryOutputCount() != 1 || m.BadCount() != 1 || m.ConstraintCount() != 1 {
		t.Fatal("section counts")
	}
	got, _ := m.Graph().Evaluate([]bool{false})
	if !slices.Equal(got, []bool{false, true, false}) {
		t.Fatal("constraint polarity", got)
	}
	got, _ = m.Graph().Evaluate([]bool{true})
	if !slices.Equal(got, []bool{true, false, true}) {
		t.Fatal("constraint polarity", got)
	}
	encoded, err := m.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadAiger(encoded, AigerLimits{})
	if err != nil || again.BadCount() != 1 || again.ConstraintCount() != 1 || !slices.Equal(again.Symbols(), m.Symbols()) {
		t.Fatal("extended round trip", err)
	}
}
func TestAigerCompactDrivers(t *testing.T) {
	// Driver array 2,4,1 encodes first 2, then +2 (4), then -3 (7).
	data := append([]byte("aig2 2 1 1 2 0\n"), 2, 4, 7)
	m, err := ReadAiger(data, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := m.Graph().Evaluate([]bool{true, false})
	if !slices.Equal(got, []bool{false, true, true}) {
		t.Fatal(got)
	}
	empty, err := ReadAiger([]byte("aig2 0 0 0 0 0\n"), AigerLimits{})
	if err != nil || empty.Validate() != nil {
		t.Fatal("empty compact", err)
	}
	for _, bad := range [][]byte{
		append([]byte("aig2 1 1 0 1 0\n"), 4),
		append([]byte("aig2 1 1 0 2 0\n"), 0, 3),
		append([]byte("aig2 1 1 0 2 0\n"), 3, 2),
	} {
		if _, err := ReadAiger(bad, AigerLimits{}); err == nil {
			t.Fatalf("accepted bad compact %q", bad)
		}
	}
}
func TestAigerMalformedAndLimits(t *testing.T) {
	cases := [][]byte{
		{}, []byte("aig 0 0 0 0 0"), []byte("aig 0 0 0 0\n"), []byte("aig 0 0 0 0 0 0 0 0 0 0\n"),
		[]byte("aig -1 0 0 0 0\n"), []byte("aig 1 0 0 0 0\n"), []byte("aig 0  0 0 0 0\n"),
		[]byte("aig 4294967296 0 0 0 0\n"), []byte("aig 0 0 0 1 0\n2\n"),
		[]byte("aig 0 0 0 1 0\n-1\n"), []byte("aig 1 0 1 0 0\n0 3\n"),
		[]byte("aig 0 0 0 1 0\n1 0\n"), []byte("aig 0 0 0 0 0\no0 wrong\n"),
		[]byte("aig 1 1 0 0 0\ni0 a\ni0 duplicate\n"), []byte("aig 1 1 0 0 0\nx0 unknown\n"),
		[]byte("aig 1 1 0 0 0\ni0 \n"), []byte("aig 1 1 0 0 0\ni0 a\x00b\n"),
		[]byte("aig 1 1 0 0 0\ni0 missing-newline"),
		append([]byte("aig 1 0 0 1 1\n2\n"), 0, 0), // self-fanin
		append([]byte("aig 1 0 0 1 1\n2\n"), 3, 0), // underflow
		append([]byte("aig 1 0 0 1 1\n2\n"), 1, 2), // second delta underflow
		append([]byte("aig 1 0 0 1 1\n2\n"), 128, 128, 128, 128, 16, 0),
		append([]byte("aig 1 0 0 1 1\n2\n"), 128, 128, 128, 128, 128, 0),
	}
	for _, data := range cases {
		if _, err := ReadAiger(data, AigerLimits{}); err == nil {
			t.Fatalf("accepted malformed %q", data)
		}
	}
	for _, data := range []string{"aag 0 0 0 0 0\n", "aig 0 0 0 0 0 0 0 1\n", "aig 0 0 0 0 0 0 0 0 1\n"} {
		if _, err := ReadAiger([]byte(data), AigerLimits{}); !errors.Is(err, ErrUnsupported) {
			t.Fatal(data, err)
		}
	}
	for _, limits := range []AigerLimits{{MaxBytes: 1}, {MaxVariables: 1}, {MaxOutputs: 1}, {MaxBytes: -1}} {
		if _, err := ReadAiger([]byte("aig 2 2 0 2 0\n2\n4\n"), limits); err == nil {
			t.Fatal("limit ignored", limits)
		}
	}
	// Many spaces must not create an unbounded array of header tokens.
	if _, err := ReadAiger(append(bytes.Repeat([]byte{' '}, 1<<20), '\n'), AigerLimits{}); err == nil {
		t.Fatal("space header")
	}
}
func TestAigerEveryTruncationAndEncoding(t *testing.T) {
	data := append([]byte("aig 3 1 1 1 1\n6\n6\n"), 2, 2)
	for end := 0; end < len(data); end++ {
		if _, err := ReadAiger(data[:end], AigerLimits{}); err == nil {
			t.Fatal("accepted truncation", end)
		}
	}
	model, err := ReadAiger(data, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	graph := model.Graph()
	buffer, err := graph.CreateBuffer(AigConst1)
	if err != nil {
		t.Fatal(err)
	}
	graph.CreateBuffer(buffer) // dangling buffers must bypass correctly in encoding
	dead := aigTestAnd(t, graph, AigLit(graph.cis[0])<<1, (AigLit(graph.cis[1]) << 1).Not())
	if dead == InvalidAigLit {
		t.Fatal("construction")
	}
	graph.Cleanup()
	encoded, err := model.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	again, err := ReadAiger(encoded, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	for a := 0; a < 4; a++ {
		in := []bool{a&1 != 0, a&2 != 0}
		want, _ := graph.Evaluate(in)
		got, _ := again.Graph().Evaluate(in)
		if !slices.Equal(got, want) {
			t.Fatal(got, want)
		}
	}
	init := model.InitialValues()
	init[0] = LatchInitOne
	if model.InitialValues()[0] != LatchInitZero {
		t.Fatal("init alias")
	}
}
func TestAigerNativeBenchmark(t *testing.T) {
	data, err := os.ReadFile("test_vectors/aiger/i10.aig")
	if err != nil {
		t.Fatal(err)
	}
	m, err := ReadAiger(data, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	if m.Graph().InputCount() != 257 || m.Graph().PrimaryOutputCount() != 224 || m.Graph().AndCount() != 2675 {
		t.Fatal("i10 dimensions", m.Graph().InputCount(), m.Graph().PrimaryOutputCount(), m.Graph().AndCount())
	}
	encoded, err := m.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	round, err := ReadAiger(encoded, AigerLimits{})
	if err != nil {
		t.Fatal(err)
	}
	input := make([]bool, 257)
	for pattern := 0; pattern < 100; pattern++ {
		for i := range input {
			input[i] = (pattern*(i+1)+i*i)%7 < 3
		}
		a, _ := m.Graph().Evaluate(input)
		b, _ := round.Graph().Evaluate(input)
		if !slices.Equal(a, b) {
			t.Fatal("i10 round trip")
		}
	}
}
func FuzzAigerRead(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte("aig 0 0 0 0 0\n"), []byte("aig 2 1 1 1 0\n2\n4\n"),
		append([]byte("aig 3 2 0 1 1\n6\n"), 1, 3),
		[]byte("aig 1 1 0 0 0 1 1\n2\n3\nc0 constraint\nb0 bad\nc\ncomment"),
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		limits := AigerLimits{MaxBytes: 65536, MaxVariables: 128, MaxOutputs: 128}
		m, err := ReadAiger(data, limits)
		if err != nil {
			return
		}
		if err := m.Validate(); err != nil {
			t.Fatal(err)
		}
		encoded, err := m.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		again, err := ReadAiger(encoded, AigerLimits{MaxBytes: 1 << 20, MaxVariables: 128, MaxOutputs: 128})
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(m.InitialValues(), again.InitialValues()) || !slices.Equal(m.Symbols(), again.Symbols()) || !bytes.Equal(m.Comment(), again.Comment()) || m.BadCount() != again.BadCount() || m.ConstraintCount() != again.ConstraintCount() {
			t.Fatal("metadata round trip")
		}
		input := make([]bool, m.Graph().InputCount())
		for assignment := 0; assignment < 16; assignment++ {
			for i := range input {
				input[i] = assignment&(1<<uint(i%4)) != 0
			}
			a, _ := m.Graph().Evaluate(input)
			b, _ := again.Graph().Evaluate(input)
			if !slices.Equal(a, b) {
				t.Fatal("function round trip")
			}
		}
	})
}
