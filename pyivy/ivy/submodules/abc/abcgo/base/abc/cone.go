package abc

// Abc_NodeConeCollect returns the cone between roots and leaves in fanin-first
// order. When includeLeaves is set, the specified leaves precede the cone.
func Abc_NodeConeCollect(roots, leaves []*Object, includeLeaves bool) []*Object {
	seen := make(map[*Object]bool, len(leaves))
	var nodes []*Object
	for _, leaf := range leaves {
		if leaf == nil {
			panic("abc: nil cone leaf")
		}
		seen[leaf] = true
		if includeLeaves {
			nodes = append(nodes, leaf)
		}
	}
	var visit func(*Object)
	visit = func(node *Object) {
		if node == nil {
			panic("abc: nil cone root or fanin")
		}
		if seen[node] {
			return
		}
		seen[node] = true
		if node.Kind == ObjectNode {
			for _, edge := range node.fanins {
				visit(edge.Object)
			}
		}
		nodes = append(nodes, node)
	}
	for _, root := range roots {
		visit(root)
	}
	return nodes
}
