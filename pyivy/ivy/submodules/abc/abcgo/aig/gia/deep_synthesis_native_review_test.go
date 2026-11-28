package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"strings"
	"testing"

	abcmain "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/misc/util"
)

func deepReviewGraph(count int32) *Gia_Man_t {
	g := Gia_ManStart(count + 8)
	a, b := Gia_ManAppendCi(g), Gia_ManAppendCi(g)
	for i := int32(0); i < count; i++ {
		a = Gia_ManAppendAnd(g, a, b)
	}
	Gia_ManAppendCo(g, a)
	return g
}

type deepReviewDispatcher struct {
	keep                                bool
	panicAt                             int
	after                               func()
	scenario, calls, roundCalls, rounds int
	last                                *Gia_Man_t
	trace                               bool
	failAt                              int
	remove                              bool
}

func (d *deepReviewDispatcher) Execute(f *abcmain.Frame, command string) int32 {
	current := abcmain.Abc_FrameReadGia(f).(*Gia_Man_t)
	if current != d.last {
		d.roundCalls = 0
		d.rounds++
		if d.trace {
			fmt.Printf("reset round%d input%d\n", d.rounds, Gia_ManAndNum(current))
		}
	}
	i := d.roundCalls
	d.roundCalls++
	d.calls++
	if d.trace {
		fmt.Printf("cmd round%d call%d batch%d input%d %s\n", d.rounds, i, abcmain.Abc_FrameIsBatchMode(), Gia_ManAndNum(current), command)
	}
	if d.panicAt > 0 && d.calls == d.panicAt {
		panic("scripted synthesis panic")
	}
	if d.after != nil {
		d.after()
	}
	if d.failAt > 0 && d.calls == d.failAt {
		return 1
	}
	if d.remove {
		f.SetResource("pGia", nil)
		return 0
	}
	if d.keep {
		d.last = current
		return 0
	}
	count := int32(6)
	switch d.scenario {
	case 1:
		count = 5
	case 2:
		count = max(1, 6-int32(i+1)/3)
	case 3:
		script := []int32{7, 6, 5, 6, 4, 5, 4, 3, 2}
		count = 3
		if i < len(script) {
			count = script[i]
		}
	}
	next := deepReviewGraph(count)
	f.SetResource("pGia", next)
	d.last = next
	return 0
}

func TestDeepSynthesisControlMatchesOriginalABC(t *testing.T) {
	exe := compileGiaSourceOracle(t, "deep_synthesis_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("native deep synthesis: %v\n%s", err, native)
	}
	old := abcmain.Abc_FrameReadGlobalFrame()
	defer abcmain.Abc_FrameSetGlobalFrame(old)
	got := hashReviewStdout(t, func() {
		for outer := 0; outer < 2; outer++ {
			for scenario := 0; scenario < 4; scenario++ {
				for _, seed := range []int32{0, 3, 11} {
					for two := int32(0); two < 2; two++ {
						for initial := int32(0); initial < 2; initial++ {
							f := abcmain.NewFrame()
							abcmain.Abc_FrameSetGlobalFrame(f)
							graph := deepReviewGraph(6)
							f.SetResource("pGia", graph)
							abcmain.Abc_FrameSetBatchMode(initial)
							d := &deepReviewDispatcher{scenario: scenario, last: graph, trace: true}
							f.Dispatcher = d
							fmt.Printf("case %d %d %d %d %d goal0\n", outer, scenario, seed, two, initial)
							var result *Gia_Man_t
							if outer == 0 {
								result = Gia_ManDeepSynOne(3, 0, 0, seed, two, 0)
							} else {
								result = Gia_ManDeepSyn(graph, 3, 3, 0, 0, seed, two, 0)
							}
							if result == nil {
								t.Fatal("synthesis unexpectedly failed")
							}
							current := abcmain.Abc_FrameReadGia(f).(*Gia_Man_t)
							expectedCalls := []int{4, 5, 19, 13}[scenario]
							if outer != 0 {
								expectedCalls *= 3
							}
							if d.calls != expectedCalls || Gia_ManAndNum(result) != []int32{6, 5, 1, 2}[scenario] || Gia_ManAndNum(current) != []int32{6, 5, 1, 3}[scenario] || abcmain.Abc_FrameIsBatchMode() != initial {
								t.Fatal("independent driver stopping/selection contract failed")
							}
							if result == current {
								t.Fatal("returned best graph aliases frame graph")
							}
							for mask := 0; mask < 4; mask++ {
								values, e := result.Evaluate([]bool{mask&1 != 0, mask&2 != 0})
								if e != nil || values[0] != (mask == 3) {
									t.Fatal("synthesis changed scripted conjunction")
								}
							}
							fmt.Printf("result calls%d best%d current%d batch%d next%d\n", d.calls, Gia_ManAndNum(result), Gia_ManAndNum(current), abcmain.Abc_FrameIsBatchMode(), util.Abc_Random(0))
							Gia_ManStop(result)
							if current.nObjs == 0 {
								t.Fatal("disposing result destroyed frame graph")
							}
							f.Close()
						}
					}
				}
			}
		}
	})
	filter := func(report string) string {
		var b strings.Builder
		for _, line := range strings.Split(report, "\n") {
			if strings.HasPrefix(line, "case ") || strings.HasPrefix(line, "cmd ") || strings.HasPrefix(line, "reset ") || strings.HasPrefix(line, "result ") {
				b.WriteString(line)
				b.WriteByte('\n')
			}
		}
		return b.String()
	}
	actual, want := filter(got), filter(string(native))
	if actual != want {
		a, b := strings.Split(actual, "\n"), strings.Split(want, "\n")
		for i := 0; i < min(len(a), len(b)); i++ {
			if a[i] != b[i] {
				t.Fatalf("deep source line%d differs\nGo:%s\nC: %s", i+1, a[i], b[i])
			}
		}
		t.Fatal("deep synthesis trace lengths differ")
	}
	t.Log("96 exact native driver configurations, 1968 commands, next random words and independently verified frame/selection/stopping")
}
