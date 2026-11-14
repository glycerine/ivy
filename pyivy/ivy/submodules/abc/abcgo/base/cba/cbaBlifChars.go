package cba

func blifBool(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

// BLIF whitespace deliberately excludes newline, which terminates a directive.
func prsBlifCharIsSpace(c byte) int32 { return blifBool(c == ' ' || c == '\t' || c == '\r') }
func Prs_CharIsStop(c byte) int32     { return blifBool(c == '#' || c == '\\' || c == '\n' || c == '=') }
func Prs_CharIsLit(c byte) int32      { return blifBool(c == '0' || c == '1' || c == '-') }
