package startingpoint

import (
	"strconv"
	"strings"
)

var mathCalls = map[string]string{
	"acos": "Acos", "asin": "Asin", "atan": "Atan", "atan2": "Atan2", "cos": "Cos", "sin": "Sin", "tan": "Tan",
	"acosh": "Acosh", "asinh": "Asinh", "atanh": "Atanh", "cosh": "Cosh", "sinh": "Sinh", "tanh": "Tanh",
	"exp": "Exp", "exp2": "Exp2", "expm1": "Expm1", "log": "Log", "log10": "Log10", "log1p": "Log1p", "log2": "Log2",
	"sqrt": "Sqrt", "cbrt": "Cbrt", "pow": "Pow", "hypot": "Hypot", "ceil": "Ceil", "floor": "Floor", "trunc": "Trunc", "round": "Round",
	"fabs": "Abs", "fmod": "Mod", "remainder": "Remainder", "copysign": "Copysign", "erf": "Erf", "erfc": "Erfc", "tgamma": "Gamma",
}
var asciiPredicates = map[string]string{
	"isdigit":  "spC >= '0' && spC <= '9'",
	"isalpha":  "(spC >= 'a' && spC <= 'z') || (spC >= 'A' && spC <= 'Z')",
	"isalnum":  "(spC >= 'a' && spC <= 'z') || (spC >= 'A' && spC <= 'Z') || (spC >= '0' && spC <= '9')",
	"isblank":  "spC == ' ' || spC == '\\t'",
	"isspace":  "spC == ' ' || (spC >= '\\t' && spC <= '\\r')",
	"iscntrl":  "(spC >= 0 && spC < 32) || spC == 127",
	"isgraph":  "spC >= 33 && spC <= 126",
	"islower":  "spC >= 'a' && spC <= 'z'",
	"isupper":  "spC >= 'A' && spC <= 'Z'",
	"isprint":  "spC >= 32 && spC <= 126",
	"ispunct":  "(spC >= 33 && spC <= 47) || (spC >= 58 && spC <= 64) || (spC >= 91 && spC <= 96) || (spC >= 123 && spC <= 126)",
	"isxdigit": "(spC >= '0' && spC <= '9') || (spC >= 'a' && spC <= 'f') || (spC >= 'A' && spC <= 'F')",
}

func (p *libraryPass) convertNumeric(c libraryCall) (string, string, bool) {
	a := c.args
	if strings.HasSuffix(c.name, "f") && p.rules[c.name].Header == "math.h" {
		base := strings.TrimSuffix(c.name, "f")
		if p.rules[base].Header == "math.h" {
			inner := c
			inner.name = base
			result, note, ok := p.convertNumeric(inner)
			if ok {
				return "float32(" + result + ")", note + " Float result rounded to float32; verify exact rounding requirements.", true
			}
			return result, note, ok
		}
	}
	if target, ok := mathCalls[c.name]; ok {
		n := 1
		switch c.name {
		case "atan2", "pow", "hypot", "fmod", "remainder", "copysign":
			n = 2
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		args := make([]string, n)
		for i := range args {
			args[i] = "float64(" + a[i] + ")"
		}
		return p.imports.use("math") + target + "(" + strings.Join(args, ", ") + ")", "Double-precision math mapping; C errno/fenv and domain handling remain review obligations.", true
	}
	if condition, ok := asciiPredicates[c.name]; ok {
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		// The template strings encode character escapes, not literal whitespace.
		condition = strings.ReplaceAll(condition, `\\`, `\`)
		return iife("spC int32", "int32", "if "+condition+" { return 1 }; return 0", "int32("+a[0]+")"), "ASCII C-locale classification; non-C locale behavior must be selected explicitly.", true
	}
	switch c.name {
	case "atoi", "atol", "atoll":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		sc := p.imports.use("strconv")
		if c.name == "atoi" {
			return sc + "Atoi(" + a[0] + ")", "Handle Go value/error results, C integer width, whitespace and trailing-text semantics downstream.", true
		}
		return sc + "ParseInt(" + a[0] + ", 10, 64)", "Handle Go value/error results and C prefix/whitespace semantics downstream.", true

	case "abs", "labs", "llabs":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		typ := "int32"
		if c.name != "abs" {
			typ = "int64"
		}
		return iife("spValue "+typ, typ, "if spValue < 0 { return -spValue }; return spValue", typ+"("+a[0]+")"), "Integer absolute value; minimum signed value remains outside the defined C domain.", true
	case "tolower", "toupper":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		body := "if spC >= 'A' && spC <= 'Z' { return spC + ('a' - 'A') }; return spC"
		if c.name == "toupper" {
			body = "if spC >= 'a' && spC <= 'z' { return spC - ('a' - 'A') }; return spC"
		}
		return iife("spC int32", "int32", body, "int32("+a[0]+")"), "ASCII C-locale case conversion; EOF/other values are preserved.", true
	case "fmin", "fmax", "fdim":
		if len(a) != 2 {
			return "", arity(c, 2), false
		}
		ma := p.imports.use("math")
		body := ""
		if c.name == "fdim" {
			body = "if " + ma + "IsNaN(spA) || " + ma + "IsNaN(spB) { return " + ma + "NaN() }; if spA > spB { return spA-spB }; return 0"
		} else {
			target := "Min"
			if c.name == "fmax" {
				target = "Max"
			}
			body = "if " + ma + "IsNaN(spA) { return spB }; if " + ma + "IsNaN(spB) { return spA }; return " + ma + target + "(spA, spB)"
		}
		return iife("spA, spB float64", "float64", body, "float64("+a[0]+")", "float64("+a[1]+")"), "Double math with explicit C NaN handling; errno/fenv still require review.", true
	case "lgamma":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		return iife("spX float64", "float64", "spValue, _ := "+p.imports.use("math")+"Lgamma(spX); return spValue", "float64("+a[0]+")"), "Log-gamma value only; any signgam/global sign use requires repair.", true
	case "exit", "_Exit":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		return p.imports.use("os") + "Exit(int(" + a[0] + "))", "os.Exit does not run Go defers or C atexit handlers; review cleanup/buffering.", true
	case "getenv":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		text := a[0]
		if v, ok := p.text(text); ok {
			text = v
		}
		return p.imports.use("os") + "Getenv(" + text + ")", "Missing environment values now return empty string, not NULL. Use LookupEnv where absence matters.", true
	case "assert":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		condition := booleanExpression(a[0], p.types, p.returnTypes())
		// pthread lock/unlock now return no status. A direct success assertion
		// around these operations becomes the operation itself, evaluated once.
		if operation, ok := p.mutexAssertion(c.rawArgs[0], a[0]); ok {
			return operation, "Go mutex methods have no integer failure status; operation retained once.", true
		}

		message := strconv.Quote(strings.TrimSpace(c.rawArgs[0]))
		return "assert(" + condition + ", " + message + ")", "Calls the package assert helper with condition text. Repair C truth values downstream; C NDEBUG/abort semantics differ.", true

	}
	return "", "", false
}
