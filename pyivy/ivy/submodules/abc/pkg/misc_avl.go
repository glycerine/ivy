package abc

// Ordered balanced-tree behavior from src/misc/avl/avl.c and avl.h.

type AVLDirection uint8

const (
	AVLForward AVLDirection = iota
	AVLBackward
)

type AVLTreeOptions struct {
	MaxEntries int
	Stop       func() bool
}

type avlNode[K, V any] struct {
	left, right *avlNode[K, V]
	key         K
	value       V
	height      int
	live        bool
}

type AVLTree[K, V any] struct {
	root       *avlNode[K, V]
	compare    func(K, K) int
	count      int
	maxEntries int
	stop       func() bool
	closed     bool
}

type AVLItem[K, V any] struct {
	Key   K
	Value V
}

type AVLSlot[K, V any] struct {
	tree *AVLTree[K, V]
	node *avlNode[K, V]
}

type AVLGenerator[K, V any] struct {
	items  []AVLItem[K, V]
	index  int
	closed bool
}

func NewAVLTree[K, V any](compare func(K, K) int, options AVLTreeOptions) (*AVLTree[K, V], error) {
	if compare == nil || options.MaxEntries < 0 {
		return nil, ErrDimension
	}
	if options.MaxEntries == 0 {
		options.MaxEntries = 1 << 20
	}
	return &AVLTree[K, V]{compare: compare, maxEntries: options.MaxEntries, stop: options.Stop}, nil
}

// AVLIntCompare is the checked Go replacement for avl_numcmp.
func AVLIntCompare(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func (tree *AVLTree[K, V]) checkOpen() error {
	if tree == nil || tree.closed || tree.compare == nil || tree.maxEntries <= 0 {
		return ErrInvariant
	}
	return nil
}

func (tree *AVLTree[K, V]) cancelled() bool {
	return tree.stop != nil && tree.stop()
}

func avlHeight[K, V any](node *avlNode[K, V]) int {
	if node == nil {
		return -1
	}
	return node.height
}

func avlSetHeight[K, V any](node *avlNode[K, V]) {
	left, right := avlHeight(node.left), avlHeight(node.right)
	if right > left {
		left = right
	}
	node.height = left + 1
}

func avlRotateLeft[K, V any](oldRoot *avlNode[K, V]) *avlNode[K, V] {
	right := oldRoot.right
	if avlHeight(right.right)-avlHeight(right.left) >= 0 {
		oldRoot.right = right.left
		right.left = oldRoot
		avlSetHeight(oldRoot)
		avlSetHeight(right)
		return right
	}
	newRoot := right.left
	oldRoot.right = newRoot.left
	right.left = newRoot.right
	newRoot.left = oldRoot
	newRoot.right = right
	avlSetHeight(oldRoot)
	avlSetHeight(right)
	avlSetHeight(newRoot)
	return newRoot
}

func avlRotateRight[K, V any](oldRoot *avlNode[K, V]) *avlNode[K, V] {
	left := oldRoot.left
	if avlHeight(left.right)-avlHeight(left.left) <= 0 {
		oldRoot.left = left.right
		left.right = oldRoot
		avlSetHeight(oldRoot)
		avlSetHeight(left)
		return left
	}
	newRoot := left.right
	oldRoot.left = newRoot.right
	left.right = newRoot.left
	newRoot.left = left
	newRoot.right = oldRoot
	avlSetHeight(oldRoot)
	avlSetHeight(left)
	avlSetHeight(newRoot)
	return newRoot
}

func avlBalance[K, V any](node *avlNode[K, V]) *avlNode[K, V] {
	if node == nil {
		return nil
	}
	avlSetHeight(node)
	balance := avlHeight(node.right) - avlHeight(node.left)
	if balance < -1 {
		return avlRotateRight(node)
	}
	if balance > 1 {
		return avlRotateLeft(node)
	}
	return node
}

func avlInsert[K, V any](root, node *avlNode[K, V], compare func(K, K) int) *avlNode[K, V] {
	if root == nil {
		return node
	}
	if compare(node.key, root.key) < 0 {
		root.left = avlInsert(root.left, node, compare)
	} else {
		root.right = avlInsert(root.right, node, compare)
	}
	return avlBalance(root)
}

// Insert always adds an entry. duplicate reports whether an equal key was
// encountered on the insertion path; equal keys follow the right branch.
func (tree *AVLTree[K, V]) Insert(key K, value V) (duplicate bool, err error) {
	if err = tree.checkOpen(); err != nil {
		return false, err
	}
	if tree.cancelled() {
		return false, ErrCancelled
	}
	if tree.count == tree.maxEntries {
		return false, ErrLimit
	}
	for node := tree.root; node != nil; {
		difference := tree.compare(key, node.key)
		if difference == 0 {
			duplicate = true
		}
		if difference < 0 {
			node = node.left
		} else {
			node = node.right
		}
	}
	entry := &avlNode[K, V]{key: key, value: value, live: true}
	tree.root = avlInsert(tree.root, entry, tree.compare)
	tree.count++
	return duplicate, nil
}

func (tree *AVLTree[K, V]) find(key K) *avlNode[K, V] {
	for node := tree.root; node != nil; {
		difference := tree.compare(key, node.key)
		if difference == 0 {
			return node
		}
		if difference < 0 {
			node = node.left
		} else {
			node = node.right
		}
	}
	return nil
}

func (tree *AVLTree[K, V]) Lookup(key K) (value V, found bool, err error) {
	if err = tree.checkOpen(); err != nil {
		return value, false, err
	}
	if tree.cancelled() {
		return value, false, ErrCancelled
	}
	node := tree.find(key)
	if node == nil {
		return value, false, nil
	}
	return node.value, true, nil
}

func (tree *AVLTree[K, V]) Contains(key K) (bool, error) {
	_, found, err := tree.Lookup(key)
	return found, err
}

func (tree *AVLTree[K, V]) FindOrAdd(key K) (slot *AVLSlot[K, V], existing bool, err error) {
	if err = tree.checkOpen(); err != nil {
		return nil, false, err
	}
	if tree.cancelled() {
		return nil, false, ErrCancelled
	}
	if node := tree.find(key); node != nil {
		return &AVLSlot[K, V]{tree: tree, node: node}, true, nil
	}
	if tree.count == tree.maxEntries {
		return nil, false, ErrLimit
	}
	var zero V
	node := &avlNode[K, V]{key: key, value: zero, live: true}
	tree.root = avlInsert(tree.root, node, tree.compare)
	tree.count++
	return &AVLSlot[K, V]{tree: tree, node: node}, false, nil
}

func (slot *AVLSlot[K, V]) check() error {
	if slot == nil || slot.tree == nil || slot.node == nil || !slot.node.live {
		return ErrInvariant
	}
	return slot.tree.checkOpen()
}

func (slot *AVLSlot[K, V]) Key() (key K, err error) {
	if err = slot.check(); err != nil {
		return key, err
	}
	return slot.node.key, nil
}

func (slot *AVLSlot[K, V]) Value() (value V, err error) {
	if err = slot.check(); err != nil {
		return value, err
	}
	return slot.node.value, nil
}

func (slot *AVLSlot[K, V]) Set(value V) error {
	if err := slot.check(); err != nil {
		return err
	}
	slot.node.value = value
	return nil
}

func avlExtractMax[K, V any](root *avlNode[K, V]) (newRoot, maximum *avlNode[K, V]) {
	if root.right == nil {
		return root.left, root
	}
	root.right, maximum = avlExtractMax(root.right)
	return avlBalance(root), maximum
}

func avlDelete[K, V any](root *avlNode[K, V], key K, compare func(K, K) int) (newRoot, removed *avlNode[K, V]) {
	if root == nil {
		return nil, nil
	}
	difference := compare(key, root.key)
	if difference < 0 {
		root.left, removed = avlDelete(root.left, key, compare)
		if removed == nil {
			return root, nil
		}
		return avlBalance(root), removed
	}
	if difference > 0 {
		root.right, removed = avlDelete(root.right, key, compare)
		if removed == nil {
			return root, nil
		}
		return avlBalance(root), removed
	}
	removed = root
	if root.left == nil {
		return root.right, removed
	}
	left, predecessor := avlExtractMax(root.left)
	predecessor.left = left
	predecessor.right = root.right
	return avlBalance(predecessor), removed
}

// Delete removes the first equal key encountered by tree search and returns
// the stored key as well as its value.
func (tree *AVLTree[K, V]) Delete(key K) (storedKey K, value V, found bool, err error) {
	if err = tree.checkOpen(); err != nil {
		return storedKey, value, false, err
	}
	if tree.cancelled() {
		return storedKey, value, false, ErrCancelled
	}
	var removed *avlNode[K, V]
	tree.root, removed = avlDelete(tree.root, key, tree.compare)
	if removed == nil {
		return storedKey, value, false, nil
	}
	storedKey, value, found = removed.key, removed.value, true
	removed.live = false
	removed.left, removed.right = nil, nil
	var zeroK K
	var zeroV V
	removed.key, removed.value = zeroK, zeroV
	tree.count--
	return storedKey, value, found, nil
}

func (tree *AVLTree[K, V]) Len() (int, error) {
	if err := tree.checkOpen(); err != nil {
		return 0, err
	}
	return tree.count, nil
}

func (tree *AVLTree[K, V]) first(last bool) (item AVLItem[K, V], found bool, err error) {
	if err = tree.checkOpen(); err != nil {
		return item, false, err
	}
	if tree.cancelled() {
		return item, false, ErrCancelled
	}
	node := tree.root
	if node == nil {
		return item, false, nil
	}
	if last {
		for node.right != nil {
			node = node.right
		}
	} else {
		for node.left != nil {
			node = node.left
		}
	}
	return AVLItem[K, V]{Key: node.key, Value: node.value}, true, nil
}

func (tree *AVLTree[K, V]) First() (AVLItem[K, V], bool, error) { return tree.first(false) }
func (tree *AVLTree[K, V]) Last() (AVLItem[K, V], bool, error)  { return tree.first(true) }

func (tree *AVLTree[K, V]) Items(direction AVLDirection) ([]AVLItem[K, V], error) {
	if err := tree.checkOpen(); err != nil {
		return nil, err
	}
	if direction != AVLForward && direction != AVLBackward {
		return nil, ErrDimension
	}
	result := make([]AVLItem[K, V], 0, tree.count)
	type frame struct {
		node    *avlNode[K, V]
		emitted bool
	}
	stack := []frame{{node: tree.root}}
	for len(stack) > 0 {
		if tree.cancelled() {
			return nil, ErrCancelled
		}
		last := len(stack) - 1
		current := &stack[last]
		if current.node == nil {
			stack = stack[:last]
			continue
		}
		if !current.emitted {
			current.emitted = true
			child := current.node.left
			if direction == AVLBackward {
				child = current.node.right
			}
			stack = append(stack, frame{node: child})
			continue
		}
		node := current.node
		stack = stack[:last]
		result = append(result, AVLItem[K, V]{Key: node.key, Value: node.value})
		child := node.right
		if direction == AVLBackward {
			child = node.left
		}
		stack = append(stack, frame{node: child})
	}
	if len(result) != tree.count {
		return nil, ErrInvariant
	}
	return result, nil
}

func (tree *AVLTree[K, V]) NewGenerator(direction AVLDirection) (*AVLGenerator[K, V], error) {
	items, err := tree.Items(direction)
	if err != nil {
		return nil, err
	}
	return &AVLGenerator[K, V]{items: items}, nil
}

func (generator *AVLGenerator[K, V]) Next() (item AVLItem[K, V], found bool, err error) {
	if generator == nil || generator.closed {
		return item, false, ErrInvariant
	}
	if generator.index == len(generator.items) {
		return item, false, nil
	}
	item = generator.items[generator.index]
	generator.index++
	return item, true, nil
}

func (generator *AVLGenerator[K, V]) Close() error {
	if generator == nil || generator.closed {
		return ErrInvariant
	}
	generator.closed = true
	generator.items = nil
	return nil
}

func (tree *AVLTree[K, V]) ForEach(direction AVLDirection, visit func(K, V) error) error {
	if visit == nil {
		return ErrDimension
	}
	items, err := tree.Items(direction)
	if err != nil {
		return err
	}
	for _, item := range items {
		if tree.cancelled() {
			return ErrCancelled
		}
		if err = visit(item.Key, item.Value); err != nil {
			return err
		}
	}
	return nil
}

func (tree *AVLTree[K, V]) Check() error {
	if err := tree.checkOpen(); err != nil {
		return err
	}
	count := 0
	var previous K
	havePrevious := false
	var walk func(*avlNode[K, V]) (int, error)
	walk = func(node *avlNode[K, V]) (int, error) {
		if node == nil {
			return -1, nil
		}
		left, err := walk(node.left)
		if err != nil {
			return 0, err
		}
		if !node.live || havePrevious && tree.compare(previous, node.key) > 0 {
			return 0, ErrInvariant
		}
		previous, havePrevious = node.key, true
		count++
		right, err := walk(node.right)
		if err != nil {
			return 0, err
		}
		height := left
		if right > height {
			height = right
		}
		height++
		if node.height != height || right-left < -1 || right-left > 1 {
			return 0, ErrInvariant
		}
		return height, nil
	}
	if _, err := walk(tree.root); err != nil {
		return err
	}
	if count != tree.count || count > tree.maxEntries {
		return ErrInvariant
	}
	return nil
}

// Close visits entries in source postorder, invalidates every outstanding
// slot, and releases the tree. Nil callbacks are allowed.
func (tree *AVLTree[K, V]) Close(freeKey func(K), freeValue func(V)) error {
	if err := tree.checkOpen(); err != nil {
		return err
	}
	stack := []*avlNode[K, V]{tree.root}
	postorder := make([]*avlNode[K, V], 0, tree.count)
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if node == nil {
			continue
		}
		postorder = append(postorder, node)
		stack = append(stack, node.left, node.right)
	}
	for index := len(postorder) - 1; index >= 0; index-- {
		node := postorder[index]
		if freeKey != nil {
			freeKey(node.key)
		}
		if freeValue != nil {
			freeValue(node.value)
		}
		node.live = false
		node.left, node.right = nil, nil
		var zeroK K
		var zeroV V
		node.key, node.value = zeroK, zeroV
	}
	tree.root = nil
	tree.count = 0
	tree.closed = true
	tree.compare = nil
	return nil
}
