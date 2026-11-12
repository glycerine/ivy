package gia

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
)

// rexSymbol supplies C's terminating NUL at the end of a Go string.
func rexSymbol(s string, i int32) byte {
	if i < 0 || int(i) > len(s) {
		panic("gia: invalid regex state")
	}
	if int(i) == len(s) {
		return 0
	}
	return s[i]
}
func Gia_ManPrintAutom(s string, transitions *Vec_Int_t) {
	for i := range s {
		fmt.Printf("%d '%c' ", i, s[i])
		next := vec.Vec_IntEntry(transitions, i)
		if next >= 0 {
			fmt.Printf("-> %d '%c' ", next, rexSymbol(s, next))
		}
		fmt.Println()
	}
}
func Gia_ManPrintReached(s string, state int32, reached *Vec_Int_t) {
	fmt.Printf("Reached from state %d '%c':  ", state, rexSymbol(s, state))
	for _, id := range vec.Vec_IntArray(reached) {
		fmt.Printf("%d '%c'  ", id, rexSymbol(s, id))
	}
	fmt.Println()
}
func Gia_ManPrintReached_rec(s string, transitions *Vec_Int_t, state int32, reached, visited *Vec_Int_t, travID int32) {
	if vec.Vec_IntEntry(visited, int(state)) == travID {
		return
	}
	vec.Vec_IntWriteEntry(visited, int(state), travID)
	c := rexSymbol(s, state)
	if Gia_SymbSpecial(c) == 0 {
		vec.Vec_IntPush(reached, state)
	}
	if c == 0 {
		return
	}
	if Gia_SymbSpecial(c) != 0 && c != '|' {
		Gia_ManPrintReached_rec(s, transitions, state+1, reached, visited, travID)
	}
	next := vec.Vec_IntEntry(transitions, int(state))
	if next >= 0 {
		Gia_ManPrintReached_rec(s, transitions, next, reached, visited, travID)
	}
}
func Gia_ManCollectReached(s string, transitions *Vec_Int_t, state int32, reached, visited *Vec_Int_t, travID int32) {
	if state != 0 && Gia_SymbSpecial(rexSymbol(s, state)) != 0 {
		panic("gia: regex starting state is special")
	}
	if vec.Vec_IntEntry(visited, int(state)) == travID {
		panic("gia: regex traversal ID reused")
	}
	vec.Vec_IntClear(reached)
	Gia_ManPrintReached_rec(s, transitions, state+1, reached, visited, travID)
}

// rexTransitions links alternation and Kleene-star epsilon edges.
func rexTransitions(s string) *Vec_Int_t {
	transitions := vec.Vec_IntStartFull(len(s))
	stack := make([]int, 0, len(s))
	pop := func() int {
		if len(stack) == 0 {
			panic("gia: unbalanced regex")
		}
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		return n
	}
	for i := range s {
		lp := i
		switch s[i] {
		case '(', '|':
			stack = append(stack, i)
		case ')':
			op := pop()
			if s[op] == '|' {
				lp = pop()
				if s[lp] != '(' {
					panic("gia: invalid regex alternation")
				}
				vec.Vec_IntWriteEntry(transitions, lp, int32(op+1))
				vec.Vec_IntWriteEntry(transitions, op, int32(i))
			} else {
				if s[op] != '(' {
					panic("gia: invalid regex group")
				}
				lp = op
			}
		}
		if i+1 < len(s) && s[i+1] == '*' {
			vec.Vec_IntWriteEntry(transitions, lp, int32(i+1))
			vec.Vec_IntWriteEntry(transitions, i+1, int32(lp))
		}
	}
	if len(stack) != 0 {
		panic("gia: unbalanced regex")
	}
	return transitions
}
