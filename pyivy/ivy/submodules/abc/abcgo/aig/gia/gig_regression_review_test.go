package gia

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func gigReviewWrite(t *testing.T, data string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "network.gig")
	if err := os.WriteFile(path, []byte(data), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func gigReviewRead(t *testing.T, path string) (*Gia_Man_t, string) {
	t.Helper()
	var p *Gia_Man_t
	report := hashReviewStdout(t, func() { p = Gia_ManReadGig(path) })
	return p, report
}
func gigReviewMask(t *testing.T, p *Gia_Man_t, output int32) uint64 {
	t.Helper()
	var mask uint64
	if Gia_ManCiNum(p) > 6 {
		t.Fatal("mask oracle supports at most six CIs")
	}
	for input := 0; input < 1<<Gia_ManCiNum(p); input++ {
		values := cofReviewValues(p, input)
		if values[Gia_ManCo(p, output).id] {
			mask |= uint64(1) << input
		}
	}
	return mask
}
func TestGigShortTruthExpansion(t *testing.T) {
	for _, tc := range []struct {
		hex  string
		want uint64
	}{{"F", 0xffff}, {"8", 0x8888}, {"AB", 0xabab}, {"0xA", 0xaaaa}, {"5", 0x5555}} {
		t.Run(tc.hex, func(t *testing.T) {
			path := gigReviewWrite(t, fmt.Sprintf("w2 = PI\nw3 = PI\nw4 = PI\nw5 = PI\nw6 = Lut4(w2,w3,w4,w5) [%s]\nw7 = PO(w6)\n", tc.hex))
			p, _ := gigReviewRead(t, path)
			if p == nil {
				t.Fatal("valid GIG rejected")
			}
			defer Gia_ManStop(p)
			if got := gigReviewMask(t, p, 0); got != tc.want {
				t.Fatalf("table%s mask%04x want%04x", tc.hex, got, tc.want)
			}
		})
	}
}
func TestGigRejectsInvalidSelectionParent(t *testing.T) {
	for _, tc := range []struct{ name, parent string }{{"PI", "w2 = PI\n"}, {"PO", "w2 = PO(0)\n"}, {"Lut4", "w2 = Lut4(0,0,0,0) [0]\n"}, {"self", ""}} {
		t.Run(tc.name, func(t *testing.T) {
			parent := "w2"
			if tc.name == "self" {
				parent = "w3"
			}
			path := gigReviewWrite(t, tc.parent+"w3 = Sel("+parent+")\nw4 = PO(0)\n")
			p, report := gigReviewRead(t, path)
			if p != nil {
				Gia_ManStop(p)
				t.Fatal("invalid unused selection was silently accepted")
			}
			if !strings.Contains(report, "select") {
				t.Fatalf("missing selection context: %s", report)
			}
		})
	}
}

func TestGigConstantWireReferences(t *testing.T) {
	file := gigReviewWrite(t, "w2 = PO(w0)\nw3 = PO(w1)\n")
	p, report := gigReviewRead(t, file)
	if p == nil {
		t.Fatalf("source-valid constant wire references rejected: %s", report)
	}
	defer Gia_ManStop(p)
	if Gia_ManCiNum(p) != 0 || Gia_ManCoNum(p) != 2 || gigReviewMask(t, p, 0) != 0 || gigReviewMask(t, p, 1) != 1 {
		t.Fatal("constant reference functions changed")
	}
}
