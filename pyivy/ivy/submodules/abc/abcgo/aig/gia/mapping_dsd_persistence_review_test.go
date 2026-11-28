package gia_test

import (
	"bytes"
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	if_ "github.com/glycerine/abcgo/map/if"
	"github.com/glycerine/abcgo/opt/dau"
)

func mappingDSDReviewPersistenceReport(p *if_.If_DsdMan_t) string {
	var result strings.Builder
	fmt.Fprintf(&result, "report %d %d %d %d\n", if_.If_DsdManVarNum(p), if_.If_DsdManLutSize(p), if_.If_DsdManObjNum(p), if_.If_DsdManTtBitNum(p))
	cell := if_.If_DsdManGetCellStr(p)
	if cell == "" {
		cell = "-"
	}
	fmt.Fprintf(&result, "cell %s\n", cell)
	for id := int32(0); id < if_.If_DsdManObjNum(p); id++ {
		if_.If_DsdManPrintOne(&result, p, id, nil, 1)
		if config := if_.If_DsdManGetFuncConfig(p, id<<1); config != nil {
			fmt.Fprintf(&result, "config %d", id)
			for _, word := range config {
				fmt.Fprintf(&result, " %016x", word)
			}
			result.WriteByte('\n')
		}
	}
	return result.String()
}
func mappingDSDReviewReadReport(t *testing.T, executable, path, copyPath string) string {
	t.Helper()
	args := []string{"read", path}
	if copyPath != "" {
		args = append(args, copyPath)
	}
	output, err := testoracle.Command(t, executable, args...).CombinedOutput()
	if err != nil {
		t.Fatalf("C read: %v %s", err, output)
	}
	start := strings.Index(string(output), "\nreport ")
	if start < 0 {
		t.Fatalf("missing C report: %s", output)
	}
	return string(output[start+1:])
}
func TestMappingDSDNativePersistenceInteroperability(t *testing.T) {
	executable := gia.CompileSourceOracleForTest(t, "if_dsd_persistence_review.c")
	for scenario := 0; scenario < 9; scenario++ {
		t.Run(strconv.Itoa(scenario), func(t *testing.T) {
			dir := t.TempDir()
			sourcePath := filepath.Join(dir, "source.dsd")
			output, err := testoracle.Command(t, executable, "write", sourcePath, strconv.Itoa(scenario)).CombinedOutput()
			if err != nil {
				t.Fatalf("C write: %v %s", err, output)
			}
			expected := mappingDSDReviewReadReport(t, executable, sourcePath, "")
			p := if_.If_DsdManLoad(sourcePath)
			if p == nil {
				t.Fatal("Go rejected original C file")
			}
			defer if_.If_DsdManFree(p, 0)
			if actual := mappingDSDReviewPersistenceReport(p); actual != expected {
				t.Fatalf("C->Go state:\n%soriginal C:\n%s", actual, expected)
			}
			goPath, cPath := filepath.Join(dir, "go.dsd"), filepath.Join(dir, "c-copy.dsd")
			if_.If_DsdManSave(p, goPath)
			if actual := mappingDSDReviewReadReport(t, executable, goPath, cPath); actual != expected {
				t.Fatalf("Go->C state:\n%soriginal C:\n%s", actual, expected)
			}
			a, err := testoracle.ReadFile(goPath)
			if err != nil {
				t.Fatal(err)
			}
			b, err := testoracle.ReadFile(cPath)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(a, b) {
				t.Fatal("C did not reserialize the Go file byte-for-byte")
			}
			// A loaded node must still hit its original structural hash entry.
			if if_.If_DsdManVarNum(p) >= 4 {
				before := if_.If_DsdManObjNum(p)
				if_.If_DsdManCompute(p, dau.Dau_DsdToTruth("D3C5{abcd}", 4), 4, make([]byte, 4), "")
				if if_.If_DsdManObjNum(p) != before {
					t.Fatal("loaded interning table lost a native prime structure")
				}
			}
		})
	}
}
