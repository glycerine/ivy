package logic

// IvyError is the general error type for Ivy.
type IvyError struct {
	Msg string
}

func (e *IvyError) Error() string {
	return e.Msg
}

// SortError is raised for sort-related errors (type mismatches, etc.).
type SortError struct {
	Msg string
}

func (e *SortError) Error() string {
	return e.Msg
}
