package gia

import (
	"fmt"
	"slices"
	"testing"
)

func fxReviewInsertionChain(n int) (*Gia_Man_t, *Vec_Wec_t, *Vec_Str_t) {
	p := Gia_ManStart(int32(n + 64))
	inputs := make([]int32, 8)
	for i := range inputs {
		inputs[i] = Gia_ManAppendCi(p)
	}
	cubes := Vec_WecAlloc(n + 1)
	prev := int32(0)
	lit := inputs[0]
	for i := 0; i < n; i++ {
		lit = Gia_ManAppendAnd(p, lit^1, inputs[(i+1)%8])
		owner := int32(8 + i)
		p.pObjs[lit>>1].Value = uint32(owner)
		row := []int32{prev<<1 | 1, int32((i+1)%8) << 1}
		slices.Sort(row)
		dst := Vec_WecPushLevel(cubes)
		for _, v := range []int32{owner, row[0], row[1]} {
			Vec_IntPush(dst, v)
		}
		prev = owner
	}
	Gia_ManAppendCo(p, lit)
	dst := Vec_WecPushLevel(cubes)
	for _, v := range []int32{int32(n + 8), 0, 2} {
		Vec_IntPush(dst, v)
	}
	return p, cubes, Vec_StrStart(n + 8)
}

func BenchmarkFxInsertion(b *testing.B) {
	for _, n := range []int{8, 64, 256} {
		b.Run(fmt.Sprintf("owners%d", n), func(b *testing.B) {
			p, cubes, phases := fxReviewInsertionChain(n)
			defer Gia_ManStop(p)
			defer Vec_WecFree(cubes)
			defer Vec_StrFree(phases)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				q := Gia_ManFxInsert(p, cubes, phases)
				Gia_ManStop(q)
			}
		})
	}
}
