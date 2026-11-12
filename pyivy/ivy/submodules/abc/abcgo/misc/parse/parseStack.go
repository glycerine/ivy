package parse

// FunctionStack stores opaque formula backend nodes. Its capacity hint replaces
// the C fixed-depth allocation; Go slices grow without corrupting parser state.
type FunctionStack struct{ values []any }
type OperatorStack struct{ values []int }
type Parse_StackFn_t = FunctionStack
type Parse_StackOp_t = OperatorStack

func NewFunctionStack(capacity int) *FunctionStack {
	if capacity < 0 {
		panic("parse: negative stack capacity")
	}
	return &FunctionStack{values: make([]any, 0, capacity)}
}
func (s *FunctionStack) Empty() bool    { return len(s.values) == 0 }
func (s *FunctionStack) Push(value any) { s.values = append(s.values, value) }
func (s *FunctionStack) Pop() any {
	if s.Empty() {
		panic("parse: function stack underflow")
	}
	i := len(s.values) - 1
	v := s.values[i]
	s.values[i] = nil
	s.values = s.values[:i]
	return v
}
func (s *FunctionStack) Clear() { clear(s.values); s.values = s.values[:0] }
func NewOperatorStack(capacity int) *OperatorStack {
	if capacity < 0 {
		panic("parse: negative stack capacity")
	}
	return &OperatorStack{values: make([]int, 0, capacity)}
}
func (s *OperatorStack) Empty() bool    { return len(s.values) == 0 }
func (s *OperatorStack) Push(value int) { s.values = append(s.values, value) }
func (s *OperatorStack) Pop() int {
	if s.Empty() {
		panic("parse: operator stack underflow")
	}
	i := len(s.values) - 1
	v := s.values[i]
	s.values = s.values[:i]
	return v
}
func (s *OperatorStack) Clear()                 { s.values = s.values[:0] }
func Parse_StackFnStart(n int) *Parse_StackFn_t { return NewFunctionStack(n) }
func Parse_StackFnIsEmpty(s *Parse_StackFn_t) int {
	if s.Empty() {
		return 1
	}
	return 0
}
func Parse_StackFnPush(s *Parse_StackFn_t, v any) { s.Push(v) }
func Parse_StackFnPop(s *Parse_StackFn_t) any     { return s.Pop() }
func Parse_StackFnFree(s *Parse_StackFn_t)        { s.Clear(); s.values = nil }
func Parse_StackOpStart(n int) *Parse_StackOp_t   { return NewOperatorStack(n) }
func Parse_StackOpIsEmpty(s *Parse_StackOp_t) int {
	if s.Empty() {
		return 1
	}
	return 0
}
func Parse_StackOpPush(s *Parse_StackOp_t, v int) { s.Push(v) }
func Parse_StackOpPop(s *Parse_StackOp_t) int     { return s.Pop() }
func Parse_StackOpFree(s *Parse_StackOp_t)        { s.Clear(); s.values = nil }
