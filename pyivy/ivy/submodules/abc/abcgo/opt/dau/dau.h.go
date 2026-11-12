package dau

// Public DAU constants and DSD-string predicates from dau.h. Strings use Go's
// explicit length instead of relying on an accessible trailing NUL byte.
const (
	DAU_MAX_VAR  = 12
	DAU_MAX_STR  = 2000
	DAU_MAX_WORD = 1 << (DAU_MAX_VAR - 6)
)

type Dau_DsdType_t int32

const (
	DAU_DSD_NONE Dau_DsdType_t = iota
	DAU_DSD_CONST0
	DAU_DSD_VAR
	DAU_DSD_AND
	DAU_DSD_XOR
	DAU_DSD_MUX
	DAU_DSD_PRIME
)

type TtCanonicizeFunc func(*Abc_TtHieMan_t, []uint64, int32, []byte, int32) uint32

func dauBool(b bool) int32 {
	if b {
		return 1
	}
	return 0
}
func Dau_DsdIsConst(p string) int32  { return dauBool(p == "0" || p == "1") }
func Dau_DsdIsConst0(p string) int32 { return dauBool(p == "0") }
func Dau_DsdIsConst1(p string) int32 { return dauBool(p == "1") }
func Dau_DsdIsVar(p string) int32 {
	if len(p) > 0 && p[0] == '!' {
		p = p[1:]
	}
	return dauBool(len(p) > 0 && p[0] >= 'a' && p[0] <= 'z')
}
func Dau_DsdReadVar(p string) int32 {
	if len(p) > 0 && p[0] == '!' {
		p = p[1:]
	}
	if len(p) == 0 {
		panic("DAU variable is empty")
	}
	return int32(p[0]) - 'a'
}
