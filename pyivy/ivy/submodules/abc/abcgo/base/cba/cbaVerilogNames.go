package cba

// verilogNameLegal mirrors the source's identifier check, including accepting
// an already escaped name before testing the reserved-name ID.
func verilogNameLegal(name string, nameID int32) int32 {
	if len(name) == 0 || name[0] == 0 {
		panic("cba: empty Verilog name")
	}
	if name[0] == '\\' {
		return 1
	}
	if nameID < 13 {
		return 0
	}
	letter := func(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_' }
	if !letter(name[0]) {
		return 0
	}
	for i := 1; i < len(name) && name[i] != 0; i++ {
		c := name[i]
		if !letter(c) && !(c >= '0' && c <= '9') && c != '$' {
			return 0
		}
	}
	return 1
}
func Prs_NameIsLegalInVerilog(name string, nameID int32) int32 { return verilogNameLegal(name, nameID) }
func Cba_NameIsLegalInVerilog(name string, nameID int32) int32 { return verilogNameLegal(name, nameID) }
