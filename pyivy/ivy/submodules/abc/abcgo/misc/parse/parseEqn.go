package parse

// ParseEquation accepts the original equation grammar: !, *, ^, + and
// parentheses, with explicit binary operators and variables in names order.
func ParseEquation(source string, names []string) (*Expression, error) {
	return parseFormula(source, names, true)
}
func CompileEquation[T any](source string, names []string, ops Operations[T]) (T, error) {
	e, err := ParseEquation(source, names)
	if err != nil {
		var zero T
		return zero, err
	}
	return Compile(e, ops, len(names))
}
