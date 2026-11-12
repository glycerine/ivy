package cut

// Cut_TruthCompose substitutes tree cut T for Node in factor cut F without
// temporarily modifying either input support or using a fixed-size C buffer.
func Cut_TruthCompose(factor *Cut_Cut_t, node int32, tree, result *Cut_Cut_t) {
	index := -1
	for i, leaf := range factor.leaves {
		if leaf == node {
			index = i
			break
		}
	}
	if index < 0 {
		panic("cut: composed node absent from factor")
	}
	clear(result.truth)
	for assignment := 0; assignment < len(result.truth)*32; assignment++ {
		treeValue := cutTruthValue(tree, result.leaves, assignment)
		factorIndex := 0
		for i, leaf := range factor.leaves {
			value := false
			if i == index {
				value = treeValue
			} else {
				found := false
				for j, mergedLeaf := range result.leaves {
					if leaf == mergedLeaf {
						value = assignment&(1<<uint(j)) != 0
						found = true
						break
					}
				}
				if !found {
					panic("cut: factor support absent from result")
				}
			}
			if value {
				factorIndex |= 1 << uint(i)
			}
		}
		if ((factor.truth[factorIndex>>5]&(1<<uint(factorIndex&31)) != 0) != factor.FCompl) != result.FCompl {
			result.truth[assignment>>5] |= 1 << uint(assignment&31)
		}
	}
}
