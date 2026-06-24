// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This should properly be in infer.go, but that file is auto-generated.

import { Checker, type positioner } from "./check.js";
import type { Type } from "./type.js";
import type { TypeParam } from "./typeparam.js";
import type { Tuple } from "./tuple.js";
import type { operand } from "./operand.js";
import type { error_ } from "./errors.js";

// infer should be an internal detail,
// but widely used packages access it using linkname.
// Notable members of the hall of shame include:
//   - github.com/goplus/gox
//
// Do not remove or change the type signature.
// See go.dev/issue/67401.
//
//go:linkname badlinkname_Checker_infer go/types.(*Checker).infer
export function badlinkname_Checker_infer(check: Checker, posn: positioner, tparams: TypeParam[], targs: Type[], params: Tuple | null, args: operand[] | null, reverse: boolean, err: error_): Type[] | null {
  return check.infer(posn, tparams, targs, params, args, reverse, err);
}

