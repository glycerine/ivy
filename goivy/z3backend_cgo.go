//go:build !tinygo

package goivy

import "fmt"

type cgoZ3Backend struct{}

func defaultZ3Backend() Z3Backend {
	return cgoZ3Backend{}
}

func newZ3BackendByCanonicalName(name string) Z3Backend {
	switch name {
	case BackendCGo:
		return cgoZ3Backend{}
	case BackendWazero, BackendJSBrowser:
		return unimplementedZ3Backend{name: name}
	default:
		panic(fmt.Sprintf("unreachable canonical Z3 backend %q", name))
	}
}

func (cgoZ3Backend) Z3BackendName() string {
	return BackendCGo
}

func (b cgoZ3Backend) NewZ3Context() *Z3Context {
	return newCGoZ3Context(b)
}

func (b cgoZ3Backend) NewInterpolationZ3Context() *Z3Context {
	return newCGoInterpolationZ3Context(b)
}

func (cgoZ3Backend) NewZ3Solver(ctx *Z3Context) *Z3Solver {
	return newCGoZ3Solver(ctx)
}
