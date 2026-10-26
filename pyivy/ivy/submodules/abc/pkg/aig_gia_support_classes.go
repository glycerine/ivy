package abc

// Support-renamed output classes from src/aig/gia/giaDup.c:
// Gia_ManIsoStrashReduceOne/Int and Gia_ManIsoStrashReduce.

// GiaSupportClasses retains each class's first source output on original CIs.
// Classes lists source CO ordinals; Outputs maps each source CO to its class.
// Supports supplies each source CO's sorted CI projection. Class membership
// means equal hashed logic after support renaming, not equal source functions
// under a common assignment, nor completeness for functional equivalence.
type GiaSupportClasses struct {
	Graph    *GiaManager
	Classes  [][]int
	Outputs  []int
	Supports [][]int
}

// ClassifyOutputsBySupport groups combinational cones by support-normalized
// phased hashed drivers. First occurrences define stable class order. Outputs
// with uniquely occurring structural support size stay singleton, matching C.
// Source constraint outputs participate as ordinary data and result roles clear.
// Representative cones keep all original CIs and raw gate kinds/name/spec.
// As in ReduceBySupport, actual zero/one-input logic is copied instead of C's
// unsafe size-only shortcut; buffers, degenerate cones and extended gates keep
// their actual semantics. Source scratch is unchanged; sequential graphs reject.
func (m *GiaManager) ClassifyOutputsBySupport(options GiaCofactorOptions) (GiaSupportClasses, error) {
	fail := func(e error) (GiaSupportClasses, error) { return GiaSupportClasses{}, e }
	if m == nil || options.MaxObjects < 0 || options.MaxMappingEntries < 0 {
		return fail(ErrDimension)
	}
	if e := m.Check(); e != nil {
		return fail(e)
	}
	if m.registers != 0 {
		return fail(ErrDimension)
	}
	stopped := func() bool { return options.Stop != nil && options.Stop() }
	if stopped() {
		return fail(ErrCancelled)
	}
	// The view only changes output-role interpretation. All helpers read source
	// storage and allocate their own scratch; neither source arrays nor roles mutate.
	view := *m
	view.constraints = 0
	normalized, e := view.ReduceBySupport(false, options)
	if e != nil {
		return fail(e)
	}
	frequencies := make([]int, len(m.cis)+1)
	for _, set := range normalized.Supports {
		frequencies[len(set)]++
	}
	classes := [][]int{}
	outputs := make([]int, len(m.cos))
	lookup := make(map[GiaLit]int)
	representatives := []int{}
	for i, co := range normalized.Graph.cos {
		if stopped() {
			return fail(ErrCancelled)
		}
		singleton := frequencies[len(normalized.Supports[i])] == 1
		driver, err := normalized.Graph.objects[co].Fanin0(co)
		if err != nil {
			return fail(err)
		}
		if group, ok := lookup[driver]; !singleton && ok {
			outputs[i] = group
			classes[group] = append(classes[group], i)
			continue
		}
		outputs[i] = len(classes)
		if !singleton {
			lookup[driver] = len(classes)
		}
		classes = append(classes, []int{i})
		representatives = append(representatives, i)
	}
	selected, e := view.ExtractOutputCones(representatives, false, options)
	if e != nil {
		return fail(e)
	}
	if stopped() {
		return fail(ErrCancelled)
	}
	return GiaSupportClasses{selected.Graph, classes, outputs, normalized.Supports}, nil
}
