package mapper

import "math"

const (
	MAP_FLOAT_LARGE = float32(math.MaxFloat32 / 10)
	MAP_FLOAT_SMALL = float32(1e-3)
	MAP_NO_VAR      = float32(-9999)
)

func boolToInt(value bool) int32 {
	if value { return 1 }
	return 0
}

func cutNext(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil { return nil }
	return cut.pNext
}

func nodeDelay(manager *Map_Man_t, node *Map_Node_t) float32 {
	index := int(node.Num)
	if manager == nil || index < 0 || index >= len(manager.pNodeDelays) { return 0 }
	return manager.pNodeDelays[index]
}

func incInt(value *int32) int32 { old := *value; *value++; return old }
func decInt(value *int32) int32 { *value--; return *value }

// Pointer tagging in the C mapper uses the low address bit as a phase. Go
// pointers cannot be tagged safely, so a complemented reference is an
// explicit, allocation-local wrapper around its regular object.
func Map_IsComplement(node *Map_Node_t) bool { return node != nil && node.complemented }

func Map_Regular(node *Map_Node_t) *Map_Node_t {
	if node == nil {
		return nil
	}
	if node.regular != nil {
		return node.regular
	}
	return node
}

func Map_Not(node *Map_Node_t) *Map_Node_t {
	if node == nil {
		return nil
	}
	base := Map_Regular(node)
	if node.complemented {
		return base
	}
	if base.neg == nil {
		base.neg = &Map_NodeStruct_t_{regular: base, complemented: true}
	}
	return base.neg
}

func Map_NotCond(node *Map_Node_t, condition bool) *Map_Node_t {
	if condition {
		return Map_Not(node)
	}
	return node
}

func Map_CutIsComplement(cut *Map_Cut_t) bool { return cut != nil && cut.complemented }

func Map_CutRegular(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil {
		return nil
	}
	if cut.regular != nil {
		return cut.regular
	}
	return cut
}

func Map_CutNot(cut *Map_Cut_t) *Map_Cut_t {
	if cut == nil {
		return nil
	}
	base := Map_CutRegular(cut)
	if cut.complemented {
		return base
	}
	if base.neg == nil {
		base.neg = &Map_CutStruct_t_{regular: base, complemented: true}
	}
	return base.neg
}

func Map_CutNotCond(cut *Map_Cut_t, condition bool) *Map_Cut_t {
	if condition {
		return Map_CutNot(cut)
	}
	return cut
}

func Map_NodeReadRef(node *Map_Node_t) int32 { return Map_Regular(node).nRefs }

func Map_NodeRef(node *Map_Node_t) { Map_Regular(node).nRefs++ }
