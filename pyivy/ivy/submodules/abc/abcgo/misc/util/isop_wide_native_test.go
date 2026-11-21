package util

import "testing"

const nativeWideIsopDigest = 0xc5fa775abb4222c8

// Keep the ordered covers, counts and result words in the same digest as
// test_vectors/perf_isop_wide.c, including the caller's initial two cubes.
func wideIsopDigest(alias bool) uint64 {
	hash := uint64(14695981039346656037)
	mix := func(value word) { hash = (hash ^ value) * 1099511628211 }
	state := word(0xd51d28de17)
	next := func() word {
		state ^= state << 13
		state ^= state >> 7
		state ^= state << 17
		return state
	}
	for trial := range 96 {
		n := int32(6 + trial%3)
		var on, dc, out, covered [4]word
		for i := range dc {
			dc[i] = next()
			on[i] = dc[i] & next()
		}
		switch trial % 4 {
		case 0:
			dc = on
		case 1:
			copy(on[2:], on[:2])
			copy(dc[2:], dc[:2])
		case 2:
			on[1], dc[1] = on[0], dc[0]
		}
		codes, count := [260]int32{7, 11}, int32(2)
		var size int32
		if alias {
			out, covered = on, dc
			size = Abc_Tt8Isop(out[:], dc[:], n, out[:])
			Abc_Tt8IsopCover(on[:], covered[:], n, covered[:], codes[:], &count)
		} else {
			size = Abc_Tt8Isop(on[:], dc[:], n, out[:])
			Abc_Tt8IsopCover(on[:], dc[:], n, covered[:], codes[:], &count)
		}
		mix(word(size))
		mix(word(count))
		for k := range out {
			mix(out[k])
			mix(covered[k])
		}
		for _, code := range codes[:count] {
			mix(word(code))
		}
	}
	return hash
}

func TestWideIsopNativeOrderedCovers(t *testing.T) {
	for _, alias := range []bool{false, true} {
		if got := wideIsopDigest(alias); got != nativeWideIsopDigest {
			t.Fatalf("aliased %t, got %016x, native %016x", alias, got, uint64(nativeWideIsopDigest))
		}
	}
}

func BenchmarkWideIsopOrderedCovers(b *testing.B) {
	for b.Loop() {
		if wideIsopDigest(false) != nativeWideIsopDigest {
			b.Fatal("wide ISOP differs from native ABC")
		}
	}
	b.ReportMetric(float64(b.Elapsed().Nanoseconds())/(96*float64(b.N)), "ns/table")
}
