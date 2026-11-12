package parse

import "testing"

func TestFunctionStackLifecycle(t *testing.T) {
	s := Parse_StackFnStart(1)
	for i := 0; i < 2000; i++ {
		Parse_StackFnPush(s, i)
	}
	for i := 1999; i >= 0; i-- {
		if got := Parse_StackFnPop(s); got != i {
			t.Fatal(i, got)
		}
	}
	if Parse_StackFnIsEmpty(s) != 1 {
		t.Fatal("not empty")
	}
	Parse_StackFnPush(s, new(int))
	Parse_StackFnFree(s)
	if s.values != nil {
		t.Fatal("retained backing store")
	}
}
func TestOperatorStackLifecycle(t *testing.T) {
	s := Parse_StackOpStart(0)
	for i := 0; i < 2000; i++ {
		s.Push(i)
	}
	for i := 1999; i >= 0; i-- {
		if s.Pop() != i {
			t.Fatal(i)
		}
	}
	s.Push(42)
	s.Clear()
	if !s.Empty() {
		t.Fatal("clear")
	}
}
func TestStackUnderflow(t *testing.T) {
	for _, pop := range []func(){func() { NewFunctionStack(0).Pop() }, func() { NewOperatorStack(0).Pop() }} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("missing underflow panic")
				}
			}()
			pop()
		}()
	}
}
