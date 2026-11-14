package cudd

// DdLevelQueue is a FIFO queue grouped by level. Items at a level are
// contiguous, and byKey enforces the source queue's key uniqueness.
func cuddLevelQueueInit(levels, itemSize, numBuckets int32) *DdLevelQueue {
	if levels < 0 {
		return nil
	}
	if numBuckets < 2 {
		numBuckets = 2
	}
	n := int32(1) << cuddComputeFloorLog2(uint32(numBuckets))
	return &DdLevelQueue{
		last:       make([]*DdQueueItem, levels),
		buckets:    make([]*DdQueueItem, n),
		levels:     levels,
		itemsize:   itemSize,
		maxsize:    n * DD_MAX_SUBTABLE_DENSITY,
		numBuckets: n,
		shift:      32 - int32(cuddComputeFloorLog2(uint32(n))),
		byKey:      make(map[any]*DdQueueItem),
	}
}

func cuddLevelQueueQuit(queue *DdLevelQueue) {
	if queue == nil {
		return
	}
	*queue = DdLevelQueue{}
}

func cuddLevelQueueEnqueue(queue *DdLevelQueue, key any, level int32) *DdQueueItem {
	if queue == nil || level < 0 || level >= queue.levels {
		return nil
	}
	if item := hashLookup(queue, key); item != nil {
		return item
	}
	item := &DdQueueItem{key: key}
	if last := queue.last[level]; last != nil {
		item.next = last.next
		last.next = item
	} else {
		previous := level - 1
		for previous >= 0 && queue.last[previous] == nil {
			previous--
		}
		if previous < 0 {
			item.next = queue.first
			queue.first = item
		} else {
			item.next = queue.last[previous].next
			queue.last[previous].next = item
		}
	}
	queue.last[level] = item
	queue.size++
	if hashInsert(queue, item) == 0 {
		return nil
	}
	return item
}

func cuddLevelQueueDequeue(queue *DdLevelQueue, level int32) {
	if queue == nil || queue.first == nil || level < 0 || level >= queue.levels {
		return
	}
	item := queue.first
	hashDelete(queue, item)
	if queue.last[level] == item {
		queue.last[level] = nil
	}
	queue.first = item.next
	item.next = queue.freelist
	queue.freelist = item
	queue.size--
}

func hashLookup(queue *DdLevelQueue, key any) *DdQueueItem {
	if queue == nil {
		return nil
	}
	return queue.byKey[key]
}

func hashInsert(queue *DdLevelQueue, item *DdQueueItem) int32 {
	if queue == nil || item == nil {
		return 0
	}
	if queue.size > queue.maxsize && hashResize(queue) == 0 {
		return 0
	}
	queue.byKey[item.key] = item
	return 1
}

func hashDelete(queue *DdLevelQueue, item *DdQueueItem) {
	if queue != nil && item != nil {
		delete(queue.byKey, item.key)
	}
}

func hashResize(queue *DdLevelQueue) int32 {
	if queue == nil || queue.numBuckets > int32(^uint32(0)>>1)/2 {
		return 0
	}
	queue.numBuckets *= 2
	queue.maxsize *= 2
	queue.shift--
	queue.buckets = make([]*DdQueueItem, queue.numBuckets)
	return 1
}
