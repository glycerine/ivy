package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func falseReviewGraph(seed uint32) (*Gia_Man_t, int32) {
	g := Gia_ManStart(128)
	rng := seed
	random := func() uint32 { rng = rng*1664525 + 1013904223; return rng >> 8 }
	var pool []int32
	for i := 0; i < 4; i++ {
		pool = append(pool, Gia_ManAppendCi(g))
	}
	for i := 0; i < 12; i++ {
		a, b := int(random())%len(pool), int(random())%len(pool)
		if a == b {
			b = (b + 1) % len(pool)
		}
		left, right := pool[a]^int32(random()&1), pool[b]^int32(random()&1)
		pool = append(pool, Gia_ManAppendAnd(g, left, right))
	}
	root := pool[len(pool)-1] >> 1
	Gia_ManAppendCo(g, pool[len(pool)-1])
	Gia_ManAppendCo(g, pool[len(pool)-2]^1)
	return g, root
}

func TestFalsePathNativeSingleRebuild(t *testing.T) {
	exe := compileGiaSourceOracle(t, "false_review.c")
	native, err := exec.Command(exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C false paths: %v\n%s", err, native)
	}
	var records strings.Builder
	for seed := uint32(0); seed < 64; seed++ {
		g, root := falseReviewGraph(seed)
		Gia_ManLevelNum(g)
		path := Gia_ManCollectPath(g, g.pObjs[root])
		fmt.Fprintf(&records, "path %d", seed)
		for _, id := range Vec_IntArray(path) {
			fmt.Fprintf(&records, " %d", id)
		}
		records.WriteByte('\n')
		out := Gia_ManCheckOne(g, 0, root, 0, 0, 0)
		count := int32(-1)
		if out != nil {
			count = out.nObjs
			cofReviewCheckRebuild(t, g, out)
		}
		fmt.Fprintf(&records, "result %d %d", seed, count)
		if out != nil {
			for bits := 0; bits < 16; bits++ {
				values := cofReviewValues(out, bits)
				packed := 0
				for i := int32(0); i < Gia_ManCoNum(out); i++ {
					if values[Gia_ManCo(out, i).id] {
						packed |= 1 << i
					}
				}
				fmt.Fprintf(&records, " %d", packed)
			}
			Gia_ManStop(out)
		}
		records.WriteByte('\n')
		Gia_ManStop(g)
	}
	esopReviewCompare(t, native, records.String())
}

func TestFalsePathNativeBatchHooksAndScratch(t *testing.T) {
	exe := compileGiaSourceOracle(t, "false_review.c")
	native, err := exec.Command(exe, "--batch").CombinedOutput()
	if err != nil {
		t.Fatalf("C batch hooks: %v\n%s", err, native)
	}
	var records strings.Builder
	for seed := uint32(0); seed < 64; seed++ {
		g, _ := falseReviewGraph(seed)
		Gia_ManLevelNum(g)
		hooks := Vec_WecStart(int(g.nObjs))
		for i, obj := range g.pObjs {
			obj.Value = uint32(71 + 17*i)
		}
		for i := int32(0); i < Gia_ManCoNum(g); i++ {
			Gia_ManCheckFalseOne(g, i, 0, hooks, 0, 0)
		}
		fmt.Fprintf(&records, "batch %d", seed)
		for i := 0; i < Vec_WecSize(hooks); i++ {
			path := Vec_WecEntry(hooks, i)
			if Vec_IntSize(path) == 0 {
				continue
			}
			fmt.Fprintf(&records, " [%d", i)
			for _, id := range Vec_IntArray(path) {
				fmt.Fprintf(&records, " %d", id)
			}
			records.WriteByte(']')
		}
		records.WriteByte('\n')
		records.WriteString("scratch")
		for _, obj := range g.pObjs {
			fmt.Fprintf(&records, " %d", obj.Value)
		}
		records.WriteByte('\n')
		Gia_ManStop(g)
	}
	esopReviewCompare(t, native, records.String())
}

func falseReviewTiedOutputs() *Gia_Man_t {
	g := Gia_ManStart(32)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	x := Gia_ManAppendAnd(g, a, b)
	root := Gia_ManAppendAnd(g, x, a^1)
	for i := int32(0); i < 5; i++ {
		Gia_ManAppendCo(g, root^(i&1))
	}
	return g
}
func TestFalsePathNativePriorityTies(t *testing.T) {
	exe := compileGiaSourceOracle(t, "false_review.c")
	native, err := exec.Command(exe, "--priority").CombinedOutput()
	if err != nil {
		t.Fatalf("C priority: %v\n%s", err, native)
	}
	g := falseReviewTiedOutputs()
	defer Gia_ManStop(g)
	got := hashReviewStdout(t, func() { out := Gia_ManCheckFalse2(g, 0, 0, 1, 0); cofReviewCheckRebuild(t, g, out); Gia_ManStop(out) })
	headers := regexp.MustCompile(`(?m)^PO +(-?[0-9]+) :`)
	wantPOs, gotPOs := headers.FindAllStringSubmatch(string(native), -1), headers.FindAllStringSubmatch(got, -1)
	if len(wantPOs) != 5 || len(gotPOs) != 5 {
		t.Fatalf("missing five queries C%v Go%v", wantPOs, gotPOs)
	}
	for i := range wantPOs {
		if gotPOs[i][1] != wantPOs[i][1] {
			t.Fatalf("query%d Go PO%s C PO%s\nGo:\n%s\nC:\n%s", i, gotPOs[i][1], wantPOs[i][1], got, native)
		}
	}
}

func TestFalsePathNativeVerboseReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "false_review.c")
	native, err := exec.Command(exe, "--verbose").CombinedOutput()
	if err != nil {
		t.Fatalf("C reports: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for verbose := int32(0); verbose < 2; verbose++ {
			for very := int32(0); very < 2; very++ {
				for mode := 0; mode < 3; mode++ {
					g := Gia_ManStart(32)
					a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
					x := Gia_ManAppendAnd(g, a, b)
					root := x
					if mode != 2 {
						root = Gia_ManAppendAnd(g, x, a^1)
					}
					count := 1
					if mode == 1 {
						count = 5
					}
					for i := 0; i < count; i++ {
						Gia_ManAppendCo(g, root^int32(i&1))
					}
					Gia_ManLevelNum(g)
					fmt.Printf("mode %d %d %d\n", mode, verbose, very)
					var out *Gia_Man_t
					if mode == 1 {
						out = Gia_ManCheckFalse2(g, 0, 0, verbose, very)
					} else {
						out = Gia_ManCheckOne(g, 0, root>>1, 0, verbose, very)
					}
					if out != nil {
						cofReviewCheckRebuild(t, g, out)
						Gia_ManStop(out)
					}
					Gia_ManStop(g)
				}
			}
		}
	})
	clock := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	for _, text := range []string{string(native), got} {
		for _, field := range clock.FindAllStringSubmatch(text, -1) {
			if len(field[1]) != 9 {
				t.Fatalf("changed source time-column width: %q", field[0])
			}
		}
	}
	want := clock.ReplaceAllString(string(native), "<time>")
	got = clock.ReplaceAllString(got, "<time>")
	if got != want {
		t.Fatalf("reports differ\nGo:\n%s\nC:\n%s", got, want)
	}
}
