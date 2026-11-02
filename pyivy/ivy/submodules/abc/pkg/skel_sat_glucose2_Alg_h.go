package abc

// Hand-checked skeleton of src/sat/glucose2/Alg.h (namespace Gluco2).

// Gluco2Seq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type Gluco2Seq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// gluco2Remove is Gluco2::remove (static inline template).
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func gluco2Remove[T comparable, V Gluco2Seq[T]](ts V, t T) {
	panic("TODO, stubbed")
}

// gluco2Find is Gluco2::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func gluco2Find[T comparable, V Gluco2Seq[T]](ts V, t T) bool {
	panic("TODO, stubbed")
}

// gluco2Copy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func gluco2Copy[T any](from T, to *T) {
	panic("TODO, stubbed")
}

// gluco2CopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func gluco2CopyVec[T any](from *Gluco2Vec[T], to *Gluco2Vec[T], append bool) {
	panic("TODO, stubbed")
}

// gluco2Append is Gluco2::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func gluco2Append[T any](from *Gluco2Vec[T], to *Gluco2Vec[T]) {
	panic("TODO, stubbed")
}
