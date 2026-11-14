package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	aiggraph "github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/gia"
	baseabc "github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func scorrTestNetwork(g *aiggraph.Graph) *Abc_Ntk_t {
	manager := gia.Gia_ManFromAig(g)
	defer gia.Gia_ManStop(manager)
	return baseabc.Abc_NtkCreateFromGias("scorr-test", nil, manager)
}

func TestScorrWritesFlopRepresentativeAndBmcFilename(t *testing.T) {
	graph := aiggraph.NewGraph()
	graph.NewInput() // primary input
	flopOutput := graph.NewInput()
	aiggraph.Aig_ManSetRegNum(graph, 1)
	aiggraph.Aig_ManReprStart(graph, aiggraph.Aig_ManObjNumMax(graph))
	aiggraph.Aig_ObjCreateRepr(graph, aiggraph.Aig_ManConst1(graph), flopOutput)
	defer aiggraph.Aig_ManReprStop(graph)

	networkGraph := aiggraph.NewGraph()
	input := networkGraph.NewInput()
	networkGraph.AddOutput(input)
	network := scorrTestNetwork(networkGraph)
	defer baseabc.Abc_NtkDelete(network)
	pi := baseabc.Abc_NtkPi(network, 0)
	pi.SetName("flop_q")

	idToName := vec.Vec_IntAlloc(int(aiggraph.Aig_ManObjNumMax(graph)))
	vec.Vec_IntFill(idToName, int(aiggraph.Aig_ManObjNumMax(graph)), -1)
	vec.Vec_IntWriteEntry(idToName, int(flopOutput.NodeID()), int32(pi.ID))
	defer vec.Vec_IntFree(idToName)
	output := filepath.Join(t.TempDir(), "pairs.eq")
	data := &Tst_Dat_t{pNetlist: network, pAig: graph, vId2Name: idToName, pFileNameOut: output, fFlopOnly: 1, fDumpBmc: 1}
	if count := Abc_NtkTestScorrWriteEquivAig(data); count != 1 {
		t.Fatalf("written pair count = %d, want 1", count)
	}
	if data.fDumpBmc != 0 {
		t.Fatal("first callback did not consume the BMC filename flag")
	}
	contents, err := os.ReadFile(Abc_NtkBmcFileName(output))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(contents); !strings.Contains(got, "flop_q=const0\n") {
		t.Fatalf("equivalence output = %q, want flop_q=const0", got)
	}
}
