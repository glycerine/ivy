package fraig

// Fraig_Regular returns the data-owning node for either edge polarity.
func Fraig_Regular(node *Fraig_Node_t) *Fraig_Node_t {
	if node != nil && node.complementOf != nil {
		return node.complementOf
	}
	return node
}

// Fraig_IsComplement reports whether the edge denotes the opposite phase.
func Fraig_IsComplement(node *Fraig_Node_t) bool {
	return node != nil && node.complementOf != nil
}

// Fraig_Not returns the canonical opposite-polarity handle for a node. FRAIG
// managers own their nodes and are not concurrently mutated by multiple users.
func Fraig_Not(node *Fraig_Node_t) *Fraig_Node_t {
	if node == nil {
		panic("fraig: cannot complement a nil edge")
	}
	if node.complementOf != nil {
		return node.complementOf
	}
	if node.complement == nil {
		node.complement = &Fraig_Node_t{complementOf: node}
	}
	return node.complement
}

func Fraig_NotCond(node *Fraig_Node_t, complement bool) *Fraig_Node_t {
	if complement {
		return Fraig_Not(node)
	}
	return node
}
