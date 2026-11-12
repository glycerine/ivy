package avl

// This is a Go port of src/misc/avl/avl.c. Rotations preserve node identity,
// allowing FindOrAdd's value slot to remain valid until that entry is deleted.
func height(n *node) int {
	if n == nil {
		return -1
	}
	return n.height
}
func update(n *node) { n.height = max(height(n.left), height(n.right)) + 1 }
func rotateLeft(n *node) *node {
	r := n.right
	n.right = r.left
	r.left = n
	update(n)
	update(r)
	return r
}
func rotateRight(n *node) *node {
	l := n.left
	n.left = l.right
	l.right = n
	update(n)
	update(l)
	return l
}
func balance(n *node) *node {
	if n == nil {
		return nil
	}
	update(n)
	if height(n.left)-height(n.right) > 1 {
		if height(n.left.left) < height(n.left.right) {
			n.left = rotateLeft(n.left)
		}
		return rotateRight(n)
	}
	if height(n.right)-height(n.left) > 1 {
		if height(n.right.right) < height(n.right.left) {
			n.right = rotateRight(n.right)
		}
		return rotateLeft(n)
	}
	return n
}

// Insert always adds an entry. It reports whether an equal key already existed.
func (t *Tree) Insert(key, value any) bool {
	found := false
	var insert func(*node) *node
	insert = func(n *node) *node {
		if n == nil {
			return &node{key: key, value: value}
		}
		c := t.compare(key, n.key)
		if c == 0 {
			found = true
		}
		if c < 0 {
			n.left = insert(n.left)
		} else {
			n.right = insert(n.right)
		}
		return balance(n)
	}
	t.root = insert(t.root)
	t.count++
	return found
}
func (t *Tree) find(key any) *node {
	for n := t.root; n != nil; {
		c := t.compare(key, n.key)
		if c == 0 {
			return n
		}
		if c < 0 {
			n = n.left
		} else {
			n = n.right
		}
	}
	return nil
}
func (t *Tree) Lookup(key any) (any, bool) {
	n := t.find(key)
	if n == nil {
		return nil, false
	}
	return n.value, true
}
func (t *Tree) Contains(key any) bool { return t.find(key) != nil }
func (t *Tree) Count() int            { return t.count }

// FindOrAdd returns a writable value slot and whether the key was present.
// A newly added entry starts with a nil value, as in the C implementation.
func (t *Tree) FindOrAdd(key any) (*any, bool) {
	if n := t.find(key); n != nil {
		return &n.value, true
	}
	var added *node
	var insert func(*node) *node
	insert = func(n *node) *node {
		if n == nil {
			added = &node{key: key}
			return added
		}
		if t.compare(key, n.key) < 0 {
			n.left = insert(n.left)
		} else {
			n.right = insert(n.right)
		}
		return balance(n)
	}
	t.root = insert(t.root)
	t.count++
	return &added.value, false
}
func (t *Tree) First() (key, value any, ok bool) {
	n := t.root
	if n == nil {
		return nil, nil, false
	}
	for n.left != nil {
		n = n.left
	}
	return n.key, n.value, true
}
func (t *Tree) Last() (key, value any, ok bool) {
	n := t.root
	if n == nil {
		return nil, nil, false
	}
	for n.right != nil {
		n = n.right
	}
	return n.key, n.value, true
}

// detachMax removes a node without copying its key/value into a different node.
func detachMax(n *node) (*node, *node) {
	if n.right == nil {
		return n.left, n
	}
	var result *node
	n.right, result = detachMax(n.right)
	return balance(n), result
}

// Delete removes one matching entry and returns its stored key and value.
func (t *Tree) Delete(key any) (storedKey, value any, ok bool) {
	var remove func(*node) *node
	remove = func(n *node) *node {
		if n == nil {
			return nil
		}
		c := t.compare(key, n.key)
		if c < 0 {
			n.left = remove(n.left)
			return balance(n)
		}
		if c > 0 {
			n.right = remove(n.right)
			return balance(n)
		}
		storedKey, value, ok = n.key, n.value, true
		if n.left == nil {
			return n.right
		}
		var replacement *node
		n.left, replacement = detachMax(n.left)
		replacement.left = n.left
		replacement.right = n.right
		return balance(replacement)
	}
	t.root = remove(t.root)
	if ok {
		t.count--
	}
	return
}
func walk(n *node, reverse bool, visit func(*node)) {
	if n == nil {
		return
	}
	first, last := n.left, n.right
	if reverse {
		first, last = last, first
	}
	walk(first, reverse, visit)
	visit(n)
	walk(last, reverse, visit)
}
func (t *Tree) Foreach(visit func(key, value any), direction int) {
	walk(t.root, direction != AVL_FORWARD, func(n *node) { visit(n.key, n.value) })
}

// Generator snapshots the ordered node list, matching the C generator. Mutating
// the tree while generating is unsupported; changes to a stored value are seen.
func (t *Tree) Generator(direction int) *Generator {
	g := &Generator{nodes: make([]*node, 0, t.count)}
	walk(t.root, direction != AVL_FORWARD, func(n *node) { g.nodes = append(g.nodes, n) })
	return g
}
func (g *Generator) Next() (key, value any, ok bool) {
	if g == nil || g.next == len(g.nodes) {
		return nil, nil, false
	}
	n := g.nodes[g.next]
	g.next++
	return n.key, n.value, true
}
func (g *Generator) Close() {
	if g != nil {
		g.nodes = nil
		g.next = 0
	}
}

// Clear visits entries in postorder before releasing the tree, as free_table did.
func (t *Tree) Clear(keyFree, valueFree func(any)) {
	var free func(*node)
	free = func(n *node) {
		if n == nil {
			return
		}
		free(n.left)
		free(n.right)
		if keyFree != nil {
			keyFree(n.key)
		}
		if valueFree != nil {
			valueFree(n.value)
		}
	}
	free(t.root)
	t.root = nil
	t.count = 0
}

// Check returns the number of height, ordering, balance or count violations.
func (t *Tree) Check() int {
	errors, count := 0, 0
	var previous any
	havePrevious := false
	var check func(*node) int
	check = func(n *node) int {
		if n == nil {
			return -1
		}
		l := check(n.left)
		if havePrevious && t.compare(previous, n.key) > 0 {
			errors++
		}
		previous = n.key
		havePrevious = true
		count++
		r := check(n.right)
		h := max(l, r) + 1
		if h != n.height {
			errors++
		}
		if l-r > 1 || r-l > 1 {
			errors++
		}
		return h
	}
	check(t.root)
	if count != t.count {
		errors++
	}
	return errors
}

// Compatibility entry points use Go output pointers and integer C status codes.
func status(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
func Avl_init_table(compare Compare) *Tree   { return New(compare) }
func Avl_insert(t *Tree, key, value any) int { return status(t.Insert(key, value)) }
func Avl_lookup(t *Tree, key any, value *any) int {
	v, ok := t.Lookup(key)
	if ok && value != nil {
		*value = v
	}
	return status(ok)
}
func Avl_is_member(t *Tree, key any) int { return status(t.Contains(key)) }
func Avl_find_or_add(t *Tree, key any, slot **any) int {
	p, ok := t.FindOrAdd(key)
	if slot != nil {
		*slot = p
	}
	return status(ok)
}
func Avl_delete(t *Tree, key, value *any) int {
	k, v, ok := t.Delete(*key)
	if ok {
		*key = k
		if value != nil {
			*value = v
		}
	}
	return status(ok)
}
func Avl_first(t *Tree, key, value *any) int {
	k, v, ok := t.First()
	if ok {
		if key != nil {
			*key = k
		}
		if value != nil {
			*value = v
		}
	}
	return status(ok)
}
func Avl_last(t *Tree, key, value *any) int {
	k, v, ok := t.Last()
	if ok {
		if key != nil {
			*key = k
		}
		if value != nil {
			*value = v
		}
	}
	return status(ok)
}
func Avl_count(t *Tree) int                          { return t.Count() }
func Avl_init_gen(t *Tree, direction int) *Generator { return t.Generator(direction) }
func Avl_gen(g *Generator, key, value *any) int {
	k, v, ok := g.Next()
	if ok {
		if key != nil {
			*key = k
		}
		if value != nil {
			*value = v
		}
	}
	return status(ok)
}
func Avl_free_gen(g *Generator)                                { g.Close() }
func Avl_foreach(t *Tree, visit func(any, any), direction int) { t.Foreach(visit, direction) }
func Avl_free_table(t *Tree, keyFree, valueFree func(any))     { t.Clear(keyFree, valueFree) }
func Avl_check_tree(t *Tree) int                               { return t.Check() }

// Avl_numcmp replaces the C pointer-to-integer comparison with integer keys.
func Avl_numcmp(x, y any) int {
	a, b := x.(int), y.(int)
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
