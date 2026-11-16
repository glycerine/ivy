package if_

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/dau"
)

func dsdReportCapture(t *testing.T, run func()) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = f
	defer func() { os.Stdout = old; f.Close() }()
	run()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// These expectations are transcribed from the source report contract, without
// compiling C or deriving the expected rows from the Go reporting routines.
func TestDsdReportOptionsAndRouting(t *testing.T) {
	p := If_DsdManAlloc(4, 0)
	defer If_DsdManFree(p, 0)
	If_DsdManCompute(p, dau.Dau_DsdToTruth("(abc)", 3), 3, make([]byte, 3), "")
	for i := 0; i < 10; i++ {
		If_DsdManCompute(p, dau.Dau_DsdToTruth("[ab]", 2), 2, make([]byte, 2), "")
	}
	if If_DsdManObjNum(p) != 4 {
		t.Fatal("unexpected setup node count")
	}
	// Each tuple is objects, primes, NPN classes, structures, nonDSD, marked.
	rows := [][6]int{{2, 0, 0, 1, 0, 0}, {0, 0, 0, 1, 0, 0}, {1, 0, 0, 1, 0, 0}, {1, 0, 0, 1, 0, 0}, {0, 0, 0, 0, 0, 0}, {4, 0, 0, 4, 0, 0}}
	var distrib strings.Builder
	distrib.WriteString("***** DSD MANAGER STATISTICS *****\nSupport     Obj   ObjNDSD            NPNNDSD                  Str   StrNDSD             Marked  \n")
	for i, r := range rows {
		if i == 5 {
			distrib.WriteString("All : ")
		} else {
			fmt.Fprintf(&distrib, "%3d : ", i)
		}
		fmt.Fprintf(&distrib, "%9d %9d   0.00 %% %9d   0.00 %%   %9d %9d   0.00 %% %9d   0.00 %%\n", r[0], r[1], r[2], r[3], r[4], r[5])
	}
	distrib.WriteString("Number of inputs = 4.  LUT size = 0.  Marks = no.  NewAsUseless = no.  Bookmark = 0.\n")
	stats := "Unique table misses        =        2\nUnique table hits          =        9\n"
	for _, label := range []string{"objects   ", "functions ", "hash table", "bound sets", "array     "} {
		stats += "Memory used for " + label + " = <storage> MB.\n"
	}
	occurrences := "The distribution of object occurrences:\n              1 :            1   ( 25.00 %)\n        10 - 19 :            1   ( 50.00 %)\nFanins: Max = 10. Ave = 2.75.\n"
	memory := regexp.MustCompile(`(?m)^(Memory used for [^=]+=) +[0-9.]+ MB\.$`)
	// The four record strings specify the source's exact ID/count/text format.
	records := []string{"     0 :  0        0 0  0\n", "     1 :  1        0 0  a\n", "     2 :  3        1 0  (abc)\n", "     3 :  2       10 0  [ab]\n"}
	for _, verbose := range []int32{0, 1} {
		for _, number := range []int32{0, 2} {
			for _, support := range []int32{0, 3} {
				for _, occurs := range []int32{0, 1} {
					for _, fileMode := range []bool{false, true} {
						t.Run(fmt.Sprintf("verbose%d/number%d/support%d/occurs%d/file%t", verbose, number, support, occurs, fileMode), func(t *testing.T) {
							target := ""
							if fileMode {
								target = filepath.Join(t.TempDir(), "report.txt")
							}
							stdout := dsdReportCapture(t, func() { If_DsdManPrint(p, target, number, support, occurs, 0, verbose) })
							wantOut, wantFile := "", ""
							prefix := ""
							if verbose != 0 {
								prefix = ifDsdReportNotations
							}
							if fileMode {
								wantFile = prefix + stats
								wantOut = distrib.String()
							} else {
								wantOut = prefix + distrib.String() + stats
							}
							if occurs != 0 {
								wantOut += occurrences
							}
							if verbose != 0 {
								order := []int{3, 2, 1, 0}
								if number == 2 {
									order = []int{2, 0}
								}
								if support == 3 {
									order = []int{2}
								}
								for rank, id := range order {
									wantOut += fmt.Sprintf("%2d : ", rank+1)
									if fileMode {
										wantFile += records[id]
									} else {
										wantOut += records[id]
									}
								}
							}
							if got := memory.ReplaceAllString(stdout, "${1} <storage> MB."); got != wantOut {
								t.Fatalf("stdout:\n%s\nwant:\n%s", got, wantOut)
							}
							if fileMode {
								b, err := os.ReadFile(target)
								if err != nil {
									t.Fatal(err)
								}
								if got := memory.ReplaceAllString(string(b), "${1} <storage> MB."); got != wantFile {
									t.Fatalf("file:\n%s\nwant:\n%s", got, wantFile)
								}
							}
						})
					}
				}
			}
		}
	}
}

func TestDsdReportNonDecomposableDescendants(t *testing.T) {
	p := If_DsdManAlloc(4, 0)
	defer If_DsdManFree(p, 0)
	root := If_DsdManCompute(p, dau.Dau_DsdToTruth("(aE8{bcd})", 4), 4, make([]byte, 4), "")
	if If_DsdManCheckNonDec_rec(p, root>>1) != 1 {
		t.Fatal("prime descendant was not counted")
	}
	for _, id := range []int32{0, 1} {
		if If_DsdManCheckNonDec_rec(p, id) != 0 {
			t.Fatal("base classified as prime")
		}
	}
	output := dsdReportCapture(t, func() { If_DsdManPrintDistrib(p) })
	if !strings.Contains(output, "All :         4         1  25.00 %         1  25.00 %           4         2  50.00 %         0   0.00 %\n") {
		t.Fatalf("prime vs containing-structure totals:\n%s", output)
	}
}

func TestDsdReportOccurrenceDecades(t *testing.T) {
	p := If_DsdManAlloc(3, 0)
	defer If_DsdManFree(p, 0)
	// Isolate histogram boundaries from canonicalization and reference updates.
	for _, count := range []int32{0, 9, 10, 19, 20, 99, 100, 999, 1000, 9999, 10000, 99999, 100000, 262143} {
		vec.Vec_PtrPush(&p.vObjs, &If_DsdObj_t{Count: count})
	}
	var out strings.Builder
	If_DsdManPrintOccurs(&out, p)
	for _, row := range []struct {
		label string
		count int
	}{
		{"0", 1}, {"9", 1}, {"10 - 19", 2}, {"20 - 29", 1}, {"90 - 99", 1}, {"100 - 199", 1}, {"900 - 999", 1}, {"1000 - 1999", 1}, {"9000 - 9999", 1}, {"10000 - 19999", 1}, {"90000 - 99999", 1}, {"100000 - 199999", 1}, {"200000 - 299999", 1},
	} {
		if !strings.Contains(out.String(), fmt.Sprintf("%15s : %12d", row.label, row.count)) {
			t.Fatalf("missing %v:\n%s", row, out.String())
		}
	}
}

func TestDsdReportOpenFailure(t *testing.T) {
	p := If_DsdManAlloc(2, 0)
	defer If_DsdManFree(p, 0)
	got := dsdReportCapture(t, func() { If_DsdManPrint(p, filepath.Join(t.TempDir(), "missing", "report"), 0, 0, 0, 0, 1) })
	if got != "cannot open output file\n" {
		t.Fatalf("open failure: %q", got)
	}
}

func TestDsdReportZeroDomainTotals(t *testing.T) {
	p := If_DsdManAlloc(0, 0)
	defer If_DsdManFree(p, 0)
	output := dsdReportCapture(t, func() { If_DsdManPrintDistrib(p) })
	want := "All :         2         0   0.00 %         0   0.00 %           2         0   0.00 %         0   0.00 %\n"
	if !strings.HasSuffix(output, want) {
		t.Fatalf("reserved objects counted incorrectly:\n%s", output)
	}
}
