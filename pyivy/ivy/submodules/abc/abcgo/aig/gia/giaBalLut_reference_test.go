package gia

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLutBalanceOriginalABCOracle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	binary := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	rng := rand.New(rand.NewSource(715))
	for run := 0; run < 8; run++ {
		src := Gia_ManStart(200)
		literals := []int32{0, 1}
		for i := 0; i < 6; i++ {
			literals = append(literals, Gia_ManAppendCi(src))
		}
		Gia_ManHashStart(src)
		for i := 0; i < 60; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			var lit int32
			switch rng.Intn(3) {
			case 0:
				lit = Gia_ManHashAnd(src, a, b)
			case 1:
				lit = Gia_ManHashXor(src, a, b)
			default:
				lit = Gia_ManHashMux(src, literals[rng.Intn(len(literals))], a, b)
			}
			literals = append(literals, lit)
		}
		for i := 0; i < 5; i++ {
			Gia_ManAppendCo(src, literals[len(literals)-1-i])
		}
		Gia_ManHashStop(src)
		data, err := MarshalAiger(src, AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.aig")
		if err := os.WriteFile(input, data, 0600); err != nil {
			t.Fatal(err)
		}
		for _, size := range []int32{3, 4, 6} {
			for _, cuts := range []int32{2, 8} {
				command := fmt.Sprintf("&r -s %s; &blut -n -K %d -C %d; &w %s", input, size, cuts, output)
				text, err := exec.Command(binary, "-c", command).CombinedOutput()
				if err != nil {
					t.Fatalf("original LUT balance %v: %s", err, text)
				}
				data, err := os.ReadFile(output)
				if err != nil {
					t.Fatalf("missing ABC output %v: %s", err, text)
				}
				original, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
				if err != nil {
					t.Fatal(err)
				}
				got := Gia_ManBalanceLut(src, size, cuts, 0)
				if Gia_ManAndNum(got) != Gia_ManAndNum(original) || Gia_ManLevelNum(got) != Gia_ManLevelNum(original) {
					t.Fatalf("run%d K%d C%d Go ands/depth %d/%d C %d/%d", run, size, cuts, Gia_ManAndNum(got), Gia_ManLevelNum(got), Gia_ManAndNum(original), Gia_ManLevelNum(original))
				}
				for mask := 0; mask < 64; mask++ {
					in := make([]bool, 6)
					for i := range in {
						in[i] = mask&(1<<i) != 0
					}
					want, e := original.Evaluate(in)
					if e != nil {
						t.Fatal(e)
					}
					actual, e := got.Evaluate(in)
					if e != nil {
						t.Fatal(e)
					}
					for i := range want {
						if actual[i] != want[i] {
							t.Fatalf("run%d mask%d output%d", run, mask, i)
						}
					}
				}
			}
		}
	}
}
