package goivy

import (
	"sort"
	"strings"
)

// --- NodeKey and Key ---

// NodeKey is a structural identity key for logic nodes.
// It is a DISTINCT TYPE (not a string alias) so the compiler enforces
// that map[NodeKey] cannot accept plain strings and vice versa.
// Two nodes with the same NodeKey are structurally equal,
// matching Python's recstruct == and hash behavior.
type NodeKey string

// String returns the NodeKey as a plain string for printing.
func (k NodeKey) String() string {
	return string(k)
}

// Key returns the structural identity key for a node.
// Use this as map key instead of the Expr pointer.
func Key(n Expr) NodeKey {
	if n == nil {
		return "(nil)"
	}
	return n.Sexp()
}

// SortKey returns the structural identity key for a sort.
func SortKey(s Sort) NodeKey {
	if s == nil {
		return "(nil)"
	}
	return s.Sexp()
}

// --- Sexp() methods on Sort types ---
// Field names are always included to prevent aliasing.

func (s *UninterpretedSort) Sexp() NodeKey {
	return NodeKey("(UninterpretedSort name:" + s.Name + ")")
}
func (s *BooleanSort) Sexp() NodeKey { return "(BooleanSort)" }
func (s *FunctionSort) Sexp() NodeKey {
	parts := make([]string, len(s.Sorts))
	for i, sub := range s.Sorts {
		parts[i] = string(sub.Sexp())
	}
	return NodeKey("(FunctionSort sorts:[" + strings.Join(parts, " ") + "])")
}
func (s *EnumeratedSort) Sexp() NodeKey {
	return NodeKey("(EnumeratedSort name:" + s.Name + " ext:[" + strings.Join(s.Extension, ",") + "])")
}
func (s *RangeSort) Sexp() NodeKey {
	return NodeKey("(RangeSort name:" + s.Name + " lb:" + s.Lb.BoundString() + " ub:" + s.Ub.BoundString() + ")")
}
func (s *TopSort) Sexp() NodeKey { return NodeKey("(TopSort name:" + s.Name + ")") }

// --- Sexp() methods on Term types ---

func (v *Variable) Sexp() NodeKey {
	return NodeKey("(Variable name:" + v.Name + " sort:" + string(v.VSort.Sexp()) + ")")
}

func (c *Const) Sexp() NodeKey {
	return c.sexp
}

func (a *Apply) Sexp() NodeKey {
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = string(t.Sexp())
	}
	return NodeKey("(Apply func:" + string(a.Func.Sexp()) + " terms:[" + strings.Join(parts, " ") + "])")
}

// --- Sexp() methods on Formula types ---

func (e *Eq) Sexp() NodeKey {
	return NodeKey("(Eq t1:" + string(e.T1.Sexp()) + " t2:" + string(e.T2.Sexp()) + ")")
}

func (n *Not) Sexp() NodeKey {
	return NodeKey("(Not body:" + string(n.Body.Sexp()) + ")")
}

func (a *And) Sexp() NodeKey {
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = string(t.Sexp())
	}
	return NodeKey("(And terms:[" + strings.Join(parts, " ") + "])")
}

func (o *Or) Sexp() NodeKey {
	parts := make([]string, len(o.Terms))
	for i, t := range o.Terms {
		parts[i] = string(t.Sexp())
	}
	return NodeKey("(Or terms:[" + strings.Join(parts, " ") + "])")
}

func (i *Implies) Sexp() NodeKey {
	return NodeKey("(Implies t1:" + string(i.T1.Sexp()) + " t2:" + string(i.T2.Sexp()) + ")")
}

func (i *Iff) Sexp() NodeKey {
	return NodeKey("(Iff t1:" + string(i.T1.Sexp()) + " t2:" + string(i.T2.Sexp()) + ")")
}

func (t *Ite) Sexp() NodeKey {
	return NodeKey("(Ite cond:" + string(t.Cond.Sexp()) + " then:" + string(t.Then.Sexp()) + " else:" + string(t.Else.Sexp()) + ")")
}

func (g *Globally) Sexp() NodeKey {
	env := "nil"
	if g.Environ != nil {
		env = *g.Environ
	}
	return NodeKey("(Globally environ:" + env + " body:" + string(g.Body.Sexp()) + ")")
}

func (e *Eventually) Sexp() NodeKey {
	env := "nil"
	if e.Environ != nil {
		env = *e.Environ
	}
	return NodeKey("(Eventually environ:" + env + " body:" + string(e.Body.Sexp()) + ")")
}

func (w *WhenOperator) Sexp() NodeKey {
	return NodeKey("(WhenOperator name:" + w.Name + " t1:" + string(w.T1.Sexp()) + " t2:" + string(w.T2.Sexp()) + ")")
}

func (c *Cond) Sexp() NodeKey {
	return NodeKey("(Cond t1:" + string(c.T1.Sexp()) + " t2:" + string(c.T2.Sexp()) + ")")
}

func VarsSexp(vars []*Variable) string {
	parts := make([]string, len(vars))
	for i, v := range vars {
		parts[i] = string(v.Sexp())
	}
	sort.Strings(parts)
	return "[" + strings.Join(parts, " ") + "]"
}

func (f *ForAll) Sexp() NodeKey {
	return NodeKey("(ForAll vars:" + VarsSexp(f.Variables) + " body:" + string(f.Body.Sexp()) + ")")
}

func (e *Exists) Sexp() NodeKey {
	return NodeKey("(Exists vars:" + VarsSexp(e.Variables) + " body:" + string(e.Body.Sexp()) + ")")
}

func (l *Lambda) Sexp() NodeKey {
	return NodeKey("(Lambda vars:" + VarsSexp(l.Variables) + " body:" + string(l.Body.Sexp()) + ")")
}

func (nb *NamedBinder) Sexp() NodeKey {
	env := "nil"
	if nb.Environ != nil {
		env = *nb.Environ
	}
	return NodeKey("(NamedBinder name:" + nb.Name + " environ:" + env + " vars:" + VarsSexp(nb.Variables) + " body:" + string(nb.Body.Sexp()) + ")")
}

// --- Sexp() on Definition ---

func (d *Definition) Sexp() NodeKey {
	return NodeKey("(Def lhs:" + string(d.Lhs.Sexp()) + " rhs:" + string(d.Rhs.Sexp()) + ")")
}

func (ds *DefinitionSchema) Sexp() NodeKey {
	return NodeKey("(DefSchema lhs:" + string(ds.Lhs.Sexp()) + " rhs:" + string(ds.Rhs.Sexp()) + ")")
}

// --- NodeMap: map from nodes (by structural equality) to nodes ---

type NodeMap struct {
	m    *omap[NodeKey, Expr]
	keys *omap[NodeKey, Expr] // original key nodes for iteration
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		m:    newOmap[NodeKey, Expr](),
		keys: newOmap[NodeKey, Expr](),
	}
}

func (nm *NodeMap) Put(key, value Expr) {
	k := Key(key)
	nm.m.set(k, value)
	nm.keys.set(k, key)
}

func (nm *NodeMap) Get(key Expr) (Expr, bool) {
	v, ok := nm.m.get2(Key(key))
	return v, ok
}

func (nm *NodeMap) Has(key Expr) bool {
	_, ok := nm.m.get2(Key(key))
	return ok
}

func (nm *NodeMap) Delete(key Expr) {
	k := Key(key)
	nm.m.delkey(k)
	nm.keys.delkey(k)
}

func (nm *NodeMap) Len() int {
	return nm.m.Len()
}

func (nm *NodeMap) Range(fn func(key, value Expr) bool) {
	for k, v := range nm.m.all() {
		if !fn(nm.keys.get(k), v) {
			return
		}
	}
}

// --- NodeSet: set of nodes by structural equality ---

type NodeSet struct {
	m *omap[NodeKey, Expr]
}

func NewNodeSet() *NodeSet {
	return &NodeSet{m: newOmap[NodeKey, Expr]()}
}

func (ns *NodeSet) Add(n Expr) {
	ns.m.set(Key(n), n)
}

func (ns *NodeSet) Has(n Expr) bool {
	_, ok := ns.m.get2(Key(n))
	return ok
}

func (ns *NodeSet) Remove(n Expr) {
	ns.m.delkey(Key(n))
}

func (ns *NodeSet) Len() int {
	return ns.m.Len()
}

func (ns *NodeSet) Range(fn func(Expr) bool) {
	for _, v := range ns.m.all() {
		if !fn(v) {
			return
		}
	}
}
