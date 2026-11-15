package gia

import "testing"

func TestCubeReachabilityMintermEnumeration(t *testing.T) {
	limit := 1
	for vars := int32(0); vars <= 6; vars++ {
		if vars > 0 {
			limit *= 3
		}
		for code := 0; code < limit; code++ {
			cube := &Gia_StaAre_t{pData: make([]uint32, 1)}
			digits := code
			for i := int32(0); i < vars; i++ {
				if digits%3 == 1 {
					Gia_StaSetValue0(cube, i)
				}
				if digits%3 == 2 {
					Gia_StaSetValue1(cube, i)
				}
				digits /= 3
			}
			// The routine adds to an existing cover; preserve unrelated bits too.
			got, want := [2]uint32{0x80000001, 0x100}, [2]uint32{0x80000001, 0x100}
			for assignment := 0; assignment < 1<<vars; assignment++ {
				digits, matches := code, true
				for i := int32(0); i < vars; i++ {
					if digits%3 == 1 && assignment&(1<<i) != 0 || digits%3 == 2 && assignment&(1<<i) == 0 {
						matches = false
						break
					}
					digits /= 3
				}
				if matches {
					want[assignment>>5] |= 1 << uint(assignment&31)
				}
			}
			Gia_ManCountMintermsInCube(cube, vars, got[:])
			if got != want {
				t.Fatalf("vars%d cube%d minterms%x want%x", vars, code, got, want)
			}
		}
	}
}

func BenchmarkCubeMintermEnumeration(b *testing.B) {
	cube := &Gia_StaAre_t{pData: make([]uint32, 1)}
	var store [2]uint32
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Gia_ManCountMintermsInCube(cube, 6, store[:])
	}
}
