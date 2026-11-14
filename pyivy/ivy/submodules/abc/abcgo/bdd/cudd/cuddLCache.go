package cudd

import (
	"math/bits"
	"sync"
)

const DD_MAX_HASHTABLE_DENSITY = 2

var localCacheRegistry = struct {
	sync.Mutex
	caches map[*DdManager]map[*DdLocalCache]struct{}
}{caches: make(map[*DdManager]map[*DdLocalCache]struct{})}

func cacheHash(key []*DdNode, slots uint32) uint32 {
	if slots == 0 {
		return 0
	}
	// CUDD hashes node addresses. Go nodes have stable manager-assigned IDs,
	// which provide the same identity-sensitive behavior without pointer casts.
	h := uint32(0)
	if len(key) != 0 {
		h = uint32(key[0].ID) * DD_P2
		for _, n := range key[1:] {
			var id uint32
			if n != nil {
				id = uint32(n.ID)
			}
			h = h*DD_P1 + id
		}
	}
	return h & (slots - 1)
}

func sameCacheKey(a, b []*DdNode) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func cuddLocalCacheInit(manager *DdManager, keySize, cacheSize, maxCacheSize uint32) *DdLocalCache {
	if manager == nil || keySize == 0 {
		return nil
	}
	if cacheSize < 2 {
		cacheSize = 2
	}
	cacheSize = 1 << (bits.Len32(cacheSize) - 1)
	if maxCacheSize < cacheSize {
		maxCacheSize = cacheSize
	}
	cache := &DdLocalCache{
		item:     make([]DdLocalCacheItem, cacheSize),
		itemsize: (keySize + 1) * 8,
		keysize:  keySize,
		slots:    cacheSize,
		shift:    int32(32 - bits.Len32(cacheSize-1)),
		lookUps:  float64(cacheSize) + 1,
		minHit:   0.1,
		maxslots: maxCacheSize,
		manager:  manager,
	}
	localCacheRegistry.Lock()
	if localCacheRegistry.caches[manager] == nil {
		localCacheRegistry.caches[manager] = make(map[*DdLocalCache]struct{})
	}
	localCacheRegistry.caches[manager][cache] = struct{}{}
	localCacheRegistry.Unlock()
	return cache
}

func cuddLocalCacheQuit(cache *DdLocalCache) {
	if cache == nil {
		return
	}
	localCacheRegistry.Lock()
	delete(localCacheRegistry.caches[cache.manager], cache)
	if len(localCacheRegistry.caches[cache.manager]) == 0 {
		delete(localCacheRegistry.caches, cache.manager)
	}
	localCacheRegistry.Unlock()
	cache.item = nil
}

func cuddLocalCacheInsert(cache *DdLocalCache, key []*DdNode, value *DdNode) {
	if cache == nil || uint32(len(key)) != cache.keysize || cache.slots == 0 {
		return
	}
	idx := cacheHash(key, cache.slots)
	copyKey := append([]*DdNode(nil), key...)
	cache.item[idx] = DdLocalCacheItem{key: copyKey, value: value}
}

func cuddLocalCacheLookup(cache *DdLocalCache, key []*DdNode) *DdNode {
	if cache == nil || uint32(len(key)) != cache.keysize || cache.slots == 0 {
		return nil
	}
	cache.lookUps++
	idx := cacheHash(key, cache.slots)
	entry := &cache.item[idx]
	if entry.value != nil && sameCacheKey(key, entry.key) {
		cache.hits++
		return entry.value
	}
	if cache.slots < cache.maxslots && cache.lookUps > 0 && cache.hits/cache.lookUps > cache.minHit {
		cuddLocalCacheResize(cache)
	}
	return nil
}

func cuddLocalCacheClearDead(manager *DdManager) {
	localCacheRegistry.Lock()
	defer localCacheRegistry.Unlock()
	for cache := range localCacheRegistry.caches[manager] {
		for i := range cache.item {
			entry := &cache.item[i]
			if entry.value == nil {
				continue
			}
			dead := Cudd_Regular(entry.value).Ref == 0
			for _, key := range entry.key {
				if key == nil || Cudd_Regular(key).Ref == 0 {
					dead = true
					break
				}
			}
			if dead {
				*entry = DdLocalCacheItem{}
			}
		}
	}
}

func cuddLocalCacheClearAll(manager *DdManager) {
	localCacheRegistry.Lock()
	defer localCacheRegistry.Unlock()
	for cache := range localCacheRegistry.caches[manager] {
		clear(cache.item)
	}
}

func cuddLocalCacheResize(cache *DdLocalCache) {
	if cache == nil || cache.slots >= cache.maxslots {
		return
	}
	newSlots := cache.slots * 2
	if newSlots > cache.maxslots {
		newSlots = cache.maxslots
	}
	if newSlots <= cache.slots {
		return
	}
	old := cache.item
	cache.item = make([]DdLocalCacheItem, newSlots)
	cache.slots = newSlots
	for _, entry := range old {
		if entry.value != nil {
			idx := cacheHash(entry.key, newSlots)
			cache.item[idx] = entry
		}
	}
}

func cuddLocalCacheAddToList(*DdLocalCache)            {}
func cuddLocalCacheRemoveFromList(cache *DdLocalCache) { cuddLocalCacheQuit(cache) }

func cuddHashTableInit(manager *DdManager, keySize, initSize uint32) *DdHashTable {
	if manager == nil || keySize == 0 {
		return nil
	}
	if initSize < 2 {
		initSize = 2
	}
	n := uint32(1) << (bits.Len32(initSize) - 1)
	return &DdHashTable{
		keysize:    keySize,
		itemsize:   (keySize+1)*8 + 8,
		bucket:     make([]*DdHashItem, n),
		numBuckets: n,
		shift:      int32(32 - bits.Len32(n-1)),
		maxsize:    n * DD_MAX_HASHTABLE_DENSITY,
		manager:    manager,
	}
}

func cuddHashTableQuit(hash *DdHashTable) {
	if hash == nil {
		return
	}
	for i := range hash.bucket {
		for item := hash.bucket[i]; item != nil; {
			next := item.next
			Cudd_Deref(item.value)
			item = next
		}
	}
	*hash = DdHashTable{}
}

func cuddHashTableInsert(hash *DdHashTable, key []*DdNode, value *DdNode, count ptrint) int32 {
	if hash == nil || uint32(len(key)) != hash.keysize || value == nil {
		return 0
	}
	if hash.size > hash.maxsize && cuddHashTableResize(hash) == 0 {
		return 0
	}
	idx := cacheHash(key, hash.numBuckets)
	item := &DdHashItem{key: append([]*DdNode(nil), key...), value: value, count: count, next: hash.bucket[idx]}
	Cudd_Ref(value)
	hash.bucket[idx] = item
	hash.size++
	return 1
}

func cuddHashTableLookup(hash *DdHashTable, key []*DdNode) *DdNode {
	if hash == nil || uint32(len(key)) != hash.keysize || hash.numBuckets == 0 {
		return nil
	}
	idx := cacheHash(key, hash.numBuckets)
	var prev *DdHashItem
	for item := hash.bucket[idx]; item != nil; item = item.next {
		if !sameCacheKey(item.key, key) {
			prev = item
			continue
		}
		item.count--
		if item.count <= 0 {
			Cudd_Deref(item.value)
			if prev == nil {
				hash.bucket[idx] = item.next
			} else {
				prev.next = item.next
			}
			hash.size--
		}
		return item.value
	}
	return nil
}

func cuddHashTableInsert1(h *DdHashTable, f, value *DdNode, count ptrint) int32 {
	return cuddHashTableInsert(h, []*DdNode{f}, value, count)
}
func cuddHashTableLookup1(h *DdHashTable, f *DdNode) *DdNode {
	return cuddHashTableLookup(h, []*DdNode{f})
}
func cuddHashTableInsert2(h *DdHashTable, f, g, value *DdNode, count ptrint) int32 {
	return cuddHashTableInsert(h, []*DdNode{f, g}, value, count)
}
func cuddHashTableLookup2(h *DdHashTable, f, g *DdNode) *DdNode {
	return cuddHashTableLookup(h, []*DdNode{f, g})
}
func cuddHashTableInsert3(h *DdHashTable, f, g, x, value *DdNode, count ptrint) int32 {
	return cuddHashTableInsert(h, []*DdNode{f, g, x}, value, count)
}
func cuddHashTableLookup3(h *DdHashTable, f, g, x *DdNode) *DdNode {
	return cuddHashTableLookup(h, []*DdNode{f, g, x})
}

func cuddHashTableResize(hash *DdHashTable) int32 {
	if hash == nil || hash.numBuckets > ^uint32(0)/2 {
		return 0
	}
	n := hash.numBuckets * 2
	newBuckets := make([]*DdHashItem, n)
	for _, head := range hash.bucket {
		for item := head; item != nil; {
			next := item.next
			idx := cacheHash(item.key, n)
			item.next = newBuckets[idx]
			newBuckets[idx] = item
			item = next
		}
	}
	hash.bucket = newBuckets
	hash.numBuckets = n
	hash.maxsize = n * DD_MAX_HASHTABLE_DENSITY
	return 1
}

func cuddHashTableAlloc(hash *DdHashTable) *DdHashItem {
	if hash == nil {
		return nil
	}
	return &DdHashItem{key: make([]*DdNode, hash.keysize)}
}
