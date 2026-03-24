package ivyutils

// Canonical strings are 1:1 with the state
// they represent, so immune to internal collection
// reordering differences: the are s-expressions
// strings in compact (not pretty printed) form.
type Canonical string

// Canonizer supports Canon calls.
type Canonizer interface {
	// Canon returns a canonical S-expression string
	// for the given implementor.
	Canon() Canonical
}
