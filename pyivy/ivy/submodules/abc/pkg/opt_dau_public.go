package abc

// Public DAU bounds and checked DSD token helpers from opt/dau/dau.h.

const (
	DauMaxVariables    = 12
	DauMaxFormulaBytes = 2000
	DauMaxTruthWords   = 1 << (DauMaxVariables - 6)
)

func DauDSDIsConstant(text string) bool {
	return text == "0" || text == "1"
}

func DauDSDIsConstantZero(text string) bool { return text == "0" }

func DauDSDIsConstantOne(text string) bool { return text == "1" }

// DauDSDIsVariable preserves the header predicate: it checks the first token,
// after one optional complement, rather than requiring the complete string to
// contain only that variable.
func DauDSDIsVariable(text string) bool {
	if len(text) != 0 && text[0] == '!' {
		text = text[1:]
	}
	return len(text) != 0 && text[0] >= 'a' && text[0] <= 'z'
}

// DauDSDReadVariable is the checked spelling of Dau_DsdReadVar.
func DauDSDReadVariable(text string) (int, error) {
	if len(text) != 0 && text[0] == '!' {
		text = text[1:]
	}
	if len(text) == 0 || text[0] < 'a' || text[0] > 'z' {
		return 0, ErrDimension
	}
	return int(text[0] - 'a'), nil
}
