package abc

// Port of src/sat/glucose/Alg.h (namespace Gluco).

// glucoNe is the `!=` operator of a comparable T, passed to remove/find as ne
// (Watcher has its own operator!= comparing cref only, which callers pass instead).
func glucoNe[T comparable](a T, b T) bool { return a != b }

// GlucoSeq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type GlucoSeq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// glucoRemove is Gluco::remove (static inline template): removes the first
// occurrence of t, keeping the order of the others.
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func glucoRemove[T any, V GlucoSeq[T]](ts V, t T, ne func(a T, b T) bool) {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	if j >= ts.Size() {
		panic("glucoRemove: element not found")
	}
	for ; j < ts.Size()-1; j++ {
		*ts.At(j) = *ts.At(j + 1)
	}
	ts.Pop()
}

// glucoFind is Gluco::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func glucoFind[T any, V GlucoSeq[T]](ts V, t T, ne func(a T, b T) bool) bool {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	return j < ts.Size()
}

// glucoCopy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func glucoCopy[T any](from T, to *T) {
	*to = from
}

// glucoCopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func glucoCopyVec[T any](from *GlucoVec[T], to *GlucoVec[T], append bool) {
	if !append {
		to.Clear(false)
	}
	for i := 0; i < from.Size(); i++ {
		to.Push()
		glucoCopy(*from.At(i), to.Last())
	}
}

// glucoAppend is Gluco::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func glucoAppend[T any](from *GlucoVec[T], to *GlucoVec[T]) {
	glucoCopyVec(from, to, true)
}
