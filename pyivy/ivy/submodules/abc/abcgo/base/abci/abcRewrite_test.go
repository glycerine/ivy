package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	abcio "github.com/glycerine/abcgo/base/io"
	"github.com/glycerine/abcgo/misc/extra"
	"github.com/glycerine/abcgo/misc/vec"
	"github.com/glycerine/abcgo/opt/cut"
)

func BenchmarkI10ReadBalanceRewrite(b *testing.B) {
	benchmarkI10Balance(b, 2, true)
}

func BenchmarkI10ReadBalance(b *testing.B) {
	benchmarkI10Balance(b, 1, false)
}

func benchmarkI10Balance(b *testing.B, passes int, rewrite bool) {
	b.Helper()
	extra.Extra_ProgressBarSetMode(false, true)
	defer extra.Extra_ProgressBarSetMode(true, false)
	b.ReportAllocs()
	var nodes, depth int32
	for b.Loop() {
		n, err := abcio.ReadAiger("../../test_vectors/i10.aig", true)
		if err != nil {
			b.Fatal(err)
		}
		for range passes {
			balanced := Abc_NtkBalance(n, 0, 0, 1)
			abc.Abc_NtkDelete(n)
			n = balanced
		}
		if rewrite && Abc_NtkRewrite(n, 0, 0, 0, 0, 0) == 0 {
			b.Fatal("rewrite failed")
		}
		nodes, depth = abc.Abc_NtkNodeNum(n), abc.Abc_AigLevel(n)
		abc.Abc_NtkDelete(n)
	}
	b.ReportMetric(float64(nodes), "ANDs/op")
	b.ReportMetric(float64(depth), "levels/op")
}

func TestRewriteI10PreservesZeroInput(t *testing.T) {
	n, err := abcio.ReadAiger("../../test_vectors/i10.aig", true)
	if err != nil {
		t.Fatal(err)
	}
	defer abc.Abc_NtkDelete(n)
	want := evaluateNetwork(t, n, 0)
	first := Abc_NtkBalance(n, 0, 0, 1)
	defer abc.Abc_NtkDelete(first)
	second := Abc_NtkBalance(first, 0, 0, 1)
	defer abc.Abc_NtkDelete(second)
	if Abc_NtkRewrite(second, 0, 0, 0, 0, 0) == 0 {
		t.Fatal("rewrite failed")
	}
	for i, output := range evaluateNetwork(t, second, 0) {
		if output != want[i] {
			t.Fatalf("all zeros: %s changed from %t to %t", second.OutputAt(int(i)).Name(), want[i], output)
		}
	}
}

func TestAbcNtkStartCutManForRewrite(t *testing.T) {
	network := abc.NewNetwork("rewrite-cuts")
	manager := Abc_NtkStartCutManForRewrite(network)
	if manager == nil {
		t.Fatal("cut manager is nil")
	}
	defer cut.Cut_ManStop(manager)
	params := cut.Cut_ManReadParams(manager)
	if params.NVarsMax != 4 || params.NKeepMax != 250 || params.NIdsMax != abc.Abc_NtkObjNumMax(network) || params.FTruth == 0 || params.FFilter == 0 {
		t.Fatalf("unexpected rewrite cut parameters: %+v", params)
	}
}

func TestAbcRwrExpWithCutRecRecognizesComplementedLeaf(t *testing.T) {
	network := abc.NewNetwork("rewrite-complemented-leaf")
	leaf := network.NewObject(abc.ObjectPI)
	leaves := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, leaf)
	defer vec.Vec_PtrFree(leaves)

	Abc_RwrExpWithCut_rec(leaf.Signal().Not(), leaves, 1)
	if !leaf.MarkA {
		t.Fatal("complemented leaf was not marked as part of the cut")
	}
}
