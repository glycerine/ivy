package vec

import (
	"bytes"
	"math"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPointerVectorsAndSimulation(t *testing.T) {
	p := Vec_PtrAlloc(0)
	a, b := new(int), new(int)
	Vec_PtrPush(p, a)
	Vec_PtrPush(p, b)
	Vec_PtrPush(p, a)
	if Vec_PtrPushUnique(p, a) != 1 || Vec_PtrFind(p, b) != 1 {
		t.Fatal("identity")
	}
	Vec_PtrRemove(p, a)
	if Vec_PtrSize(p) != 2 || Vec_PtrEntry(p, 0) != a {
		t.Fatal("remove last occurrence")
	}
	Vec_PtrPushFirst(p, b)
	Vec_PtrReorder(p, 1)
	if Vec_PtrEntryLast(p) != b {
		t.Fatal("reorder")
	}
	// Byte slices model C pointers by identity, not by equal contents.
	x, y := []byte{1}, []byte{1}
	Vec_PtrClear(p)
	Vec_PtrPush(p, x)
	if Vec_PtrFind(p, y) != -1 || Vec_PtrFind(p, x) != 0 {
		t.Fatal("slice identity")
	}
	sim := Vec_PtrAllocTruthTables(6)
	if Vec_PtrEntry(sim, 0).([]uint32)[0] != 0xaaaaaaaa || Vec_PtrEntry(sim, 5).([]uint32)[0] != 0 || Vec_PtrEntry(sim, 5).([]uint32)[1] != 0xffffffff {
		t.Fatal("truth table masks")
	}
	Vec_PtrDoubleSimInfo(sim)
	Vec_PtrReallocSimInfo(sim)
	if len(sim.data) != 12 || Vec_PtrReadWordsSimInfo(sim) != 4 || Vec_PtrEntry(sim, 0).([]uint32)[0] != 0xaaaaaaaa {
		t.Fatal("simulation resizing")
	}
	Vec_PtrFillSimInfo(sim, 1, 4)
	Vec_PtrCleanSimInfo(sim, 2, 4)
	if Vec_PtrEntry(sim, 0).([]uint32)[1] != 0xffffffff || Vec_PtrEntry(sim, 0).([]uint32)[2] != 0 {
		t.Fatal("range fill")
	}
}
func TestBitVectorsBoundaryGrowth(t *testing.T) {
	p := Vec_BitAlloc(1)
	for i := 0; i < 1000; i++ {
		Vec_BitPush(p, i&1)
	}
	if Vec_BitSize(p) != 1000 || Vec_BitCount(p) != 500 {
		t.Fatal("bit growth")
	}
	Vec_BitShrink(p, 999)
	if Vec_BitCount(p) != 499 {
		t.Fatal("tail masking")
	}
	Vec_BitFillExtra(p, 1050, 1)
	if Vec_BitCount(p) != 550 {
		t.Fatal("fill across boundaries")
	}
	Vec_BitReset(p)
	if Vec_BitCount(p) != 0 {
		t.Fatal("reset")
	}
	full := Vec_BitStartFull(33)
	if Vec_BitSize(full) != 64 || Vec_BitCount(full) != 64 {
		t.Fatal("rounded start size")
	}
}
func TestWordHexAndRecordClasses(t *testing.T) {
	p := Vec_WrdAllocArrayCopy([]uint64{0x123456789abcdef0, 0x0123456789abcdef, 0x123456789abcdef0}, 3)
	var m *Vec_Int_t
	if Vec_WrdUniqueCount(p, 1, &m) != 2 || !reflect.DeepEqual(Vec_IntArray(m), []int32{0, 1, 0}) {
		t.Fatal("word records")
	}
	var out bytes.Buffer
	Gia_ManSimPatWriteOne(&out, Vec_WrdArray(p), 1)
	if out.String() != "0FEDCBA987654321\n" {
		t.Fatal(out.String())
	}
	path := filepath.Join(t.TempDir(), "patterns.hex")
	Vec_WrdDumpHex(path, p, 1, 0)
	words := 0
	q := Vec_WrdReadHex(path, &words, 0)
	if words != 1 || Vec_WrdEqual(p, q) != 1 {
		t.Fatal("hex roundtrip")
	}
	truth := Vec_WrdStartTruthTables(7)
	if Vec_WrdEntry(truth, 0) != 0xaaaaaaaaaaaaaaaa || Vec_WrdEntry(truth, 12) != 0 || Vec_WrdEntry(truth, 13) != ^uint64(0) {
		t.Fatal("64-bit truth masks")
	}
}
func TestByteVectorBinaryRoundtrip(t *testing.T) {
	p := Vec_StrAlloc(0)
	Vec_StrPutI_ne(p, -123)
	Vec_StrPutI(p, 0x7fffffff)
	Vec_StrPutW(p, 0xfedcba9876543210)
	Vec_StrPutF(p, float32(-1.25))
	Vec_StrPutD(p, math.Inf(1))
	Vec_StrPutS(p, "hello")
	Vec_StrPutC(p, 0xff)
	pos := 0
	if Vec_StrGetI_ne(p, &pos) != -123 || Vec_StrGetI(p, &pos) != 0x7fffffff || Vec_StrGetW(p, &pos) != 0xfedcba9876543210 || Vec_StrGetF(p, &pos) != -1.25 || !math.IsInf(Vec_StrGetD(p, &pos), 1) || Vec_StrGetS(p, &pos) != "hello" || Vec_StrGetC(p, &pos) != 0xff || pos != len(p.data) {
		t.Fatal("binary roundtrip")
	}
	Vec_StrClear(p)
	Vec_StrPrintNumStar(p, -7, 4)
	Vec_StrPrintF(p, "-%s-%d", "x", 9)
	if string(Vec_StrArray(p)) != "-007-x-9" {
		t.Fatal(string(Vec_StrArray(p)))
	}
}
func TestLevelVectorsMarksAndOwnership(t *testing.T) {
	p := Vec_WecAlloc(0)
	Vec_WecPush(p, 4, 10)
	Vec_WecPush(p, 1, 2)
	Vec_WecPushUnique(p, 1, 2)
	if Vec_WecSizeSize(p) != 2 || Vec_WecSizeUsed(p) != 2 {
		t.Fatal("level sizes")
	}
	q := Vec_WecDup(p)
	Vec_IntWriteEntry(Vec_WecEntry(q, 1), 0, 99)
	if Vec_WecEntryEntry(p, 1, 0) != 2 {
		t.Fatal("deep copy")
	}
	levels := Vec_IntAllocArrayCopy([]int32{1, 4}, 2)
	Vec_WecMarkLevels(p, levels)
	if Vec_WecIntHasMark(Vec_WecEntry(p, 4)) != 1 {
		t.Fatal("mark")
	}
	Vec_WecUnmarkLevels(p, levels)
	Vec_WecRemoveEmpty(p)
	if len(p.data) != 2 || Vec_WecEntryEntry(p, 0, 0) != 2 {
		t.Fatal("remove empty")
	}
	Vec_WecInsertLevel(p, 1)
	if Vec_WecLevelSize(p, 1) != 0 || Vec_WecEntryEntry(p, 2, 0) != 10 {
		t.Fatal("insert")
	}
	outer := Vec_VecAlloc(0)
	Vec_VecPushInt(outer, 2, 7)
	Vec_VecClear(outer)
	if Vec_VecSize(outer) != 3 || Vec_VecSizeSize(outer) != 0 {
		t.Fatal("pointer levels retain shape")
	}
}
func TestIndexedPriorityQueue(t *testing.T) {
	costs := []float32{1, 4, 2, 3}
	p := Vec_QueAlloc(0)
	Vec_QueSetPriority(p, &costs)
	for i := range costs {
		Vec_QuePush(p, i)
	}
	Vec_QueCheck(p)
	if Vec_QuePop(p) != 1 {
		t.Fatal("maximum")
	}
	costs[0] = 10
	Vec_QueUpdate(p, 0)
	Vec_QueCheck(p)
	if Vec_QuePop(p) != 0 || Vec_QuePop(p) != 3 || Vec_QuePop(p) != 2 {
		t.Fatal("updated order")
	}
	// Arbitrary sparse member IDs must work, not just a dense prefix.
	Vec_QueSetPriority(Vec_QueAlloc(1), nil)
	q := Vec_QueAlloc(1)
	Vec_QuePush(q, 100)
	Vec_QuePush(q, 3)
	Vec_QueCheck(q)
	Vec_QueClear(q)
	if Vec_QueIsMember(q, 100) != 0 || Vec_QueSize(q) != 0 {
		t.Fatal("clear")
	}
}
func TestPagedWordMemoryHash(t *testing.T) {
	p := Vec_MemAlloc(2, 2)
	Vec_MemHashAlloc(p, 2)
	first := []uint64{1, 2}
	if Vec_MemHashInsert(p, first) != 0 || Vec_MemHashInsert(p, first) != 0 {
		t.Fatal("interning")
	}
	stable := Vec_MemReadEntry(p, 0)
	for i := 1; i < 100; i++ {
		if Vec_MemHashInsert(p, []uint64{uint64(i + 1), uint64(i)}) != i {
			t.Fatal("record id")
		}
	}
	if stable[0] != 1 || stable[1] != 2 || Vec_MemPageNum(p) != 25 {
		t.Fatal("stable pages")
	}
	q := Vec_MemDup(p)
	Vec_MemWriteEntry(q, 0, []uint64{8, 9})
	if stable[0] != 1 {
		t.Fatal("deep copy")
	}
	Vec_MemShrink(p, 1)
	if Vec_MemHashInsert(p, first) != 0 || Vec_MemPageNum(p) != 1 {
		t.Fatal("shrink rehash")
	}
	truth := Vec_MemAllocForTT(4, 0)
	Vec_MemAddMuxTT(truth, 4)
	if truth.nEntries != 3 || Vec_MemReadEntry(truth, 2)[0] != 0xcacacacacacacaca {
		t.Fatal("truth records")
	}
}
