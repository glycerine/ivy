package compiler

import (
	"fmt"
	"iter"
	"sync/atomic"

	rb "github.com/glycerine/rbtree"
)

// insMap is a map that iterates in insertion order.
type insMap[K comparable, V any] struct {
	version int64

	tree *rb.Tree
	idx  map[K]rb.Iterator

	// ordercache caches the pointers in contiguous memory.
	ordercache   []*ikv[K, V]
	cacheversion int64

	// nextInsertSeq ensures we sort by insertion time.
	nextInsertSeq uint64
}

type ikv[K comparable, V any] struct {
	seq uint64 // insertion order sequence
	key K
	val V
}

// newInsMap makes a new insMap.
func newInsMap[K comparable, V any]() *insMap[K, V] {
	return &insMap[K, V]{
		idx: make(map[K]rb.Iterator),
		tree: rb.NewTree(func(a, b rb.Item) int {
			aseq := a.(*ikv[K, V]).seq
			bseq := b.(*ikv[K, V]).seq
			if aseq < bseq {
				return -1
			}
			if aseq > bseq {
				return 1
			}
			return 0
		}),
	}
}

// cached returns the raw internal ikv slice.
func (s *insMap[K, V]) cached() []*ikv[K, V] {
	n := s.tree.Len()
	nc := len(s.ordercache)
	vers := atomic.LoadInt64(&s.version)
	if nc == n && s.cacheversion == vers {
		return s.ordercache
	}
	s.ordercache = nil
	s.cacheversion = vers
	for it := s.tree.Min(); !it.Limit(); it = it.Next() {
		kv := it.Item().(*ikv[K, V])
		s.ordercache = append(s.ordercache, kv)
	}
	return s.ordercache
}

func (s *insMap[K, V]) Len() int {
	return len(s.idx)
}

func (s *insMap[K, V]) String() string {
	r := "insMap{"
	for k, v := range s.all() {
		r += fmt.Sprintf("%v:%v, ", k, v)
	}
	return r + "}"
}

// delkey deletes a key from the insMap and returns the next iterator in the tree.
func (s *insMap[K, V]) delkey(key K) (found bool, next rb.Iterator) {
	if s.idx == nil || isNil(key) {
		return false, s.tree.Limit()
	}
	it, ok := s.idx[key]
	if !ok {
		return false, s.tree.Limit()
	}
	found = true
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	next = it.Next()
	s.tree.DeleteWithIterator(it)
	delete(s.idx, key)
	return
}

// deleteWithIter deletes by iterator and returns the next iterator.
func (s *insMap[K, V]) deleteWithIter(it rb.Iterator) (found bool, next rb.Iterator) {
	if it.Limit() || s.idx == nil {
		return false, s.tree.Limit()
	}
	kv := it.Item().(*ikv[K, V])
	if _, ok := s.idx[kv.key]; !ok {
		return false, s.tree.Limit()
	}
	found = true
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	next = it.Next()
	s.tree.DeleteWithIterator(it)
	delete(s.idx, kv.key)
	return
}

func (s *insMap[K, V]) deleteAll() {
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0
	s.tree.DeleteAll()
	s.idx = make(map[K]rb.Iterator)
	s.nextInsertSeq = 0
}

func (s *insMap[K, V]) set(key K, val V) (newlyAdded bool) {
	if isNil(key) {
		return false
	}
	atomic.AddInt64(&s.version, 1)
	s.ordercache = nil
	s.cacheversion = 0

	if s.idx == nil {
		s.idx = make(map[K]rb.Iterator)
	}
	it, ok := s.idx[key]

	if !ok {
		newlyAdded = true
		seq := s.nextInsertSeq
		s.nextInsertSeq++
		item := &ikv[K, V]{seq: seq, key: key, val: val}
		_, it2 := s.tree.InsertGetIt(item)
		s.idx[key] = it2
		return
	}
	it.Item().(*ikv[K, V]).val = val
	return
}

// all iterates in insertion order.
func (s *insMap[K, V]) all() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if s == nil || s.tree == nil {
			return
		}

		vers := atomic.LoadInt64(&s.version)
		if len(s.ordercache) == s.tree.Len() && s.cacheversion == vers {
			for _, kv := range s.ordercache {
				lastSeq := kv.seq
				if !yield(kv.key, kv.val) {
					return
				}
				if atomic.LoadInt64(&s.version) != vers {
					s.resumeFrom(lastSeq+1, yield)
					return
				}
			}
			return
		}

		it := s.tree.Min()
		for !it.Limit() {
			kv := it.Item().(*ikv[K, V])
			lastSeq := kv.seq
			if !yield(kv.key, kv.val) {
				return
			}
			if atomic.LoadInt64(&s.version) != vers {
				vers = atomic.LoadInt64(&s.version)
				it = s.tree.FindGE(&ikv[K, V]{seq: lastSeq + 1})
			} else {
				it = it.Next()
			}
		}
	}
}

// resumeFrom handles the re-sync logic after a mutation.
func (s *insMap[K, V]) resumeFrom(nextSeq uint64, yield func(K, V) bool) {
	it := s.tree.FindGE(&ikv[K, V]{seq: nextSeq})
	for !it.Limit() {
		kv := it.Item().(*ikv[K, V])
		lastSeq := kv.seq
		vers := atomic.LoadInt64(&s.version)
		if !yield(kv.key, kv.val) {
			return
		}
		if atomic.LoadInt64(&s.version) != vers {
			it = s.tree.FindGE(&ikv[K, V]{seq: lastSeq + 1})
		} else {
			it = it.Next()
		}
	}
}

func (s *insMap[K, V]) allikv() iter.Seq2[K, *ikv[K, V]] {
	return func(yield func(K, *ikv[K, V]) bool) {
		it := s.tree.Min()
		for !it.Limit() {
			kv := it.Item().(*ikv[K, V])
			lastSeq := kv.seq
			vers := atomic.LoadInt64(&s.version)
			if !yield(kv.key, kv) {
				return
			}
			if atomic.LoadInt64(&s.version) != vers {
				it = s.tree.FindGE(&ikv[K, V]{seq: lastSeq + 1})
			} else {
				it = it.Next()
			}
		}
	}
}

func (s *insMap[K, V]) get2(key K) (val V, found bool) {
	if s.idx == nil || isNil(key) {
		return
	}
	if it, ok := s.idx[key]; ok {
		return it.Item().(*ikv[K, V]).val, true
	}
	return
}

func (s *insMap[K, V]) get(key K) (val V) {
	v, _ := s.get2(key)
	return v
}

func (s *insMap[K, V]) getikv(key K) (kv *ikv[K, V], found bool) {
	if s.idx == nil || isNil(key) {
		return
	}
	if it, ok := s.idx[key]; ok {
		return it.Item().(*ikv[K, V]), true
	}
	return
}
