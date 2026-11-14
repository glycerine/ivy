package abc

import (
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

func TestNetworkSourceAccessorsPreserveExistingState(t *testing.T) {
	network := NewNetwork("source")
	sequenceModel := &util.Abc_Cex_t{}
	Abc_NtkSetSeqModel(network, sequenceModel)
	if Abc_NtkReadSeqModel(network) != sequenceModel {
		t.Fatal("sequence model accessor did not preserve the assigned pointer")
	}

	exdc := NewNetwork("exdc")
	Abc_NtkSetExdc(network, exdc)
	if Abc_NtkReadExdc(network) != exdc {
		t.Fatal("EXDC accessor did not preserve the assigned network")
	}

	copyObject := network.NewObject(ObjectNode)
	Abc_ObjSetCopy(network.constant, copyObject)
	if Abc_ObjReadCopy(network.constant) != copyObject {
		t.Fatal("copy accessor did not preserve the assigned object")
	}
}
