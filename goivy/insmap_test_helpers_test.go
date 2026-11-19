package goivy

import "sort"

// nodeKeySetOf converts a map literal used by older tests into the
// ordered *InsMap that free-symbol sets now are, in sorted key order so
// tests stay deterministic.
func nodeKeySetOf(m map[NodeKey]Expr) *InsMap[NodeKey, Expr] {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	r := NewInsMap[NodeKey, Expr]()
	for _, k := range keys {
		r.Set(NodeKey(k), m[NodeKey(k)])
	}
	return r
}
