package logic

import (
	"strings"
)

// --- Sexp() methods on Sort types ---
// Field names are always included to prevent aliasing.

func (s *UninterpretedSort) Sexp() string {
	return "(UninterpretedSort name:" + s.Name + ")"
}
func (s *BooleanSort) Sexp() string { return "(BooleanSort)" }
func (s *FunctionSort) Sexp() string {
	parts := make([]string, len(s.Sorts))
	for i, sub := range s.Sorts {
		parts[i] = sub.Sexp()
	}
	return "(FunctionSort sorts:[" + strings.Join(parts, " ") + "])"
}
func (s *EnumeratedSort) Sexp() string {
	return "(EnumeratedSort name:" + s.Name + " ext:[" + strings.Join(s.Extension, ",") + "])"
}
func (s *RangeSort) Sexp() string {
	return "(RangeSort name:" + s.Name + " lb:" + s.Lb + " ub:" + s.Ub + ")"
}
func (s *TopSort) Sexp() string { return "(TopSort name:" + s.Name + ")" }

// --- Sexp() methods on Term types ---

func (v *Variable) Sexp() string {
	return "(Variable name:" + v.Name + " sort:" + v.VSort.Sexp() + ")"
}

func (c *Symbol) Sexp() string {
	return "(Symbol name:" + c.Name + " sort:" + c.CSort.Sexp() + ")"
}

func (a *Apply) Sexp() string {
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = t.Sexp()
	}
	return "(Apply func:" + a.Func.Sexp() + " terms:[" + strings.Join(parts, " ") + "])"
}

// --- Sexp() methods on Formula types ---

func (e *Eq) Sexp() string {
	return "(Eq t1:" + e.T1.Sexp() + " t2:" + e.T2.Sexp() + ")"
}

func (n *Not) Sexp() string {
	return "(Not body:" + n.Body.Sexp() + ")"
}

func (a *And) Sexp() string {
	parts := make([]string, len(a.Terms))
	for i, t := range a.Terms {
		parts[i] = t.Sexp()
	}
	return "(And terms:[" + strings.Join(parts, " ") + "])"
}

func (o *Or) Sexp() string {
	parts := make([]string, len(o.Terms))
	for i, t := range o.Terms {
		parts[i] = t.Sexp()
	}
	return "(Or terms:[" + strings.Join(parts, " ") + "])"
}

func (i *Implies) Sexp() string {
	return "(Implies t1:" + i.T1.Sexp() + " t2:" + i.T2.Sexp() + ")"
}

func (i *Iff) Sexp() string {
	return "(Iff t1:" + i.T1.Sexp() + " t2:" + i.T2.Sexp() + ")"
}

func (t *Ite) Sexp() string {
	return "(Ite cond:" + t.Cond.Sexp() + " then:" + t.Then.Sexp() + " else:" + t.Else.Sexp() + ")"
}

func (g *Globally) Sexp() string {
	env := "nil"
	if g.Environ != nil {
		env = *g.Environ
	}
	return "(Globally environ:" + env + " body:" + g.Body.Sexp() + ")"
}

func (e *Eventually) Sexp() string {
	env := "nil"
	if e.Environ != nil {
		env = *e.Environ
	}
	return "(Eventually environ:" + env + " body:" + e.Body.Sexp() + ")"
}

func (w *WhenOperator) Sexp() string {
	return "(WhenOperator name:" + w.Name + " t1:" + w.T1.Sexp() + " t2:" + w.T2.Sexp() + ")"
}

func (c *Cond) Sexp() string {
	return "(Cond t1:" + c.T1.Sexp() + " t2:" + c.T2.Sexp() + ")"
}

func varsSexp(vars []*Variable) string {
	parts := make([]string, len(vars))
	for i, v := range vars {
		parts[i] = v.Sexp()
	}
	return "[" + strings.Join(parts, " ") + "]"
}

func (f *ForAll) Sexp() string {
	return "(ForAll vars:" + varsSexp(f.Variables) + " body:" + f.Body.Sexp() + ")"
}

func (e *Exists) Sexp() string {
	return "(Exists vars:" + varsSexp(e.Variables) + " body:" + e.Body.Sexp() + ")"
}

func (l *Lambda) Sexp() string {
	return "(Lambda vars:" + varsSexp(l.Variables) + " body:" + l.Body.Sexp() + ")"
}

func (nb *NamedBinder) Sexp() string {
	env := "nil"
	if nb.Environ != nil {
		env = *nb.Environ
	}
	return "(NamedBinder name:" + nb.Name + " environ:" + env + " vars:" + varsSexp(nb.Variables) + " body:" + nb.Body.Sexp() + ")"
}

// --- Sexp() on Definition ---

func (d *Definition) Sexp() string {
	return "(Def lhs:" + d.Lhs.Sexp() + " rhs:" + d.Rhs.Sexp() + ")"
}

func (ds *DefinitionSchema) Sexp() string {
	return "(DefSchema lhs:" + ds.Lhs.Sexp() + " rhs:" + ds.Rhs.Sexp() + ")"
}

// --- NodeKey and Key ---

// NodeKey is a structural identity key for logic nodes.
// Two nodes with the same NodeKey are structurally equal,
// matching Python's recstruct == and hash behavior.
type NodeKey = string

// Key returns the structural identity key for a node.
// Use this as map key instead of the Node pointer.
func Key(n Node) NodeKey {
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

// --- NodeMap: map from nodes (by structural equality) to nodes ---

type NodeMap struct {
	m    *omap[NodeKey, Node]
	keys *omap[NodeKey, Node] // original key nodes for iteration
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		m:    newOmap[NodeKey, Node](),
		keys: newOmap[NodeKey, Node](),
	}
}

func (nm *NodeMap) Put(key, value Node) {
	k := Key(key)
	nm.m.set(k, value)
	nm.keys.set(k, key)
}

func (nm *NodeMap) Get(key Node) (Node, bool) {
	v, ok := nm.m.get2(Key(key))
	return v, ok
}

func (nm *NodeMap) Has(key Node) bool {
	_, ok := nm.m.get2(Key(key))
	return ok
}

func (nm *NodeMap) Delete(key Node) {
	k := Key(key)
	nm.m.delkey(k)
	nm.keys.delkey(k)
}

func (nm *NodeMap) Len() int {
	return nm.m.Len()
}

func (nm *NodeMap) Range(fn func(key, value Node) bool) {
	for k, v := range nm.m.all() {
		if !fn(nm.keys.get(k), v) {
			return
		}
	}
}

// --- NodeSet: set of nodes by structural equality ---

type NodeSet struct {
	m *omap[NodeKey, Node]
}

func NewNodeSet() *NodeSet {
	return &NodeSet{m: newOmap[NodeKey, Node]()}
}

func (ns *NodeSet) Add(n Node) {
	ns.m.set(Key(n), n)
}

func (ns *NodeSet) Has(n Node) bool {
	_, ok := ns.m.get2(Key(n))
	return ok
}

func (ns *NodeSet) Remove(n Node) {
	ns.m.delkey(Key(n))
}

func (ns *NodeSet) Len() int {
	return ns.m.Len()
}

func (ns *NodeSet) Range(fn func(Node) bool) {
	for _, v := range ns.m.all() {
		if !fn(v) {
			return
		}
	}
}
