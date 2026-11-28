package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"regexp"
	"testing"
)

func fxReviewMappedPair() *Gia_Man_t {
	p := Gia_ManStart(32)
	a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
	x := Gia_ManAppendAnd(p, a, b)
	y, z := Gia_ManAppendAnd(p, x, c), Gia_ManAppendAnd(p, x, c^1)
	Gia_ManAppendCo(p, y)
	Gia_ManAppendCo(p, z^1)
	p.VMapping = Vec_IntStart(int(p.nObjs))
	for _, root := range []int32{y >> 1, z >> 1} {
		Vec_IntWriteEntry(p.VMapping, int(root), int32(Vec_IntSize(p.VMapping)))
		for _, value := range []int32{3, a >> 1, b >> 1, c >> 1, root} {
			Vec_IntPush(p.VMapping, value)
		}
	}
	return p
}

func TestFxNativeDriverReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "fx_report_review.c")
	native, err := testoracle.Command(t, exe).CombinedOutput()
	if err != nil {
		t.Fatalf("C driver reports: %v\n%s", err, native)
	}
	got := hashReviewStdout(t, func() {
		for mode := int32(0); mode < 4; mode++ {
			p := fxReviewMappedPair()
			fmt.Printf("mode%d\n", mode)
			q := Gia_ManPerformFx(p, 4, 0, 0, mode&1, mode>>1&1)
			cofReviewCheckRebuild(t, p, q)
			fmt.Printf("result ci%d co%d and%d\n", Gia_ManCiNum(q), Gia_ManCoNum(q), Gia_ManAndNum(q))
			Gia_ManStop(q)
			Gia_ManStop(p)
		}
	})
	clock := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	for _, output := range []string{string(native), got} {
		for _, field := range clock.FindAllStringSubmatch(output, -1) {
			if len(field[1]) != 9 {
				t.Fatalf("timing width changed: %q", field[0])
			}
		}
	}
	esopReviewCompare(t, []byte(clock.ReplaceAllString(string(native), "Time = <clock> sec")), clock.ReplaceAllString(got, "Time = <clock> sec"))
}
