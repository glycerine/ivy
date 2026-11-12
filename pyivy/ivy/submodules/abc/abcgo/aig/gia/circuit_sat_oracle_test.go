package gia

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"testing"
)

func TestCircuitSATOriginalABCOracle(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	binary := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../abc"))
	if _, err := os.Stat(binary); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	rng := rand.New(rand.NewSource(992))
	for run := 0; run < 12; run++ {
		graph := Gia_ManStart(200)
		Gia_ManHashStart(graph)
		var literals []int32
		for i := 0; i < 6; i++ {
			literals = append(literals, Gia_ManAppendCi(graph))
		}
		for i := 0; i < 50; i++ {
			a, b := literals[rng.Intn(len(literals))]^int32(rng.Intn(2)), literals[rng.Intn(len(literals))]^int32(rng.Intn(2))
			var lit int32
			if i%3 == 0 {
				lit = Gia_ManHashXor(graph, a, b)
			} else {
				lit = Gia_ManHashAnd(graph, a, b)
			}
			literals = append(literals, lit)
		}
		for _, lit := range literals {
			if lit >= 2 {
				Gia_ManAppendCo(graph, lit)
				Gia_ManAppendCo(graph, lit^1)
			}
		}
		Gia_ManHashStop(graph)
		data, err := MarshalAiger(graph, AigerWriteOptions{})
		if err != nil {
			t.Fatal(err)
		}
		input := filepath.Join(t.TempDir(), "input.aig")
		if err := os.WriteFile(input, data, 0600); err != nil {
			t.Fatal(err)
		}
		text, err := exec.Command(binary, "-c", fmt.Sprintf("&r -s %s; &sat -c -C 1000 -v", input)).CombinedOutput()
		if err != nil {
			t.Fatalf("original circuit SAT: %v\n%s", err, text)
		}
		solver := Cbs_ManAlloc(graph)
		counts := make(map[int32]int)
		for _, id := range Vec_IntArray(graph.vCos) {
			counts[Cbs_ManSolve(solver, Gia_ObjChild0(graph.pObjs[id]))]++
		}
		for status, label := range map[int32]string{1: "Unsat", 0: "Sat", -1: "Undef"} {
			pattern := regexp.MustCompile(label + `\s+calls\s+(\d+)`)
			match := pattern.FindStringSubmatch(string(text))
			if match == nil {
				t.Fatalf("missing source SAT stats %s\n%s", label, text)
			}
			want, _ := strconv.Atoi(match[1])
			if counts[status] != want {
				t.Fatalf("source circuit SAT counts run%d %s: Go%d C%d\n%s", run, label, counts[status], want, text)
			}
		}
	}
}
