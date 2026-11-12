package cut

func mergeLeaves(a, b []int32, limit int32) []int32 {
	out := make([]int32, 0, min(int(limit), len(a)+len(b)))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
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
		if len(out) > int(limit) {
			return nil
		}
	}
	return out
}
func Cut_CutMergeTwo(p *Cut_Man_t, a, b *Cut_Cut_t) *Cut_Cut_t {
	leaves := mergeLeaves(a.leaves, b.leaves, p.params.NVarsMax)
	if leaves == nil {
		return nil
	}
	c := Cut_CutAlloc(p)
	c.leaves = leaves
	if len(leaves) == 1 {
		p.stats.Trivial++
	}
	c.USign = a.USign | b.USign
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
