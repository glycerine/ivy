package glucose

/*******************************************************************************************[Map.h]
Copyright (c) 2006-2010, Niklas Sorensson

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and
associated documentation files (the "Software"), to deal in the Software without restriction,
including without limitation the rights to use, copy, modify, merge, publish, distribute,
sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or
substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT
NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT
OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
**************************************************************************************************/

import "hash/maphash"

var defaultMapSeed = maphash.MakeSeed()

type Pair[K any, V any] struct {
	Key  K
	Data V
}

// Map retains source bucket iteration and caller-supplied hash/equality.
// count caches an aggregate over independent buckets, not a slice length.
type Map[K any, V any] struct {
	hash    func(K) uint32
	equal   func(K, K) bool
	buckets [][]Pair[K, V]
	count   int
}

var mapPrimes = []int{31, 73, 151, 313, 643, 1291, 2593, 5233, 10501, 21013, 42073, 84181, 168451, 337219, 674701, 1349473, 2699299, 5398891, 10798093, 21596719, 43193641, 86387383, 172775299, 345550609, 691101253}

func NewHashMap[K any, V any](hash func(K) uint32, equal func(K, K) bool) *Map[K, V] {
	if hash == nil || equal == nil {
		panic("map requires hash and equality")
	}
	return &Map[K, V]{hash: hash, equal: equal}
}
func NewMap[K comparable, V any]() *Map[K, V] {
	return NewHashMap[K, V](func(k K) uint32 {
		switch x := any(k).(type) {
		case int:
			return uint32(x)
		case int32:
			return uint32(x)
		case uint32:
			return x
		case int64:
			return uint32(x)
		case uint64:
			return uint32(x)
		case Lit:
			return uint32(x)
		case Var:
			return uint32(x)
		}
		return uint32(maphash.Comparable(defaultMapSeed, k))
	}, func(a, b K) bool { return a == b })
}
func (m *Map[K, V]) Elems() int                { return m.count }
func (m *Map[K, V]) BucketCount() int          { return len(m.buckets) }
func (m *Map[K, V]) Bucket(i int) []Pair[K, V] { return m.buckets[i] }
func (m *Map[K, V]) Peek(k K) (V, bool) {
	if m.count > 0 {
		for _, p := range m.buckets[int(m.hash(k)%uint32(len(m.buckets)))] {
			if m.equal(p.Key, k) {
				return p.Data, true
			}
		}
	}
	var zero V
	return zero, false
}
func (m *Map[K, V]) Has(k K) bool { _, ok := m.Peek(k); return ok }
func (m *Map[K, V]) At(k K) *V {
	if m.count > 0 {
		ps := m.buckets[int(m.hash(k)%uint32(len(m.buckets)))]
		for i := range ps {
			if m.equal(ps[i].Key, k) {
				return &ps[i].Data
			}
		}
	}
	panic("map key missing")
}
func (m *Map[K, V]) Insert(k K, v V) {
	if m.Has(k) {
		panic("duplicate map key")
	}
	if m.count+1 > len(m.buckets) {
		n := mapPrimes[len(mapPrimes)-1]
		for _, p := range mapPrimes {
			if p > len(m.buckets) {
				n = p
				break
			}
		}
		if n <= len(m.buckets) {
			panic(OutOfMemoryException{})
		}
		old := m.buckets
		m.buckets = make([][]Pair[K, V], n)
		for _, ps := range old {
			for _, p := range ps {
				i := int(m.hash(p.Key) % uint32(n))
				m.buckets[i] = append(m.buckets[i], p)
			}
		}
	}
	i := int(m.hash(k) % uint32(len(m.buckets)))
	m.buckets[i] = append(m.buckets[i], Pair[K, V]{k, v})
	m.count++
}
func (m *Map[K, V]) Remove(k K) {
	if m.count > 0 {
		i := int(m.hash(k) % uint32(len(m.buckets)))
		ps := m.buckets[i]
		for j, p := range ps {
			if m.equal(p.Key, k) {
				ps[j] = ps[len(ps)-1]
				ps[len(ps)-1] = Pair[K, V]{}
				m.buckets[i] = ps[:len(ps)-1]
				m.count--
				return
			}
		}
	}
	panic("map key missing")
}
func (m *Map[K, V]) Clear() { m.buckets = nil; m.count = 0 }
func (m *Map[K, V]) MoveTo(to *Map[K, V]) {
	if m == to {
		return
	}
	to.buckets = m.buckets
	to.count = m.count
	m.buckets = nil
	m.count = 0
}
