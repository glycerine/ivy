package abc

// Port of src/sat/glucose2/Alg.h (namespace Gluco2).

// gluco2Ne is the `!=` operator of a comparable T, passed to remove/find as ne
// (Watcher has its own operator!= comparing cref only, which callers pass instead).
func gluco2Ne[T comparable](a T, b T) bool { return a != b }

// Gluco2Seq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type Gluco2Seq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// gluco2Remove is Gluco2::remove (static inline template): removes the first
// occurrence of t, keeping the order of the others.
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func gluco2Remove[T any, V Gluco2Seq[T]](ts V, t T, ne func(a T, b T) bool) {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	if j >= ts.Size() {
		panic("gluco2Remove: element not found")
	}
	for ; j < ts.Size()-1; j++ {
		*ts.At(j) = *ts.At(j + 1)
	}
	ts.Pop()
}

// gluco2Find is Gluco2::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func gluco2Find[T any, V Gluco2Seq[T]](ts V, t T, ne func(a T, b T) bool) bool {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	return j < ts.Size()
}

// gluco2Copy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func gluco2Copy[T any](from T, to *T) {
	*to = from
}

// gluco2CopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func gluco2CopyVec[T any](from *Gluco2Vec[T], to *Gluco2Vec[T], append bool) {
	if !append {
		to.Clear(false)
	}
	for i := 0; i < from.Size(); i++ {
		to.Push()
		gluco2Copy(*from.At(i), to.Last())
	}
}

// gluco2Append is Gluco2::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func gluco2Append[T any](from *Gluco2Vec[T], to *Gluco2Vec[T]) {
	gluco2CopyVec(from, to, true)
}
