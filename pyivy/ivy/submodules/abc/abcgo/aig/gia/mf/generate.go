package mf

import (
	"fmt"
	"github.com/glycerine/abcgo/aig/ioa"
)

// GenerateCNF ports Mf_ManGenerateCnf's source defaults and preprocessing. It
// performs real-XOR/MUX coarsening unless object IDs or choices suppress it.
func GenerateCNF(network Network, lutSize int, objectIDs, addOr, withMapping bool) (*CNF, error) {
	if network == nil {
		return nil, fmt.Errorf("mf: nil network")
	}
	options := DefaultOptions()
	options.LUTSize = lutSize
	options.ObjectIDs = objectIDs
	choices := false
	if provider, ok := network.(interface{ Sibling(int32) int32 }); ok {
		for i := 0; i < network.ObjectCount(); i++ {
			if provider.Sibling(int32(i)) != 0 {
				choices = true
				break
			}
		}
	}
	if !objectIDs && !choices {
		coarse, err := Coarsen(network, 3)
		if err != nil {
			return nil, err
		}
		network = coarse
	}
	mapping, err := GenerateMapping(network, options)
	if err != nil {
		return nil, err
	}
	return DeriveCNF(mapping, objectIDs, addOr, withMapping)
}

// GenerateModelCNF is the source Gia_ManSatokoCreate configuration:
// Mf_ManGenerateCnf(network,8,0,1,0,0).
func GenerateModelCNF(model *ioa.Model) (*CNF, error) {
	network, err := NetworkFromModel(model)
	if err != nil {
		return nil, err
	}
	return GenerateCNF(network, 8, false, true, false)
}
