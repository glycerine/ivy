package abc

// Hand-checked skeleton of src/sat/bsat2/Alg.h (namespace Minisat).

// MinisatSeq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type MinisatSeq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// minisatRemove is Minisat::remove (static inline template).
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func minisatRemove[T comparable, V MinisatSeq[T]](ts V, t T) {
	panic("TODO, stubbed")
}

// minisatFind is Minisat::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func minisatFind[T comparable, V MinisatSeq[T]](ts V, t T) bool {
	panic("TODO, stubbed")
}

// minisatCopy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func minisatCopy[T any](from T, to *T) {
	panic("TODO, stubbed")
}

// minisatCopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func minisatCopyVec[T any](from *MinisatVec[T], to *MinisatVec[T], append bool) {
	panic("TODO, stubbed")
}

// minisatAppend is Minisat::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func minisatAppend[T any](from *MinisatVec[T], to *MinisatVec[T]) {
	panic("TODO, stubbed")
}
