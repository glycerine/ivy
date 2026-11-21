package cut

func mergeLeaves(a, b []int32, limit int32, out []int32) []int32 {
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(a) == int(limit) {
		if !cutContains(a, b) {
			return nil
		}
		return a
	}
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		if len(out) == int(limit) {
			return nil
		}
		var v int32
		if j == len(b) || i < len(a) && a[i] < b[j] {
			v = a[i]
			i++
		} else if i == len(a) || b[j] < a[i] {
			v = b[j]
			j++
		} else {
			v = a[i]
			i++
			j++
		}
		out = append(out, v)
	}
	return out
}
func Cut_CutMergeTwo(p *Cut_Man_t, a, b *Cut_Cut_t) *Cut_Cut_t {
	var scratch [CUT_SIZE_MAX]int32
	leaves := mergeLeaves(a.leaves, b.leaves, p.params.NVarsMax, scratch[:0])
	if leaves == nil {
		return nil
	}
	p.countAllocation()
	return newMergedCut(p, leaves, a.USign|b.USign)
}

func newMergedCut(p *Cut_Man_t, leaves []int32, sign uint32) *Cut_Cut_t {
	c := p.newCut()
	c.setLeaves(leaves)
	if len(leaves) == 1 {
		p.stats.Trivial++
	}
	c.USign = sign
	return c
}
func cutContains(a, b []int32) bool {
	i := 0
	for _, v := range b {
		for i < len(a) && a[i] < v {
			i++
		}
		if i == len(a) || a[i] != v {
			return false
		}
		i++
	}
	return true
}
