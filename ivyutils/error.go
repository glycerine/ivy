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

// Locatable is an interface for objects that can provide a LocationTuple.
// This allows extractLocation to work with AST nodes without importing the ast package.
type Locatable interface {
	GetLinenoLT() *LocationTuple
}

// extractLocation converts various location types to *LocationTuple.
// Corresponds to Python: ast.lineno if hasattr(ast,'lineno') else Location()
// Python falls back to Location() (empty LocationTuple), never None.
func extractLocation(lineno interface{}) *LocationTuple {
	if lineno == nil {
		return &LocationTuple{} // Python: Location() = LocationTuple([None, None])
	}
	switch v := lineno.(type) {
	case *LocationTuple:
		return v
	case Locatable:
		lt := v.GetLinenoLT()
		if lt != nil {
			return lt
		}
		return &LocationTuple{}
	default:
		return &LocationTuple{} // Python: fallback to Location()
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

// HasFilename is implemented by errors that carry their own filename.
// Corresponds to Python's hasattr(e, 'filename') check in ErrorList.__repr__.
type HasFilename interface {
	GetFilename() string
}

// GetFilename returns the filename from the error's location, if any.
// Satisfies the HasFilename interface so IvyError works with ErrorList.
func (e *IvyError) GetFilename() string {
	if e.Lineno != nil {
		return e.Lineno.Filename
	}
	return ""
}

// Error implements the error interface.
// Corresponds to Python's ErrorList.__repr__:
//
//	pre = (self.filename + ': ') if hasattr(self,'filename') else ''
//	return '\n'.join((repr(e) if hasattr(e,'filename') else pre + str(e)) for e in self.errors)
func (e *ErrorList) Error() string {
	pre := ""
	if e.Filename != "" {
		pre = e.Filename + ": "
	}
	var parts []string
	for _, err := range e.Errors {
		// Python: hasattr(e, 'filename') — check if error has its own filename
		if hf, ok := err.(HasFilename); ok && hf.GetFilename() != "" {
			parts = append(parts, err.Error())
		} else {
			parts = append(parts, pre+err.Error())
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
// Corresponds to Python: warn(ast, msg) which does:
//
//	print(str(IvyError(ast,msg)).replace('error: ','warning: '))
//
// This creates a full IvyError (with reference chain) then replaces
// all "error: " with "warning: ".
func Warn(lineno interface{}, msg string) {
	// Temporarily ensure catch is true to prevent panic in NewIvyError
	oldCatch := Catch.GetBool()
	Catch.Value = true
	e := NewIvyError(lineno, msg)
	Catch.Value = oldCatch
	fmt.Println(strings.ReplaceAll(e.Error(), "error: ", "warning: "))
}

// ParseErrorListVar is a transitional global; use IvyUtilsConfig.ParseErrorList instead.
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
