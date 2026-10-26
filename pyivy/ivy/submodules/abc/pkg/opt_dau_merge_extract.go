package abc

// Pure-subtree extraction from opt/dau/dauMerge.c.

import "strings"

type dauDSDMergeDefinitions struct {
	shared, maxVariables, maxBytes int
	values                         []string
}

func (d *dauDSDMergeDefinitions) add(text string) (string, error) {
	index := d.shared + len(d.values)
	if index >= d.maxVariables || index >= 26 || len(text) > d.maxBytes {
		return "", ErrLimit
	}
	d.values = append(d.values, text)
	return string(byte('a' + index)), nil
}

func dauDSDMergeGroupSpelling(formula *DauDSDFormula, index int) (prefix, suffix string, children []int) {
	node := &formula.nodes[index]
	complement := ""
	if node.complement {
		complement = "!"
	}
	switch node.kind {
	case DauDSDAnd:
		return complement + "(", ")", node.children
	case DauDSDXor:
		return complement + "[", "]", node.children
	case DauDSDMux:
		if len(node.bindings) == 0 {
			return complement + "<", ">", node.children
		}
		var local strings.Builder
		local.WriteString(complement)
		local.WriteByte('<')
		for _, child := range node.children {
			local.WriteString(formula.render(child, false))
		}
		local.WriteString(">{")
		return local.String(), "}", node.bindings
	case DauDSDPrime:
		spelling := formula.render(index, false)
		brace := strings.IndexByte(spelling, '{')
		return spelling[:brace+1], "}", node.children
	default:
		return "", "", nil
	}
}

// dauDSDMergeExtract replaces pure direct subtrees with definition variables.
// Definitions are appended in source traversal order and never alias the
// formula. shared+len(definitions) is the next placeholder variable.
func dauDSDMergeExtract(formula *DauDSDFormula, statuses []int, index int, definitions *dauDSDMergeDefinitions, stop func() bool) (string, error) {
	if formula == nil || definitions == nil || index < 0 || index >= len(formula.nodes) || len(statuses) != len(formula.nodes) {
		return "", ErrDimension
	}
	if stop != nil && stop() {
		return "", ErrCancelled
	}
	node := &formula.nodes[index]
	if node.kind == DauDSDVariable || node.kind == DauDSDConstant {
		return formula.render(index, false), nil
	}
	if statuses[index] == 3 {
		return "", ErrInvariant
	}
	prefix, suffix, children := dauDSDMergeGroupSpelling(formula, index)
	if len(children) < 2 {
		return "", ErrInvariant
	}
	parts := make([]string, 0, len(children))
	status := statuses[index]
	individual := status == 1 || node.kind == DauDSDMux || node.kind == DauDSDPrime
	if status == 2 && !individual {
		pure := make([]string, 0, len(children))
		for _, child := range children {
			if statuses[child] == 3 {
				pure = append(pure, formula.render(child, false))
				continue
			}
			part, err := dauDSDMergeExtract(formula, statuses, child, definitions, stop)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
		open, close := "(", ")"
		if node.kind == DauDSDXor {
			open, close = "[", "]"
		}
		placeholder, err := definitions.add(open + strings.Join(pure, "") + close)
		if err != nil {
			return "", err
		}
		parts = append(parts, placeholder)
	} else {
		for _, child := range children {
			if statuses[child] == 3 {
				placeholder, err := definitions.add(formula.render(child, false))
				if err != nil {
					return "", err
				}
				parts = append(parts, placeholder)
				continue
			}
			part, err := dauDSDMergeExtract(formula, statuses, child, definitions, stop)
			if err != nil {
				return "", err
			}
			parts = append(parts, part)
		}
	}
	result := prefix + strings.Join(parts, "") + suffix
	if len(result) > definitions.maxBytes {
		return "", ErrLimit
	}
	return result, nil
}
