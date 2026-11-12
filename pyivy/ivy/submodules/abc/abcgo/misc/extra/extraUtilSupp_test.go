package extra

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestSuppOriginalParityCover(t *testing.T) {
	hash := sha256.New()
	write32 := func(v int32) { var b [4]byte; binary.LittleEndian.PutUint32(b[:], uint32(v)); hash.Write(b[:]) }
	write64 := func(v uint64) { var b [8]byte; binary.LittleEndian.PutUint64(b[:], v); hash.Write(b[:]) }
	for n := int32(1); n <= 8; n++ {
		for m := int32(0); m <= n; m++ {
			vectors := Abc_SuppGen(m, n)
			pairs := Abc_SuppGenPairs(vectors, n)
			write32(int32(vec.Vec_WrdSize(vectors)))
			write32(int32(vec.Vec_WrdSize(pairs)))
			for _, v := range vec.Vec_WrdArray(vectors) {
				write64(v)
			}
			for _, v := range vec.Vec_WrdArray(pairs) {
				write64(v)
			}
			var matrix [64]uint64
			rows := Abc_SuppMinimize(matrix[:], pairs, n, 0)
			write32(rows)
			for _, v := range matrix[:rows] {
				write64(v)
			}
			if suppVerify(vectors, matrix[:], n, rows) != 0 {
				t.Fatalf("invalid parity encoding m%d n%d", m, n)
			}
		}
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != "d98b67a937954f27c1b36d8d589176e594ba40c56c1f3785ce1d30c6e78c6f03" {
		t.Fatalf("original C generation/matrix mismatch %s", got)
	}
}
func TestSuppParityMaskMatchesMarkers(t *testing.T) {
	seed := uint64(1)
	for iteration := 0; iteration < 1000; iteration++ {
		p := vec.Vec_WrdAlloc(30)
		var original []uint64
		for i := 0; i < 30; i++ {
			seed = seed*6364136223846793005 + 1
			v := uint64(1) + seed%127
			original = append(original, v)
			vec.Vec_WrdPush(p, v)
		}
		mask := Abc_SuppFindOne(p, 7)
		for i, v := range original {
			if vec.Vec_WrdArray(p)[i]>>7&1 != uint64(bits.OnesCount64(v&mask)&1) {
				t.Fatalf("mask differs from toggled parity at case%d", iteration)
			}
		}
	}
}
func bruteSuppGreedy(rows []uint64, n int32) int32 {
	removed := make([]bool, n)
	remaining := n
	rows = slices.Clone(rows)
	for {
		best := int32(-1)
		bestDouble := int(^uint(0) >> 1)
		for v := int32(0); v < n; v++ {
			if removed[v] {
				continue
			}
			single, double := 0, 0
			for _, row := range rows {
				if row>>v&1 == 0 {
					continue
				}
				switch bits.OnesCount64(row) {
				case 1:
					single++
				case 2:
					double++
				}
			}
			if single == 0 && double < bestDouble {
				best, bestDouble = v, double
			}
		}
		if best == -1 {
			return remaining
		}
		removed[best] = true
		remaining--
		for i := range rows {
			rows[i] &^= uint64(1) << best
		}
	}
}
func TestSuppRemovalGreedyAgainstRecomputation(t *testing.T) {
	seed := uint64(9)
	for iteration := 0; iteration < 1000; iteration++ {
		p := vec.Vec_WrdAlloc(20)
		n := int32(iteration%8 + 1)
		for i := 0; i < 20; i++ {
			seed = seed*6364136223846793005 + 1
			vec.Vec_WrdPush(p, 1+seed%((1<<n)-1))
		}
		original := slices.Clone(vec.Vec_WrdArray(p))
		want := bruteSuppGreedy(original, n)
		if got := Abc_SuppSolve(p, n); got != want {
			t.Fatalf("case%d got%d want%d", iteration, got, want)
		}
		if !slices.Equal(original, vec.Vec_WrdArray(p)) {
			t.Fatal("solver mutated caller matrix")
		}
	}
	empty := vec.Vec_WrdAlloc(0)
	if Abc_SuppSolve(empty, 64) != 0 {
		t.Fatal("unused variables should all be removed")
	}
}
func TestSuppMinReaderAndZeroDifference(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "cover.min")
	text := "header\nINPUT F-COVER 2\n10, 01, 10 ROW0\n01 10 01 ROW1\nINPUT R-COVER\n"
	if err := os.WriteFile(name, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	cubes, n, err := Abc_SuppReadMin(name)
	if err != nil || n != 3 || !slices.Equal(vec.Vec_WrdArray(cubes), []uint64{5, 2}) {
		t.Fatal(cubes, n, err)
	}
	diff := Abc_SuppDiffMatrix(cubes)
	if !slices.Equal(vec.Vec_WrdArray(diff), []uint64{7}) {
		t.Fatal("difference matrix")
	}
	for _, text := range []string{"missing marker", "INPUT F-COVER 2 10 ROW0", "INPUT F-COVER 1 00 ROW0"} {
		os.WriteFile(name, []byte(text), 0600)
		if _, _, err = Abc_SuppReadMin(name); err == nil {
			t.Fatalf("accepted %q", text)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("zero difference should report impossibility")
		}
	}()
	zero := vec.Vec_WrdStart(1)
	Abc_SuppMinimize(make([]uint64, 64), zero, 3, 0)
}
