package gia

import "testing"

func TestFxRetrieveCovers(t *testing.T) {
	for _, reverse := range []int32{0, 1} {
		p := Gia_ManStart(10)
		a, b, c := Gia_ManAppendCi(p), Gia_ManAppendCi(p), Gia_ManAppendCi(p)
		x := Gia_ManAppendAnd(p, a^1, b)
		y := Gia_ManAppendAnd(p, x, c)
		Gia_ManAppendCo(p, y)
		p.VMapping = Vec_IntStart(int(p.nObjs))
		for _, lut := range []struct{ id, a, b int32 }{{x >> 1, a >> 1, b >> 1}, {y >> 1, x >> 1, c >> 1}} {
			Vec_IntWriteEntry(p.VMapping, int(lut.id), int32(Vec_IntSize(p.VMapping)))
			for _, value := range []int32{2, lut.a, lut.b, lut.id} {
				Vec_IntPush(p.VMapping, value)
			}
		}
		var phases *Vec_Str_t
		cubes := Gia_ManFxRetrieve(p, &phases, reverse)
		for assignment := 0; assignment < 8; assignment++ {
			values := make([]bool, 5)
			for i := 0; i < 3; i++ {
				values[i] = assignment&(1<<i) != 0
			}
			for owner := 3; owner < 5; owner++ {
				var sum bool
				for i := 0; i < Vec_WecSize(cubes); i++ {
					cube := Vec_WecEntry(cubes, i)
					if int(Vec_IntEntry(cube, 0)) != owner {
						continue
					}
					product := true
					for j := 1; j < Vec_IntSize(cube); j++ {
						lit := Vec_IntEntry(cube, j)
						product = product && (values[lit>>1] != (lit&1 != 0))
					}
					sum = sum || product
				}
				values[owner] = sum != (Vec_StrEntry(phases, owner) != 0)
			}
			want := assignment&1 == 0 && assignment&2 != 0 && assignment&4 != 0
			if values[4] != want {
				t.Fatalf("reverse=%d assignment=%d got=%v want=%v", reverse, assignment, values[4], want)
			}
		}
	}
}
