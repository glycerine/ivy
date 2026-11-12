package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"slices"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestPrimeOriginalCubePairs(t *testing.T) {
	hash := sha256.New()
	write32 := func(v int32) { var b [4]byte; binary.LittleEndian.PutUint32(b[:], uint32(v)); hash.Write(b[:]) }
	for n := int32(2); n <= 10; n++ {
		_, primes := primeSieve(n)
		p := Tab_ManAlloc(n, int32(vec.Vec_IntSize(primes)))
		Tab_ManStart(p, primes)
		pairs := Tab_ManCollectDist1(p, 0)
		for _, v := range []int32{n, int32(vec.Vec_IntSize(primes)), int32(vec.Vec_IntSize(pairs))} {
			write32(v)
		}
		for _, v := range vec.Vec_IntArray(primes) {
			write32(v)
		}
		for _, v := range p.pValues {
			var b [8]byte
			binary.LittleEndian.PutUint64(b[:], v)
			hash.Write(b[:])
		}
		for _, v := range vec.Vec_IntArray(pairs) {
			write32(v)
		}
		again := Tab_ManCollectDist1(p, 0)
		if !slices.Equal(vec.Vec_IntArray(pairs), vec.Vec_IntArray(again)) {
			t.Fatal("second collection changed pairs")
		}
		for k := 0; k < vec.Vec_IntSize(pairs); k += 2 {
			a, b := vec.Vec_IntEntry(pairs, k), vec.Vec_IntEntry(pairs, k+1)
			x, y := vec.Vec_IntEntry(primes, int(a)), vec.Vec_IntEntry(primes, int(b))
			if x^y == 0 || (x^y)&((x^y)-1) != 0 {
				t.Fatal("non-distance-1 pair")
			}
		}
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "06501e6a91ad7ea34b5f393c2906d4c9fb1dc2451603034a3b9fac3316e7f906" {
		t.Fatalf("original C table/pair mismatch %s", got)
	}
}
func TestPrimeSieveTrialDivision(t *testing.T) {
	for n := int32(1); n <= 13; n++ {
		composite, primes := primeSieve(n)
		at := 0
		for v := int32(0); v < 1<<n; v++ {
			prime := v >= 2
			for d := int32(2); d <= v/max(1, d); d++ {
				if v%d == 0 {
					prime = false
					break
				}
			}
			if (vec.Vec_BitEntry(composite, int(v)) == 0) != prime {
				t.Fatalf("sieve %d", v)
			}
			if prime {
				if vec.Vec_IntEntry(primes, at) != v {
					t.Fatal("prime order")
				}
				at++
			}
		}
		if at != vec.Vec_IntSize(primes) {
			t.Fatal("prime count")
		}
	}
}
func TestPrimeCubeComparisonAndFree(t *testing.T) {
	a, b := []int32{3, 0, 2, 4}, []int32{3, 0, 3, 4}
	if Tab_ManCheckEqual(a, b, 2, 2) != 1 || Tab_ManCheckEqual2(a, b, 2, 2) != 1 || Tab_ManCheckEqual(a, b, 0, 0) != 0 {
		t.Fatal("cube equality")
	}
	out := vec.Vec_IntAlloc(4)
	if Tab_ManCubeFree(a, b, out) != 2 || !slices.Equal(vec.Vec_IntArray(out), []int32{4, 7}) {
		t.Fatal("cube-free divisor")
	}
	if Tab_ManCubeFree(a, a, out) != 3 || vec.Vec_IntSize(out) != 0 {
		t.Fatal("duplicate cube handling")
	}
	p := Tab_ManAlloc(3, 2)
	count, starts := Tab_ManCountItems(p, 1)
	if count != 2 || !slices.Equal(vec.Vec_IntArray(starts), []int32{0, 1}) {
		t.Fatal("empty-cube count")
	}
	Tab_ManCube(p, 0)[0] = 3
	Tab_ManCube(p, 1)[0] = 3
	count, _ = Tab_ManCountItems(p, 1)
	if count != 14 {
		t.Fatal("distance-two variant count")
	}
	Tab_ManFree(p)
	Tab_ManFree(p)
	if p.pTable != nil || p.pCubes != nil {
		t.Fatal("free retained storage")
	}
}

func TestPrimeEndOmissionAndDistanceTwo(t *testing.T) {
	if Tab_ManCheckEqual2([]int32{2, 0, 2}, []int32{2, 0, 3}, 2, 2) != 1 {
		t.Fatal("last literal omission")
	}
	cubes := vec.Vec_IntAlloc(4)
	for _, v := range []int32{1, 3, 5, 7} {
		vec.Vec_IntPush(cubes, v)
	}
	p := Tab_ManAlloc(3, 4)
	Tab_ManStart(p, cubes)
	// Original distance-2 item counting used n*n/2 and failed its own item-count
	// assertion. Building all 1+n+choose(n,2) variants must terminate and be repeatable.
	a := Tab_ManCollectDist1(p, 1)
	b := Tab_ManCollectDist1(p, 1)
	if !slices.Equal(vec.Vec_IntArray(a), vec.Vec_IntArray(b)) {
		t.Fatal("distance-two repeatability")
	}
}
