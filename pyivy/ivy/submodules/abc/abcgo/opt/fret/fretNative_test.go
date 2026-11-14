package fret

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	abc "github.com/glycerine/abcgo/base/abc"
	vec "github.com/glycerine/abcgo/misc/vec"
)

func TestTrace2FollowsFirstCriticalFanin(t *testing.T) {
	network := abc.NewNetwork("trace2-critical-fanin")
	pi := abc.Abc_NtkCreatePi(network)
	critical := abc.Abc_NtkCreateNode(network)
	noncritical := abc.Abc_NtkCreatePi(network)
	root := abc.Abc_NtkCreateNode(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{{critical, pi}, {root, critical}, {root, noncritical}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	root.Level, critical.Level, noncritical.Level, pi.Level = 2, 1, 0, 0
	pManMR = &MinRegMan_t{
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
		vTimeEdges: make([]vec.Vec_Ptr_t, abc.Abc_NtkObjNumMax(network)+1),
	}
	t.Cleanup(func() { pManMR = nil })

	output := captureStdout(t, func() { trace2(root) })
	want := []string{
		"node " + fmt.Sprint(abc.Abc_ObjId(root)),
		"node " + fmt.Sprint(abc.Abc_ObjId(critical)),
		"node " + fmt.Sprint(abc.Abc_ObjId(pi)),
	}
	if got := strings.Count(output, "node "); got != len(want) {
		t.Fatalf("trace emitted %d nodes, want %d:\n%s", got, len(want), output)
	}
	previous := -1
	for _, node := range want {
		position := strings.Index(output, node)
		if position <= previous {
			t.Fatalf("trace output does not contain ordered node %q:\n%s", node, output)
		}
		previous = position
	}
}

func TestMinRegReportsRejectedDelayConstraint(t *testing.T) {
	network := abc.NewNetwork("minreg-delay-guard")
	pi := abc.Abc_NtkCreatePi(network)
	first, second := abc.Abc_NtkCreateNode(network), abc.Abc_NtkCreateNode(network)
	po := abc.Abc_NtkCreatePo(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{{first, pi}, {second, first}, {po, second}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	output := captureStdout(t, func() {
		got := Abc_FlowRetime_MinReg(network, 1, 0, 0, 0, 0, 0, 5, 1, 0)
		if got != network {
			t.Fatalf("rejected delay constraint returned %p, want original %p", got, network)
		}
	})
	if !strings.Contains(output, "max delay constraint = 1") || !strings.Contains(output, "must be > current max delay") {
		t.Fatalf("delay rejection output = %q", output)
	}
}

func captureStdout(t *testing.T, run func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	os.Stdout = w
	defer w.Close()
	defer func() { os.Stdout = old }()
	run()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestFastFlowPreorderDistancesOnSimpleAIGPath(t *testing.T) {
	network := abc.NewNetwork("flow-preorder-path")
	pi := abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	po := abc.Abc_NtkCreatePo(network)
	abc.Abc_ObjAddFanin(node, pi)
	abc.Abc_ObjAddFanin(po, node)
	pManMR = &MinRegMan_t{
		fIsForward:    1,
		pNtk:          network,
		nNodes:        abc.Abc_NtkObjNum(network),
		pDataArray:    make([]Flow_Data_t, abc.Abc_NtkObjNum(network)),
		vSinkDistHist: vec.Vec_IntStart(20),
	}
	t.Cleanup(func() { pManMR = nil })

	dfsfast_preorder(network)
	want := map[*abc.Abc_Obj_t]Flow_Data_t{
		pi:   {e_dist: 1, r_dist: 2},
		node: {e_dist: 2, r_dist: 3},
		po:   {r_dist: 1},
	}
	for obj, expected := range want {
		got := pManMR.pDataArray[abc.Abc_ObjId(obj)]
		if got.e_dist != expected.e_dist || got.r_dist != expected.r_dist {
			t.Errorf("object %d distances=(%d,%d), want (%d,%d)", abc.Abc_ObjId(obj), got.e_dist, got.r_dist, expected.e_dist, expected.r_dist)
		}
	}
}

func TestPlainBackwardFlowFindsPathToPrimaryInput(t *testing.T) {
	network := abc.NewNetwork("plain-backward-flow-path")
	pi := abc.Abc_NtkCreatePi(network)
	node := abc.Abc_NtkCreateNode(network)
	bi := abc.Abc_NtkCreateBi(network)
	latch := abc.Abc_NtkCreateLatch(network)
	for _, edge := range [][2]*abc.Abc_Obj_t{{node, pi}, {bi, node}, {latch, bi}} {
		abc.Abc_ObjAddFanin(edge[0], edge[1])
	}
	pManMR = &MinRegMan_t{
		fIsForward: 0,
		pNtk:       network,
		pDataArray: make([]Flow_Data_t, abc.Abc_NtkObjNumMax(network)+1),
	}
	t.Cleanup(func() { pManMR = nil })

	if got := dfsplain_e(latch, nil); got != 1 {
		t.Fatal("plain backward DFS did not find the PI-to-latch path")
	}
	if got := FTEST(bi, FLOW); got == 0 {
		t.Fatal("successful augmenting path did not mark the BI as carrying flow")
	}
	if got := FGETPRED(bi); got != nil {
		t.Fatalf("source-side flow predecessor = %p, want nil", got)
	}
}

func TestInitConflictOrderMatchesVecPtrReorder(t *testing.T) {
	network := abc.NewNetwork("init-conflict-order")
	pi0 := abc.Abc_NtkCreatePi(network)
	pi1 := abc.Abc_NtkCreatePi(network)
	node0 := abc.Abc_NtkCreateNode(network)
	node1 := abc.Abc_NtkCreateNode(network)
	got := initConflictNodeOrder([]*abc.Abc_Obj_t{node0, node1}, []*abc.Abc_Obj_t{pi0, pi1})
	want := []*abc.Abc_Obj_t{pi0, pi1, node0, node1}
	if len(got) != len(want) {
		t.Fatalf("conflict order length = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("conflict order[%d] = %p, want %p", i, got[i], want[i])
		}
	}
}

func TestBiasLagMatchesSourceConditional(t *testing.T) {
	for _, test := range []struct {
		lag     int32
		isLatch bool
		wantLag int32
	}{
		{lag: 0, isLatch: false, wantLag: 0},
		{lag: 0, isLatch: true, wantLag: 1},
		{lag: 1, isLatch: false, wantLag: 1},
		{lag: 1, isLatch: true, wantLag: 0},
	} {
		if got := biasNextLag(test.lag, test.isLatch); got != test.wantLag {
			t.Errorf("biasNextLag(%d, %v) = %d, want %d", test.lag, test.isLatch, got, test.wantLag)
		}
	}
}
