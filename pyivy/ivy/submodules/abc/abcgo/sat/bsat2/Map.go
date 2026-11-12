package bsat2

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

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}
type Map[K comparable, V any] struct {
	buckets [][]Pair[K, V]
	size    int
	hash    func(K) uint32
}

var mapPrimes = []int{31, 73, 151, 313, 643, 1291, 2593, 5233, 10501, 21013, 42073, 84181, 168451, 337219, 674701, 1349473, 2699299, 5398891, 10798093, 21596719, 43193641, 86387383, 172775299, 345550609, 691101253}

func NewMap[K comparable, V any](hash func(K) uint32) *Map[K, V] {
	if hash == nil {
		panic("nil map hash")
	}
	return &Map[K, V]{hash: hash}
}
func (m *Map[K, V]) Len() int         { return m.size }
func (m *Map[K, V]) BucketCount() int { return len(m.buckets) }
func (m *Map[K, V]) Bucket(index int) []Pair[K, V] {
	return append([]Pair[K, V](nil), m.buckets[index]...)
}
func (m *Map[K, V]) rehash() {
	capacity := 0
	for _, prime := range mapPrimes {
		if prime > len(m.buckets) {
			capacity = prime
			break
		}
	}
	if capacity == 0 {
		panic("hash table capacity overflow")
	}
	old := m.buckets
	m.buckets = make([][]Pair[K, V], capacity)
	for _, bucket := range old {
		for _, pair := range bucket {
			i := m.hash(pair.Key) % uint32(capacity)
			m.buckets[i] = append(m.buckets[i], pair)
		}
	}
}
func (m *Map[K, V]) Get(key K) (V, bool) {
	if len(m.buckets) > 0 {
		for _, pair := range m.buckets[m.hash(key)%uint32(len(m.buckets))] {
			if pair.Key == key {
				return pair.Value, true
			}
		}
	}
	var zero V
	return zero, false
}
func (m *Map[K, V]) Insert(key K, value V) {
	if _, exists := m.Get(key); exists {
		panic("duplicate map insertion")
	}
	if m.size+1 > len(m.buckets) {
		m.rehash()
	}
	i := m.hash(key) % uint32(len(m.buckets))
	m.buckets[i] = append(m.buckets[i], Pair[K, V]{key, value})
	m.size++
}
func (m *Map[K, V]) Set(key K, value V) {
	if len(m.buckets) > 0 {
		bucket := m.buckets[m.hash(key)%uint32(len(m.buckets))]
		for i := range bucket {
			if bucket[i].Key == key {
				bucket[i].Value = value
				return
			}
		}
	}
	m.Insert(key, value)
}
func (m *Map[K, V]) Remove(key K) bool {
	if len(m.buckets) == 0 {
		return false
	}
	i := m.hash(key) % uint32(len(m.buckets))
	bucket := m.buckets[i]
	for j, pair := range bucket {
		if pair.Key == key {
			bucket[j] = bucket[len(bucket)-1]
			var zero Pair[K, V]
			bucket[len(bucket)-1] = zero
			m.buckets[i] = bucket[:len(bucket)-1]
			m.size--
			return true
		}
	}
	return false
}
func (m *Map[K, V]) Clear() { m.buckets = nil; m.size = 0 }
func (m *Map[K, V]) MoveTo(to *Map[K, V]) {
	if m == to {
		return
	}
	*to = *m
	m.buckets = nil
	m.size = 0
}
