package fra

import "github.com/glycerine/abcgo/proof/framgr"

// classPartition shares the FRAIG manager's candidate-class state.
type classPartition = framgr.ClassPartition

func newClassPartition(nodeCount int) *classPartition { return framgr.NewClassPartition(nodeCount) }
