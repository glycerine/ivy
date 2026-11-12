package gia

import (
	"fmt"
	"math/bits"
	"os"
)

func Gia_EnumFirstUnused(used []int32, nVars int32) int32 {
	for i := int32(0); i < nVars; i++ {
		if used[i] == 0 {
			return i
		}
	}
	return -1
}

// Gia_EnumPerms_rec enumerates perfect matchings by pairing the first unused
// element with each remaining element; pair order therefore stays canonical.
func Gia_EnumPerms_rec(used []int32, nVars int32, permutation []int32, nPerm int32, count *int32, file *os.File, nLogVars int32) {
	if nVars < 0 || nVars%2 != 0 || nPerm < 0 || nPerm > nVars {
		panic("gia: invalid matching dimensions")
	}
	if nPerm == nVars {
		if file != nil {
			for i := int32(0); i < nLogVars; i++ {
				fmt.Fprintf(file, "%c", '0'+((*count>>uint(nLogVars-1-i))&1))
			}
			fmt.Fprint(file, " ")
			for i := int32(0); i < nVars; i++ {
				for k := int32(0); k < nVars; k++ {
					bit := byte('0')
					if permutation[i] == k {
						bit = '1'
					}
					fmt.Fprintf(file, "%c", bit)
				}
			}
			fmt.Fprintln(file)
		} else if *count < 20 {
			fmt.Printf("%5d : ", *count)
			for i := int32(0); i < nVars; i += 2 {
				fmt.Printf("%d %d  ", permutation[i], permutation[i+1])
			}
			fmt.Println()
		}
		*count++
		return
	}
	first := Gia_EnumFirstUnused(used, nVars)
	if first < 0 {
		panic("gia: no unused matching element")
	}
	permutation[nPerm] = first
	used[first] = 1
	for i := int32(0); i < nVars; i++ {
		if used[i] != 0 {
			continue
		}
		permutation[nPerm+1] = i
		used[i] = 1
		Gia_EnumPerms_rec(used, nVars, permutation, nPerm+2, count, file, nLogVars)
		used[i] = 0
	}
	used[first] = 0
}
func Gia_EnumPerms(nVars int32) {
	if nVars < 0 || nVars%2 != 0 {
		panic("gia: matching size must be nonnegative and even")
	}
	used, permutation := make([]int32, nVars), make([]int32, nVars)
	count := int32(0)
	fmt.Printf("Printing sets of pairs for %d objects:\n", nVars)
	Gia_EnumPerms_rec(used, nVars, permutation, 0, &count, nil, -1)
	if count > 20 {
		fmt.Println("...")
	}
	fmt.Printf("Finished enumerating %d sets of pairs.\n", count)
	nLogVars := int32(bits.Len32(uint32(count - 1)))
	fmt.Printf("Need %d variables to encode %d sets.\n", nLogVars, count)
	file, err := os.Create("pairset.pla")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fmt.Fprintf(file, ".i %d\n.o %d\n", nLogVars, nVars*nVars)
	count = 0
	Gia_EnumPerms_rec(used, nVars, permutation, 0, &count, file, nLogVars)
	fmt.Fprintln(file, ".e")
	fmt.Printf("Finished dumping file %q.\n", "pairset.pla")
}
