package abc

import (
	"github.com/glycerine/abcgo/aig/miniaig"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestSerializeMappedNetwork(t *testing.T) {
	n, _, gates, _, node, _ := abcSCLFixture(t)
	node.Data = mio.Mio_LibraryReadGateByName(gates, "and2", "")
	result := abcSerializeMapped(n)
	row := vec.Vec_IntArray(vec.Vec_WecEntry(result, node.Temp))
	if len(row) < 3 || row[0] != miniaig.ABC_OPER_BIT_AND || row[1] != int32(node.Temp) {
		t.Fatalf("node record %v", row)
	}
	last := vec.Vec_IntArray(vec.Vec_WecEntryLast(result))
	if len(last) != 2 || last[0] != miniaig.ABC_OPER_BIT_BUF {
		t.Fatalf("output record %v", last)
	}
}
