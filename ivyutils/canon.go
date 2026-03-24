package ivyutils

type Canonical string

type Canonizer interface {
	Canon() Canonical
}
