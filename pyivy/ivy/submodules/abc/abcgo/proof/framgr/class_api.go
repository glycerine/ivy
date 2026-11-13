package framgr

// ClassPartition is the shared FRA candidate-class manager.
type ClassPartition = classPartition

// NewClassPartition creates candidate storage for the given number of nodes.
func NewClassPartition(nodeCount int) *ClassPartition { return newClassPartition(nodeCount) }
