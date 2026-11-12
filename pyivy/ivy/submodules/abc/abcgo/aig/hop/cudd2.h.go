// CUDD-operation recording API, ported from ABC cudd2.h (Alan Mishchenko).
package hop

// BDDRef represents an opaque BDD identity and its complemented-edge bit.
// Identity must be comparable (normally a pointer or an integer handle).
// uintptr handles are also accepted directly with ABC's low-bit convention.
type BDDRef struct {
	Identity   any
	Complement bool
}
