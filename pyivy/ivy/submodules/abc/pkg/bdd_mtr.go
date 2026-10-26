package abc

// Ordered multiway and variable-group trees from src/bdd/mtr.

import (
	"strconv"
	"strings"
)

type MTRFlags uint32

const (
	MTRDefault  MTRFlags = 0
	MTRTerminal MTRFlags = 1
	MTRSoft     MTRFlags = 2
	MTRFixed    MTRFlags = 4
	MTRNewNode  MTRFlags = 8
	MTRMaxHigh           = uint32(1<<31 - 1)
)

const mtrAllFlags = MTRTerminal | MTRSoft | MTRFixed | MTRNewNode

type MTRLimits struct {
	MaxNodes, MaxDepth, MaxInputBytes int
	Stop                              func() bool
}

func normalizeMTRLimits(limits MTRLimits) (MTRLimits, error) {
	if limits.MaxNodes < 0 || limits.MaxDepth < 0 || limits.MaxInputBytes < 0 {
		return limits, ErrDimension
	}
	if limits.MaxNodes == 0 {
		limits.MaxNodes = 1 << 20
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = 1 << 16
	}
	if limits.MaxInputBytes == 0 {
		limits.MaxInputBytes = 1 << 26
	}
	return limits, nil
}

// MTRNode retains the source's scalar node fields. Relationship accessors are
// read-only so callers cannot split the parent/child/sibling invariants.
type MTRNode struct {
	Flags            MTRFlags
	Low, Size, Index uint32
	parent           *MTRNode
	child            *MTRNode
	elder            *MTRNode
	younger          *MTRNode
	owner            *MTRTree
	live             bool
}

func (node *MTRNode) Parent() *MTRNode  { return node.parent }
func (node *MTRNode) Child() *MTRNode   { return node.child }
func (node *MTRNode) Elder() *MTRNode   { return node.elder }
func (node *MTRNode) Younger() *MTRNode { return node.younger }
func (node *MTRNode) Has(flag MTRFlags) bool {
	return node != nil && node.Flags&flag != 0
}
func (node *MTRNode) Set(flag MTRFlags) error {
	if node == nil || !node.live || flag&^mtrAllFlags != 0 {
		return ErrDimension
	}
	node.Flags |= flag
	return nil
}
func (node *MTRNode) Reset(flag MTRFlags) error {
	if node == nil || !node.live || flag&^mtrAllFlags != 0 {
		return ErrDimension
	}
	node.Flags &^= flag
	return nil
}

type MTRTree struct {
	root   *MTRNode
	limits MTRLimits
	nodes  int
	group  bool
	closed bool
}

func NewMTRTree(limits MTRLimits) (*MTRTree, error) {
	limits, err := normalizeMTRLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.Stop != nil && limits.Stop() {
		return nil, ErrCancelled
	}
	if limits.MaxNodes < 1 {
		return nil, ErrLimit
	}
	tree := &MTRTree{limits: limits, nodes: 1}
	tree.root = &MTRNode{owner: tree, live: true}
	return tree, nil
}

func NewMTRGroupTree(lower, size int, limits MTRLimits) (*MTRTree, error) {
	if lower < 0 || size < 0 || uint64(lower) > uint64(MTRMaxHigh) || uint64(size) > uint64(MTRMaxHigh) || uint64(lower)+uint64(size) > uint64(MTRMaxHigh)+1 {
		return nil, ErrDimension
	}
	tree, err := NewMTRTree(limits)
	if err != nil {
		return nil, err
	}
	tree.root.Low, tree.root.Size = uint32(lower), uint32(size)
	tree.group = true
	return tree, nil
}

func (tree *MTRTree) check() error {
	if tree == nil || tree.closed || tree.root == nil || !tree.root.live || tree.nodes < 1 || tree.nodes > tree.limits.MaxNodes {
		return ErrInvariant
	}
	return nil
}

func (tree *MTRTree) Root() *MTRNode {
	if tree == nil || tree.closed {
		return nil
	}
	return tree.root
}

func (tree *MTRTree) NodeCount() int {
	if tree == nil || tree.closed {
		return 0
	}
	return tree.nodes
}

func (tree *MTRTree) stopped() bool { return tree.limits.Stop != nil && tree.limits.Stop() }

func (tree *MTRTree) newNode() (*MTRNode, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	if tree.nodes >= tree.limits.MaxNodes {
		return nil, ErrLimit
	}
	if tree.stopped() {
		return nil, ErrCancelled
	}
	tree.nodes++
	return &MTRNode{owner: tree, live: true}, nil
}

func mtrDetached(node *MTRNode) bool {
	return node != nil && node.live && node.parent == nil && node.elder == nil && node.younger == nil
}

func mtrContains(root, target *MTRNode) bool {
	for node := root; node != nil; node = node.younger {
		if node == target || mtrContains(node.child, target) {
			return true
		}
	}
	return false
}

func mtrDepth(node *MTRNode) int {
	depth := 0
	for ; node != nil; node = node.parent {
		depth++
	}
	return depth
}

func (tree *MTRTree) MakeFirstChild(parent, child *MTRNode) error {
	if err := tree.check(); err != nil {
		return err
	}
	if parent == nil || !parent.live || parent.owner != tree || !mtrDetached(child) || child.owner != tree || parent == child || mtrContains(child, parent) {
		return ErrDimension
	}
	if mtrDepth(parent) >= tree.limits.MaxDepth {
		return ErrLimit
	}
	if tree.stopped() {
		return ErrCancelled
	}
	child.parent, child.elder, child.younger = parent, nil, parent.child
	if parent.child != nil {
		parent.child.elder = child
	}
	parent.child = child
	return nil
}

func (tree *MTRTree) MakeLastChild(parent, child *MTRNode) error {
	if err := tree.check(); err != nil {
		return err
	}
	if parent == nil || !parent.live || parent.owner != tree || !mtrDetached(child) || child.owner != tree || parent == child || mtrContains(child, parent) {
		return ErrDimension
	}
	if mtrDepth(parent) >= tree.limits.MaxDepth {
		return ErrLimit
	}
	if tree.stopped() {
		return ErrCancelled
	}
	child.parent = parent
	if parent.child == nil {
		parent.child = child
		return nil
	}
	last := parent.child
	for last.younger != nil {
		last = last.younger
	}
	last.younger, child.elder = child, last
	return nil
}

func (tree *MTRTree) CreateFirstChild(parent *MTRNode) (*MTRNode, error) {
	child, err := tree.newNode()
	if err != nil {
		return nil, err
	}
	if err = tree.MakeFirstChild(parent, child); err != nil {
		tree.nodes--
		child.live = false
		return nil, err
	}
	return child, nil
}

func (tree *MTRTree) CreateLastChild(parent *MTRNode) (*MTRNode, error) {
	child, err := tree.newNode()
	if err != nil {
		return nil, err
	}
	if err = tree.MakeLastChild(parent, child); err != nil {
		tree.nodes--
		child.live = false
		return nil, err
	}
	return child, nil
}

func (tree *MTRTree) MakeNextSibling(first, second *MTRNode) error {
	if err := tree.check(); err != nil {
		return err
	}
	if first == nil || !first.live || first.owner != tree || first.parent == nil || !mtrDetached(second) || second.owner != tree || first == second || mtrContains(second, first) {
		return ErrDimension
	}
	if tree.stopped() {
		return ErrCancelled
	}
	second.parent, second.elder, second.younger = first.parent, first, first.younger
	if first.younger != nil {
		first.younger.elder = second
	}
	first.younger = second
	return nil
}

func (tree *MTRTree) Clone(expansion int, limits MTRLimits) (*MTRTree, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	if expansion < 1 {
		return nil, ErrDimension
	}
	limits, err := normalizeMTRLimits(limits)
	if err != nil {
		return nil, err
	}
	if limits.MaxNodes < tree.nodes {
		return nil, ErrLimit
	}
	copyTree := &MTRTree{limits: limits, group: tree.group}
	var cloneChain func(*MTRNode, *MTRNode, int) (*MTRNode, error)
	cloneChain = func(source, parent *MTRNode, depth int) (*MTRNode, error) {
		if source == nil {
			return nil, nil
		}
		if depth > limits.MaxDepth || uint64(source.Low)*uint64(expansion) > uint64(MTRMaxHigh) || uint64(source.Size)*uint64(expansion) > uint64(MTRMaxHigh) || uint64(source.Index)*uint64(expansion) > uint64(MTRMaxHigh) {
			return nil, ErrLimit
		}
		if limits.Stop != nil && limits.Stop() {
			return nil, ErrCancelled
		}
		node := &MTRNode{Flags: source.Flags, Low: source.Low * uint32(expansion), Size: source.Size * uint32(expansion), Index: source.Index * uint32(expansion), parent: parent, owner: copyTree, live: true}
		copyTree.nodes++
		child, e := cloneChain(source.child, node, depth+1)
		if e != nil {
			return nil, e
		}
		node.child = child
		younger, e := cloneChain(source.younger, parent, depth)
		if e != nil {
			return nil, e
		}
		node.younger = younger
		if younger != nil {
			younger.elder = node
		}
		return node, nil
	}
	copyTree.root, err = cloneChain(tree.root, nil, 1)
	if err != nil {
		return nil, err
	}
	return copyTree, nil
}

func mtrRangeInside(root *MTRNode, low, size uint32) bool {
	return size > 0 && uint64(low) >= uint64(root.Low) && uint64(low)+uint64(size) <= uint64(root.Low)+uint64(root.Size)
}

func (tree *MTRTree) MakeGroup(low, size uint32, flags MTRFlags) (*MTRNode, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	if flags&^mtrAllFlags != 0 || size == 0 || !mtrRangeInside(tree.root, low, size) {
		return nil, ErrDimension
	}
	if tree.stopped() {
		return nil, ErrCancelled
	}
	var makeGroup func(*MTRNode, int) (*MTRNode, error)
	makeGroup = func(root *MTRNode, depth int) (*MTRNode, error) {
		if depth > tree.limits.MaxDepth {
			return nil, ErrLimit
		}
		if root.Low == low && root.Size == size {
			root.Flags = flags
			return root, nil
		}
		previous, first := (*MTRNode)(nil), root.child
		for first != nil && uint64(low) >= uint64(first.Low)+uint64(first.Size) {
			previous, first = first, first.younger
		}
		if first != nil && uint64(low) >= uint64(first.Low) && uint64(low)+uint64(size) <= uint64(first.Low)+uint64(first.Size) {
			return makeGroup(first, depth+1)
		}
		if first != nil {
			end := uint64(low) + uint64(size)
			firstEnd := uint64(first.Low) + uint64(first.Size)
			if (end > uint64(first.Low) && end < firstEnd) || uint64(low) > uint64(first.Low) && uint64(low) < firstEnd {
				return nil, ErrDimension
			}
		}
		if depth >= tree.limits.MaxDepth {
			return nil, ErrLimit
		}
		newNode, err := tree.newNode()
		if err != nil {
			return nil, err
		}
		newNode.Low, newNode.Size, newNode.Flags, newNode.parent = low, size, flags, root
		end := uint64(low) + uint64(size)
		if first == nil || end <= uint64(first.Low) {
			newNode.elder, newNode.younger = previous, first
			if first != nil {
				first.elder = newNode
			}
			if previous != nil {
				previous.younger = newNode
			} else {
				root.child = newNode
			}
			return newNode, nil
		}
		last := first
		for last.younger != nil && uint64(last.younger.Low)+uint64(last.younger.Size) <= end {
			last = last.younger
		}
		if last.younger != nil && uint64(last.younger.Low) < end {
			tree.nodes--
			newNode.live = false
			return nil, ErrDimension
		}
		after := last.younger
		newNode.child, newNode.elder, newNode.younger = first, previous, after
		first.elder = nil
		last.younger = nil
		if after != nil {
			after.elder = newNode
		}
		if previous != nil {
			previous.younger = newNode
		} else {
			root.child = newNode
		}
		for node := first; node != nil; node = node.younger {
			node.parent = newNode
		}
		return newNode, nil
	}
	return makeGroup(tree.root, 1)
}

func (tree *MTRTree) FindGroup(low, size uint32) *MTRNode {
	if tree.check() != nil || size == 0 || !mtrRangeInside(tree.root, low, size) {
		return nil
	}
	root := tree.root
	for root != nil {
		if root.Low == low && root.Size == size {
			return root
		}
		node := root.child
		for node != nil && uint64(low) >= uint64(node.Low)+uint64(node.Size) {
			node = node.younger
		}
		if node == nil || uint64(low)+uint64(size) > uint64(node.Low)+uint64(node.Size) {
			return nil
		}
		root = node
	}
	return nil
}

func (tree *MTRTree) DissolveGroup(group *MTRNode) (*MTRNode, error) {
	if err := tree.check(); err != nil {
		return nil, err
	}
	if group == nil || !group.live || group.owner != tree || group.parent == nil || group.Flags&MTRTerminal != 0 || group.child == nil {
		return nil, ErrDimension
	}
	if tree.stopped() {
		return nil, ErrCancelled
	}
	parent, first := group.parent, group.child
	last := first
	for {
		last.parent = parent
		if last.younger == nil {
			break
		}
		last = last.younger
	}
	last.younger = group.younger
	if group.younger != nil {
		group.younger.elder = last
	}
	first.elder = group.elder
	if group.elder == nil {
		parent.child = first
	} else {
		group.elder.younger = first
	}
	group.parent, group.child, group.elder, group.younger, group.owner, group.live = nil, nil, nil, nil, nil, false
	tree.nodes--
	return parent, nil
}

func mtrCanShift(node *MTRNode, shift int64) bool {
	low := int64(node.Low) + shift
	if low < 0 || node.Size == 0 || uint64(low)+uint64(node.Size)-1 > uint64(MTRMaxHigh) {
		return false
	}
	if node.Flags&MTRTerminal == 0 {
		for child := node.child; child != nil; child = child.younger {
			if !mtrCanShift(child, shift) {
				return false
			}
		}
	}
	return true
}

func mtrShift(node *MTRNode, shift int64) {
	node.Low = uint32(int64(node.Low) + shift)
	if node.Flags&MTRTerminal == 0 {
		for child := node.child; child != nil; child = child.younger {
			mtrShift(child, shift)
		}
	}
}

func (tree *MTRTree) SwapGroups(first, second *MTRNode) error {
	if err := tree.check(); err != nil {
		return err
	}
	if first == nil || second == nil || !first.live || !second.live || first.owner != tree || second.owner != tree {
		return ErrDimension
	}
	if second.younger == first {
		first, second = second, first
	} else if first.younger != second {
		return ErrDimension
	}
	parent := first.parent
	if parent == nil || second.parent != parent || !mtrCanShift(first, int64(second.Size)) || !mtrCanShift(second, -int64(first.Size)) {
		return ErrDimension
	}
	if tree.stopped() {
		return ErrCancelled
	}
	before, after := first.elder, second.younger
	if before == nil {
		parent.child = second
	} else {
		before.younger = second
	}
	if after != nil {
		after.elder = first
	}
	second.elder, second.younger = before, first
	first.elder, first.younger = second, after
	mtrShift(first, int64(second.Size))
	mtrShift(second, -int64(first.Size))
	return nil
}

func (tree *MTRTree) FormatGroups() (string, error) {
	if err := tree.Validate(); err != nil {
		return "", err
	}
	var hasZeroSize func(*MTRNode) bool
	hasZeroSize = func(node *MTRNode) bool {
		for ; node != nil; node = node.younger {
			if node.Size == 0 || hasZeroSize(node.child) {
				return true
			}
		}
		return false
	}
	if hasZeroSize(tree.root) {
		return "", ErrDimension
	}
	var builder strings.Builder
	var write func(*MTRNode)
	write = func(root *MTRNode) {
		builder.WriteByte('(')
		builder.WriteString(strconv.FormatUint(uint64(root.Low), 10))
		if root.Flags&MTRTerminal != 0 || root.child == nil {
			builder.WriteByte(',')
		} else {
			for child := root.child; child != nil; child = child.younger {
				write(child)
			}
		}
		builder.WriteString(strconv.FormatUint(uint64(root.Low+root.Size-1), 10))
		if root.Flags != MTRDefault {
			builder.WriteByte('|')
			if root.Flags&MTRFixed != 0 {
				builder.WriteByte('F')
			}
			if root.Flags&MTRNewNode != 0 {
				builder.WriteByte('N')
			}
			if root.Flags&MTRSoft != 0 {
				builder.WriteByte('S')
			}
		}
		builder.WriteByte(')')
	}
	write(tree.root)
	builder.WriteByte('\n')
	return builder.String(), nil
}

// FormatTree returns a deterministic relation dump; stable traversal IDs
// replace the source's process-specific pointer addresses.
func (tree *MTRTree) FormatTree() (string, error) {
	if err := tree.Validate(); err != nil {
		return "", err
	}
	ids := make(map[*MTRNode]int, tree.nodes)
	var assignIDs func(*MTRNode)
	assignIDs = func(node *MTRNode) {
		for ; node != nil; node = node.younger {
			ids[node] = len(ids) + 1
			assignIDs(node.child)
		}
	}
	assignIDs(tree.root)
	order := make([]*MTRNode, 0, tree.nodes)
	var collect func(*MTRNode)
	collect = func(node *MTRNode) {
		for ; node != nil; node = node.younger {
			order = append(order, node)
			if node.Flags&MTRTerminal == 0 {
				collect(node.child)
			}
		}
	}
	collect(tree.root)
	relation := func(node *MTRNode) string {
		if node == nil {
			return "0"
		}
		return strconv.Itoa(ids[node])
	}
	var builder strings.Builder
	for _, node := range order {
		builder.WriteString("N=")
		builder.WriteString(relation(node))
		builder.WriteString(" C=")
		builder.WriteString(relation(node.child))
		builder.WriteString(" Y=")
		builder.WriteString(relation(node.younger))
		builder.WriteString(" E=")
		builder.WriteString(relation(node.elder))
		builder.WriteString(" P=")
		builder.WriteString(relation(node.parent))
		builder.WriteString(" F=")
		builder.WriteString(strconv.FormatUint(uint64(node.Flags), 16))
		builder.WriteString(" L=")
		builder.WriteString(strconv.FormatUint(uint64(node.Low), 10))
		builder.WriteString(" S=")
		builder.WriteString(strconv.FormatUint(uint64(node.Size), 10))
		builder.WriteByte('\n')
	}
	return builder.String(), nil
}

func ParseMTRGroups(data []byte, leaves int, limits MTRLimits) (*MTRTree, error) {
	limits, err := normalizeMTRLimits(limits)
	if err != nil {
		return nil, err
	}
	if leaves < 0 || len(data) > limits.MaxInputBytes {
		if leaves < 0 {
			return nil, ErrDimension
		}
		return nil, ErrLimit
	}
	fields := strings.Fields(string(data))
	if len(fields)%3 != 0 {
		return nil, ErrFormat
	}
	tree, err := NewMTRGroupTree(0, leaves, limits)
	if err != nil {
		return nil, err
	}
	for index := 0; index < len(fields); index += 3 {
		if tree.stopped() {
			tree.Close()
			return nil, ErrCancelled
		}
		low, e1 := strconv.Atoi(fields[index])
		size, e2 := strconv.Atoi(fields[index+1])
		if e1 != nil || e2 != nil || low < 0 || size < 1 || uint64(low)+uint64(size) > uint64(leaves) {
			tree.Close()
			return nil, ErrFormat
		}
		if len(fields[index+2]) > 32 {
			tree.Close()
			return nil, ErrFormat
		}
		var flags MTRFlags
		for _, ch := range fields[index+2] {
			switch ch {
			case 'D':
			case 'F':
				flags |= MTRFixed
			case 'N':
				flags |= MTRNewNode
			case 'S':
				flags |= MTRSoft
			case 'T':
				flags |= MTRTerminal
			default:
				tree.Close()
				return nil, ErrFormat
			}
		}
		if _, err = tree.MakeGroup(uint32(low), uint32(size), flags); err != nil {
			tree.Close()
			return nil, err
		}
	}
	return tree, nil
}

func (tree *MTRTree) Validate() error {
	if err := tree.check(); err != nil {
		return err
	}
	seen := make(map[*MTRNode]bool, tree.nodes)
	count := 0
	var walk func(*MTRNode, *MTRNode, int, bool) error
	walk = func(node, parent *MTRNode, depth int, groupRanges bool) error {
		var elder *MTRNode
		lastEnd := uint64(0)
		for ; node != nil; node = node.younger {
			if tree.stopped() {
				return ErrCancelled
			}
			if depth > tree.limits.MaxDepth || seen[node] || !node.live || node.owner != tree || node.parent != parent || node.elder != elder || node.Flags&^mtrAllFlags != 0 {
				return ErrInvariant
			}
			if node.Size > 0 && uint64(node.Low)+uint64(node.Size)-1 > uint64(MTRMaxHigh) {
				return ErrInvariant
			}
			if groupRanges && elder != nil && uint64(node.Low) < lastEnd {
				return ErrInvariant
			}
			if groupRanges && parent != nil && !mtrRangeInside(parent, node.Low, node.Size) {
				return ErrInvariant
			}
			seen[node] = true
			count++
			if err := walk(node.child, node, depth+1, groupRanges && node.Flags&MTRTerminal == 0); err != nil {
				return err
			}
			lastEnd = uint64(node.Low) + uint64(node.Size)
			elder = node
		}
		return nil
	}
	if tree.root.parent != nil || tree.root.elder != nil || tree.root.younger != nil {
		return ErrInvariant
	}
	if err := walk(tree.root, nil, 1, tree.group); err != nil {
		return err
	}
	if count != tree.nodes {
		return ErrInvariant
	}
	return nil
}

func (tree *MTRTree) Close() {
	if tree == nil || tree.closed {
		return
	}
	var clear func(*MTRNode)
	clear = func(node *MTRNode) {
		for node != nil {
			next := node.younger
			clear(node.child)
			node.parent = nil
			node.child = nil
			node.elder = nil
			node.younger = nil
			node.owner = nil
			node.live = false
			node = next
		}
	}
	clear(tree.root)
	tree.root = nil
	tree.nodes = 0
	tree.closed = true
}
