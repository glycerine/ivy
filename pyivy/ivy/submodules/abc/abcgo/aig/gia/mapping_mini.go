package gia

// Gia_ManBuildFromMiniInt expands a compact AND list using the given leaves.
// The compact vector is scratch storage and is cleared after an AND network.
func Gia_ManBuildFromMiniInt(p *Gia_Man_t, leaves, mini *Vec_Int_t, hash int32) int32 {
	n := Vec_IntSize(mini)
	if n == 0 {
		panic("gia: empty mini AIG")
	}
	phase := Vec_IntEntry(mini, n-1)
	if phase < 0 || phase > 1 {
		panic("gia: invalid mini AIG output phase")
	}
	if n == 1 {
		return phase
	}
	if n == 2 {
		if Vec_IntEntry(mini, 0) != 0 || Vec_IntSize(leaves) != 1 {
			panic("gia: invalid mini AIG variable")
		}
		return Vec_IntEntry(leaves, 0) ^ phase
	}
	if n%2 == 0 {
		panic("gia: malformed mini AIG")
	}
	nLeaves := Vec_IntSize(leaves)
	var result int32
	for i := 0; i < n-1; i += 2 {
		resolve := func(lit int32) int32 {
			if lit < 0 {
				panic("gia: negative mini literal")
			}
			id := int(lit >> 1)
			var value int32
			if id < nLeaves {
				value = Vec_IntEntry(leaves, id)
			} else {
				index := id - nLeaves
				if index >= i/2 {
					panic("gia: forward mini reference")
				}
				value = Vec_IntEntry(mini, index)
			}
			return value ^ (lit & 1)
		}
		a, b := resolve(Vec_IntEntry(mini, i)), resolve(Vec_IntEntry(mini, i+1))
		if hash != 0 {
			result = Gia_ManHashAnd(p, a, b)
		} else if a == b {
			result = a
		} else {
			result = Gia_ManAppendAnd(p, a, b)
		}
		Vec_IntWriteEntry(mini, i/2, result)
	}
	result ^= phase
	Vec_IntClear(mini)
	return result
}
