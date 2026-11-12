// Port of ABC src/aig/miniaig/abcOper.h (Alan Mishchenko, UC Berkeley).
package miniaig

const (
	ABC_OPER_NONE         = 0
	ABC_OPER_PI           = 1
	ABC_OPER_PO           = 2
	ABC_OPER_CI           = 3
	ABC_OPER_CO           = 4
	ABC_OPER_FON          = 5
	ABC_OPER_BOX          = 6
	ABC_OPER_CONST_F      = 7
	ABC_OPER_CONST_T      = 8
	ABC_OPER_CONST_X      = 9
	ABC_OPER_CONST_Z      = 10
	ABC_OPER_BIT_BUF      = 11
	ABC_OPER_BIT_INV      = 12
	ABC_OPER_BIT_AND      = 13
	ABC_OPER_BIT_NAND     = 14
	ABC_OPER_BIT_OR       = 15
	ABC_OPER_BIT_NOR      = 16
	ABC_OPER_BIT_XOR      = 17
	ABC_OPER_BIT_NXOR     = 18
	ABC_OPER_BIT_SHARP    = 19
	ABC_OPER_BIT_SHARPL   = 20
	ABC_OPER_BIT_MUX      = 21
	ABC_OPER_BIT_MAJ      = 22
	ABC_OPER_ABC          = 23
	ABC_OPER_BA           = 24
	ABC_OPER_BO           = 25
	ABC_OPER_BX           = 26
	ABC_OPER_BN           = 27
	ABC_OPER_BAO          = 28
	ABC_OPER_BOA          = 29
	ABC_OPER_RED_AND      = 30
	ABC_OPER_RED_NAND     = 31
	ABC_OPER_RED_OR       = 32
	ABC_OPER_RED_NOR      = 33
	ABC_OPER_RED_XOR      = 34
	ABC_OPER_RED_NXOR     = 35
	ABC_OPER_LOGIC_NOT    = 36
	ABC_OPER_LOGIC_AND    = 37
	ABC_OPER_LOGIC_NAND   = 38
	ABC_OPER_LOGIC_OR     = 39
	ABC_OPER_LOGIC_NOR    = 40
	ABC_OPER_LOGIC_XOR    = 41
	ABC_OPER_LOGIC_XNOR   = 42
	ABC_OPER_SEL_NMUX     = 43
	ABC_OPER_SEL_SEL      = 44
	ABC_OPER_SEL_PSEL     = 45
	ABC_OPER_SEL_ENC      = 46
	ABC_OPER_SEL_PENC     = 47
	ABC_OPER_SEL_DEC      = 48
	ABC_OPER_SEL_EDEC     = 49
	ABC_OPER_ARI_ADD      = 50
	ABC_OPER_ARI_SUB      = 51
	ABC_OPER_ARI_MUL      = 52
	ABC_OPER_ARI_SMUL     = 53
	ABC_OPER_ARI_DIV      = 54
	ABC_OPER_ARI_MOD      = 55
	ABC_OPER_ARI_REM      = 56
	ABC_OPER_ARI_POW      = 57
	ABC_OPER_ARI_MIN      = 58
	ABC_OPER_ARI_SQRT     = 59
	ABC_OPER_ARI_ABS      = 60
	ABC_OPER_COMP_SLESS   = 61
	ABC_OPER_COMP_LESS    = 62
	ABC_OPER_COMP_LESSEQU = 63
	ABC_OPER_COMP_MOREEQU = 64
	ABC_OPER_COMP_MORE    = 65
	ABC_OPER_COMP_EQU     = 66
	ABC_OPER_COMP_NOTEQU  = 67
	ABC_OPER_SHIFT_L      = 68
	ABC_OPER_SHIFT_R      = 69
	ABC_OPER_SHIFT_LA     = 70
	ABC_OPER_SHIFT_RA     = 71
	ABC_OPER_SHIFT_ROTL   = 72
	ABC_OPER_SHIFT_ROTR   = 73
	ABC_OPER_NODE         = 74
	ABC_OPER_LUT          = 75
	ABC_OPER_GATE         = 76
	ABC_OPER_TABLE        = 77
	ABC_OPER_TRI          = 78
	ABC_OPER_RAM          = 79
	ABC_OPER_RAMR         = 80
	ABC_OPER_RAMW         = 81
	ABC_OPER_RAMWC        = 82
	ABC_OPER_RAML         = 83
	ABC_OPER_RAMS         = 84
	ABC_OPER_RAMBOX       = 85
	ABC_OPER_LATCH        = 86
	ABC_OPER_LATCHRS      = 87
	ABC_OPER_DFF          = 88
	ABC_OPER_DFFRSE       = 89
	ABC_OPER_DFFLAST      = 90
	ABC_OPER_SLICE        = 91
	ABC_OPER_CONCAT       = 92
	ABC_OPER_ZEROPAD      = 93
	ABC_OPER_SIGNEXT      = 94
	ABC_OPER_LOGIC_IMPL   = 95
	ABC_OPER_ARI_SQUARE   = 96
	ABC_OPER_CONST        = 97
	ABC_OPER_ARI_ADDSUB   = 98
	ABC_OPER_LAST         = 99
)

func Abc_OperName(t int) string {
	switch t {
	case ABC_OPER_NONE:
		return ""
	case ABC_OPER_PI:
		return "pi"
	case ABC_OPER_PO:
		return "po"
	case ABC_OPER_CI:
		return "ci"
	case ABC_OPER_CO:
		return "co"
	case ABC_OPER_FON:
		return "fon"
	case ABC_OPER_BOX:
		return "box"
	case ABC_OPER_BIT_BUF:
		return "buf"
	case ABC_OPER_BIT_INV:
		return "~"
	case ABC_OPER_BIT_MUX:
		return "mux"
	case ABC_OPER_BIT_MAJ:
		return "maj"
	case ABC_OPER_BIT_AND:
		return "&"
	case ABC_OPER_BIT_OR:
		return "|"
	case ABC_OPER_BIT_XOR:
		return "^"
	case ABC_OPER_BIT_NAND:
		return "~&"
	case ABC_OPER_BIT_NOR:
		return "~|"
	case ABC_OPER_BIT_NXOR:
		return "~^"
	case ABC_OPER_RED_AND:
		return "&"
	case ABC_OPER_RED_OR:
		return "|"
	case ABC_OPER_RED_XOR:
		return "^"
	case ABC_OPER_RED_NAND:
		return "~&"
	case ABC_OPER_RED_NOR:
		return "~|"
	case ABC_OPER_RED_NXOR:
		return "~^"
	case ABC_OPER_LOGIC_NOT:
		return "!"
	case ABC_OPER_LOGIC_IMPL:
		return "=>"
	case ABC_OPER_LOGIC_AND:
		return "&&"
	case ABC_OPER_LOGIC_OR:
		return "||"
	case ABC_OPER_LOGIC_XOR:
		return "^^"
	case ABC_OPER_ARI_ADD:
		return "+"
	case ABC_OPER_ARI_SUB:
		return "-"
	case ABC_OPER_ARI_MUL:
		return "*"
	case ABC_OPER_ARI_SMUL:
		return "*"
	case ABC_OPER_ARI_DIV:
		return "/"
	case ABC_OPER_ARI_REM:
		return "%"
	case ABC_OPER_ARI_MOD:
		return "mod"
	case ABC_OPER_ARI_POW:
		return "**"
	case ABC_OPER_ARI_MIN:
		return "-"
	case ABC_OPER_ARI_SQRT:
		return "sqrt"
	case ABC_OPER_ARI_SQUARE:
		return "squar"
	case ABC_OPER_COMP_EQU:
		return "=="
	case ABC_OPER_COMP_NOTEQU:
		return "!="
	case ABC_OPER_COMP_LESS:
		return "<"
	case ABC_OPER_COMP_MORE:
		return ">"
	case ABC_OPER_COMP_LESSEQU:
		return "<="
	case ABC_OPER_COMP_MOREEQU:
		return ">="
	case ABC_OPER_SHIFT_L:
		return "<<"
	case ABC_OPER_SHIFT_R:
		return ">>"
	case ABC_OPER_SHIFT_LA:
		return "<<<"
	case ABC_OPER_SHIFT_RA:
		return ">>>"
	case ABC_OPER_SHIFT_ROTL:
		return "rotL"
	case ABC_OPER_SHIFT_ROTR:
		return "rotR"
	case ABC_OPER_DFFRSE:
		return "DFFRSE"
	case ABC_OPER_SLICE:
		return "[:]"
	case ABC_OPER_CONCAT:
		return "{}"
	case ABC_OPER_ZEROPAD:
		return "zPad"
	case ABC_OPER_SIGNEXT:
		return "sExt"
	case ABC_OPER_CONST:
		return "const"
	case ABC_OPER_TABLE:
		return "table"
	case ABC_OPER_LUT:
		return "lut"
	case ABC_OPER_LAST:
		return ""
	default:
		panic("unknown ABC operator")
	}
}
