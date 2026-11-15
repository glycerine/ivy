package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutReportsMatchOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "cut_enumeration_review.c")
	for _, mode := range []string{"reports", "driver"} {
		t.Run(mode, func(t *testing.T) {
			native, err := exec.Command(exe, mode).CombinedOutput()
			if err != nil {
				t.Fatalf("native cut report: %v\n%s", err, native)
			}
			g := cutEnumReviewGraph(0)
			got := hashReviewStdout(t, func() {
				if mode == "driver" {
					Gia_StoComputeCuts(g)
					return
				}
				p := Gia_StoAlloc(g, 4, 10, 1, 1, 0)
				defer Gia_StoFree(p)
				cut := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 12}, Cost: 7, CostLev: 9, nTreeLeaves: 1}
				Gia_CutPrint(p, 3, &cut)
				Gia_CutPrint(p, 3, nil)
			})
			want := string(native)
			if mode == "driver" {
				timing := regexp.MustCompile(`Time = +[0-9]+\.[0-9]{2} sec\n$`)
				if !timing.MatchString(got) {
					t.Fatalf("missing source-format elapsed report: %q", got)
				}
				got, want = timing.ReplaceAllString(got, "Time = <elapsed>\n"), timing.ReplaceAllString(want, "Time = <elapsed>\n")
			}
			if got != want {
				t.Fatalf("cut report differs\nGo:%s\nC:%s", got, want)
			}
		})
	}
}

func BenchmarkCutTruthScratch(b *testing.B) {
	for _, size := range []int32{4, 8} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			g := Gia_ManStart(16)
			for i := 0; i < 8; i++ {
				Gia_ManAppendCi(g)
			}
			p := Gia_StoAlloc(g, size, 10, 1, 1, 0)
			defer Gia_StoFree(p)
			words := 1 << uint(max(0, size-6))
			left, right := make([]uint64, words), make([]uint64, words)
			for i := range left {
				left[i] = 0x8888888888888888
				right[i] = 0x6666666666666666
			}
			a := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{1, 3}, iFunc: int32(vec.Vec_MemHashInsert(p.vTtMem, left) << 1)}
			c := Gia_Cut_t{nLeaves: 2, pLeaves: [8]int32{2, 3}, iFunc: int32(vec.Vec_MemHashInsert(p.vTtMem, right) << 1)}
			initial := Gia_Cut_t{nLeaves: 3, pLeaves: [8]int32{1, 2, 3}}
			out := initial
			Gia_CutComputeTruth(p, &a, &c, 0, 0, &out, 0)
			if out.nLeaves != 3 {
				b.Fatal("benchmark minimized relevant variables")
			}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out = initial
				Gia_CutComputeTruth(p, &a, &c, 0, 0, &out, 0)
			}
		})
	}
}

// The driver-only report parser is also used to check finite statistics for
// supported empty interfaces without requiring C's undefined zero divisions.
func cutEnumReportLine(report, prefix string) string {
	for _, line := range strings.Split(report, "\n") {
		if strings.HasPrefix(line, prefix) {
			return line
		}
	}
	return ""
}
