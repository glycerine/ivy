package super

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func originalABC(t *testing.T) string {
	t.Helper()
	path, err := testoracle.NativeABCAbs()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = testoracle.Stat(path); err != nil {
		t.Skip("original ABC executable unavailable")
	}
	return path
}
func TestANDEnumerationAgainstOriginalABC(t *testing.T) {
	binary := originalABC(t)
	for _, item := range []struct{ inputs, levels int }{{2, 1}, {2, 2}, {3, 2}} {
		t.Run(strconv.Itoa(item.inputs)+"x"+strconv.Itoa(item.levels), func(t *testing.T) {
			native, err := GenerateAND(item.inputs, item.levels)
			if err != nil {
				t.Fatal(err)
			}
			dir := t.TempDir()
			command := testoracle.Command(t, binary, "-q", "super2 -I "+strconv.Itoa(item.inputs)+" -L "+strconv.Itoa(item.levels))
			command.Dir = dir
			if log, err := command.CombinedOutput(); err != nil {
				t.Fatalf("original super2: %v\n%s", err, log)
			}
			source, err := testoracle.ReadFile(filepath.Join(dir, "superI"+strconv.Itoa(item.inputs)+"L"+strconv.Itoa(item.levels)))
			if err != nil {
				t.Fatal(err)
			}
			var reference []uint32
			first := true
			for _, line := range strings.Split(string(source), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if first {
					first = false
					continue
				}
				field := strings.Fields(line)[0]
				truth, err := strconv.ParseUint(field, 2, 32)
				if err != nil {
					t.Fatal(err)
				}
				reference = append(reference, uint32(truth))
			}
			values := make([]uint32, len(native.Gates))
			for i, g := range native.Gates {
				values[i] = native.canonical(g.Truth)
			}
			sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
			if !reflect.DeepEqual(values, reference) {
				t.Fatalf("AND native=%v source=%v", values, reference)
			}
		})
	}
}
