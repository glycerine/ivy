package mf

import (
	"bufio"
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func originalABCCNF(t *testing.T, model *ioa.Model) *CNF {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	abc := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../abc"))
	if _, err := os.Stat(abc); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	dir := t.TempDir()
	input, output := filepath.Join(dir, "input.aig"), filepath.Join(dir, "output.cnf")
	f, err := os.Create(input)
	if err != nil {
		t.Fatal(err)
	}
	err = ioa.Write(f, model, ioa.WriteOptions{})
	closeErr := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	command := fmt.Sprintf("&r -s %s; &write_cnf -K 8 %s", input, output)
	if text, err := exec.Command(abc, "-c", command).CombinedOutput(); err != nil {
		t.Fatalf("ABC differential oracle: %v\n%s", err, text)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	result := &CNF{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	var clause []int32
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "c") || strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "p") {
			var count int
			if _, err := fmt.Sscanf(line, "p cnf %d %d", &result.Variables, &count); err != nil {
				t.Fatal(err)
			}
			continue
		}
		r := strings.NewReader(line)
		for {
			var lit int32
			if _, err := fmt.Fscan(r, &lit); err != nil {
				break
			}
			if lit == 0 {
				result.Clauses = append(result.Clauses, clause)
				clause = nil
				continue
			}
			if lit > 0 {
				clause = append(clause, 2*(lit-1))
			} else {
				clause = append(clause, 2*(-lit-1)+1)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	return result
}
func randomizedCNFModel(seed int64, inputs, gates int) *ioa.Model {
	rng := rand.New(rand.NewSource(seed))
	m := &ioa.Model{}
	for i := 0; i < inputs; i++ {
		m.Inputs = append(m.Inputs, uint32(2*(i+1)))
	}
	for i := 0; i < gates; i++ {
		id := uint32(inputs + i + 1)
		a, b := uint32(rng.Intn(int(id-1))+1)<<1, uint32(rng.Intn(int(id-1))+1)<<1
		for a == b {
			b = uint32(rng.Intn(int(id-1))+1) << 1
		}
		a |= uint32(rng.Intn(2))
		b |= uint32(rng.Intn(2))
		if a < b {
			a, b = b, a
		}
		m.ANDs = append(m.ANDs, ioa.And{Literal: id << 1, Left: a, Right: b})
	}
	m.MaxVar = uint32(inputs + gates)
	m.Outputs = []uint32{m.MaxVar << 1, uint32(inputs+gates-1)<<1 | 1, uint32(inputs+gates/2) << 1}
	return m
}
func TestGenerateCNFOriginalABCExactDifferential(t *testing.T) {
	for _, test := range []struct {
		seed          int64
		inputs, gates int
	}{{91, 4, 10}, {92, 10, 45}, {93, 12, 70}, {94, 16, 110}, {95, 20, 150}, {96, 9, 80}} {
		t.Run(fmt.Sprintf("seed%d", test.seed), func(t *testing.T) {
			m := randomizedCNFModel(test.seed, test.inputs, test.gates)
			want := originalABCCNF(t, m)
			got, err := GenerateModelCNF(m)
			if err != nil {
				t.Fatal(err)
			}
			if got.Variables != want.Variables || !reflect.DeepEqual(got.Clauses, want.Clauses) {
				t.Fatalf("source differential seed%d: vars%d want%d clauses%v want%v", test.seed, got.Variables, want.Variables, got.Clauses, want.Clauses)
			}
		})
	}
}

func reconvergentCNFModel(inputs int) *ioa.Model {
	m := &ioa.Model{}
	for i := 0; i < inputs; i++ {
		m.Inputs = append(m.Inputs, uint32(2*(i+1)))
	}
	and := func(a, b uint32) uint32 {
		if a == b {
			return a
		}
		if a == b^1 {
			return 0
		}
		id := uint32(inputs + len(m.ANDs) + 1)
		if a < b {
			a, b = b, a
		}
		m.ANDs = append(m.ANDs, ioa.And{Literal: id << 1, Left: a, Right: b})
		return id << 1
	}
	xor := func(a, b uint32) uint32 { l, r := and(a, b^1), and(a^1, b); return and(l^1, r^1) ^ 1 }
	mux := func(c, a, b uint32) uint32 { l, r := and(c, a), and(c^1, b); return and(l^1, r^1) ^ 1 }
	accumulator := m.Inputs[0]
	var checkpoints []uint32
	for i := 1; i < inputs; i++ {
		switch i % 3 {
		case 0:
			accumulator = xor(accumulator, m.Inputs[i])
		case 1:
			accumulator = and(accumulator, m.Inputs[i])
		case 2:
			accumulator = mux(m.Inputs[i], accumulator, m.Inputs[i-1]^1)
		}
		if i%4 == 0 {
			checkpoints = append(checkpoints, accumulator)
		}
	}
	for i, branch := range checkpoints {
		accumulator = xor(accumulator, mux(m.Inputs[i], branch, m.Inputs[inputs-i-1]))
	}
	m.Outputs = append([]uint32{accumulator}, checkpoints...)
	m.MaxVar = uint32(inputs + len(m.ANDs))
	return m
}
func TestGenerateCNFReconvergentWideOriginalABC(t *testing.T) {
	for _, inputs := range []int{10, 16, 24, 32} {
		t.Run(fmt.Sprintf("inputs%d", inputs), func(t *testing.T) {
			m := reconvergentCNFModel(inputs)
			want := originalABCCNF(t, m)
			got, err := GenerateModelCNF(m)
			if err != nil {
				t.Fatal(err)
			}
			if got.Variables != want.Variables || !reflect.DeepEqual(got.Clauses, want.Clauses) {
				t.Fatalf("original ABC pipeline differs: inputs%d vars%d/%d clauses%d/%d", inputs, got.Variables, want.Variables, len(got.Clauses), len(want.Clauses))
			}
		})
	}
}

func TestGenerateCNFRandomStressOriginalABC(t *testing.T) {
	for seed := int64(200); seed < 224; seed++ {
		t.Run(fmt.Sprintf("seed%d", seed), func(t *testing.T) {
			m := randomizedCNFModel(seed, 10+int(seed%15), 90+int(seed%8)*23)
			want := originalABCCNF(t, m)
			got, err := GenerateModelCNF(m)
			if err != nil {
				t.Fatal(err)
			}
			if got.Variables != want.Variables || !reflect.DeepEqual(got.Clauses, want.Clauses) {
				t.Fatalf("original ABC pipeline differs seed%d vars%d/%d clauses%d/%d", seed, got.Variables, want.Variables, len(got.Clauses), len(want.Clauses))
			}
		})
	}
}
