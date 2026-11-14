package gia

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
	"testing"
)

func TestMappingTransferRemapsPrunesAndPreservesMuxSigns(t *testing.T) {
	p, q := Gia_ManStart(1), Gia_ManStart(1)
	a, b, c, unused := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	qc, qa, qb := Gia_ManAppendCi(q), Gia_ManAppendCi(q), Gia_ManAppendCi(q)
	roots, copies := make([]int32, 7), make([]int32, 6)
	for i := range roots {
		control := c
		if i == 6 {
			control = unused
		}
		roots[i] = Gia_ManAppendMux(p, control, a^int32(i&1), b^int32((i>>1)&1))
	}
	for i := 5; i >= 0; i-- {
		copies[i] = Gia_ManAppendMux(q, qc, qa^int32(i&1), qb^int32((i>>1)&1))
	}
	for i := range copies {
		Gia_ManAppendCo(p, roots[i])
		Gia_ManAppendCo(q, copies[i])
	}
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for i, root := range roots {
		id := root >> 1
		Vec_IntWriteEntry(p.VMapping, int(id), int32(Vec_IntSize(p.VMapping)))
		trailer := id
		if i&1 != 0 {
			trailer = -id
		}
		for _, entry := range []int32{4, a >> 1, unused >> 1, b >> 1, c >> 1, trailer} {
			Vec_IntPush(p.VMapping, entry)
		}
	}
	Gia_ManFillValue(p)
	Gia_ManConst0(p).Value = 0
	for old, newLit := range map[int32]int32{a >> 1: qa, b >> 1: qb, c >> 1: qc} {
		Gia_ManObj(p, old).Value = uint32(newLit)
	}
	for i, copied := range copies {
		Gia_ManObj(p, roots[i]>>1).Value = uint32(copied &^ 1)
	}
	p.VPacking = Vec_IntAlloc(10)
	Vec_IntPush(p.VPacking, 3)
	for size, k := 1, 0; size <= 3; size++ {
		Vec_IntPush(p.VPacking, int32(size))
		for j := 0; j < size; j++ {
			Vec_IntPush(p.VPacking, roots[k]>>1)
			k++
		}
	}
	originalMapping := slices.Clone(Vec_IntArray(p.VMapping))
	originalPacking := slices.Clone(Vec_IntArray(p.VPacking))
	q.VMapping = Vec_IntStart(int(q.nObjs))
	oldDestinationMapping := q.VMapping
	Gia_ManTransferMapping(q, p)
	Gia_ManTransferPacking(q, p)
	if q.VMapping == oldDestinationMapping || Gia_ManLutNum(q) != 6 {
		t.Fatal("mapping was not replaced or dangling LUT was retained")
	}
	for i, copied := range copies {
		id := copied >> 1
		trailer := id
		if i&1 != 0 {
			trailer = -id
		}
		if !slices.Equal(Gia_ObjLutFanins(q, id), []int32{qa >> 1, qb >> 1, qc >> 1}) || Gia_ObjLutMuxId(q, id) != trailer {
			t.Fatalf("LUT %d: remapping, dangling fanin pruning or MUX sign failed", i)
		}
	}
	wantPacking := []int32{3, 1, copies[0] >> 1, 2, copies[1] >> 1, copies[2] >> 1, 3, copies[3] >> 1, copies[4] >> 1, copies[5] >> 1}
	if !slices.Equal(Vec_IntArray(q.VPacking), wantPacking) {
		t.Fatalf("packing %v, want %v", Vec_IntArray(q.VPacking), wantPacking)
	}
	if !slices.Equal(Vec_IntArray(p.VMapping), originalMapping) || !slices.Equal(Vec_IntArray(p.VPacking), originalPacking) {
		t.Fatal("transfer modified source records")
	}
	var got strings.Builder
	for _, record := range []struct {
		name string
		data *Vec_Int_t
	}{{"mapping", q.VMapping}, {"packing", q.VPacking}} {
		got.WriteString(record.name)
		for _, entry := range Vec_IntArray(record.data) {
			fmt.Fprintf(&got, " %d", entry)
		}
		got.WriteByte('\n')
	}
	Vec_IntWriteEntry(q.VPacking, 2, 0)
	Vec_IntWriteEntry(q.VMapping, 0, 42)
	if !slices.Equal(Vec_IntArray(p.VMapping), originalMapping) || !slices.Equal(Vec_IntArray(p.VPacking), originalPacking) {
		t.Fatal("destination records alias source records")
	}
	// Compare complete destination records with the original C routines, not
	// another implementation of the transfer algorithm in the test.
	exe := compileGiaSourceOracle(t, "mapping_transfer.c")
	want, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C mapping transfer: %v: %s", err, want)
	}
	if got.String() != string(want) {
		t.Fatalf("Go:\n%sC:\n%s", got.String(), want)
	}
}
