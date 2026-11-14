package abc

import "github.com/glycerine/abcgo/misc/nm"

// Abc_NtkResetNames replaces a network's name manager with an empty one sized
// for its CI, CO, and box objects, matching ABC's prepare-permutation setup.
func Abc_NtkResetNames(n *Network) {
	nm.Nm_ManFree(n.names)
	n.names = nm.Nm_ManCreate(Abc_NtkCiNum(n) + Abc_NtkCoNum(n) + Abc_NtkBoxNum(n))
}
