// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// infer should be an internal detail,
// but widely used packages access it using linkname.
// Notable members of the hall of shame include:
//   - github.com/goplus/gox
//
// Do not remove or change the type signature.
// See go.dev/issue/67401.
//
//go:linkname badlinkname_Checker_infer go/types.(*Checker).infer
export function badlinkname_Checker_infer(check, posn, tparams, targs, params, args, reverse, err) {
    return check.infer(posn, tparams, targs, params, args, reverse, err);
}
