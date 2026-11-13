package acec

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

// Set ABC_BINARY to an ABC executable built from this checkout to replay the
// checked-in results. The same graph generators drive the ordinary Go tests.
func TestABCDifferential(t *testing.T) {
	abc := os.Getenv("ABC_BINARY")
	if abc != "" {
		var err error
		abc, err = filepath.Abs(abc)
		if err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile("test_vectors/abc_acec_results.txt")
	if err != nil {
		t.Fatal(err)
	}
	results := map[string]string{}
	for _, line := range strings.Split(string(data), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			results[k] = v
		}
	}
	cases := []struct {
		name        string
		count, base int
		graph       func(int) *Gia_Man_t
	}{
		{"mutation", 4, -1, integrationMutatedTree},
		{"input_polarity", 5, -1, integrationPerturbedAdder},
		{"signed_chain", 64, 0, acecSignedChain},
		{"output_inversion", 3, -1, func(i int) *Gia_Man_t { return integrationAdderTree(i != -1, i) }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			expected := results[c.name]
			if len(expected) != c.count {
				t.Fatalf("invalid vector length %d", len(expected))
			}
			dir := t.TempDir()
			base := filepath.Join(dir, "base.aig")
			if err := gia.Gia_AigerWrite(c.graph(c.base), base, 0, 0, 0); err != nil {
				t.Fatal(err)
			}
			for i := 0; i < c.count; i++ {
				var params Acec_ParCec_t
				Acec_ManCecSetDefaultParams(&params)
				status := Acec_Solve(c.graph(c.base), c.graph(i), &params)
				if status != int32(expected[i]-'0') {
					t.Fatalf("case %d: Go returned %d, recorded ABC returned %c", i, status, expected[i])
				}
				if abc == "" {
					continue
				}
				path := filepath.Join(dir, fmt.Sprintf("case%d.aig", i))
				if err := gia.Gia_AigerWrite(c.graph(i), path, 0, 0, 0); err != nil {
					t.Fatal(err)
				}
				out, err := exec.Command(abc, "-c", fmt.Sprintf("&acec %s %s", base, path)).CombinedOutput()
				if err != nil {
					t.Fatalf("case %d: %v\n%s", i, err, out)
				}
				got := byte('?')
				if strings.Contains(string(out), "Networks are equivalent.") {
					got = '1'
				} else if strings.Contains(string(out), "Networks are NOT EQUIVALENT.") {
					got = '0'
				}
				if got != expected[i] {
					t.Fatalf("case %d: ABC returned %c, expected %c\n%s", i, got, expected[i], out)
				}
			}
		})
	}
}
