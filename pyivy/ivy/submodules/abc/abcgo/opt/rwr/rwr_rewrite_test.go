package rwr

import (
	"os"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/bool/dec"
)

type rewriteTestNetwork struct {
	marks []int
	calls int
}

func (n *rewriteTestNetwork) Lookup(id int32) (any, bool)     { return id, id >= 0 }
func (n *rewriteTestNetwork) Complement(v any) any            { return ^v.(int32) }
func (n *rewriteTestNetwork) FanoutCount(any) int             { return 2 }
func (n *rewriteTestNetwork) MarkBoundary(_ []any, delta int) { n.marks = append(n.marks, delta) }
func (n *rewriteTestNetwork) LabelMFFC(any) int32             { return 3 }
func (n *rewriteTestNetwork) CountGraph(_ any, _ *dec.Dec_Graph_t, saved, level int32) int32 {
	if saved != 3 || level != 12 {
		panic("incorrect MFFC or required level")
	}
	n.calls++
	return 1
}

func TestRewriteCutsChoosesLibraryImplementation(t *testing.T) {
	p := Rwr_ManStart(0)
	n := &rewriteTestNetwork{}
	truth := uint16(0x6996)
	if len(p.vClasses[p.pMap[truth]]) == 0 {
		t.Fatal("missing XOR4 library class")
	}
	cuts := []rwrCut{{truth, []int32{-1, 1, 2, 3}}, {truth, []int32{0, 1, 2, 3}}}
	choice := p.rwrRewriteCuts(n, nil, cuts, 12, false)
	if choice.graph == nil || choice.gain != 2 || n.calls == 0 || len(n.marks) != 2 || n.marks[0] != 1 || n.marks[1] != -1 {
		t.Fatalf("choice=%+v calls=%d marks=%v", choice, n.calls, n.marks)
	}
	if p.nCutsBad != 1 || p.nCutsGood != 1 || p.nNodesConsidered != 1 || p.nNodesRewritten != 1 || p.nNodesGained != 2 || p.nScores[p.pMap[truth]] != 1 {
		t.Fatal("rewrite statistics do not match source")
	}
	for i, leaf := range choice.fanins {
		if dec.Dec_GraphLeafFunc(choice.graph, int32(i)) != leaf {
			t.Fatalf("best graph lost leaf %d", i)
		}
	}
}

func TestRewriteCutsKeepsFirstEqualGainCut(t *testing.T) {
	p := Rwr_ManStart(0)
	n := &rewriteTestNetwork{}
	truth := uint16(0x6996)
	first := rwrCut{truth, []int32{0, 1, 2, 3}}
	second := rwrCut{truth, []int32{4, 5, 6, 7}}
	choice := p.rwrRewriteCuts(n, nil, []rwrCut{first, second}, 12, false)
	if choice.gain != 2 {
		t.Fatalf("gain=%d, want 2", choice.gain)
	}
	for i, want := range []any{int32(0), int32(1), int32(2), int32(3)} {
		if choice.fanins[i] != want {
			t.Fatalf("fanins=%v; equal-gain second cut replaced first at index %d", choice.fanins, i)
		}
	}
}

func TestRewriteScoresTheManagersWinningImplementation(t *testing.T) {
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	truth := uint16(0x6996)
	net := &rewriteTestNetwork{}
	gain := Rwr_RewriteCuts(p, net, nil, []Rwr_Cut{{truth, []int32{0, 1, 2, 3}}}, 12, 0, 0)
	class := p.vClasses[p.pMap[truth]]
	winner := class[0] // Each candidate costs one AND. C keeps the first tie.
	if gain != 2 || winner.nScore != 1 || winner.nGain != 2 || winner.nAdded != 1 || winner.graph != Rwr_ManReadDecs(p) {
		t.Fatalf("gain %d first candidate score/gain/added %d/%d/%d, returned its graph %v", gain, winner.nScore, winner.nGain, winner.nAdded, winner.graph == Rwr_ManReadDecs(p))
	}
	Rwr_ScoresClean(p)
	if winner.nScore != 0 || winner.nGain != 0 || winner.nAdded != 0 || p.nScores[p.pMap[truth]] != 1 {
		t.Fatal("score cleanup must reset candidate scores and preserve the class rewrite count")
	}
}

func TestCleanAndReportCandidateScores(t *testing.T) {
	p := Rwr_ManStart(0)
	defer Rwr_ManStop(p)
	class := p.pMap[0x6996]
	node := p.vClasses[class][0]
	node.nScore, node.nAdded, node.nGain = 2, 3, 7
	p.nScores[class] = 2
	file, err := os.CreateTemp(t.TempDir(), "scores")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	os.Stdout = file
	defer func() { os.Stdout = previous; file.Close() }()
	Rwr_ScoresReport(p)
	data, err := os.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "Total gain =      7") || !strings.Contains(string(data), "S=    2. A=    3. G=     7.") {
		t.Errorf("missing gains or candidate detail in report %q", data)
	}
	Rwr_ScoresClean(p)
	if node.nScore != 0 || node.nAdded != 0 || node.nGain != 0 || p.nScores[class] != 2 {
		t.Errorf("after cleanup candidate %d/%d/%d, class rewrites %d", node.nScore, node.nAdded, node.nGain, p.nScores[class])
	}
}
