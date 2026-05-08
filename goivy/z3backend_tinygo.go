//go:build tinygo

package goivy

import "fmt"

type tinyGoZ3Backend struct {
	name string
}

func defaultZ3Backend() Z3Backend {
	return tinyGoZ3Backend{name: BackendCGo}
}

func newZ3BackendByCanonicalName(name string) Z3Backend {
	switch name {
	case BackendCGo, BackendWazero, BackendJSBrowser:
		return tinyGoZ3Backend{name: name}
	default:
		panic(fmt.Sprintf("unreachable canonical Z3 backend %q", name))
	}
}

func (b tinyGoZ3Backend) Z3BackendName() string {
	return b.name
}

func (tinyGoZ3Backend) NewZ3Context() *Z3Context {
	return &Z3Context{}
}

func (tinyGoZ3Backend) NewInterpolationZ3Context() *Z3Context {
	return &Z3Context{}
}

func (tinyGoZ3Backend) NewZ3Solver(ctx *Z3Context) *Z3Solver {
	return &Z3Solver{ctx: ctx}
}
