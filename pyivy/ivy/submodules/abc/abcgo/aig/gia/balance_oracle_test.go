package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
)

func TestBalanceOriginalABCOracle(t *testing.T) {
	binary := testoracle.NativeABCPath()
	if _, err := testoracle.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	rng := rand.New(rand.NewSource(715))
	for run := 0; run < 12; run++ {
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
		for _, simple := range []int32{0, 1} {
			for _, area := range []bool{false, true} {
				options := "-N 20"
				if !area {
					options = "-d -s"
				}
				if simple != 0 {
					options += " -a"
				}
				// relative names: &w writes the model name, the path read
				cmd := testoracle.Command(t, binary, "-c", fmt.Sprintf("&r -s input.aig; &b %s; &w output.aig", options))
				cmd.Dir = dir
				text, err := cmd.CombinedOutput()
				if err != nil {
					t.Fatalf("original balance oracle: %v\n%s", err, text)
				}
				data, err := testoracle.ReadFile(output)
				if err != nil {
					t.Fatalf("missing ABC output: %v\n%s", err, text)
				}
				original, err := ParseAiger(data, AigerReadOptions{SkipStrash: true})
				if err != nil {
					t.Fatal(err)
				}
				var got *Gia_Man_t
				if area {
					got = Gia_ManAreaBalance(src, simple, 20, 0, 0)
				} else {
					got = Gia_ManBalance(src, simple, 1, 0)
				}
				if Gia_ManAndNum(got) != Gia_ManAndNum(original) || Gia_ManLevelNum(got) != Gia_ManLevelNum(original) {
					t.Fatalf("source balance structure differs run%d simple%d area%v: Go ands=%d depth=%d C ands=%d depth=%d", run, simple, area, Gia_ManAndNum(got), Gia_ManLevelNum(got), Gia_ManAndNum(original), Gia_ManLevelNum(original))
				}
			}
		}
	}
}
