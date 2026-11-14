package cba

import "strings"

type sopInfo struct {
	kind  Cba_ObjType_t
	label string
}

var sourceSopTypes = map[string]sopInfo{
	" 0\n": {CBA_BOX_CF, "CBA_BOX_C0"}, " 1\n": {CBA_BOX_CT, "CBA_BOX_C1"},
	"1 1\n": {CBA_BOX_BUF, "CBA_BOX_BUF"}, "0 1\n": {CBA_BOX_INV, "CBA_BOX_INV"},
	"11 1\n": {CBA_BOX_AND, "CBA_BOX_AND"}, "00 1\n": {CBA_BOX_NOR, "CBA_BOX_NOR"},
	"00 0\n": {CBA_BOX_OR, "CBA_BOX_OR"}, "-1 1\n1- 1\n": {CBA_BOX_OR, "CBA_BOX_OR"}, "1- 1\n-1 1\n": {CBA_BOX_OR, "CBA_BOX_OR"},
	"01 1\n10 1\n": {CBA_BOX_XOR, "CBA_BOX_XOR"}, "10 1\n01 1\n": {CBA_BOX_XOR, "CBA_BOX_XOR"},
	"11 1\n00 1\n": {CBA_BOX_XNOR, "CBA_BOX_XNOR"}, "00 1\n11 1\n": {CBA_BOX_XNOR, "CBA_BOX_XNOR"},
	"10 1\n": {CBA_BOX_SHARP, "CBA_BOX_SHARP"}, "01 1\n": {CBA_BOX_SHARPL, "CBA_BOX_SHARPL"},
}

func sourceSop(s string) sopInfo {
	if i := strings.IndexByte(s, 0); i >= 0 {
		s = s[:i]
	}
	v, ok := sourceSopTypes[s]
	if !ok {
		panic("cba: unsupported SOP")
	}
	return v
}
func Ptr_SopToType(s string) Cba_ObjType_t { return sourceSop(s).kind }
func Ptr_SopToTypeName(s string) string    { return sourceSop(s).label }
func Ptr_TypeToName(kind Cba_ObjType_t) string {
	names := map[Cba_ObjType_t]string{CBA_BOX_CF: "const0", CBA_BOX_CT: "const1", CBA_BOX_CX: "constX", CBA_BOX_CZ: "constZ", CBA_BOX_BUF: "buf", CBA_BOX_INV: "not", CBA_BOX_AND: "and", CBA_BOX_NAND: "nand", CBA_BOX_OR: "or", CBA_BOX_NOR: "nor", CBA_BOX_XOR: "xor", CBA_BOX_XNOR: "xnor", CBA_BOX_MUX: "mux", CBA_BOX_MAJ: "maj", CBA_BOX_SHARP: "sharp", CBA_BOX_SHARPL: "sharpl", CBA_BOX_TRI: "bufifl"}
	v, ok := names[kind]
	if !ok {
		panic("cba: unsupported type name")
	}
	return v
}
func Ptr_TypeToSop(kind Cba_ObjType_t) string {
	sops := map[Cba_ObjType_t]string{CBA_BOX_CF: " 0\n", CBA_BOX_CT: " 1\n", CBA_BOX_CX: " 0\n", CBA_BOX_CZ: " 0\n", CBA_BOX_BUF: "1 1\n", CBA_BOX_INV: "0 1\n", CBA_BOX_AND: "11 1\n", CBA_BOX_NAND: "11 0\n", CBA_BOX_OR: "00 0\n", CBA_BOX_NOR: "00 1\n", CBA_BOX_XOR: "01 1\n10 1\n", CBA_BOX_XNOR: "00 1\n11 1\n", CBA_BOX_SHARP: "10 1\n", CBA_BOX_SHARPL: "01 1\n", CBA_BOX_MUX: "11- 1\n0-1 1\n", CBA_BOX_MAJ: "11- 1\n1-1 1\n-11 1\n"}
	v, ok := sops[kind]
	if !ok {
		panic("cba: unsupported type SOP")
	}
	return v
}
