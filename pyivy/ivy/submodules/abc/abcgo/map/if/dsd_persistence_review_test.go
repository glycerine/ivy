package if_

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

// A source-layout AND3 record, including deliberately nonzero C padding.
func dsdReviewSourceFile() []byte {
	data := []byte("dsd1")
	for _, value := range []uint32{3, 2, 3, 3, 2, 3 | 3<<3 | 1<<8 | 7<<9 | 3<<27, 2, 2, 2, 0xdeadbeef, 0, 0, 1, 0, 0, 0} {
		data = binary.LittleEndian.AppendUint32(data, value)
	}
	return data
}
func TestDsdSaveWritesOriginalBinaryFormat(t *testing.T) {
	p := If_DsdManAlloc(3, 2)
	defer If_DsdManFree(p, 0)
	If_DsdManCompute(p, dau.Dau_DsdToTruth("(abc)", 3), 3, make([]byte, 3), "")
	path := filepath.Join(t.TempDir(), "manager.dsd")
	If_DsdManSave(p, path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 4 || string(data[:4]) != DSD_VERSION {
		t.Fatalf("Save wrote%q; want original dsd1 binary header", data)
	}
}
func TestDsdLoadReadsSourceRecordsAndRebuildsInterning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "source.dsd")
	if err := os.WriteFile(path, dsdReviewSourceFile(), 0600); err != nil {
		t.Fatal(err)
	}
	p := If_DsdManLoad(path)
	if p == nil {
		t.Fatal("Load rejected valid original binary AND3 record")
	}
	defer If_DsdManFree(p, 0)
	if p.nVars != 3 || p.LutSize != 2 || If_DsdManObjNum(p) != 3 || p.pStore != path {
		t.Fatal("loaded geometry/path differs from source")
	}
	o := If_DsdVecObj(&p.vObjs, 2)
	if o.Count != 7 || o.fMark != 1 || o.nSupp != 3 || !slices.Equal(o.pFans, []int32{2, 2, 2}) {
		t.Fatalf("loaded node%+v", o)
	}
	truth := []word{0x8080808080808080}
	if !slices.Equal(If_DsdManComputeTruth(p, 4, nil), truth) {
		t.Fatal("loaded function differs")
	}
	if lit := If_DsdManCompute(p, truth, 3, make([]byte, 3), ""); lit != 4 || If_DsdManObjNum(p) != 3 || o.Count != 8 {
		t.Fatal("loaded nodes were not interned")
	}
}
func TestDsdPersistenceRetainsPrimeAndCellState(t *testing.T) {
	for _, configured := range []bool{false, true} {
		t.Run(map[bool]string{false: "decomposition", true: "cell"}[configured], func(t *testing.T) {
			p := If_DsdManAlloc(4, 2)
			defer If_DsdManFree(p, 0)
			for _, text := range []string{"(ab)", "[abc]", "E8{abc}", "E8{a(bc)d}", "17E8{abcd}", "D3C5{abcd}"} {
				n := dau.Dau_DsdFindVarNum(text)
				If_DsdManCompute(p, dau.Dau_DsdToTruth(text, n), n, make([]byte, n), "")
			}
			if configured {
				Id_DsdManTuneStr1(p, "{abcd}", 0, 0)
			}
			path := filepath.Join(t.TempDir(), "roundtrip.dsd")
			If_DsdManSave(p, path)
			q := If_DsdManLoad(path)
			if q == nil {
				t.Fatal("saved manager could not be loaded")
			}
			defer If_DsdManFree(q, 0)
			if q.nVars != p.nVars || q.LutSize != p.LutSize || q.nConfigWords != p.nConfigWords || q.nTtBits != p.nTtBits || q.pCellStr != p.pCellStr || If_DsdManObjNum(q) != If_DsdManObjNum(p) {
				t.Fatal("roundtrip manager metadata differs")
			}
			for id := int32(2); id < If_DsdManObjNum(p); id++ {
				a, b := If_DsdVecObj(&p.vObjs, id), If_DsdVecObj(&q.vObjs, id)
				if a.Type != b.Type || a.nSupp != b.nSupp || a.Count != b.Count || a.fMark != b.fMark || !slices.Equal(a.pFans, b.pFans) {
					t.Fatalf("roundtrip node%d differs", id)
				}
				if !slices.Equal(If_DsdManComputeTruth(p, id<<1, nil), If_DsdManComputeTruth(q, id<<1, nil)) {
					t.Fatalf("roundtrip truth%d differs", id)
				}
				if !slices.Equal(If_DsdManGetFuncConfig(p, id<<1), If_DsdManGetFuncConfig(q, id<<1)) {
					t.Fatalf("roundtrip config%d differs", id)
				}
			}
			for v := int32(3); v <= p.nVars; v++ {
				if vec.Vec_MemEntryNum(p.vTtMem[v]) != vec.Vec_MemEntryNum(q.vTtMem[v]) || Vec_PtrSize(p.vTtDecs[v]) != Vec_PtrSize(q.vTtDecs[v]) {
					t.Fatalf("prime cache%d dimensions differ", v)
				}
				for i := 0; i < vec.Vec_MemEntryNum(p.vTtMem[v]); i++ {
					if !slices.Equal(vec.Vec_MemReadEntry(p.vTtMem[v], i), vec.Vec_MemReadEntry(q.vTtMem[v], i)) {
						t.Fatal("prime truth changed")
					}
				}
				for i := 0; i < Vec_PtrSize(p.vTtDecs[v]); i++ {
					a := Vec_PtrEntry(p.vTtDecs[v], i).(*Vec_Int_t)
					b := Vec_PtrEntry(q.vTtDecs[v], i).(*Vec_Int_t)
					if !slices.Equal(Vec_IntArray(a), Vec_IntArray(b)) {
						t.Fatal("prime decomposition changed")
					}
				}
			}
			second := filepath.Join(t.TempDir(), "second.dsd")
			If_DsdManSave(q, second)
			a, _ := os.ReadFile(path)
			b, _ := os.ReadFile(second)
			if !bytes.Equal(a, b) {
				t.Fatal("binary roundtrip is not stable")
			}
		})
	}
}
func TestDsdSaveUsesDefaultStorePath(t *testing.T) {
	t.Chdir(t.TempDir())
	p := If_DsdManAlloc(3, 0)
	defer If_DsdManFree(p, 0)
	If_DsdManSave(p, "")
	if data, err := os.ReadFile("03.dsd"); err != nil || len(data) < 4 || string(data[:4]) != "dsd1" {
		t.Fatalf("default store not written: %v", err)
	}
}

func TestDsdLoadRejectsTruncatedAndInconsistentRecords(t *testing.T) {
	// Invalid-file diagnostics are expected; capture them to keep verbose test
	// output from looking like an application failure.
	output, err := os.CreateTemp(t.TempDir(), "invalid-load-output")
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = output
	defer func() { os.Stdout = original; output.Close() }()
	valid := dsdReviewSourceFile()
	path := filepath.Join(t.TempDir(), "invalid.dsd")
	reject := func(label string, data []byte) {
		t.Helper()
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
		if p := If_DsdManLoad(path); p != nil {
			If_DsdManFree(p, 0)
			t.Fatalf("accepted invalid DSD %s", label)
		}
	}
	for size := 0; size < len(valid); size++ {
		reject("truncated prefix", valid[:size])
	}
	for _, mutation := range []struct {
		name   string
		offset int
		value  uint32
	}{
		{"variables", 4, 17}, {"negative variables", 4, 0xffffffff}, {"LUT size", 8, 1}, {"node count", 12, 1}, {"huge count", 12, 0x7fffffff},
		{"huge node", 16, 0x7fffffff}, {"node ID", 20, 3}, {"node type", 24, 7 | 3<<3 | 3<<27}, {"support", 24, 3 | 2<<3 | 3<<27},
		{"self fanin", 28, 4}, {"negative fanin", 28, 0xffffffff}, {"constant fanin", 28, 0}, {"excess decompositions", 48, 1}, {"config width", 52, 0},
	} {
		data := append([]byte(nil), valid...)
		binary.LittleEndian.PutUint32(data[mutation.offset:], mutation.value)
		reject(mutation.name, data)
	}
	prime := append([]byte(nil), valid[:44]...)
	prime = binary.LittleEndian.AppendUint32(prime, 0)
	prime = append(prime, valid[44:]...)
	binary.LittleEndian.PutUint32(prime[24:], 6|3<<3|3<<27)
	reject("missing prime truth", prime)
	duplicate := append([]byte(nil), valid[:44]...)
	record := append([]byte(nil), valid[16:44]...)
	binary.LittleEndian.PutUint32(record[4:], 3)
	duplicate = append(duplicate, record...)
	duplicate = append(duplicate, valid[44:]...)
	binary.LittleEndian.PutUint32(duplicate[12:], 4)
	reject("duplicate canonical node", duplicate)
	duplicateTruth := append([]byte(nil), valid[:44]...)
	duplicateTruth = binary.LittleEndian.AppendUint32(duplicateTruth, 2)
	duplicateTruth = binary.LittleEndian.AppendUint64(duplicateTruth, 0x1717171717171717)
	duplicateTruth = binary.LittleEndian.AppendUint64(duplicateTruth, 0x1717171717171717)
	duplicateTruth = append(duplicateTruth, valid[48:]...)
	reject("duplicate prime truth", duplicateTruth)
}

func TestDsdPersistenceOmitsReservedBaseRecords(t *testing.T) {
	p := If_DsdManAlloc(2, 0)
	defer If_DsdManFree(p, 0)
	for id := int32(0); id < 2; id++ {
		o := If_DsdVecObj(&p.vObjs, id)
		o.Count = 9
		o.fMark = 1
	}
	path := filepath.Join(t.TempDir(), "bases.dsd")
	If_DsdManSave(p, path)
	q := If_DsdManLoad(path)
	if q == nil {
		t.Fatal("reserved-base-only file did not load")
	}
	defer If_DsdManFree(q, 0)
	for id := int32(0); id < 2; id++ {
		o := If_DsdVecObj(&q.vObjs, id)
		if o.Count != 0 || o.fMark != 0 {
			t.Fatal("source format does not store base counts/marks")
		}
	}
}

func TestDsdDecodedStateOwnsRetainedStorage(t *testing.T) {
	p := If_DsdManAlloc(4, 2)
	defer If_DsdManFree(p, 0)
	If_DsdManCompute(p, dau.Dau_DsdToTruth("D3C5{abcd}", 4), 4, make([]byte, 4), "")
	Id_DsdManTuneStr1(p, "{abcd}", 0, 0)
	var input bytes.Buffer
	if err := ifDsdWrite(p, &input); err != nil {
		t.Fatal(err)
	}
	data := input.Bytes()
	q, err := ifDsdRead(data, "owned.dsd")
	if err != nil {
		t.Fatal(err)
	}
	defer If_DsdManFree(q, 0)
	var expected bytes.Buffer
	if err := ifDsdWrite(q, &expected); err != nil {
		t.Fatal(err)
	}
	clear(data)
	var actual bytes.Buffer
	if err := ifDsdWrite(q, &actual); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual.Bytes(), expected.Bytes()) {
		t.Fatal("decoded state aliases borrowed input bytes")
	}
}
