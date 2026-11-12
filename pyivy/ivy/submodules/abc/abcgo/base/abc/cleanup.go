package abc

import "fmt"

// Abc_NtkReduceNodes removes logic nodes absent from the retained set.
func Abc_NtkReduceNodes(n *Network, nodes []*Object) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: node reduction requires a logic network")
	}
	keep := make(map[*Object]bool, len(nodes))
	for _, node := range nodes {
		if node == nil || node.network != n || !n.contains(node) {
			panic("abc: invalid retained node")
		}
		keep[node] = true
	}
	var removed int32
	for _, node := range n.ObjectsOfKind(ObjectNode) {
		if !keep[node] {
			Abc_NtkDeleteObj(node)
			removed++
		}
	}
	if Abc_NtkCheck(n) == 0 {
		fmt.Println("Abc_NtkCleanup: The network check has failed.")
	}
	return removed
}

// Abc_NtkCleanup removes logic nodes that cannot reach an output.
func Abc_NtkCleanup(n *Network, verbose int32) int32 {
	if Abc_NtkIsLogic(n) == 0 {
		panic("abc: cleanup requires a logic network")
	}
	removed := Abc_NtkReduceNodes(n, Abc_NtkDfs(n, 0))
	if verbose != 0 {
		fmt.Printf("Cleanup removed %d dangling nodes.\n", removed)
	}
	return removed
}
