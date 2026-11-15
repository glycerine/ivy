package gia

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestExplicitReachabilityNativeReports(t *testing.T) {
	exe := compileGiaSourceOracle(t, "era_review.c")
	native, err := exec.Command(exe, "--verbose").CombinedOutput()
	if err != nil {
		t.Fatalf("C progress: %v\n%s", err, native)
	}
	g := Gia_ManStart(256)
	defer Gia_ManStop(g)
	var r, next [13]int32
	for i := range r {
		r[i] = Gia_ManAppendCi(g)
	}
	Gia_ManHashAlloc(g)
	carry := int32(1)
	for i := range r {
		next[i] = Gia_ManHashXor(g, r[i], carry)
		carry = Gia_ManHashAnd(g, carry, r[i])
	}
	Gia_ManAppendCo(g, r[12])
	for _, v := range next {
		Gia_ManAppendCo(g, v)
	}
	Gia_ManHashStop(g)
	Gia_ManSetRegNum(g, 13)
	goText := hashReviewStdout(t, func() {
		if got := Gia_ManCollectReachable(g, 20000, 0, 0, 1); got != 1 {
			t.Fatalf("counter traversal%d", got)
		}
	})
	// Match the native report fields and widths, while allowing the measured
	// elapsed time and Go's representation-dependent memory estimate to vary.
	timeField := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	memoryField := regexp.MustCompile(`Mem =([ ]*[0-9]+\.[0-9]{2}) MB\.`)
	normalize := func(text string) string {
		fields := timeField.FindAllStringSubmatch(text, -1)
		if len(fields) != 2 {
			t.Errorf("expected progress and summary time fields: %q", text)
		}
		for _, field := range fields {
			if len(field[1]) != 9 {
				t.Errorf("time field width%d, want9: %q", len(field[1]), field[0])
			}
		}
		memories := memoryField.FindAllStringSubmatch(text, -1)
		if len(memories) != 1 {
			t.Errorf("missing progress memory: %q", text)
		}
		for _, field := range memories {
			if len(field[1]) != 9 {
				t.Errorf("memory field width%d, want9", len(field[1]))
			}
		}
		text = timeField.ReplaceAllString(text, "Time = <elapsed> sec")
		return memoryField.ReplaceAllString(text, "Mem = <memory> MB.")
	}
	if got, want := normalize(goText), normalize(string(native)); got != want {
		t.Fatalf("Go report\n%s\nC report\n%s", got, want)
	}
	if !strings.Contains(goText, "Reachability analysis traversed 8192 states with depth 8192.") {
		t.Fatal(fmt.Sprintf("unexpected counter summary: %s", goText))
	}
}
