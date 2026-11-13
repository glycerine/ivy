package rwr

// rwrClassifyForest implements the canonical-node insertion in
// Rwr_ManAddNode followed by Rwr_ManPreprocess's truth-ordered class lists.
// Leaves are excluded because Rwr_ManStart loads the library with
// fPrecompute == 0, so they are not inserted in the hash table.
func rwrClassifyForest(forest []rwrForestNode, canonical []uint16, classes []byte) ([][]int, []uint16) {
	if len(canonical) != 1<<16 || len(classes) != 1<<16 {
		panic("rwr: invalid canonicalization tables")
	}
	table := make([][]int, 1<<16)
	for i := 5; i < len(forest); i++ {
		truth := forest[i].truth
		if canonical[truth] == truth {
			table[truth] = append(table[truth], i)
		}
	}
	result := make([][]int, 222)
	inverse := make([]uint16, 222)
	for truth, list := range table {
		for _, i := range list {
			class := classes[truth]
			if int(class) >= len(result) {
				panic("rwr: invalid NPN class")
			}
			result[class] = append(result[class], i)
			inverse[class] = canonical[truth]
		}
	}
	return result, inverse
}
