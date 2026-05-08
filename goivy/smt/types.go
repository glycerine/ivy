package smt

// CheckResult mirrors Z3_lbool.
type CheckResult int32

const (
	Unsat   CheckResult = -1
	Unknown CheckResult = 0
	Sat     CheckResult = 1
)

func (r CheckResult) String() string {
	switch r {
	case Sat:
		return "sat"
	case Unsat:
		return "unsat"
	default:
		return "unknown"
	}
}
