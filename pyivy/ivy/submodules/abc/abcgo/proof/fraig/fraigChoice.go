package fraig

import "fmt"

// Fraig_ManAddChoices creates associative alternatives up to the C limit.
func Fraig_ManAddChoices(p *Fraig_Man_t, verbose, limit int32) {
	before, inputs := Fraig_ManReadNodeNum(p), Fraig_ManReadInputNum(p)
	maxNodes := inputs + limit*(before-inputs)
	fmt.Printf("Limit = %d, Before = %d\n", maxNodes, before)
	Fraig_ManCheckConsistency(p)
	for i := inputs + 1; i < Fraig_ManReadNodeNum(p) && maxNodes > Fraig_ManReadNodeNum(p); i++ {
		node := Fraig_ManReadIthNode(p, i)
		assert(node != nil, "choice node exists")
		x, c := Fraig_NodeReadOne(node), Fraig_NodeReadTwo(node)
		if Fraig_NodeIsAnd(x) != 0 && !Fraig_IsComplement(x) {
			a, b := Fraig_NodeReadOne(x), Fraig_NodeReadTwo(x)
			Fraig_NodeAnd(p, a, Fraig_NodeAnd(p, b, c))
			Fraig_NodeAnd(p, b, Fraig_NodeAnd(p, a, c))
		}
		a, x := Fraig_NodeReadOne(node), Fraig_NodeReadTwo(node)
		if Fraig_NodeIsAnd(x) != 0 && !Fraig_IsComplement(x) {
			b, c := Fraig_NodeReadOne(x), Fraig_NodeReadTwo(x)
			Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, b), c)
			Fraig_NodeAnd(p, Fraig_NodeAnd(p, a, c), b)
		}
	}
	Fraig_ManCheckConsistency(p)
	after := Fraig_ManReadNodeNum(p)
	fmt.Printf("Nodes before = %6d. Nodes with associative choices = %6d. Increase = %4.2f %%.\n", before, after, float64(after-before)*100/float64(before-inputs))
	fmt.Println("Distributive = 0")
}
