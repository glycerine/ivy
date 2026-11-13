package cnf

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestWriteQuantifiedCNFPlainAndGzip(t *testing.T) {
	data := NewMappedCNF(3, [][]int32{{0, 3}, {4}}, nil, nil, nil)
	all, exists := vec.Vec_IntAlloc(1), vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(all, 0)
	vec.Vec_IntPush(exists, 1)
	vec.Vec_IntPush(exists, 2)
	for _, suffix := range []string{".cnf", ".cnf.gz"} {
		name := filepath.Join(t.TempDir(), "formula"+suffix)
		Cnf_DataWriteIntoFile(data, name, 0, all, exists)
		file, err := os.Open(name)
		if err != nil {
			t.Fatal(err)
		}
		var reader io.Reader = file
		if strings.HasSuffix(suffix, ".gz") {
			z, err := gzip.NewReader(file)
			if err != nil {
				file.Close()
				t.Fatal(err)
			}
			defer z.Close()
			reader = z
		}
		content, err := io.ReadAll(reader)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range []string{"p cnf 3 2\n", "a 1 0\n", "e 2 3 0\n", "1 -2 0\n", "3 0\n"} {
			if !strings.Contains(string(content), line) {
				t.Fatalf("%s missing %q in %q", suffix, line, content)
			}
		}
	}
}

func TestReadRejectsIncorrectClauseCount(t *testing.T) {
	for _, input := range []string{"p cnf 1 2\n1 0\n", "p cnf 1 0\n1 0\n"} {
		name := filepath.Join(t.TempDir(), "invalid.cnf")
		if err := os.WriteFile(name, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
		if Cnf_DataReadFromFile(name) != nil {
			t.Fatalf("accepted inconsistent header: %q", input)
		}
	}
}
