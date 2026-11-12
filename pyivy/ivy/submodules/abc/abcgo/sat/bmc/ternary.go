package bmc

// Three-valued simulation states match the source encoding stored in two-bit
// fields: zero, one, and not yet determined.
const (
	absZero int32 = iota + 1
	absOne
	absUnknown
)

func ternaryNot(value int32) int32 {
	switch value {
	case absZero:
		return absOne
	case absOne:
		return absZero
	case absUnknown:
		return absUnknown
	default:
		panic("bmc: invalid ternary state")
	}
}

func ternaryAnd(left, right int32) int32 {
	if left < absZero || left > absUnknown || right < absZero || right > absUnknown {
		panic("bmc: invalid ternary state")
	}
	if left == absZero || right == absZero {
		return absZero
	}
	if left == absOne && right == absOne {
		return absOne
	}
	return absUnknown
}
