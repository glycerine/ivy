package abc

// Port of src/sat/bsat2/Alg.h (namespace Minisat).

// minisatNe is the `!=` operator of a comparable T, passed to remove/find as ne
// (Watcher has its own operator!= comparing cref only, which callers pass instead).
func minisatNe[T comparable](a T, b T) bool { return a != b }

// MinisatSeq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type MinisatSeq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// minisatRemove is Minisat::remove (static inline template): removes the first
// occurrence of t, keeping the order of the others.
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func minisatRemove[T any, V MinisatSeq[T]](ts V, t T, ne func(a T, b T) bool) {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	if j >= ts.Size() {
		panic("minisatRemove: element not found")
	}
	for ; j < ts.Size()-1; j++ {
		*ts.At(j) = *ts.At(j + 1)
	}
	ts.Pop()
}

// minisatFind is Minisat::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func minisatFind[T any, V MinisatSeq[T]](ts V, t T, ne func(a T, b T) bool) bool {
	j := 0
	for ; j < ts.Size() && ne(*ts.At(j), t); j++ {
	}
	return j < ts.Size()
}

// minisatCopy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func minisatCopy[T any](from T, to *T) {
	*to = from
}

// minisatCopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func minisatCopyVec[T any](from *MinisatVec[T], to *MinisatVec[T], append bool) {
	if !append {
		to.Clear(false)
	}
	for i := 0; i < from.Size(); i++ {
		to.Push()
		minisatCopy(*from.At(i), to.Last())
	}
}

// minisatAppend is Minisat::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func minisatAppend[T any](from *MinisatVec[T], to *MinisatVec[T]) {
	minisatCopyVec(from, to, true)
}
