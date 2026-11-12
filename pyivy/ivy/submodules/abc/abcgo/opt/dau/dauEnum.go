package dau

import (
	"fmt"
	"github.com/glycerine/abcgo/misc/vec"
	"os"
	"strings"
)

func Dau_EnumLift(name string, shift int32) string {
	data := []byte(name)
	for i, c := range data {
		if c >= 'a' && c <= 'z' {
			data[i] = byte(int32(c) + shift)
		}
	}
	return string(data)
}
func Dau_EnumLift2(name string, shift int32) string { return Dau_EnumLift(name, shift) }
func enumOperand(name string, complement int32) string {
	name = strings.TrimPrefix(name, "*")
	if complement != 0 {
		return "!" + name
	}
	return name
}
func Dau_EnumCombineTwo(out *vec.Vec_Ptr_t, star, xor int32, a, b string, shift, negA, negB int32) {
	open, close := "(", ")"
	if xor != 0 {
		open, close = "[", "]"
	}
	prefix := ""
	if star != 0 {
		prefix = "*"
	}
	vec.Vec_PtrPush(out, prefix+open+enumOperand(a, negA)+enumOperand(Dau_EnumLift(b, shift), negB)+close)
}
func Dau_EnumCombineThree(out *vec.Vec_Ptr_t, star int32, c, a, b string, shiftA, shiftB, negC, negA, negB int32) {
	prefix := ""
	if star != 0 {
		prefix = "*"
	}
	vec.Vec_PtrPush(out, prefix+"<"+enumOperand(c, negC)+enumOperand(Dau_EnumLift(a, shiftA), negA)+enumOperand(Dau_EnumLift2(b, shiftB), negB)+">")
}
func Dau_EnumTestDump(sets *vec.Vec_Ptr_t, name string) {
	f, err := os.Create(name)
	if err != nil {
		return
	}
	defer f.Close()
	for v, entry := range vec.Vec_PtrArray(sets) {
		if _, err := fmt.Fprintf(f, "VARIABLE NUMBER %d:\n", v); err != nil {
			panic(err)
		}
		for _, name := range vec.Vec_PtrArray(entry.(*vec.Vec_Ptr_t)) {
			if _, err := fmt.Fprintln(f, name.(string)); err != nil {
				panic(err)
			}
		}
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
}

// EnumerateDSD preserves the source's partition and complement-pruning order.
func EnumerateDSD(maxVars int32) *vec.Vec_Ptr_t {
	if maxVars < 1 || maxVars > 26 {
		panic("dau: DSD enumeration requires1..26variables")
	}
	sets := vec.Vec_PtrAlloc(int(maxVars) + 1)
	vec.Vec_PtrPush(sets, vec.Vec_PtrAlloc(0))
	one := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(one, "*a")
	vec.Vec_PtrPush(sets, one)
	entry := func(i int32) *vec.Vec_Ptr_t { return vec.Vec_PtrEntry(sets, int(i)).(*vec.Vec_Ptr_t) }
	for v := int32(2); v <= maxVars; v++ {
		out := vec.Vec_PtrAlloc(100)
		for i := int32(1); i < v; i++ {
			for j := i; j < v; j++ {
				if i+j != v {
					continue
				}
				for ai, a0 := range vec.Vec_PtrArray(entry(i)) {
					a := a0.(string)
					for bi, b0 := range vec.Vec_PtrArray(entry(j)) {
						b := b0.(string)
						Dau_EnumCombineTwo(out, 0, 0, a, b, i, 0, 0)
						if a[0] != '*' {
							Dau_EnumCombineTwo(out, 0, 0, a, b, i, 1, 0)
						}
						if b[0] != '*' && !(i == j && ai == bi) {
							Dau_EnumCombineTwo(out, 0, 0, a, b, i, 0, 1)
						}
						if a[0] != '*' && b[0] != '*' {
							Dau_EnumCombineTwo(out, 0, 0, a, b, i, 1, 1)
						}
						Dau_EnumCombineTwo(out, dauBool(a[0] == '*' || b[0] == '*'), 1, a, b, i, 0, 0)
					}
				}
			}
		}
		for k := int32(1); k < v; k++ {
			for i := int32(1); i < v; i++ {
				for j := i; j < v; j++ {
					if k+i+j != v {
						continue
					}
					for _, c0 := range vec.Vec_PtrArray(entry(k)) {
						c := c0.(string)
						for ai, a0 := range vec.Vec_PtrArray(entry(i)) {
							a := a0.(string)
							for bi, b0 := range vec.Vec_PtrArray(entry(j)) {
								b := b0.(string)
								star := dauBool(a[0] == '*' && b[0] == '*')
								Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 0, 0, 0)
								if a[0] != '*' {
									Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 0, 1, 0)
								}
								if b[0] != '*' && !(i == j && ai == bi) {
									Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 0, 0, 1)
								}
								if c[0] != '*' && !(i == j && ai == bi) {
									Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 1, 0, 0)
									if a[0] != '*' {
										Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 1, 1, 0)
									}
									if b[0] != '*' {
										Dau_EnumCombineThree(out, star, c, a, b, k, k+i, 1, 0, 1)
									}
								}
							}
						}
					}
				}
			}
		}
		vec.Vec_PtrPush(sets, out)
	}
	return sets
}
func Dau_EnumTest() {
	sets := EnumerateDSD(10)
	Dau_EnumTestDump(sets, "_npn/npn/dsd10.txt")
	for v, entry := range vec.Vec_PtrArray(sets) {
		if v > 4 {
			break
		}
		fmt.Printf("VARIABLE NUMBER %d:\n", v)
		for _, name := range vec.Vec_PtrArray(entry.(*vec.Vec_Ptr_t)) {
			fmt.Println(name)
		}
	}
	for v, entry := range vec.Vec_PtrArray(sets) {
		fmt.Printf("%d=%d ", v, vec.Vec_PtrSize(entry.(*vec.Vec_Ptr_t)))
	}
	fmt.Println()
}
