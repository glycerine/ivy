package bbl

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"testing"
)

func TestBinaryRoundTrip(t *testing.T) {
	p := Bbl_ManSimpleDemo()
	if err := p.Check(); err != nil {
		t.Fatal(err)
	}
	var wire bytes.Buffer
	if err := p.WriteBinary(&wire); err != nil {
		t.Fatal(err)
	}
	q, err := ParseBinary(wire.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if q.Name != "hadder" || len(q.Objects) != 7 {
		t.Fatal(q)
	}
	for i, o := range q.Objects {
		if o.ID != i || o.OriginalID != p.Objects[i].ID || o.SOP != p.Objects[i].SOP {
			t.Fatal(i, o)
		}
		for j, f := range o.Fanins {
			if f.OriginalID != p.Objects[i].Fanins[j].ID {
				t.Fatal("wrong edge")
			}
		}
	}
	var again bytes.Buffer
	if err := q.WriteBinary(&again); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wire.Bytes(), again.Bytes()) {
		t.Fatal("wire round trip differs")
	}
	path := filepath.Join(t.TempDir(), "net.bblif")
	if err := Bbl_ManDumpBinaryBlif(p, path); err != nil {
		t.Fatal(err)
	}
	if _, err := Bbl_ManReadBinaryBlif(path); err != nil {
		t.Fatal(err)
	}
	var text bytes.Buffer
	if err := q.WriteBLIF(&text); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{".model hadder", ".inputs 0", ".outputs 3", ".names 5 3\n1 1", ".end\n"} {
		if !strings.Contains(text.String(), want) {
			t.Fatal(text.String())
		}
	}
}

func TestSparseIdsForwardEdgesAndSopDedup(t *testing.T) {
	p := Bbl_ManStart("sparse")
	if err := Bbl_ManCreateObject(p, BBL_OBJ_NODE, 100, 1, "1 1\n"); err != nil {
		t.Fatal(err)
	}
	if err := Bbl_ManCreateObject(p, BBL_OBJ_CI, 999, 0, ""); err != nil {
		t.Fatal(err)
	}
	if err := Bbl_ManAddFanin(p, 100, 999); err != nil {
		t.Fatal(err)
	}
	if err := Bbl_ManCreateObject(p, BBL_OBJ_NODE, 101, 1, "1 1\n"); err != nil {
		t.Fatal(err)
	}
	if err := Bbl_ManAddFanin(p, 101, 999); err != nil {
		t.Fatal(err)
	}
	if p.Objects[0].function != p.Objects[2].function || len(p.functions) != 1 {
		t.Fatal("SOP was not shared")
	}
	Bbl_ObjMark(p.Objects[0])
	var b bytes.Buffer
	p.WriteBinary(&b)
	q, err := ParseBinary(b.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if !q.Objects[0].Marked || q.Objects[0].Fanins[0] != q.Objects[1] || q.Objects[1].OriginalID != 999 {
		t.Fatal(q.Objects)
	}
	if Bbl_ManObjFirst(q) != q.Objects[0] || Bbl_ManObjNext(q, q.Objects[2]) != nil || Bbl_ObjFaninFirst(q.Objects[0]) != q.Objects[1] || Bbl_ObjFaninNext(q.Objects[0], q.Objects[1]) != nil {
		t.Fatal("iterator mismatch")
	}
}

func TestTruthSopRoundTrip(t *testing.T) {
	for n := 0; n <= 4; n++ {
		for pattern := 0; pattern < 1<<(1<<n) && pattern < 65536; pattern++ {
			input := []uint32{uint32(pattern)}
			s := Bbl_ManTruthToSop(input, n)
			got := Bbl_ManSopToTruth(s, n)
			if len(got) != 1 {
				t.Fatal(n, pattern, s)
			}
			for bit := 0; bit < 32; bit++ {
				want := uint32(pattern) >> uint(bit%(1<<n)) & 1
				if got[0]>>uint(bit)&1 != want {
					t.Fatalf("n=%d pattern=%x sop=%q got=%x", n, pattern, s, got)
				}
			}
		}
	}
	random := rand.New(rand.NewSource(42))
	for _, n := range []int{5, 6, 8, 12, 16} {
		words := (1 << n) / 32
		truth := make([]uint32, words)
		for i := range truth {
			truth[i] = random.Uint32()
		}
		s := Bbl_ManTruthToSop(truth, n)
		if !equalTruth(truth, Bbl_ManSopToTruth(s, n)) {
			t.Fatal("wide truth mismatch", n)
		}
	}
	if got := Bbl_ManSopToTruth("00 0\n", 2); len(got) != 1 || got[0] != 0xeeeeeeee {
		t.Fatal(got)
	}
	for _, s := range []string{"", "a 1\n", "1 2\n", "1 1\n0 0\n"} {
		if Bbl_ManSopToTruth(s, 1) != nil {
			t.Fatal("invalid SOP accepted", s)
		}
	}
}

func TestInvalidNetworksAndBinary(t *testing.T) {
	p := Bbl_ManStart("test")
	if Bbl_ManCreateObject(p, BBL_OBJ_CI, 1, 1, "") == nil {
		t.Fatal("invalid CI")
	}
	Bbl_ManCreateObject(p, BBL_OBJ_NODE, 2, 1, "1 1\n")
	if p.Check() == nil {
		t.Fatal("missing edge accepted")
	}
	if Bbl_ManAddFanin(p, 2, 99) == nil {
		t.Fatal("missing fanin accepted")
	}
	p = Bbl_ManSimpleDemo()
	var b bytes.Buffer
	p.WriteBinary(&b)
	data := b.Bytes()
	for i := 0; i < len(data); i++ {
		if _, err := ParseBinary(data[:i]); err == nil {
			t.Fatalf("truncation %d accepted", i)
		}
	}
	if _, err := ParseBinary(append(append([]byte{}, data...), 0)); err == nil {
		t.Fatal("trailing bytes accepted")
	}
	bad := append([]byte{}, data...)
	binary.LittleEndian.PutUint32(bad, ^uint32(0))
	if _, err := ParseBinary(bad); err == nil {
		t.Fatal("negative pool accepted")
	}
	if err := p.WriteBinary(shortWriter{}); !errors.Is(err, io.ErrShortWrite) {
		t.Fatal(err)
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return 0, nil }

func TestPaddedFunctionRecordsReserialize(t *testing.T) {
	var objects, functions, data bytes.Buffer
	for i, handle := range []int32{0, 12} {
		for _, x := range []int32{int32(i), handle, 0} {
			binary.Write(&objects, binary.LittleEndian, x)
		}
	}
	binary.Write(&functions, binary.LittleEndian, int32(2))
	functions.Write([]byte{' ', '0', '\n', 0, 0, 0, 0, 0})
	binary.Write(&functions, binary.LittleEndian, int32(1))
	functions.Write([]byte{' ', '1', '\n', 0})
	for _, pool := range [][]byte{make([]byte, 16), objects.Bytes(), functions.Bytes()} {
		binary.Write(&data, binary.LittleEndian, int32(len(pool)))
		data.Write(pool)
	}
	p, err := ParseBinary(data.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	var encoded bytes.Buffer
	if err := p.WriteBinary(&encoded); err != nil {
		t.Fatal(err)
	}
	q, err := ParseBinary(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if q.Objects[0].SOP != " 0\n" || q.Objects[1].SOP != " 1\n" {
		t.Fatal(q.Objects)
	}
}
