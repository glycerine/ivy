This project is goivy. goivy is meant to be a MECHANICAL PORT of 
the Ivy python project into Go.

A. Source of truth: the python version 

The python version of Ivy is the source of truth, and
all new features and bug fixes should conform to its example. It is:
/Users/jaten/pyivy/ivy

The Go must follow the Python for all execution flow.

The goivy Go port, which should conform to the original Python Ivy logic:
/Users/jaten/go/src/github.com/glycerine/goivy

B. MECHANICAL PORT RULES:

1. Python class names -> Go struct names: SAME NAME. App stays App. Atom stays Atom. Symbol
stays Symbol. Variable stays Variable. Do not rename.

2. Python function names -> Go function names: SAME NAME with capital first letter.
substitute_ast -> SubstituteAst. Do not invent new names.

3. Python field names -> Go field names: SAME NAME with capital first letter. self.rep ->
Rep. self.args -> Args. Do not rename fields.

4. One Python file -> one Go file with the same base name. ivy_logic_utils.py ->
ivy_logic_utils.go. Do not reorganize into different packages.

5. Do not merge Python classes. If Python has App and Atom as separate classes, Go has App
and Atom as separate structs.

6. Do not omit functions. If a Python function exists, a Go function must exist 
with the same name (different capitalization and substituting PascalCase for snake_case is allowed).

7. Do not add abstractions, interfaces, or helper types that don't exist in Python.

8. When in doubt, translate literally. A wrong but literal translation is easier to fix than a creative one.

