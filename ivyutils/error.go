package ivyutils

import (
	"fmt"
	"os"
	"strings"
)

// IvyError is the general error type for Ivy with location and reference chain support.
// Corresponds to Python's IvyError exception class in ivy_utils.py.
type IvyError struct {
	Lineno *LocationTuple
	Msg    string
}

// NewIvyError creates a new IvyError. If the global Catch parameter is false,
// it prints the error and panics (matching Python's assert False behavior).
// The lineno argument can be: *LocationTuple, nil, or any other value (ignored).
// Corresponds to Python: IvyError(ast, msg)
func NewIvyError(lineno interface{}, msg string) *IvyError {
	loc := extractLocation(lineno)
	e := &IvyError{Lineno: loc, Msg: msg}
	if !Catch.GetBool() {
		fmt.Println(e.Error())
		panic("IvyError: " + msg) // Python does assert False
	}
	return e
}

// Error implements the error interface.
// Corresponds to Python's IvyError.__str__ with recursive reference chain.
func (e *IvyError) Error() string {
	return e.recur(e.Lineno)
}

func (e *IvyError) recur(lineno *LocationTuple) string {
	if lineno != nil && lineno.Reference != nil {
		res := e.recur(lineno.Reference)
		ref := &LocationTuple{Filename: lineno.Filename, Line: lineno.Line}
		return res + "\n" + ref.String() + "error: instantiated here"
	}
	prefix := ""
	if lineno != nil {
		prefix = lineno.String()
	}
	return prefix + "error: " + e.Msg
}

// extractLocation converts various location types to *LocationTuple.
func extractLocation(lineno interface{}) *LocationTuple {
	if lineno == nil {
		return nil
	}
	switch v := lineno.(type) {
	case *LocationTuple:
		return v
	default:
		return nil
	}
}

// IvyUndefined is raised for undefined references.
// Corresponds to Python's IvyUndefined(IvyError) subclass.
type IvyUndefined struct {
	IvyError
}

// NewIvyUndefined creates a new IvyUndefined error.
// Corresponds to Python: IvyUndefined(ast, name) which calls
// super().__init__(ast, "undefined: " + name)
func NewIvyUndefined(lineno interface{}, name string) *IvyUndefined {
	loc := extractLocation(lineno)
	e := &IvyUndefined{
		IvyError: IvyError{Lineno: loc, Msg: "undefined: " + name},
	}
	if !Catch.GetBool() {
		fmt.Println(e.Error())
		panic("IvyUndefined: " + name)
	}
	return e
}

// ErrorList holds a list of errors.
// Corresponds to Python's ErrorList(IvyError) class.
type ErrorList struct {
	Errors   []error
	Filename string
}

// NewErrorList creates a new ErrorList.
// Corresponds to Python: ErrorList(errors)
func NewErrorList(errors []error) *ErrorList {
	return &ErrorList{Errors: errors}
}

// Error implements the error interface.
// Corresponds to Python's ErrorList.__repr__: joins errors with newlines,
// conditionally prefixes with self.filename.
func (e *ErrorList) Error() string {
	var parts []string
	for _, err := range e.Errors {
		if ivyErr, ok := err.(*IvyError); ok && ivyErr.Lineno != nil && ivyErr.Lineno.Filename != "" {
			parts = append(parts, err.Error())
		} else if e.Filename != "" {
			parts = append(parts, e.Filename+": "+err.Error())
		} else {
			parts = append(parts, err.Error())
		}
	}
	return strings.Join(parts, "\n")
}

// WithErrorPrinter runs fn, catching IvyError panics and printing them before exiting.
// Corresponds to Python's ErrorPrinter context manager.
func WithErrorPrinter(fn func()) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case *IvyError:
				fmt.Println(e.Error())
				os.Exit(1)
			case *IvyUndefined:
				fmt.Println(e.Error())
				os.Exit(1)
			default:
				panic(r) // re-panic non-IvyError
			}
		}
	}()
	fn()
}

// Warn prints a warning message with location information.
// Corresponds to Python: warn(ast, msg) which creates an IvyError string
// and replaces "error:" with "warning:".
func Warn(lineno interface{}, msg string) {
	loc := extractLocation(lineno)
	prefix := ""
	if loc != nil {
		prefix = loc.String()
	}
	fmt.Println(prefix + "warning: " + msg)
}

// ParseErrorListVar is a global list for collecting parse errors.
// Corresponds to Python's module-level error_list = [].
var ParseErrorListVar []error

// ParseWith runs a parse function, collecting errors into ParseErrorListVar.
// Returns an ErrorList if any errors were collected.
// Corresponds to Python's parse_with(s, parser, lexer).
func ParseWith(parseFn func(string) (interface{}, error), s string) (interface{}, error) {
	ParseErrorListVar = nil
	result, err := parseFn(s)
	if err != nil {
		return nil, err
	}
	if len(ParseErrorListVar) > 0 {
		return nil, &ErrorList{Errors: ParseErrorListVar}
	}
	return result, nil
}

// PError reports a parse error by appending to ParseErrorListVar.
// Corresponds to Python's p_error(token).
func PError(lineno int, value string, msg string) {
	if value != "" {
		ParseErrorListVar = append(ParseErrorListVar, &IvyError{
			Lineno: Location("", lineno),
			Msg:    msg,
		})
	} else {
		ParseErrorListVar = append(ParseErrorListVar, &IvyError{
			Msg: "unexpected end of input",
		})
	}
}
