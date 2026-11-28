package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"regexp"
	"testing"
)

func TestFramesNativeReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "frames_review.c")
	native, err := testoracle.Command(t, exe, "--reports").CombinedOutput()
	if err != nil {
		t.Fatalf("C frame reports: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for mode := 0; mode < 5; mode++ {
			fmt.Printf("report %d\n", mode)
			g := framesReviewLiveInitGraph()
			p := Gia_ParFra_t{NFrames: 3, FVerbose: 1}
			var q *Gia_Man_t
			switch mode {
			case 0:
				q = Gia_ManFrames2(g, &p)
			case 4:
				Gia_ManStop(g)
				g = Gia_ManStart(8)
				g.pName = "prefix"
				r := Gia_ManAppendCi(g)
				Gia_ManAppendCo(g, r)
				Gia_ManAppendCo(g, r^1)
				Gia_ManSetRegNum(g, 1)
				q = Gia_ManFramesInitSpecial(g, 4, 1)
			default:
				if mode == 1 {
					p.FInit = 1
				}
				if mode == 2 {
					p.FDisableSt, p.FOrPos = 1, 1
				}
				q = Gia_ManFrames(g, &p)
			}
			Gia_ManStop(q)
			Gia_ManStop(g)
		}
	})
	clock := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	for _, text := range []string{string(native), got} {
		fields := clock.FindAllStringSubmatch(text, -1)
		if len(fields) != 2 {
			t.Fatalf("expected two clock fields, got%d", len(fields))
		}
		for _, field := range fields {
			if len(field[1]) != 9 {
				t.Fatalf("changed timing width%q", field[0])
			}
		}
	}
	// The empty-prefix level average is undefined. C printf can retain a
	// NaN sign while Go prints NaN; compare that IEEE value independently
	// of its platform spelling, retaining the rest of the report exactly.
	nan := regexp.MustCompile(`\((?:-nan|NaN)\)`)
	want := nan.ReplaceAllString(clock.ReplaceAllString(string(native), "<time>"), "(NaN)")
	actual := nan.ReplaceAllString(clock.ReplaceAllString(got, "<time>"), "(NaN)")
	esopReviewCompare(t, []byte(want), actual)
}
