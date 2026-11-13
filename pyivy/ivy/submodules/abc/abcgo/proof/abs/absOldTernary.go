package abs

// Ternary and visitation encodings from absOldSim.c.
const (
	SAIG_ZER int32 = 1
	SAIG_ONE int32 = 2
	SAIG_UND int32 = 3

	SAIG_ZER_NEW int32 = 0
	SAIG_ONE_NEW int32 = 1
	SAIG_ZER_OLD int32 = 2
	SAIG_ONE_OLD int32 = 3
)

func Saig_ManSimInfoNot(value int32) int32 {
	switch value {
	case SAIG_ZER:
		return SAIG_ONE
	case SAIG_ONE:
		return SAIG_ZER
	default:
		return SAIG_UND
	}
}

func Saig_ManSimInfoAnd(a, b int32) int32 {
	if a == SAIG_ZER || b == SAIG_ZER {
		return SAIG_ZER
	}
	if a == SAIG_ONE && b == SAIG_ONE {
		return SAIG_ONE
	}
	return SAIG_UND
}

func Saig_ManSimInfo2IsOld(value int32) int32 {
	if value == SAIG_ZER_OLD || value == SAIG_ONE_OLD {
		return 1
	}
	return 0
}

func Saig_ManSimInfo2SetOld(value int32) int32 {
	switch value {
	case SAIG_ZER_NEW:
		return SAIG_ZER_OLD
	case SAIG_ONE_NEW:
		return SAIG_ONE_OLD
	default:
		panic("abstraction: marking an invalid or already visited value")
	}
}

func Saig_ManSimInfo2Not(value int32) int32 {
	if value < SAIG_ZER_NEW || value > SAIG_ONE_OLD {
		panic("abstraction: invalid visited simulation value")
	}
	return value ^ 1
}

func Saig_ManSimInfo2And(a, b int32) int32 {
	if a == SAIG_ZER_NEW || b == SAIG_ZER_NEW {
		return SAIG_ZER_NEW
	}
	if a == SAIG_ONE_NEW && b == SAIG_ONE_NEW {
		return SAIG_ONE_NEW
	}
	panic("abstraction: AND requires unvisited binary values")
}
