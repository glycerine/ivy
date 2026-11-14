package cba

func prsCharResult(v bool) int32 {
	if v {
		return 1
	}
	return 0
}
func Prs_CharIsSpace(c byte) int32 {
	return prsCharResult(c == ' ' || c == '\t' || c == '\r' || c == '\n')
}
func Prs_CharIsDigit(c byte) int32 { return prsCharResult(c >= '0' && c <= '9') }
func Prs_CharIsDigitB(c byte) int32 {
	return prsCharResult(c == '0' || c == '1' || c == 'x' || c == 'z')
}
func Prs_CharIsDigitH(c byte) int32 {
	return prsCharResult(c >= '0' && c <= '9' || c >= 'A' && c <= 'F' || c >= 'a' && c <= 'f')
}
func Prs_CharIsChar(c byte) int32  { return prsCharResult(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z') }
func Prs_CharIsSymb1(c byte) int32 { return prsCharResult(Prs_CharIsChar(c) != 0 || c == '_') }
func Prs_CharIsSymb2(c byte) int32 {
	return prsCharResult(Prs_CharIsSymb1(c) != 0 || Prs_CharIsDigit(c) != 0 || c == '$')
}
