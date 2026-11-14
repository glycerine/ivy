package cudd

// Internal constants and table types shared by the native CUDD routines.
// The public manager and node implementations live in package core.

const (
	DD_MAXREF DdHalfWord = ^DdHalfWord(0)

	DD_DEFAULT_RESIZE                     = 10
	DD_MEM_CHUNK                          = 1022
	DD_ONE_VAL                            = 1.0
	DD_ZERO_VAL                           = 0.0
	DD_EPSILON                            = 1e-12
	DD_PLUS_INF_VAL                       = 1e301
	DD_MINUS_INF_VAL                      = -1e301
	DD_MAX_SUBTABLE_DENSITY               = 4
	DD_GC_FRAC_LO                         = DD_MAX_SUBTABLE_DENSITY * 0.25
	DD_GC_FRAC_HI                         = DD_MAX_SUBTABLE_DENSITY * 1.0
	DD_GC_FRAC_MIN                        = 0.2
	DD_MIN_HIT                            = 30
	DD_MAX_LOOSE_FRACTION                 = 5
	DD_MAX_CACHE_FRACTION                 = 3
	DD_STASH_FRACTION                     = 64
	DD_MAX_CACHE_TO_SLOTS_RATIO           = 4
	DD_SIFT_MAX_VAR                       = 1000
	DD_SIFT_MAX_SWAPS                     = 2000000
	DD_DEFAULT_RECOMB                     = 0
	DD_MAX_REORDER_GROWTH                 = 1.1
	DD_FIRST_REORDER                      = 4004
	DD_DYN_RATIO                          = 2
	DD_P1                                 = 12582917
	DD_P2                                 = 4256249
	DD_P3                                 = 741457
	DD_P4                                 = 1610612741
	DD_BDD_AND_ABSTRACT_TAG               = 0x06
	DD_BDD_XOR_EXIST_ABSTRACT_TAG         = 0x0a
	DD_ADD_BDD_DO_INTERVAL_TAG            = 0x22
	DD_BDD_CLIPPING_AND_ABSTRACT_UP_TAG   = 0x26
	DD_BDD_CLIPPING_AND_ABSTRACT_DOWN_TAG = 0x2a
	DD_BDD_COMPOSE_RECUR_TAG              = 0x2e
	DD_ADD_COMPOSE_RECUR_TAG              = 0x42
	DD_ADD_NON_SIM_COMPOSE_TAG            = 0x46
	DD_EQUIV_DC_TAG                       = 0x4a
	DD_ZDD_ITE_TAG                        = 0x4e
	DD_ADD_ITE_CONSTANT_TAG               = 0x62
	DD_ADD_EVAL_CONST_TAG                 = 0x66
	DD_BDD_ITE_CONSTANT_TAG               = 0x6a
	DD_ADD_OUT_SUM_TAG                    = 0x6e
	DD_BDD_LEQ_UNLESS_TAG                 = 0x82
	DD_ADD_TRIANGLE_TAG                   = 0x86
	CUDD_GEN_CUBES                        = 0
	CUDD_GEN_PRIMES                       = 1
	CUDD_GEN_NODES                        = 2
	CUDD_GEN_ZDD_PATHS                    = 3
	CUDD_GEN_EMPTY                        = 0
	CUDD_GEN_NONEMPTY                     = 1
)

type DdGen struct {
	manager *DdManager
	genType int32
	status  int32
	data    struct {
		cubes struct {
			cube  []int32
			value CUDD_VALUE_TYPE
		}
		primes struct {
			cube []int32
			ub   *DdNode
		}
		nodes struct{ size int32 }
	}
	stack struct {
		sp    int32
		stack []*DdNode
	}
	node *DdNode
}

type DdHook struct {
	f    DD_HFP
	next *DdHook
}
type ptrint = int
type ptruint = uintptr
type DdNodePtr = *DdNode

type DdLocalCacheItem struct {
	value *DdNode
	count ptrint
	key   []*DdNode
}
type DdLocalCache struct {
	item                     []DdLocalCacheItem
	itemsize, keysize, slots uint32
	shift                    int32
	lookUps, minHit, hits    float64
	maxslots                 uint32
	manager                  *DdManager
	next                     *DdLocalCache
}
type DdHashItem struct {
	next  *DdHashItem
	count ptrint
	value *DdNode
	key   []*DdNode
}
type DdHashTable struct {
	keysize, itemsize uint32
	bucket            []*DdHashItem
	nextFree          *DdHashItem
	memoryList        []*DdHashItem
	numBuckets        uint32
	shift             int32
	size, maxsize     uint32
	manager           *DdManager
}
type DdCache struct {
	f, g  *DdNode
	h     ptruint
	data  *DdNode
	count ptrint
	hash  uint32
}
type DdSubtable struct {
	nodelist                         []*DdNode
	shift                            int32
	slots, keys, maxKeys, dead, next uint32
	bindVar                          int32
	varType                          Cudd_VariableType
	pairIndex, varHandled            int32
	varToBeGrouped                   Cudd_LazyGroupType
}
type Move struct {
	x, y  DdHalfWord
	flags uint32
	size  int32
	next  *Move
}
type DdQueueItem struct {
	next, cnext *DdQueueItem
	key         any
}
type DdLevelQueue struct {
	first                                              *DdQueueItem
	last                                               []*DdQueueItem
	freelist                                           *DdQueueItem
	buckets                                            []*DdQueueItem
	levels, itemsize, size, maxsize, numBuckets, shift int32
	byKey                                              map[any]*DdQueueItem
}
