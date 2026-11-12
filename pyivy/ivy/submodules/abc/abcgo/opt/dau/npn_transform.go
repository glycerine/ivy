package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"io"
)

type Dtt_FunImpl_t struct{ Type, NP1, FI1, NP2, FI2 int32 }

func Dtt_FunImplFI2Str(fi, np int32, functions *vec.Vec_Int_t) string {
	var p, n [5]int32
	for i := range p {
		p[i] = np & 7
		np >>= 3
		n[i] = np & 1
		np >>= 1
	}
	return fmt.Sprintf("[%08x(%03d),%d%d%d%d%d,%d%d%d%d%d]", uint32(vec.Vec_IntEntry(functions, int(fi))), fi, p[0], p[1], p[2], p[3], p[4], n[0], n[1], n[2], n[3], n[4])
}
func Dtt_FunImpl2Str(kind int32, a, b string) string {
	switch kind {
	case 0:
		return "(" + a + "&" + b + ")"
	case 1:
		return "(~" + a + "&" + b + ")"
	case 2:
		return "(" + a + "&~" + b + ")"
	case 3:
		return "~(~" + a + "&~" + b + ")"
	case 4:
		return "(" + a + "^" + b + ")"
	case 5:
		return "~(" + a + "&" + b + ")"
	case 6:
		return "~(~" + a + "&" + b + ")"
	case 7:
		return "~(" + a + "&~" + b + ")"
	case 8:
		return "(~" + a + "&~" + b + ")"
	case 9:
		return "~(" + a + "^" + b + ")"
	}
	panic("dau: invalid implementation type")
}
func Dtt_ComposeNP(first, second int32) int32 {
	np := int32(0)
	for i := 0; i < 5; i++ {
		shift := (second & 7) * 4
		np |= (first >> shift & 7) << uint(4*i)
		np |= ((first>>shift ^ second) & 8) << uint(4*i)
		second >>= 4
	}
	return np
}
func Dtt_MakePI(np int32) string {
	for i := 0; i < 5; i++ {
		if np&7 == 0 {
			s := string(byte('a' + i))
			if np&8 != 0 {
				s = "~" + s
			}
			return s
		}
		np >>= 4
	}
	panic("dau: NP has no primary input")
}
func Dtt_ProcessType(kind *int32, fanin int32) {
	if fanin == 3 {
		if *kind < 5 {
			*kind += 5
		} else {
			*kind -= 5
		}
		return
	}
	if fanin != 1 && fanin != 2 {
		panic("dau: invalid fanin phase")
	}
	switch {
	case *kind == 0 || *kind == 5:
		*kind += fanin
	case *kind == fanin:
		*kind = 0
	case *kind+fanin == 3:
		*kind = 8
	case *kind == 3:
		if fanin == 1 {
			*kind = 7
		} else {
			*kind = 6
		}
	case *kind == 4:
		*kind = 9
	case *kind == fanin+5:
		*kind = 5
	case *kind+fanin == 8:
		*kind = 3
	case *kind == 8:
		if fanin == 1 {
			*kind = 2
		} else {
			*kind = 1
		}
	case *kind == 9:
		*kind = 4
	default:
		panic("dau: invalid implementation type")
	}
}
func Dtt_Check(fun, goal, current uint32, kind *int32) int32 {
	if goal == 0 {
		return dauBool(current == fun || ^current == fun)
	}
	var positive, negative uint32
	switch *kind {
	case 0, 5:
		positive, negative = current&fun, ^current&fun
	case 1, 6:
		positive, negative = ^current&fun, current&fun
	case 2, 7:
		positive, negative = current&^fun, ^current&^fun
	case 3, 8:
		positive, negative = current|fun, ^current|fun
	case 4, 9:
		positive, negative = current^fun, ^current^fun
	default:
		panic("dau: invalid implementation type")
	}
	if negative == goal {
		Dtt_ProcessType(kind, 1)
		return 1
	}
	return dauBool(positive == goal)
}
func Dtt_MakeFormula(fun uint32, impl *Dtt_FunImpl_t, library [][]*Dtt_FunImpl_t, np int32, print bool, w io.Writer) string {
	var result string
	if impl.FI1 == 0 {
		return Dtt_MakeFormulaFI2(fun, impl, library, np, Dtt_MakePI(Dtt_ComposeNP(impl.NP1, np)), print, w)
	}
	for _, first := range library[impl.FI1] {
		a := Dtt_MakeFormula(fun, first, library, Dtt_ComposeNP(impl.NP1, np), false, w)
		result = Dtt_MakeFormulaFI2(fun, impl, library, np, a, print, w)
	}
	return result
}
func Dtt_MakeFormulaFI2(fun uint32, impl *Dtt_FunImpl_t, library [][]*Dtt_FunImpl_t, np int32, a string, print bool, w io.Writer) string {
	emit := func(b string) string {
		s := Dtt_FunImpl2Str(impl.Type, a, b)
		if print {
			if _, err := fmt.Fprintf(w, "%08x = %s\n", fun, s); err != nil {
				panic(err)
			}
		}
		return s
	}
	if impl.FI2 == 0 {
		return emit(Dtt_MakePI(Dtt_ComposeNP(impl.NP2, np)))
	}
	var result string
	for _, second := range library[impl.FI2] {
		result = emit(Dtt_MakeFormula(fun, second, library, Dtt_ComposeNP(impl.NP2, np), false, w))
	}
	return result
}
