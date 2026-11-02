package abc

// Hand-checked skeleton of src/sat/glucose/Alg.h (namespace Gluco).

// GlucoSeq is what Alg.h's remove/find/copy templates require of their container
// argument V (vec<T>, Clause): size(), operator[] and pop().
type GlucoSeq[T any] interface {
	Size() int
	At(i int) *T
	Pop()
}

// glucoRemove is Gluco::remove (static inline template).
// C: template<class V, class T> static inline void remove(V& ts, const T& t)
func glucoRemove[T comparable, V GlucoSeq[T]](ts V, t T) {
	panic("TODO, stubbed")
}

// glucoFind is Gluco::find (static inline template).
// C: template<class V, class T> static inline bool find(V& ts, const T& t)
func glucoFind[T comparable, V GlucoSeq[T]](ts V, t T) bool {
	panic("TODO, stubbed")
}

// glucoCopy is the overload copy(const T& from, T& to) (static inline template).
// C: template<class T> static inline void copy(const T& from, T& to)
func glucoCopy[T any](from T, to *T) {
	panic("TODO, stubbed")
}

// glucoCopyVec is the overload copy(const vec<T>& from, vec<T>& to, bool append = false).
// C: template<class T> static inline void copy(const vec<T>& from, vec<T>& to, bool append = false)
func glucoCopyVec[T any](from *GlucoVec[T], to *GlucoVec[T], append bool) {
	panic("TODO, stubbed")
}

// glucoAppend is Gluco::append (static inline template).
// C: template<class T> static inline void append(const vec<T>& from, vec<T>& to)
func glucoAppend[T any](from *GlucoVec[T], to *GlucoVec[T]) {
	panic("TODO, stubbed")
}
