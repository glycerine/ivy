package gia

import (
	"slices"
	"testing"
)

func TestMappingTransferRemapsPrunesAndPreservesMuxSigns(t *testing.T) {
	p, q := Gia_ManStart(1), Gia_ManStart(1)
	defer Gia_ManStop(p)
	defer Gia_ManStop(q)
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
	// Exact source record order: copied roots are 21,18,15,12,9,6;
	// the remapped CI order is qa=2, qb=3, qc=1.
	wantMapping := []int32{
		0, 0, 0, 0, 0, 0, 53, 0, 0, 48, 0, 0, 43, 0,
		0, 38, 0, 0, 33, 0, 0, 28, 0, 0, 0, 0, 0, 0,
		3, 2, 3, 1, 21,
		3, 2, 3, 1, -18,
		3, 2, 3, 1, 15,
		3, 2, 3, 1, -12,
		3, 2, 3, 1, 9,
		3, 2, 3, 1, -6,
	}
	if !slices.Equal(Vec_IntArray(q.VMapping), wantMapping) {
		t.Fatalf("mapping %v, want %v", Vec_IntArray(q.VMapping), wantMapping)
	}
	Vec_IntWriteEntry(q.VPacking, 2, 0)
	Vec_IntWriteEntry(q.VMapping, 0, 42)
	if !slices.Equal(Vec_IntArray(p.VMapping), originalMapping) || !slices.Equal(Vec_IntArray(p.VPacking), originalPacking) {
		t.Fatal("destination records alias source records")
	}
}

func TestMappingTransfersWithoutSourceMetadataAreNoops(t *testing.T) {
	src, dst := Gia_ManStart(1), Gia_ManStart(1)
	defer Gia_ManStop(src)
	defer Gia_ManStop(dst)
	dst.VMapping, dst.VPacking = Vec_IntStart(1), Vec_IntStart(1)
	mapping, packing := dst.VMapping, dst.VPacking
	output := hashReviewStdout(t, func() {
		Gia_ManTransferMapping(dst, src)
		Gia_ManTransferPacking(dst, src)
	})
	if output != "" || dst.VMapping != mapping || dst.VPacking != packing || !slices.Equal(Vec_IntArray(mapping), []int32{0}) || !slices.Equal(Vec_IntArray(packing), []int32{0}) {
		t.Fatal("absent source metadata replaced destination metadata")
	}
}
