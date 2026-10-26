package abc

// Associative DSD brace removal from opt/dau/dauMerge.c.

// DauDSDRemoveBraces removes redundant associative DSD groups. Like the
// source operation, it flattens an uncomplemented AND inside an AND and any
// XOR inside an XOR. A complemented nested XOR transfers its phase to the
// first flattened fanin, preserving its Boolean function. MUX and prime
// boundaries are retained. The input string is never modified.
func DauDSDRemoveBraces(text string, options DauDSDOptions) (string, error) {
	o, err := normalizeDauDSDOptions(options)
	if err != nil {
		return "", err
	}
	if len(text) > o.MaxBytes {
		return "", ErrLimit
	}
	// The source compaction pass cancels adjacent complements after removing
	// spaces. Collapse them before parsing so inputs such as !!a retain that
	// accepted behavior without adding double-negation nodes to the arena.
	compact := make([]byte, 0, len(text))
	for i := 0; i < len(text); {
		if text[i] != '!' {
			compact = append(compact, text[i])
			i++
			continue
		}
		j := i
		for j < len(text) && text[j] == '!' {
			j++
		}
		if (j-i)&1 != 0 {
			compact = append(compact, '!')
		}
		i = j
	}
	formula, err := ParseDauDSD(string(compact), o)
	if err != nil {
		return "", err
	}
	var simplify func(int) error
	simplify = func(index int) error {
		if o.Stop != nil && o.Stop() {
			return ErrCancelled
		}
		node := &formula.nodes[index]
		for _, child := range node.children {
			if err := simplify(child); err != nil {
				return err
			}
		}
		for _, child := range node.bindings {
			if err := simplify(child); err != nil {
				return err
			}
		}
		if node.kind != DauDSDAnd && node.kind != DauDSDXor {
			return nil
		}
		children := make([]int, 0, len(node.children))
		for _, childIndex := range node.children {
			child := &formula.nodes[childIndex]
			if child.kind != node.kind || node.kind == DauDSDAnd && child.complement {
				children = append(children, childIndex)
				continue
			}
			if node.kind == DauDSDXor && child.complement {
				// !(x XOR y ...) == !x XOR y ... . The parsed arena is a
				// tree, so toggling this child cannot affect another branch.
				first := child.children[0]
				formula.nodes[first].complement = !formula.nodes[first].complement
			}
			children = append(children, child.children...)
		}
		node.children = children
		return nil
	}
	if err := simplify(formula.root); err != nil {
		return "", err
	}
	result := formula.String()
	if len(result) > o.MaxBytes {
		return "", ErrLimit
	}
	return result, nil
}
