package vec

import "testing"

func TestMemHashNativeSmallTables(t *testing.T) {
	for hint, want := range []int{1, 1, 3, 3, 5} {
		mem := Vec_MemAlloc(1, 2)
		Vec_MemHashAlloc(mem, hint)
		if got := Vec_IntSize(mem.vTable); got != want {
			t.Errorf("hint %d got %d buckets, native has %d", hint, got, want)
		}
		if got := Vec_MemHashKey(mem, []uint64{1}); got != 1699%uint32(want) {
			t.Errorf("hint %d returned hash bucket %d", hint, got)
		}
	}
}

const nativeMemHashDigest uint64 = 0xfce2fa148b5a01ac

func memHashInputs() (data [128][17]uint64) {
	state := uint64(0x18532dc021)
	for i := range data {
		for j := range data[i] {
			state ^= state << 13
			state ^= state >> 7
			state ^= state << 17
			data[i][j] = state
		}
	}
	return data
}

func memHashDigest(data *[128][17]uint64) uint64 {
	hash := uint64(14695981039346656037)
	for _, width := range []int{1, 2, 4, 5, 9, 17} {
		for _, hint := range []int{0, 1, 2, 3, 4, 17, 32} {
			mem := Vec_MemAlloc(width, 4)
			Vec_MemHashAlloc(mem, hint)
			sum := uint64(Vec_IntSize(mem.vTable))
			for i := range data {
				key := Vec_MemHashKey(mem, data[i][:])
				id := Vec_MemHashInsert(mem, data[i][:])
				prior := Vec_MemHashInsert(mem, data[i/3][:])
				sum = (sum ^ uint64(key)) * 1099511628211
				sum = (sum ^ uint64(id)) * 1099511628211
				sum = (sum ^ uint64(prior)) * 1099511628211
			}
			hash = (hash ^ sum) * 1099511628211
		}
	}
	return hash
}

func TestMemHashNativeWidthsAndGrowth(t *testing.T) {
	data := memHashInputs()
	if hash := memHashDigest(&data); hash != nativeMemHashDigest {
		t.Fatalf("hash keys and record IDs got %016x, native %016x", hash, nativeMemHashDigest)
	}
}

func BenchmarkMemHashRecords(b *testing.B) {
	data := memHashInputs()
	for b.Loop() {
		if memHashDigest(&data) != nativeMemHashDigest {
			b.Fatal("paged hash differs from native")
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/(6*7*128*float64(b.N)), "ns/record")
}
