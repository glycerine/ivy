package gia

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func gigReviewCompile(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	src := filepath.Join(filepath.Dir(file), "../../../src/aig/gia/giaGig.c")
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) <= 310 || strings.TrimSpace(lines[310]) != "assert( k == 4 );" {
		t.Fatal("narrow LUT6 correction no longer matches source; review the fixture")
	}
	return compileGiaSourceOracle(t, "gig_review.c")
}

func TestGigNativeSuccessReport(t *testing.T) {
	exe := gigReviewCompile(t)
	file := gigReviewWrite(t, "w2 = PI\nw3 = PI\nw4 = Lut4(w2,w3,-,-) [8888]\nw5 = PO(w4)\n")
	native, err := exec.Command(exe, "--raw", file).CombinedOutput()
	if err != nil {
		t.Fatalf("C GIG report: %v %s", err, native)
	}
	prefix, _, ok := strings.Cut(string(native), "graph ")
	if !ok {
		t.Fatal("missing native graph record")
	}
	p, report := gigReviewRead(t, file)
	if p == nil {
		t.Fatal("Go read failed")
	}
	Gia_ManStop(p)
	clock := regexp.MustCompile(`Time =([ ]*[0-9]+\.[0-9]{2}) sec`)
	if clock.ReplaceAllString(prefix, "Time = <clock> sec") != clock.ReplaceAllString(report, "Time = <clock> sec") {
		t.Fatalf("Go report%q C%q", report, prefix)
	}
	for _, field := range clock.FindAllStringSubmatch(report, -1) {
		if len(field[1]) != 9 {
			t.Fatal("clock field width changed")
		}
	}
}
