package goivy

import "fmt"

type unimplementedZ3Backend struct {
	name string
}

func (b unimplementedZ3Backend) Z3BackendName() string {
	return b.name
}

func (b unimplementedZ3Backend) NewZ3Context() *Z3Context {
	panic(fmt.Sprintf("Z3 backend %q is selected but not wired yet", b.name))
}

func (b unimplementedZ3Backend) NewInterpolationZ3Context() *Z3Context {
	panic(fmt.Sprintf("Z3 backend %q interpolation context is selected but not wired yet", b.name))
}

func (b unimplementedZ3Backend) NewZ3Solver(ctx *Z3Context) *Z3Solver {
	panic(fmt.Sprintf("Z3 backend %q solver is selected but not wired yet", b.name))
}
