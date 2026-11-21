package ioa

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDenseAigerValidationCoversPropertiesAndFallbacks(t *testing.T) {
	model := func() *Model {
		return &Model{
			MaxVar: 3, Inputs: []uint32{2}, Latches: []Latch{{Literal: 4, Next: 6, Init: 4}},
			ANDs: []And{{Literal: 6, Left: 2, Right: 5}}, Outputs: []uint32{7},
			Bad: []uint32{6}, Constraints: []uint32{4}, Fairness: []uint32{3}, Justice: [][]uint32{{2, 5}},
			Symbols: map[Symbol]string{{Kind: 'i', Index: 0}: "in"},
		}
	}
	for _, tc := range []struct {
		name  string
		edit  func(*Model)
		dense bool
		valid bool
	}{
		{"ordered", func(*Model) {}, true, true},
		{"huge sparse range", func(m *Model) { m.MaxVar = 0x7fffffff }, false, true},
		{"duplicate definition", func(m *Model) { m.Latches[0].Literal = 2 }, false, false},
		{"self dependency", func(m *Model) { m.ANDs[0].Right = 7 }, false, false},
		{"latch next", func(m *Model) { m.Latches[0].Next = 8 }, true, false},
		{"latch reset", func(m *Model) { m.Latches[0].Init = 2 }, true, false},
		{"output", func(m *Model) { m.Outputs[0] = 8 }, true, false},
		{"bad", func(m *Model) { m.Bad[0] = 9 }, true, false},
		{"constraint", func(m *Model) { m.Constraints[0] = 8 }, true, false},
		{"justice", func(m *Model) { m.Justice[0][1] = 9 }, true, false},
		{"fairness", func(m *Model) { m.Fairness[0] = 8 }, true, false},
		{"symbol", func(m *Model) { m.Symbols[Symbol{Kind: 'i', Index: 1}] = "missing" }, true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m := model()
			tc.edit(m)
			if m.HasDenseTopologicalOrder() != tc.dense || (m.Check() == nil) != tc.valid {
				t.Fatalf("dense %t, validation %v", m.HasDenseTopologicalOrder(), m.Check())
			}
		})
	}
}

func BenchmarkI10ModelCheck(b *testing.B) {
	data, err := os.ReadFile("../../../i10.aig")
	if err != nil {
		b.Fatal(err)
	}
	m, err := Read(data)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := m.Check(); err != nil {
			b.Fatal(err)
		}
	}
}

func TestAigerIntegerCodec(t *testing.T) {
	for _, v := range []uint32{0, 1, 127, 128, 16383, 16384, 0x7fffffff, 0xffffffff} {
		var encoded []byte
		Ioa_WriteAigerEncodeStr(&encoded, v)
		tail := append(append([]byte(nil), encoded...), 42)
		got, err := Ioa_ReadAigerDecode(&tail)
		if err != nil || got != v || !bytes.Equal(tail, []byte{42}) {
			t.Fatal(v, got, err, tail)
		}
		buffer := make([]byte, 12)
		n := Ioa_WriteAigerEncode(buffer, 3, v)
		if !bytes.Equal(buffer[3:n], encoded) {
			t.Fatal(v)
		}
	}
	for _, bad := range [][]byte{{}, {128}, {255, 255, 255, 255, 16}, {128, 128, 128, 128, 128}} {
		tail := append([]byte(nil), bad...)
		before := append([]byte(nil), tail...)
		if _, err := Ioa_ReadAigerDecode(&tail); err == nil || !bytes.Equal(before, tail) {
			t.Fatal("malformed integer accepted or consumed", bad)
		}
	}
}
func TestCompactLiterals(t *testing.T) {
	for _, literals := range [][]uint32{nil, {0}, {1, 2, 1, 127, 0, 1024, 1023}} {
		encoded, err := Ioa_WriteEncodeLiterals(literals)
		if err != nil {
			t.Fatal(err)
		}
		tail := encoded
		decoded, err := Ioa_WriteDecodeLiterals(&tail, int32(len(literals)))
		if err != nil || len(tail) != 0 || len(decoded) != len(literals) {
			t.Fatal(decoded, err)
		}
		for i, v := range literals {
			if decoded[i] != v {
				t.Fatal(i)
			}
		}
	}
	if _, err := Ioa_WriteEncodeLiterals([]uint32{0, 0xffffffff}); err == nil {
		t.Fatal("compact overflow accepted")
	}
	tail := []byte{0, 3}
	if _, err := Ioa_WriteDecodeLiterals(&tail, 2); err == nil {
		t.Fatal("negative literal accepted")
	}
}
func TestAigerReferenceFixtures(t *testing.T) {
	fixtures := [][]byte{
		[]byte("aag 0 0 0 1 0\n0\n"), []byte("aag 0 0 0 1 0\n1\n"),
		[]byte("aag 3 2 0 1 1\n2\n4\n6\n6 2 4\n"),
		append([]byte("aig 3 2 0 1 1\n6\n"), 2, 2),
		[]byte("aag 5 1 1 0 3 1 1\n2\n4 10 0\n4\n3\n6 5 3\n8 4 2\n10 9 7\n"),
	}
	for _, fixture := range fixtures {
		m, err := Read(fixture)
		if err != nil {
			t.Fatalf("%q: %v", fixture, err)
		}
		for _, options := range []WriteOptions{{}, {ASCII: true}, {Compact: true}} {
			encoded, err := Marshal(m, options)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := Read(encoded)
			if err != nil {
				t.Fatalf("round trip %q: %v", encoded, err)
			}
			a, _ := m.Normalize()
			b, _ := decoded.Normalize()
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("round trip differs:\n%+v\n%+v", a, b)
			}
		}
	}
}
func TestAigerExtendedRoundTrip(t *testing.T) {
	m := &Model{MaxVar: 5, Inputs: []uint32{2}, Latches: []Latch{{4, 10, 4}}, Outputs: []uint32{11, 0, 1}, ANDs: []And{{6, 5, 3}, {8, 4, 2}, {10, 9, 7}}, Bad: []uint32{4}, Constraints: []uint32{3}, Justice: [][]uint32{{2, 4}, {10}}, Fairness: []uint32{5}, Symbols: map[Symbol]string{{'i', 0}: "enable", {'l', 0}: "counter", {'o', 1}: "false", {'b', 0}: "reach", {'c', 0}: "hold", {'j', 1}: "live", {'f', 0}: "fair"}, Name: "counter", Comment: []byte("\nsource fixture\n")}
	for _, options := range []WriteOptions{{Symbols: true}, {ASCII: true, Symbols: true}} {
		encoded, err := Marshal(m, options)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := Read(encoded)
		if err != nil {
			t.Fatal(err)
		}
		a, _ := m.Normalize()
		b, _ := decoded.Normalize()
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("extended roundtrip\n%+v\n%+v", a, b)
		}
	}
	if _, err := Marshal(m, WriteOptions{Compact: true}); err == nil {
		t.Fatal("compact silently dropped extended metadata")
	}
}
func TestAigerSparseUnorderedNormalization(t *testing.T) {
	m := &Model{MaxVar: 20, Inputs: []uint32{8, 16}, Outputs: []uint32{40}, ANDs: []And{{40, 20, 16}, {20, 8, 17}}}
	normalized, err := m.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	if normalized.MaxVar != 4 || normalized.ANDs[0].Literal != 6 || normalized.ANDs[1].Literal != 8 {
		t.Fatal(normalized)
	}
	if m.MaxVar != 20 || m.ANDs[0].Literal != 40 {
		t.Fatal("normalization mutated input")
	}
	data, err := Marshal(m, WriteOptions{})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalized, decoded) {
		t.Fatalf("%+v != %+v", normalized, decoded)
	}
}
func TestMalformedAiger(t *testing.T) {
	for _, bad := range [][]byte{
		nil, []byte("aig 0 0 0 0 0"), []byte("aig 3 1 0 0 0\n"), []byte("aig 2 0 0 0 0\n"), []byte("aig 1 0 0 1 1\n2\n"),
		[]byte("aig 2147483647 2147483647 0 0 0\n"), []byte("aag 1 1 0 1 0\n2\n4\n"),
		[]byte("aag 1 0 0 1 1\n2\n2 2 0\n"), []byte("aag 1 1 0 0 1\n2\n2 0 1\n"),
		[]byte("aag 1 0 1 0 0\n2 0 3\n"), []byte("aag 1 1 0 1 0\n2\n2\ni1 bad\n"),
		[]byte("aag 0 0 0 1 0\n0\nx0 nope\n"), []byte("aag 0 0 0 0 0\ncnmissing terminator"),
		append([]byte("aig 1 0 0 1 1\n2\n"), 3, 0),
	} {
		if _, err := Read(bad); err == nil {
			t.Fatalf("accepted malformed %q", bad)
		}
	}
}

type shortAigerWriter struct{}

func (shortAigerWriter) Write(data []byte) (int, error) { return 0, nil }
func TestAigerFileAndMemoryWrappers(t *testing.T) {
	m := &Model{Outputs: []uint32{1}}
	dir := t.TempDir()
	path := filepath.Join(dir, "truth.aig")
	if err := Ioa_WriteAiger(m, path, 0, 1); err != nil {
		t.Fatal(err)
	}
	decoded, err := Ioa_ReadAiger(path, 1)
	if err != nil || decoded.Outputs[0] != 1 {
		t.Fatal(decoded, err)
	}
	var size int32
	data, err := Ioa_WriteAigerIntoMemory(m, &size)
	if err != nil || size != int32(len(data)) {
		t.Fatal(size, err)
	}
	if _, err := Ioa_ReadAigerFromMemory(data, size, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := Ioa_ReadAigerFromMemory(data, size+1, 1); err == nil {
		t.Fatal("invalid length accepted")
	}
	if !errors.Is(Write(shortAigerWriter{}, m, WriteOptions{}), io.ErrShortWrite) {
		t.Fatal("short write not reported")
	}
	if Ioa_FileSize(path) == 0 {
		t.Fatal("file size")
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
func FuzzReadAiger(f *testing.F) {
	for _, seed := range [][]byte{[]byte("aig 0 0 0 1 0\n1\n"), append([]byte("aig 3 2 0 1 1\n6\n"), 2, 2), []byte("aag 1 0 1 1 0\n2 3 2\n2\n")} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}
		m, err := Read(data)
		if err != nil {
			return
		}
		encoded, err := Marshal(m, WriteOptions{Symbols: true})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := Read(encoded); err != nil {
			t.Fatal(err)
		}
	})
}

func evaluateModelOutput(m *Model, assignment int) []bool {
	values := map[uint32]bool{0: false}
	for i, lit := range m.Inputs {
		values[lit>>1] = assignment&(1<<uint(i)) != 0
	}
	gates := map[uint32]And{}
	for _, a := range m.ANDs {
		gates[a.Literal>>1] = a
	}
	var eval func(uint32) bool
	eval = func(lit uint32) bool {
		value, ok := values[lit>>1]
		if !ok {
			a := gates[lit>>1]
			value = eval(a.Left) && eval(a.Right)
			values[lit>>1] = value
		}
		return value != (lit&1 != 0)
	}
	result := make([]bool, len(m.Outputs))
	for i, lit := range m.Outputs {
		result[i] = eval(lit)
	}
	return result
}
func TestAigerABCReference(t *testing.T) {
	binary, err := filepath.Abs("../../../abc")
	if err != nil {
		t.Fatal(err)
	}
	if info, err := os.Stat(binary); err != nil || info.Mode()&0111 == 0 {
		t.Skip("ABC reference binary unavailable")
	}
	m := &Model{MaxVar: 3, Inputs: []uint32{2, 4}, Outputs: []uint32{6, 0, 1}, ANDs: []And{{6, 2, 4}}}
	for _, compact := range []int32{0, 1} {
		dir := t.TempDir()
		input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
		if err := Ioa_WriteAiger(m, input, 0, compact); err != nil {
			t.Fatal(err)
		}
		command := fmt.Sprintf("read_aiger %s; write_aiger %s", input, output)
		if out, err := exec.Command(binary, "-c", command).CombinedOutput(); err != nil {
			t.Fatalf("ABC reference: %v\n%s", err, out)
		}
		result, err := Ioa_ReadAiger(output, 1)
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			if !reflect.DeepEqual(evaluateModelOutput(m, assignment), evaluateModelOutput(result, assignment)) {
				t.Fatal("ABC roundtrip changed truth", compact, assignment)
			}
		}
	}
}
func TestAigerCommentPayloadRoundTrip(t *testing.T) {
	for _, name := range []string{"", "example"} {
		for _, comment := range [][]byte{[]byte("plain comment"), []byte("\nblank first line\n"), []byte{0, 1, 2}} {
			m := &Model{Name: name, Comment: comment}
			encoded, err := Marshal(m, WriteOptions{})
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := Read(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if decoded.Name != name || !bytes.Equal(decoded.Comment, comment) {
				t.Fatalf("name=%q payload=%q -> %q", name, comment, decoded.Comment)
			}
		}
	}
}

// The native fixture appends a single comment marker after this graph body.
func i10WriterModel(t testing.TB) *Model {
	t.Helper()
	data, err := os.ReadFile("../../../i10.aig")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Read(data)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestI10BinaryBodyMatchesNativeMemoryWriter(t *testing.T) {
	m := i10WriterModel(t)
	m.Name, m.Comment = "", nil
	data, err := Ioa_WriteAigerIntoMemoryStr(m)
	if err != nil {
		t.Fatal(err)
	}
	const native = "c0fe91cbd03dba79ffaa2aa442264965cbe47e9d94d996ec5db0009231e3c441"
	if got := fmt.Sprintf("%x", sha256.Sum256(data)); len(data) != 8720 || got != native {
		t.Fatalf("binary body has %d bytes and SHA-256 %s", len(data), got)
	}
}

func BenchmarkI10BinaryMarshal(b *testing.B) {
	m := i10WriterModel(b)
	b.ReportAllocs()
	for b.Loop() {
		encoded, err := Marshal(m, WriteOptions{Symbols: true})
		if err != nil || len(encoded) == 0 {
			b.Fatal("invalid encoding", err)
		}
	}
}

func BenchmarkI10MemoryWriter(b *testing.B) {
	m := i10WriterModel(b)
	m.Name, m.Comment = "", nil
	b.ReportAllocs()
	for b.Loop() {
		encoded, err := Ioa_WriteAigerIntoMemoryStr(m)
		if err != nil || len(encoded) != 8720 {
			b.Fatal("invalid binary graph body", len(encoded), err)
		}
	}
}

func TestDenseNormalizeOwnsAllFieldsAndOrdersFaninPairs(t *testing.T) {
	m := &Model{
		MaxVar: 3, Inputs: []uint32{2}, Latches: []Latch{{4, 7, 4}}, ANDs: []And{{6, 2, 5}},
		Outputs: []uint32{7}, Bad: []uint32{6}, Constraints: []uint32{5}, Justice: [][]uint32{{}, {0, 1, 2}},
		Fairness: []uint32{3}, Symbols: map[Symbol]string{{'i', 0}: "enable"}, Comment: []byte("payload"),
	}
	copy, err := m.Normalize()
	if err != nil {
		t.Fatal(err)
	}
	if copy.ANDs[0] != (And{6, 5, 2}) || copy.Justice[0] != nil {
		t.Fatal("normalized representation", copy.ANDs, copy.Justice)
	}
	encoded, err := Marshal(m, WriteOptions{Symbols: true})
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := Marshal(copy, WriteOptions{Symbols: true})
	if err != nil || !bytes.Equal(encoded, canonical) {
		t.Fatal("binary encoding differs with reversed fanins", err)
	}
	copy.Inputs[0], copy.Latches[0].Init, copy.ANDs[0].Left = 0, 0, 0
	copy.Outputs[0], copy.Bad[0], copy.Constraints[0], copy.Fairness[0] = 0, 0, 0, 0
	copy.Justice[1][2], copy.Justice[0], copy.Comment[0] = 0, []uint32{1}, 'x'
	delete(copy.Symbols, Symbol{'i', 0})
	if m.Inputs[0] != 2 || m.Latches[0].Init != 4 || m.ANDs[0] != (And{6, 2, 5}) ||
		m.Outputs[0] != 7 || m.Bad[0] != 6 || m.Constraints[0] != 5 || m.Fairness[0] != 3 ||
		m.Justice[1][2] != 2 || len(m.Justice[0]) != 0 || string(m.Comment) != "payload" || m.Symbols[Symbol{'i', 0}] != "enable" {
		t.Fatal("normalized copy aliases its source")
	}
}

func TestDenseAndSparseSymbolTablesKeepAigerOrder(t *testing.T) {
	for _, slots := range []int{1, 32} {
		m := &Model{
			MaxVar: uint32(slots + 1), Latches: []Latch{{Literal: uint32(2*slots + 2)}},
			Outputs: []uint32{0}, Bad: []uint32{0}, Constraints: []uint32{0},
			Justice: [][]uint32{{1}}, Fairness: []uint32{1}, Symbols: make(map[Symbol]string),
		}
		for i := range slots {
			m.Inputs = append(m.Inputs, uint32(2*i+2))
		}
		for _, k := range []byte("fjcboli") {
			m.Symbols[Symbol{k, 0}] = string(k)
		}
		want := "i0 i\nl0 l\no0 o\nb0 b\nc0 c\nj0 j\nf0 f\n"
		for _, ascii := range []bool{false, true} {
			got, err := Marshal(m, WriteOptions{ASCII: ascii, Symbols: true})
			if err != nil || !bytes.HasSuffix(got, []byte(want)) {
				t.Fatalf("slots %d ASCII %v: data %q, error %v", slots, ascii, got, err)
			}
		}
	}
}
