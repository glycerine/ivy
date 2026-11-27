package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"io"
	"os"
	"strings"
	"testing"
)

// cecCore.c:296-298 prints rounds, frames and words in that order.
func TestSimulationSequentialReportSourceDimensions(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, r := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, r)
	gia.Gia_ManSetRegNum(g, 1)
	var p Cec_ParSim_t
	Cec_ManSimSetDefaultParams(&p)
	p.FSeqSimulate = 1
	p.NRounds = 0
	p.NFrames = 3
	p.NWords = 7
	old := os.Stdout
	rd, wr, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = wr
	Cec_ManSimulation(g, &p)
	os.Stdout = old
	wr.Close()
	out, err := io.ReadAll(rd)
	rd.Close()
	if err != nil {
		t.Fatal(err)
	}
	// rounds, frames and words each in their place (CBUGS_FIXED.md #22; the
	// source reported the rounds as frames and the frames as words)
	if !strings.Contains(string(out), "Performing 0 rounds of random simulation of 3 frames with 7 words.\n") {
		t.Fatalf("source report: %s", out)
	}
}
