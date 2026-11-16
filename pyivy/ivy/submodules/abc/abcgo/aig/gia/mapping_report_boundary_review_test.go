package gia

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	abcutil "github.com/glycerine/abcgo/misc/util"
)

func TestMappingEmptyDenominatorsRetainSourceNaN(t *testing.T) {
	for _, bridge := range []bool{false, true} {
		t.Run(fmt.Sprintf("bridge%t", bridge), func(t *testing.T) {
			old := abcutil.PrintBridge
			defer func() { abcutil.PrintBridge = old }()
			var text strings.Builder
			abcutil.PrintBridge = nil
			if bridge {
				abcutil.PrintBridge = func(message string) { text.WriteString(message) }
			}
			for _, graph := range []bool{false, true} {
				p := Gia_ManStart(1)
				if graph {
					a, b := Gia_ManAppendCi(p), Gia_ManAppendCi(p)
					root := Gia_ManAppendAnd(p, a, b) >> 1
					p.VMapping = Vec_IntStart(int(p.nObjs))
					Vec_IntWriteEntry(p.VMapping, int(root), int32(Vec_IntSize(p.VMapping)))
					for _, item := range []int32{2, 1, 2, root} {
						Vec_IntPush(p.VMapping, item)
					}
				} else {
					p.VMapping = Vec_IntStart(int(p.nObjs))
				}
				text.Reset()
				output := hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, "") })
				if bridge {
					output = text.String()
				}
				// C divides by the actual CO count, even when there are none.
				// Go's formatter spells the resulting IEEE non-number NaN.
				if !strings.Contains(output, "(NaN)") {
					t.Errorf("graph%t: undefined average became a finite value: %q", graph, output)
				}
				Gia_ManStop(p)
			}
			text.Reset()
			output := hashReviewStdout(t, func() { Gia_ManPrintNodeProfile([]int32{0, 0, 0, 0}, 3) })
			if bridge {
				output = text.String()
			}
			want := "LUT = 0 : 2=0 NaN %  3=0 NaN %  Ave = 0.00\n"
			if output != want {
				t.Errorf("empty profile: got%q want%q", output, want)
			}
		})
	}
}

func TestMappingStatsUnnamedInitialContinuation(t *testing.T) {
	giaMappingStatsState.Lock()
	oldName, oldSince := giaMappingStatsState.name, giaMappingStatsState.since
	giaMappingStatsState.name, giaMappingStatsState.since = "", 0
	giaMappingStatsState.Unlock()
	defer func() {
		giaMappingStatsState.Lock()
		giaMappingStatsState.name, giaMappingStatsState.since = oldName, oldSince
		giaMappingStatsState.Unlock()
	}()
	p := mappingReviewAverageGraph()
	defer Gia_ManStop(p)
	p.SetIdentity("", "")
	file := filepath.Join(t.TempDir(), "unnamed.txt")
	before := abcutil.Abc_Clock()
	hashReviewStdout(t, func() {
		Gia_ManPrintMappingStats(p, file)
		Gia_ManPrintMappingStats(p, file)
	})
	after := abcutil.Abc_Clock()
	giaMappingStatsState.Lock()
	stored := giaMappingStatsState.since
	giaMappingStatsState.Unlock()
	if stored < before || stored > after {
		t.Fatalf("saved clock%d outside [%d,%d]", stored, before, after)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	// Source's initial remembered name is empty; an unnamed first graph
	// follows the continuation branch, rather than inventing a header.
	elapsed := regexp.MustCompile(`[0-9]+\.[0-9]{2}`)
	if got := elapsed.ReplaceAllString(string(data), "<elapsed>"); got != " 2 2 <elapsed> 2 2 <elapsed>" {
		t.Fatalf("unnamed append: %q", got)
	}
	first := elapsed.FindString(string(data))
	seconds, err := strconv.ParseFloat(first, 64)
	if err != nil {
		t.Fatal(err)
	}
	// The initial source clock is zero, so its first continuation measures
	// from the common clock origin, not from a fabricated wall-time value.
	unit := float64(abcutil.CLOCKS_PER_SEC)
	if seconds < float64(before)/unit-0.0051 || seconds > float64(after)/unit+0.0051 {
		t.Fatalf("initial elapsed%f outside clock interval [%f,%f]", seconds, float64(before)/unit, float64(after)/unit)
	}
}

func TestMappingStatsOpenFailurePreservesHistory(t *testing.T) {
	giaMappingStatsState.Lock()
	oldName, oldSince := giaMappingStatsState.name, giaMappingStatsState.since
	giaMappingStatsState.name, giaMappingStatsState.since = "previous", 123
	giaMappingStatsState.Unlock()
	defer func() {
		giaMappingStatsState.Lock()
		giaMappingStatsState.name, giaMappingStatsState.since = oldName, oldSince
		giaMappingStatsState.Unlock()
	}()
	p := mappingReviewAverageGraph()
	defer Gia_ManStop(p)
	p.SetIdentity("retry", "")
	dir := t.TempDir()
	stderr, err := os.Create(filepath.Join(dir, "stderr.txt"))
	if err != nil {
		t.Fatal(err)
	}
	oldStderr := os.Stderr
	os.Stderr = stderr
	defer func() { os.Stderr = oldStderr; stderr.Close() }()
	// A nonexistent parent fails even when the tests run as root.
	badPath := filepath.Join(dir, "missing", "stats.txt")
	output := hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, badPath) })
	diagnostic, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output, "Mapping (K=2)") || !strings.Contains(string(diagnostic), "gia: cannot open mapping statistics:") || !strings.Contains(string(diagnostic), badPath) {
		t.Fatalf("stdout%q stderr%q", output, diagnostic)
	}
	giaMappingStatsState.Lock()
	name, since := giaMappingStatsState.name, giaMappingStatsState.since
	giaMappingStatsState.Unlock()
	if name != "previous" || since != 123 {
		t.Fatal("failed open changed report history")
	}
	path := filepath.Join(dir, "retry.txt")
	hashReviewStdout(t, func() { Gia_ManPrintMappingStats(p, path) })
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "\nretry  2 2 2 " {
		t.Fatalf("successful retry: %q", data)
	}
}

func TestMappingNodeProfileAcceptsEmptySentinel(t *testing.T) {
	for i, counts := range [][]int32{nil, {}, {9, 7, 5}} {
		for _, sizeMax := range []int32{-1, -2, -1 << 31} {
			for _, bridge := range []bool{false, true} {
				t.Run(fmt.Sprintf("counts%d/max%d/bridge%t", i, sizeMax, bridge), func(t *testing.T) {
					old := abcutil.PrintBridge
					defer func() { abcutil.PrintBridge = old }()
					var text strings.Builder
					abcutil.PrintBridge = nil
					if bridge {
						abcutil.PrintBridge = func(message string) { text.WriteString(message) }
					}
					output := hashReviewStdout(t, func() { Gia_ManPrintNodeProfile(counts, sizeMax) })
					if bridge {
						output = text.String()
					}
					if want := "LUT = 0 : Ave = 0.00\n"; output != want {
						t.Fatalf("got%q want%q", output, want)
					}
				})
			}
		}
	}
}
