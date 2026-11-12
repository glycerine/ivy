package ioa

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

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
