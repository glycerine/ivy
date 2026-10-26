package abc

// Standalone round-trip, parser, path, limit, concurrency and fuzz tests for saigIoa.c.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func saigBLIFFixture(t testing.TB) *AigManager {
	t.Helper()
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	state, _ := graph.CreateCI()
	first, err := graph.And(a.Not(), b)
	if err != nil {
		t.Fatal(err)
	}
	property, _ := graph.And(first, state.Not())
	next, _ := graph.Or(a, state.Not())
	graph.CreateCO(property.Not())
	graph.CreateCO(AigConst0)
	graph.CreateCO(next)
	if err = graph.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	graph.Name, graph.Spec = "sequential", "source.aig"
	return graph
}

func saigBLIFCompareSemantics(t testing.TB, left, right *AigManager) {
	t.Helper()
	if left.InputCount() != right.InputCount() || left.OutputCount() != right.OutputCount() || left.RegisterCount() != right.RegisterCount() {
		t.Fatalf("partitions left=%d/%d/%d right=%d/%d/%d", left.InputCount(), left.OutputCount(), left.RegisterCount(), right.InputCount(), right.OutputCount(), right.RegisterCount())
	}
	if left.InputCount() > 16 {
		t.Fatal("semantic fixture too wide")
	}
	for assignment := 0; assignment < 1<<left.InputCount(); assignment++ {
		inputs := make([]bool, left.InputCount())
		for index := range inputs {
			inputs[index] = assignment&(1<<index) != 0
		}
		got, gotErr := right.Evaluate(inputs)
		want, wantErr := left.Evaluate(inputs)
		if gotErr != nil || wantErr != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("assignment=%d got=%v want=%v gotErr=%v wantErr=%v", assignment, got, want, gotErr, wantErr)
		}
	}
}

func TestSAIGBLIFExactRoundTripNamesAndPaths(t *testing.T) {
	source := saigBLIFFixture(t)
	before := source.Clone()
	data, err := MarshalSAIGBLIF(source, SAIGBLIFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, fragment := range []string{
		"# BLIF file written by procedure Saig_ManDumpBlif()\n",
		".model sequential\n.inputs pi0 pi1\n.outputs po0 po1\n",
		".latch li0 lo0 0\n",
		".names n0\n 1\n",
		".end\n",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("missing %q in\n%s", fragment, text)
		}
	}
	parsed, err := ParseSAIGBLIF(data, "memory.blif", SAIGBLIFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Name != source.Name || parsed.Spec != "memory.blif" || parsed.ConstraintCount() != 0 {
		t.Fatalf("metadata=%q/%q constraints=%d", parsed.Name, parsed.Spec, parsed.ConstraintCount())
	}
	saigBLIFCompareSemantics(t, source, parsed)
	again, err := MarshalSAIGBLIF(parsed, SAIGBLIFOptions{})
	if err != nil || !bytes.Equal(data, again) {
		t.Fatalf("unstable round trip err=%v\nfirst:\n%s\nsecond:\n%s", err, data, again)
	}
	if !reflect.DeepEqual(source.objects, before.objects) || !reflect.DeepEqual(source.cis, before.cis) || !reflect.DeepEqual(source.cos, before.cos) {
		t.Fatal("marshal mutated source")
	}

	for _, id := range append(append([]AigID{0}, source.cis...), source.cos...) {
		name, nameErr := SAIGObjectName(source, id)
		if nameErr != nil || name == "" {
			t.Fatalf("id=%d name=%q err=%v", id, name, nameErr)
		}
	}
	if number, numberErr := SAIGBLIFReadNumber("n00042"); numberErr != nil || number != 42 {
		t.Fatalf("number=%d err=%v", number, numberErr)
	}
	if _, numberErr := SAIGBLIFReadNumber("bad"); !errors.Is(numberErr, ErrFormat) {
		t.Fatalf("bad number err=%v", numberErr)
	}

	path := filepath.Join(t.TempDir(), "roundtrip.blif")
	if err = WriteSAIGBLIF(source, path, SAIGBLIFOptions{}); err != nil {
		t.Fatal(err)
	}
	onDisk, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(onDisk, data) {
		t.Fatalf("disk err=%v equal=%v", err, bytes.Equal(onDisk, data))
	}
	fromDisk, err := ReadSAIGBLIF(path, SAIGBLIFOptions{})
	if err != nil || fromDisk.Spec != path {
		t.Fatalf("read err=%v spec=%q", err, fromDisk.Spec)
	}
	saigBLIFCompareSemantics(t, source, fromDisk)
}

func TestSAIGBLIFMalformedLimitsAndCancellation(t *testing.T) {
	source := saigBLIFFixture(t)
	data, err := MarshalSAIGBLIF(source, SAIGBLIFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	malformed := []string{
		"",
		".model x\n.inputs\n.outputs po0\n.end\n",
		".model x\n.inputs pi1\n.outputs po0\n.names pi0 po0\n1 1\n.end\n",
		".model x\n.inputs pi0\n.outputs po0\n.latch li0 lo0 2\n.end\n",
		".model x\n.inputs pi0\n.outputs po0\n.names n9 po0\n1 1\n.end\n",
		".model x\n.inputs pi0\n.outputs po0\n.names pi0 pi0 n2\n00 1\n.names n2 po0\n1 1\n.end\n",
		".model x\n.inputs pi0\n.outputs po0\n.names pi0 pi0 n2\n1- 1\n.names n2 po0\n1 1\n.end\n",
		".model x\n.inputs pi0\n.outputs po0\n.names pi0 po0\n1 1\n",
	}
	for index, input := range malformed {
		if graph, parseErr := ParseSAIGBLIF([]byte(input), "bad", SAIGBLIFOptions{}); graph != nil || parseErr == nil {
			t.Fatalf("case=%d graph=%v err=%v", index, graph, parseErr)
		}
	}
	for _, options := range []SAIGBLIFOptions{
		{MaxInputBytes: -1}, {MaxOutputBytes: -1}, {MaxObjects: -1}, {MaxTokens: -1}, {MaxTokenBytes: -1}, {MaxWork: -1},
	} {
		if output, marshalErr := MarshalSAIGBLIF(source, options); output != nil || !errors.Is(marshalErr, ErrDimension) {
			t.Fatalf("marshal options=%+v output=%v err=%v", options, output, marshalErr)
		}
		if graph, parseErr := ParseSAIGBLIF(data, "x", options); graph != nil || !errors.Is(parseErr, ErrDimension) {
			t.Fatalf("parse options=%+v graph=%v err=%v", options, graph, parseErr)
		}
	}
	for _, options := range []SAIGBLIFOptions{
		{MaxOutputBytes: len(data) - 1}, {MaxObjects: 1}, {MaxWork: 1}, {Stop: func() bool { return true }},
	} {
		if output, marshalErr := MarshalSAIGBLIF(source, options); output != nil || !(errors.Is(marshalErr, ErrLimit) || errors.Is(marshalErr, ErrCancelled)) {
			t.Fatalf("marshal options=%+v output=%v err=%v", options, output, marshalErr)
		}
	}
	for _, options := range []SAIGBLIFOptions{
		{MaxInputBytes: len(data) - 1}, {MaxObjects: 1}, {MaxTokens: 1}, {MaxTokenBytes: 1}, {MaxWork: 1}, {Stop: func() bool { return true }},
	} {
		if graph, parseErr := ParseSAIGBLIF(data, "x", options); graph != nil || !(errors.Is(parseErr, ErrLimit) || errors.Is(parseErr, ErrCancelled)) {
			t.Fatalf("parse options=%+v graph=%v err=%v", options, graph, parseErr)
		}
	}
	if output, marshalErr := MarshalSAIGBLIF(nil, SAIGBLIFOptions{}); output != nil || !errors.Is(marshalErr, ErrDimension) {
		t.Fatalf("nil marshal output=%v err=%v", output, marshalErr)
	}
	empty := NewAigManager()
	empty.Name = "empty"
	if output, marshalErr := MarshalSAIGBLIF(empty, SAIGBLIFOptions{}); output != nil || !errors.Is(marshalErr, ErrDimension) {
		t.Fatalf("empty marshal output=%v err=%v", output, marshalErr)
	}
	unnamed := source.Clone()
	unnamed.Name = "bad name"
	if output, marshalErr := MarshalSAIGBLIF(unnamed, SAIGBLIFOptions{}); output != nil || !errors.Is(marshalErr, ErrDimension) {
		t.Fatalf("unnamed output=%v err=%v", output, marshalErr)
	}
	buffered := NewAigManager()
	input, _ := buffered.CreateCI()
	buffer, _ := buffered.CreateBuffer(input)
	buffered.CreateCO(buffer)
	buffered.Name = "buffer"
	if output, marshalErr := MarshalSAIGBLIF(buffered, SAIGBLIFOptions{}); output != nil || !errors.Is(marshalErr, ErrUnsupported) {
		t.Fatalf("buffer output=%v err=%v", output, marshalErr)
	}
	if err = WriteSAIGBLIF(source, "", SAIGBLIFOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatalf("empty write path err=%v", err)
	}
	if graph, readErr := ReadSAIGBLIF("", SAIGBLIFOptions{}); graph != nil || !errors.Is(readErr, ErrDimension) {
		t.Fatalf("empty read graph=%v err=%v", graph, readErr)
	}
}

func TestSAIGBLIFDepthAndConcurrentFrozenSource(t *testing.T) {
	deep := NewAigManager()
	root, _ := deep.CreateCI()
	for index := 0; index < 10000; index++ {
		input, _ := deep.CreateCI()
		root, _ = deep.And(root.NotIf(index&1 != 0), input.NotIf(index&2 != 0))
	}
	deep.CreateCO(root)
	deep.Name = "deep"
	data, err := MarshalSAIGBLIF(deep, SAIGBLIFOptions{MaxObjects: 50000, MaxOutputBytes: 8 << 20, MaxWork: 1000000})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseSAIGBLIF(data, "deep.blif", SAIGBLIFOptions{MaxInputBytes: 8 << 20, MaxObjects: 50000, MaxTokens: 1000000, MaxWork: 1000000})
	if err != nil || parsed.InputCount() != deep.InputCount() || parsed.AndCount() != deep.AndCount() {
		t.Fatalf("deep parsed=%v err=%v", parsed, err)
	}

	source := saigBLIFFixture(t)
	baseline, err := MarshalSAIGBLIF(source, SAIGBLIFOptions{})
	if err != nil {
		t.Fatal(err)
	}
	var failures atomic.Int32
	var wait sync.WaitGroup
	for worker := 0; worker < 16; worker++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			encoded, marshalErr := MarshalSAIGBLIF(source, SAIGBLIFOptions{})
			decoded, parseErr := ParseSAIGBLIF(encoded, "parallel", SAIGBLIFOptions{})
			if marshalErr != nil || parseErr != nil || !bytes.Equal(encoded, baseline) || decoded.RegisterCount() != 1 {
				failures.Add(1)
			}
		}()
	}
	wait.Wait()
	if failures.Load() != 0 {
		t.Fatalf("concurrent failures=%d", failures.Load())
	}
}

func FuzzSAIGBLIFRoundTripAndParser(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte(".model malformed\n.end\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		parseOptions := SAIGBLIFOptions{MaxInputBytes: 4096, MaxObjects: 128, MaxTokens: 2048, MaxTokenBytes: 256, MaxWork: 8192}
		_, _ = ParseSAIGBLIF(data, "fuzz", parseOptions)
		if len(data) == 0 {
			return
		}
		at := func(index int) byte { return data[index%len(data)] }
		graph := NewAigManager()
		inputs := make([]AigLit, 3)
		for index := range inputs {
			inputs[index], _ = graph.CreateCI()
		}
		pool := append([]AigLit(nil), inputs...)
		for gate := 0; gate < 1+int(at(0)%5); gate++ {
			left := pool[int(at(2*gate+1))%len(pool)].NotIf(at(gate+1)&1 != 0)
			right := pool[int(at(2*gate+2))%len(pool)].NotIf(at(gate+2)&2 != 0)
			root, err := graph.And(left, right)
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, root)
		}
		graph.CreateCO(pool[len(pool)-1].NotIf(at(3)&4 != 0))
		graph.CreateCO(pool[int(at(4))%len(pool)].NotIf(at(5)&8 != 0))
		if err := graph.SetRegisterCount(1); err != nil {
			t.Fatal(err)
		}
		graph.Name = "fuzz"
		encoded, err := MarshalSAIGBLIF(graph, SAIGBLIFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := ParseSAIGBLIF(encoded, "fuzz", SAIGBLIFOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 1<<graph.InputCount(); assignment++ {
			values := make([]bool, graph.InputCount())
			for index := range values {
				values[index] = assignment&(1<<index) != 0
			}
			got, gotErr := decoded.Evaluate(values)
			want, wantErr := graph.Evaluate(values)
			if gotErr != nil || wantErr != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("assignment=%d got=%v want=%v gotErr=%v wantErr=%v\n%s", assignment, got, want, gotErr, wantErr, encoded)
			}
		}
	})
}
