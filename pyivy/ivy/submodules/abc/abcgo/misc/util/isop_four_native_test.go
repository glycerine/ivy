package util

import "testing"

func isopFourDigest(size int) uint64 {
	const repeat uint64 = 0x0001000100010001
	state := uint32(0x27159385)
	hash := uint64(14695981039346656037)
	for i := range size {
		state ^= state << 13
		state ^= state >> 17
		state ^= state << 5
		exact := uint16(i*40503 + 7919)
		for _, allowed := range [2]uint16{exact, uint16(state)} {
			count := int32(7)
			result := Abc_Tt6Isop(uint64(exact&allowed)*repeat, uint64(allowed)*repeat, 4, &count)
			hash = (hash ^ result) * 1099511628211
			hash = (hash ^ uint64(count)) * 1099511628211
		}
	}
	return hash
}

func TestIsopFourNativeTruthAndCounts(t *testing.T) {
	if got := isopFourDigest(65536); got != 0x8e445a19aae238cd {
		t.Fatalf("four-input ISOP truth and counts = %016x, native 8e445a19aae238cd", got)
	}
}

func benchmarkIsopFour(b *testing.B, size int, want uint64) {
	b.Helper()
	b.ReportAllocs()
	for b.Loop() {
		if got := isopFourDigest(size); got != want {
			b.Fatalf("four-input ISOP digest = %016x, want %016x", got, want)
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/(float64(b.N)*float64(size*2)), "ns/pair")
}
func BenchmarkIsopFourDiverse(b *testing.B)  { benchmarkIsopFour(b, 65536, 0x8e445a19aae238cd) }
func BenchmarkIsopFourRepeated(b *testing.B) { benchmarkIsopFour(b, 128, 0x3b6cdca3e521626d) }

func TestIsopFourConcurrentPackedCounts(t *testing.T) {
	const first = 0x69a4
	var counts [4]int32
	var packed uint32
	for i := range counts {
		truth := uint64(first+i) * isopHalfRepeat
		var cover [16]int32
		if got := Abc_Tt6IsopCover(truth, truth, 4, cover[:], &counts[i]); got != truth {
			t.Fatal("exact cover differs from input")
		}
		packed |= uint32(counts[i]) << uint(8*i)
	}
	slot := &isop4ExactCounts[first>>2]
	for range 16 {
		slot.Store(0)
		start := make(chan struct{})
		results := make(chan bool, len(counts))
		for i, want := range counts {
			go func() {
				<-start
				truth := uint64(first+i) * isopHalfRepeat
				cubes := int32(3)
				got := Abc_Tt6Isop(truth, truth, 4, &cubes)
				results <- got == truth && cubes == want+3
			}()
		}
		close(start)
		for range counts {
			if !<-results {
				t.Fatal("concurrent exact truth or count changed")
			}
		}
		if got := slot.Load(); got != packed {
			t.Fatalf("packed counts %08x, want %08x", got, packed)
		}
	}
}
