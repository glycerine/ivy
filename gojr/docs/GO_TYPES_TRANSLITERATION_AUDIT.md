# Go Types Transliteration Audit

This file is the audit baseline for the fresh symbol-by-symbol TypeScript transliteration of `/usr/local/go1.27rc1/src/go/types`.

Target TypeScript files currently included in the inventory are only files under `gojr/src/go/types/`. The legacy frontend checker is deliberately excluded and must not count as the `go/types` port.

Current same-kind exact-symbol audit: 826/1121 upstream go/types declarations have a matching symbol in the current TypeScript inventory; 295 are missing by exact symbol name.

Rule for this pass: every upstream `go/types` declaration must have a corresponding same-kind TypeScript declaration, and function bodies must be ported with visibly isomorphic control flow before being marked complete. `Present` means a same-kind exact or receiver-qualified symbol name exists in the fresh port; it does not prove faithful logic yet. `Missing` means the fresh port has not yet reached symbol-level correspondence.

Generated from Go source using /private/tmp/go_types_inventory_127.json; TypeScript symbols were collected with the TypeScript compiler API from 53 fresh-port files.

## go/types

### alias.go

#### Types
- `Alias` (struct) - Present: ClassDeclaration Alias (gojr/src/go/types/alias.ts) - structure: not yet 1:1-audited

#### Functions
- `NewAlias` @ /usr/local/go1.27rc1/src/go/types/alias.go:38:1 - Present: FunctionDeclaration NewAlias (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Unalias` @ /usr/local/go1.27rc1/src/go/types/alias.go:86:1 - Present: FunctionDeclaration Unalias (gojr/src/go/types/alias.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `unalias` @ /usr/local/go1.27rc1/src/go/types/alias.go:93:1 - Present: FunctionDeclaration unalias (gojr/src/go/types/alias.ts) - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=2, defer=0, go=0, call=0, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `asNamed` @ /usr/local/go1.27rc1/src/go/types/alias.go:109:1 - Present: FunctionDeclaration asNamed (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Alias.Obj` @ /usr/local/go1.27rc1/src/go/types/alias.go:47:1 - Present: Method Alias.Obj (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.String` @ /usr/local/go1.27rc1/src/go/types/alias.go:49:1 - Present: Method Alias.String (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.Underlying` @ /usr/local/go1.27rc1/src/go/types/alias.go:56:1 - Present: Method Alias.Underlying (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.Origin` @ /usr/local/go1.27rc1/src/go/types/alias.go:60:1 - Present: Method Alias.Origin (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.TypeParams` @ /usr/local/go1.27rc1/src/go/types/alias.go:64:1 - Present: Method Alias.TypeParams (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.SetTypeParams` @ /usr/local/go1.27rc1/src/go/types/alias.go:68:1 - Present: Method Alias.SetTypeParams (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.TypeArgs` @ /usr/local/go1.27rc1/src/go/types/alias.go:75:1 - Present: Method Alias.TypeArgs (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.Rhs` @ /usr/local/go1.27rc1/src/go/types/alias.go:79:1 - Present: Method Alias.Rhs (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newAlias` @ /usr/local/go1.27rc1/src/go/types/alias.go:116:1 - Present: Method Checker.newAlias (gojr/src/go/types/check.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newAliasInstance` @ /usr/local/go1.27rc1/src/go/types/alias.go:136:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=1, defer=0, go=0, call=9, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Alias.cleanup` @ /usr/local/go1.27rc1/src/go/types/alias.go:147:1 - Present: Method Alias.cleanup (gojr/src/go/types/alias.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### api.go

#### Types
- `Error` (struct) - Present: ClassDeclaration Error (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `ArgumentError` (struct) - Present: ClassDeclaration ArgumentError (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `Importer` (interface) - Present: InterfaceDeclaration Importer (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `ImportMode` (*ast.Ident) - Present: TypeAliasDeclaration ImportMode (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `ImporterFrom` (interface) - Present: InterfaceDeclaration ImporterFrom (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `Config` (struct) - Present: ClassDeclaration Config (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `Info` (struct) - Present: ClassDeclaration Info (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `TypeAndValue` (struct) - Present: ClassDeclaration TypeAndValue (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `Instance` (struct) - Present: ClassDeclaration Instance (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited
- `Initializer` (struct) - Present: ClassDeclaration Initializer (gojr/src/go/types/api.ts) - structure: not yet 1:1-audited

#### Functions
- `srcimporter_setUsesCgo` @ /usr/local/go1.27rc1/src/go/types/api.go:194:1 - Present: FunctionDeclaration srcimporter_setUsesCgo (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `Error.Error` @ /usr/local/go1.27rc1/src/go/types/api.go:69:1 - Present: Method Error.Error (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*ArgumentError.Error` @ /usr/local/go1.27rc1/src/go/types/api.go:79:1 - Present: Method ArgumentError.Error (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*ArgumentError.Unwrap` @ /usr/local/go1.27rc1/src/go/types/api.go:80:1 - Present: Method ArgumentError.Unwrap (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Info.recordTypes` @ /usr/local/go1.27rc1/src/go/types/api.go:330:1 - Present: Method Info.recordTypes (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Info.TypeOf` @ /usr/local/go1.27rc1/src/go/types/api.go:336:1 - Present: Method Info.TypeOf (gojr/src/go/types/api.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=3, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Info.ObjectOf` @ /usr/local/go1.27rc1/src/go/types/api.go:355:1 - Present: Method Info.ObjectOf (gojr/src/go/types/api.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Info.PkgNameOf` @ /usr/local/go1.27rc1/src/go/types/api.go:368:1 - Present: Method Info.PkgNameOf (gojr/src/go/types/api.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.IsVoid` @ /usr/local/go1.27rc1/src/go/types/api.go:389:1 - Present: Method TypeAndValue.IsVoid (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.IsType` @ /usr/local/go1.27rc1/src/go/types/api.go:394:1 - Present: Method TypeAndValue.IsType (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.IsBuiltin` @ /usr/local/go1.27rc1/src/go/types/api.go:400:1 - Present: Method TypeAndValue.IsBuiltin (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.IsValue` @ /usr/local/go1.27rc1/src/go/types/api.go:407:1 - Present: Method TypeAndValue.IsValue (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.IsNil` @ /usr/local/go1.27rc1/src/go/types/api.go:417:1 - Present: Method TypeAndValue.IsNil (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.Addressable` @ /usr/local/go1.27rc1/src/go/types/api.go:423:1 - Present: Method TypeAndValue.Addressable (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.Assignable` @ /usr/local/go1.27rc1/src/go/types/api.go:429:1 - Present: Method TypeAndValue.Assignable (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `TypeAndValue.HasOk` @ /usr/local/go1.27rc1/src/go/types/api.go:435:1 - Present: Method TypeAndValue.HasOk (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Instance.String` @ /usr/local/go1.27rc1/src/go/types/api.go:448:1 - Present: Method Instance.String (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Initializer.String` @ /usr/local/go1.27rc1/src/go/types/api.go:460:1 - Present: Method Initializer.String (gojr/src/go/types/api.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=6, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Config.Check` @ /usr/local/go1.27rc1/src/go/types/api.go:484:1 - Present: Method Config.Check (gojr/src/go/types/api.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### api_predicates.go

#### Functions
- `AssertableTo` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:19:1 - Present: FunctionDeclaration AssertableTo (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=4, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `AssignableTo` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:33:1 - Present: FunctionDeclaration AssignableTo (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `ConvertibleTo` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:44:1 - Present: FunctionDeclaration ConvertibleTo (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `Implements` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:53:1 - Present: FunctionDeclaration Implements (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=5, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Satisfies` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:70:1 - Present: FunctionDeclaration Satisfies (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Identical` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:88:1 - Present: FunctionDeclaration Identical (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `IdenticalIgnoreTags` @ /usr/local/go1.27rc1/src/go/types/api_predicates.go:95:1 - Present: FunctionDeclaration IdenticalIgnoreTags (gojr/src/go/types/api_predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### array.go

#### Types
- `Array` (struct) - Present: ClassDeclaration Array (gojr/src/go/types/array.ts) - structure: not yet 1:1-audited

#### Functions
- `NewArray` @ /usr/local/go1.27rc1/src/go/types/array.go:18:1 - Present: FunctionDeclaration NewArray (gojr/src/go/types/array.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Array.Len` @ /usr/local/go1.27rc1/src/go/types/array.go:22:1 - Present: Method Array.Len (gojr/src/go/types/array.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Array.Elem` @ /usr/local/go1.27rc1/src/go/types/array.go:25:1 - Present: Method Array.Elem (gojr/src/go/types/array.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Array.Underlying` @ /usr/local/go1.27rc1/src/go/types/array.go:27:1 - Present: Method Array.Underlying (gojr/src/go/types/array.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Array.String` @ /usr/local/go1.27rc1/src/go/types/array.go:28:1 - Present: Method Array.String (gojr/src/go/types/array.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### assignments.go

#### Functions
- `operandTypes` @ /usr/local/go1.27rc1/src/go/types/assignments.go:281:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `varTypes` @ /usr/local/go1.27rc1/src/go/types/assignments.go:289:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `measure` @ /usr/local/go1.27rc1/src/go/types/assignments.go:346:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.assignment` @ /usr/local/go1.27rc1/src/go/types/assignments.go:24:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=14, for=0, range=0, switch=2, typeSwitch=0, select=0, branch=0, assign=12, return=7, defer=0, go=0, call=33, binary=13, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.initConst` @ /usr/local/go1.27rc1/src/go/types/assignments.go:120:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=3, defer=0, go=0, call=12, binary=6, unary=4, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.initVar` @ /usr/local/go1.27rc1/src/go/types/assignments.go:155:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=2, defer=0, go=0, call=11, binary=5, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.lhsVar` @ /usr/local/go1.27rc1/src/go/types/assignments.go:187:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=8, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=7, return=6, defer=0, go=0, call=14, binary=11, unary=6, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.assignVar` @ /usr/local/go1.27rc1/src/go/types/assignments.go:251:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=10, binary=6, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.typesSummary` @ /usr/local/go1.27rc1/src/go/types/assignments.go:303:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=1, switch=2, typeSwitch=0, select=0, branch=1, assign=9, return=1, defer=0, go=0, call=10, binary=8, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.assignError` @ /usr/local/go1.27rc1/src/go/types/assignments.go:353:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=6, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.returnError` @ /usr/local/go1.27rc1/src/go/types/assignments.go:367:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=0, defer=0, go=0, call=11, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.initVars` @ /usr/local/go1.27rc1/src/go/types/assignments.go:387:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=14, for=0, range=4, switch=0, typeSwitch=0, select=0, branch=0, assign=13, return=3, defer=0, go=0, call=19, binary=22, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.assignVars` @ /usr/local/go1.27rc1/src/go/types/assignments.go:473:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=8, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=3, defer=0, go=0, call=16, binary=13, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.shortVarDecl` @ /usr/local/go1.27rc1/src/go/types/assignments.go:532:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=8, for=0, range=3, switch=0, typeSwitch=0, select=0, branch=3, assign=20, return=1, defer=0, go=0, call=26, binary=9, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### badlinkname.go

#### Functions
- `badlinkname_Checker_infer` @ /usr/local/go1.27rc1/src/go/types/badlinkname.go:20:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### basic.go

#### Constants
- `Invalid` - Present: EnumMember Invalid in BasicKind (gojr/src/go/types/basic.ts); Variable Invalid (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Bool` - Present: EnumMember Bool in BasicKind (gojr/src/go/types/basic.ts); Variable Bool (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Int` - Present: EnumMember Int in BasicKind (gojr/src/go/types/basic.ts); Variable Int (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Int8` - Present: EnumMember Int8 in BasicKind (gojr/src/go/types/basic.ts); Variable Int8 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Int16` - Present: EnumMember Int16 in BasicKind (gojr/src/go/types/basic.ts); Variable Int16 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Int32` - Present: EnumMember Int32 in BasicKind (gojr/src/go/types/basic.ts); Variable Int32 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Int64` - Present: EnumMember Int64 in BasicKind (gojr/src/go/types/basic.ts); Variable Int64 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uint` - Present: EnumMember Uint in BasicKind (gojr/src/go/types/basic.ts); Variable Uint (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uint8` - Present: EnumMember Uint8 in BasicKind (gojr/src/go/types/basic.ts); Variable Uint8 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uint16` - Present: EnumMember Uint16 in BasicKind (gojr/src/go/types/basic.ts); Variable Uint16 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uint32` - Present: EnumMember Uint32 in BasicKind (gojr/src/go/types/basic.ts); Variable Uint32 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uint64` - Present: EnumMember Uint64 in BasicKind (gojr/src/go/types/basic.ts); Variable Uint64 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Uintptr` - Present: EnumMember Uintptr in BasicKind (gojr/src/go/types/basic.ts); Variable Uintptr (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Float32` - Present: EnumMember Float32 in BasicKind (gojr/src/go/types/basic.ts); Variable Float32 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Float64` - Present: EnumMember Float64 in BasicKind (gojr/src/go/types/basic.ts); Variable Float64 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Complex64` - Present: EnumMember Complex64 in BasicKind (gojr/src/go/types/basic.ts); Variable Complex64 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Complex128` - Present: EnumMember Complex128 in BasicKind (gojr/src/go/types/basic.ts); Variable Complex128 (gojr/src/go/types/basic.ts) - logic: declaration-only
- `String` - Present: EnumMember String in BasicKind (gojr/src/go/types/basic.ts); Variable String (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UnsafePointer` - Present: EnumMember UnsafePointer in BasicKind (gojr/src/go/types/basic.ts); Variable UnsafePointer (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedBool` - Present: EnumMember UntypedBool in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedBool (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedInt` - Present: EnumMember UntypedInt in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedInt (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedRune` - Present: EnumMember UntypedRune in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedRune (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedFloat` - Present: EnumMember UntypedFloat in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedFloat (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedComplex` - Present: EnumMember UntypedComplex in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedComplex (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedString` - Present: EnumMember UntypedString in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedString (gojr/src/go/types/basic.ts) - logic: declaration-only
- `UntypedNil` - Present: EnumMember UntypedNil in BasicKind (gojr/src/go/types/basic.ts); Variable UntypedNil (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Byte` - Present: Variable Byte (gojr/src/go/types/basic.ts) - logic: declaration-only
- `Rune` - Present: Variable Rune (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsBoolean` - Present: Variable IsBoolean (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsInteger` - Present: Variable IsInteger (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsUnsigned` - Present: Variable IsUnsigned (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsFloat` - Present: Variable IsFloat (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsComplex` - Present: Variable IsComplex (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsString` - Present: Variable IsString (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsUntyped` - Present: Variable IsUntyped (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsOrdered` - Present: Variable IsOrdered (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsNumeric` - Present: Variable IsNumeric (gojr/src/go/types/basic.ts) - logic: declaration-only
- `IsConstType` - Present: Variable IsConstType (gojr/src/go/types/basic.ts) - logic: declaration-only

#### Types
- `BasicKind` (*ast.Ident) - Present: EnumDeclaration BasicKind (gojr/src/go/types/basic.ts) - structure: not yet 1:1-audited
- `BasicInfo` (*ast.Ident) - Present: TypeAliasDeclaration BasicInfo (gojr/src/go/types/basic.ts) - structure: not yet 1:1-audited
- `Basic` (struct) - Present: ClassDeclaration Basic (gojr/src/go/types/basic.ts) - structure: not yet 1:1-audited

#### Methods
- `*Basic.Kind` @ /usr/local/go1.27rc1/src/go/types/basic.go:76:1 - Present: Method Basic.Kind (gojr/src/go/types/basic.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Basic.Info` @ /usr/local/go1.27rc1/src/go/types/basic.go:79:1 - Present: Method Basic.Info (gojr/src/go/types/basic.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Basic.Name` @ /usr/local/go1.27rc1/src/go/types/basic.go:82:1 - Present: Method Basic.Name (gojr/src/go/types/basic.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Basic.Underlying` @ /usr/local/go1.27rc1/src/go/types/basic.go:84:1 - Present: Method Basic.Underlying (gojr/src/go/types/basic.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Basic.String` @ /usr/local/go1.27rc1/src/go/types/basic.go:85:1 - Present: Method Basic.String (gojr/src/go/types/basic.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### builtins.go

#### Functions
- `sliceElem` @ /usr/local/go1.27rc1/src/go/types/builtins.go:992:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=4, defer=0, go=0, call=7, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `makeSig` @ /usr/local/go1.27rc1/src/go/types/builtins.go:1135:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=9, binary=1, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `arrayPtrDeref` @ /usr/local/go1.27rc1/src/go/types/builtins.go:1151:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.builtin` @ /usr/local/go1.27rc1/src/go/types/builtins.go:23:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=128, for=0, range=9, switch=6, typeSwitch=6, select=0, branch=4, assign=154, return=73, defer=1, go=0, call=362, binary=126, unary=46, composite=2, funcLiteral=9 - logic: not yet 1:1-audited
- `*Checker.hasVarSize` @ /usr/local/go1.27rc1/src/go/types/builtins.go:1017:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=8, return=7, defer=2, go=0, call=18, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.applyTypeFunc` @ /usr/local/go1.27rc1/src/go/types/builtins.go:1085:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=9, return=6, defer=0, go=0, call=14, binary=3, unary=1, composite=1, funcLiteral=1 - logic: not yet 1:1-audited

### call.go

#### Variables
- `cgoPrefixes` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Methods
- `*Checker.funcInst` @ /usr/local/go1.27rc1/src/go/types/call.go:34:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=21, return=5, defer=0, go=0, call=37, binary=17, unary=5, composite=3, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.instantiateSignature` @ /usr/local/go1.27rc1/src/go/types/call.go:132:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=1, go=0, call=30, binary=6, unary=0, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*Checker.callExpr` @ /usr/local/go1.27rc1/src/go/types/call.go:172:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=20, for=0, range=0, switch=3, typeSwitch=0, select=0, branch=2, assign=38, return=12, defer=0, go=0, call=73, binary=26, unary=7, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.exprList` @ /usr/local/go1.27rc1/src/go/types/call.go:355:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=4, binary=2, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.genericExprList` @ /usr/local/go1.27rc1/src/go/types/call.go:377:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=0, range=3, switch=0, typeSwitch=0, select=0, branch=0, assign=21, return=1, defer=1, go=0, call=28, binary=19, unary=13, composite=4, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.arguments` @ /usr/local/go1.27rc1/src/go/types/call.go:474:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=19, for=3, range=3, switch=0, typeSwitch=1, select=0, branch=0, assign=47, return=5, defer=0, go=0, call=84, binary=27, unary=3, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.selector` @ /usr/local/go1.27rc1/src/go/types/call.go:686:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=33, for=0, range=1, switch=1, typeSwitch=2, select=0, branch=14, assign=60, return=2, defer=0, go=0, call=91, binary=34, unary=11, composite=3, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.use` @ /usr/local/go1.27rc1/src/go/types/call.go:1002:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.useLHS` @ /usr/local/go1.27rc1/src/go/types/call.go:1007:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.useN` @ /usr/local/go1.27rc1/src/go/types/call.go:1009:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.use1` @ /usr/local/go1.27rc1/src/go/types/call.go:1019:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=1, assign=7, return=1, defer=0, go=0, call=5, binary=6, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### chan.go

#### Constants
- `SendRecv` - Present: EnumMember SendRecv in ChanDir (gojr/src/go/types/chan.ts) - logic: declaration-only
- `SendOnly` - Present: EnumMember SendOnly in ChanDir (gojr/src/go/types/chan.ts) - logic: declaration-only
- `RecvOnly` - Present: EnumMember RecvOnly in ChanDir (gojr/src/go/types/chan.ts) - logic: declaration-only

#### Types
- `Chan` (struct) - Present: ClassDeclaration Chan (gojr/src/go/types/chan.ts) - structure: not yet 1:1-audited
- `ChanDir` (*ast.Ident) - Present: EnumDeclaration ChanDir (gojr/src/go/types/chan.ts) - structure: not yet 1:1-audited

#### Functions
- `NewChan` @ /usr/local/go1.27rc1/src/go/types/chan.go:27:1 - Present: FunctionDeclaration NewChan (gojr/src/go/types/chan.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Chan.Dir` @ /usr/local/go1.27rc1/src/go/types/chan.go:32:1 - Present: Method Chan.Dir (gojr/src/go/types/chan.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Chan.Elem` @ /usr/local/go1.27rc1/src/go/types/chan.go:35:1 - Present: Method Chan.Elem (gojr/src/go/types/chan.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Chan.Underlying` @ /usr/local/go1.27rc1/src/go/types/chan.go:37:1 - Present: Method Chan.Underlying (gojr/src/go/types/chan.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Chan.String` @ /usr/local/go1.27rc1/src/go/types/chan.go:38:1 - Present: Method Chan.String (gojr/src/go/types/chan.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### check.go

#### Constants
- `debug` - Present: Variable debug (gojr/src/go/types/check.ts) - logic: declaration-only
- `tracePos` - Present: Variable tracePos (gojr/src/go/types/check.ts) - logic: declaration-only

#### Variables
- `nopos` - Present: Variable nopos (gojr/src/go/types/check.ts) - logic: declaration-only
- `noposn` - Present: Variable noposn (gojr/src/go/types/check.ts) - logic: declaration-only

#### Types
- `exprInfo` (struct) - Present: ClassDeclaration exprInfo (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `environment` (struct) - Present: ClassDeclaration environment (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `importKey` (struct) - Present: ClassDeclaration importKey (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `dotImportKey` (struct) - Present: ClassDeclaration dotImportKey (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `action` (struct) - Present: ClassDeclaration action (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `actionDesc` (struct) - Present: ClassDeclaration actionDesc (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `Checker` (struct) - Present: ClassDeclaration Checker (gojr/src/go/types/check.ts); InterfaceDeclaration Checker (gojr/src/go/types/cycles.ts); InterfaceDeclaration Checker (gojr/src/go/types/decl.ts); InterfaceDeclaration Checker (gojr/src/go/types/errors.ts); InterfaceDeclaration Checker (gojr/src/go/types/errsupport.ts); InterfaceDeclaration Checker (gojr/src/go/types/format.ts); InterfaceDeclaration Checker (gojr/src/go/types/initorder.ts); InterfaceDeclaration Checker (gojr/src/go/types/instantiate.ts); InterfaceDeclaration Checker (gojr/src/go/types/labels.ts); InterfaceDeclaration Checker (gojr/src/go/types/lookup.ts); InterfaceDeclaration Checker (gojr/src/go/types/mono.ts); InterfaceDeclaration Checker (gojr/src/go/types/named.ts); InterfaceDeclaration Checker (gojr/src/go/types/recording.ts); InterfaceDeclaration Checker (gojr/src/go/types/resolver.ts); InterfaceDeclaration Checker (gojr/src/go/types/subst.ts) - structure: not yet 1:1-audited
- `cleaner` (interface) - Present: InterfaceDeclaration cleaner (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `bailout` (struct) - Present: ClassDeclaration bailout (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited

#### Functions
- `NewChecker` @ /usr/local/go1.27rc1/src/go/types/check.go:236:1 - Present: FunctionDeclaration NewChecker (gojr/src/go/types/check.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=6, binary=2, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `versionMax` @ /usr/local/go1.27rc1/src/go/types/check.go:359:1 - Present: FunctionDeclaration versionMax (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `instantiatedIdent` @ /usr/local/go1.27rc1/src/go/types/check.go:546:1 - Present: FunctionDeclaration instantiatedIdent (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=2, select=0, branch=0, assign=5, return=2, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*environment.lookupScope` @ /usr/local/go1.27rc1/src/go/types/check.go:62:1 - Present: Method environment.lookupScope (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=4, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*environment.lookup` @ /usr/local/go1.27rc1/src/go/types/check.go:72:1 - Present: Method environment.lookup (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*action.describef` @ /usr/local/go1.27rc1/src/go/types/check.go:102:1 - Present: Method action.describef (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.addDeclDep` @ /usr/local/go1.27rc1/src/go/types/check.go:174:1 - Present: Method Checker.addDeclDep (gojr/src/go/types/check.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.rememberUntyped` @ /usr/local/go1.27rc1/src/go/types/check.go:185:1 - Present: Method Checker.rememberUntyped (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.later` @ /usr/local/go1.27rc1/src/go/types/check.go:200:1 - Present: Method Checker.later (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.push` @ /usr/local/go1.27rc1/src/go/types/check.go:207:1 - Present: Method Checker.push (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.pop` @ /usr/local/go1.27rc1/src/go/types/check.go:216:1 - Present: Method Checker.pop (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.needsCleanup` @ /usr/local/go1.27rc1/src/go/types/check.go:230:1 - Present: Method Checker.needsCleanup (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.initFiles` @ /usr/local/go1.27rc1/src/go/types/check.go:268:1 - Present: Method Checker.initFiles (gojr/src/go/types/check.ts) - control-flow shape: if=5, for=0, range=2, switch=1, typeSwitch=0, select=0, branch=1, assign=24, return=0, defer=0, go=0, call=18, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.pushPos` @ /usr/local/go1.27rc1/src/go/types/check.go:367:1 - Present: Method Checker.pushPos (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.popPos` @ /usr/local/go1.27rc1/src/go/types/check.go:372:1 - Present: Method Checker.popPos (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.handleBailout` @ /usr/local/go1.27rc1/src/go/types/check.go:379:1 - Present: Method Checker.handleBailout (gojr/src/go/types/check.ts) - control-flow shape: if=2, for=1, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=14, binary=5, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.Files` @ /usr/local/go1.27rc1/src/go/types/check.go:409:1 - Present: Method Checker.Files (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=1, go=0, call=2, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.checkFiles` @ /usr/local/go1.27rc1/src/go/types/check.go:430:1 - Present: Method Checker.checkFiles (gojr/src/go/types/check.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=12, return=0, defer=0, go=0, call=23, binary=1, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.processDelayed` @ /usr/local/go1.27rc1/src/go/types/check.go:494:1 - Present: Method Checker.processDelayed (gojr/src/go/types/check.ts) - control-flow shape: if=3, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=0, defer=0, go=0, call=8, binary=4, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.cleanup` @ /usr/local/go1.27rc1/src/go/types/check.go:524:1 - Present: Method Checker.cleanup (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordTypeAndValueInSyntax` @ /usr/local/go1.27rc1/src/go/types/check.go:534:1 - Present: Method Checker.recordTypeAndValueInSyntax (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordCommaOkTypesInSyntax` @ /usr/local/go1.27rc1/src/go/types/check.go:540:1 - Present: Method Checker.recordCommaOkTypesInSyntax (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### const.go

#### Functions
- `representableConst` @ /usr/local/go1.27rc1/src/go/types/const.go:78:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=16, for=0, range=0, switch=5, typeSwitch=0, select=0, branch=0, assign=20, return=32, defer=0, go=0, call=56, binary=86, unary=4, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `fitsFloat32` @ /usr/local/go1.27rc1/src/go/types/const.go:221:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `roundFloat32` @ /usr/local/go1.27rc1/src/go/types/const.go:227:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=4, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `fitsFloat64` @ /usr/local/go1.27rc1/src/go/types/const.go:236:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `roundFloat64` @ /usr/local/go1.27rc1/src/go/types/const.go:241:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.overflow` @ /usr/local/go1.27rc1/src/go/types/const.go:22:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=4, defer=0, go=0, call=22, binary=8, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.representable` @ /usr/local/go1.27rc1/src/go/types/const.go:251:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.representation` @ /usr/local/go1.27rc1/src/go/types/const.go:266:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=4, defer=0, go=0, call=9, binary=3, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.invalidConversion` @ /usr/local/go1.27rc1/src/go/types/const.go:289:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.convertUntyped` @ /usr/local/go1.27rc1/src/go/types/const.go:301:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=8, binary=3, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### context.go

#### Types
- `Context` (struct) - Present: ClassDeclaration Context (gojr/src/go/types/context.ts) - structure: not yet 1:1-audited
- `ctxtEntry` (struct) - Present: ClassDeclaration ctxtEntry (gojr/src/go/types/context.ts) - structure: not yet 1:1-audited

#### Functions
- `NewContext` @ /usr/local/go1.27rc1/src/go/types/context.go:58:1 - Present: FunctionDeclaration NewContext (gojr/src/go/types/context.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Context.instanceHash` @ /usr/local/go1.27rc1/src/go/types/context.go:68:1 - Present: Method Context.instanceHash (gojr/src/go/types/context.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=11, binary=3, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Context.lookup` @ /usr/local/go1.27rc1/src/go/types/context.go:90:1 - Present: Method Context.lookup (gojr/src/go/types/context.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=1, go=0, call=5, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Context.update` @ /usr/local/go1.27rc1/src/go/types/context.go:111:1 - Present: Method Context.update (gojr/src/go/types/context.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=1, go=0, call=7, binary=3, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Context.getID` @ /usr/local/go1.27rc1/src/go/types/context.go:137:1 - Present: Method Context.getID (gojr/src/go/types/context.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=1, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### conversions.go

#### Functions
- `isUintptr` @ /usr/local/go1.27rc1/src/go/types/conversions.go:297:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isUnsafePointer` @ /usr/local/go1.27rc1/src/go/types/conversions.go:302:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isPointer` @ /usr/local/go1.27rc1/src/go/types/conversions.go:307:1 - Present: FunctionDeclaration isPointer (gojr/src/go/types/lookup.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isBytesOrRunes` @ /usr/local/go1.27rc1/src/go/types/conversions.go:312:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=2, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.conversion` @ /usr/local/go1.27rc1/src/go/types/conversions.go:20:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=13, for=0, range=0, switch=2, typeSwitch=0, select=0, branch=0, assign=19, return=9, defer=0, go=0, call=47, binary=23, unary=7, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*operand.convertibleTo` @ /usr/local/go1.27rc1/src/go/types/conversions.go:139:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=29, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=0, assign=23, return=29, defer=0, go=0, call=49, binary=40, unary=5, composite=0, funcLiteral=5 - logic: not yet 1:1-audited

### cycles.go

#### Methods
- `*Checker.directCycles` @ /usr/local/go1.27rc1/src/go/types/cycles.go:14:1 - Present: Method Checker.directCycles interface method (gojr/src/go/types/cycles.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.directCycle` @ /usr/local/go1.27rc1/src/go/types/cycles.go:42:1 - Present: Method Checker.directCycle interface method (gojr/src/go/types/cycles.ts) - control-flow shape: if=6, for=1, range=3, switch=0, typeSwitch=0, select=0, branch=4, assign=8, return=0, defer=0, go=0, call=10, binary=3, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isComplete` @ /usr/local/go1.27rc1/src/go/types/cycles.go:111:1 - Present: Method Checker.isComplete interface method (gojr/src/go/types/cycles.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=7, return=3, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### decl.go

#### Types
- `decl` (interface) - Present: InterfaceDeclaration decl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited
- `importDecl` (struct) - Present: ClassDeclaration importDecl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited
- `constDecl` (struct) - Present: ClassDeclaration constDecl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited
- `varDecl` (struct) - Present: ClassDeclaration varDecl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited
- `typeDecl` (struct) - Present: ClassDeclaration typeDecl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited
- `funcDecl` (struct) - Present: ClassDeclaration funcDecl (gojr/src/go/types/decl.ts) - structure: not yet 1:1-audited

#### Functions
- `pathString` @ /usr/local/go1.27rc1/src/go/types/decl.go:37:1 - Present: FunctionDeclaration pathString (gojr/src/go/types/decl.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `firstInSrc` @ /usr/local/go1.27rc1/src/go/types/decl.go:303:1 - Present: FunctionDeclaration firstInSrc (gojr/src/go/types/decl.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.declare` @ /usr/local/go1.27rc1/src/go/types/decl.go:16:1 - Present: Method Checker.declare interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=9, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.objDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:49:1 - Present: Method Checker.objDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=9, for=0, range=0, switch=0, typeSwitch=2, select=0, branch=0, assign=9, return=2, defer=4, go=0, call=35, binary=8, unary=3, composite=1, funcLiteral=3 - logic: not yet 1:1-audited
- `*Checker.validCycle` @ /usr/local/go1.27rc1/src/go/types/decl.go:171:1 - Present: Method Checker.validCycle interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=10, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=1, assign=10, return=4, defer=1, go=0, call=25, binary=12, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.cycleError` @ /usr/local/go1.27rc1/src/go/types/decl.go:254:1 - Present: Method Checker.cycleError interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=6, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=3, defer=0, go=0, call=22, binary=11, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `importDecl.node` @ /usr/local/go1.27rc1/src/go/types/decl.go:331:1 - Present: Method importDecl.node (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `constDecl.node` @ /usr/local/go1.27rc1/src/go/types/decl.go:332:1 - Present: Method constDecl.node (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `varDecl.node` @ /usr/local/go1.27rc1/src/go/types/decl.go:333:1 - Present: Method varDecl.node (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `typeDecl.node` @ /usr/local/go1.27rc1/src/go/types/decl.go:334:1 - Present: Method typeDecl.node (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `funcDecl.node` @ /usr/local/go1.27rc1/src/go/types/decl.go:335:1 - Present: Method funcDecl.node (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.walkDecls` @ /usr/local/go1.27rc1/src/go/types/decl.go:337:1 - Present: Method Checker.walkDecls interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.walkDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:343:1 - Present: Method Checker.walkDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=1, switch=2, typeSwitch=2, select=0, branch=0, assign=7, return=0, defer=0, go=0, call=12, binary=4, unary=0, composite=5, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.constDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:387:1 - Present: Method Checker.constDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=1, defer=1, go=0, call=11, binary=3, unary=3, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.varDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:433:1 - Present: Method Checker.varDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=7, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=10, binary=11, unary=3, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isImportedConstraint` @ /usr/local/go1.27rc1/src/go/types/decl.go:487:1 - Present: Method Checker.isImportedConstraint interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=3, binary=7, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.typeDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:496:1 - Present: Method Checker.typeDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=11, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=17, return=1, defer=3, go=0, call=33, binary=21, unary=9, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*Checker.collectTypeParams` @ /usr/local/go1.27rc1/src/go/types/decl.go:584:1 - Present: Method Checker.collectTypeParams interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=2, for=0, range=4, switch=0, typeSwitch=0, select=0, branch=0, assign=11, return=0, defer=1, go=0, call=10, binary=2, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.bound` @ /usr/local/go1.27rc1/src/go/types/decl.go:639:1 - Present: Method Checker.bound interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=8, return=2, defer=0, go=0, call=2, binary=3, unary=2, composite=4, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.declareTypeParam` @ /usr/local/go1.27rc1/src/go/types/decl.go:662:1 - Present: Method Checker.declareTypeParam interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=4, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.collectMethods` @ /usr/local/go1.27rc1/src/go/types/decl.go:675:1 - Present: Method Checker.collectMethods interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=5, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=5, return=1, defer=0, go=0, call=25, binary=9, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.checkFieldUniqueness` @ /usr/local/go1.27rc1/src/go/types/decl.go:732:1 - Present: Method Checker.checkFieldUniqueness interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=3, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=0, defer=0, go=0, call=11, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.funcDecl` @ /usr/local/go1.27rc1/src/go/types/decl.go:762:1 - Present: Method Checker.funcDecl interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=11, binary=7, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.declStmt` @ /usr/local/go1.27rc1/src/go/types/decl.go:795:1 - Present: Method Checker.declStmt interface method (gojr/src/go/types/decl.ts) - control-flow shape: if=4, for=0, range=6, switch=1, typeSwitch=1, select=0, branch=1, assign=18, return=0, defer=0, go=0, call=37, binary=4, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

### errors.go

#### Constants
- `invalidArg` - Present: Variable invalidArg (gojr/src/go/types/errors.ts) - logic: declaration-only
- `invalidOp` - Present: Variable invalidOp (gojr/src/go/types/errors.ts) - logic: declaration-only

#### Types
- `errorDesc` (struct) - Present: ClassDeclaration errorDesc (gojr/src/go/types/errors.ts) - structure: not yet 1:1-audited
- `error_` (struct) - Present: ClassDeclaration error_ (gojr/src/go/types/errors.ts) - structure: not yet 1:1-audited
- `positioner` (interface) - Present: InterfaceDeclaration positioner (gojr/src/go/types/check.ts) - structure: not yet 1:1-audited
- `atPos` (*ast.SelectorExpr) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `posSpan` (struct) - Present: ClassDeclaration posSpan (gojr/src/go/types/errors.ts) - structure: not yet 1:1-audited

#### Functions
- `assert` @ /usr/local/go1.27rc1/src/go/types/errors.go:18:1 - Present: FunctionDeclaration assert (gojr/src/go/types/errors.ts); FunctionDeclaration assert (gojr/src/go/types/util.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `inNode` @ /usr/local/go1.27rc1/src/go/types/errors.go:283:1 - Present: FunctionDeclaration inNode (gojr/src/go/types/errors.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=3, binary=3, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `spanOf` @ /usr/local/go1.27rc1/src/go/types/errors.go:294:1 - Present: FunctionDeclaration spanOf (gojr/src/go/types/errors.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=4, return=5, defer=0, go=0, call=6, binary=1, unary=0, composite=4, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.newError` @ /usr/local/go1.27rc1/src/go/types/errors.go:47:1 - Present: Method Checker.newError interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.addf` @ /usr/local/go1.27rc1/src/go/types/errors.go:60:1 - Present: Method error_.addf (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.addAltDecl` @ /usr/local/go1.27rc1/src/go/types/errors.go:65:1 - Present: Method error_.addAltDecl (gojr/src/go/types/errors.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=4, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.empty` @ /usr/local/go1.27rc1/src/go/types/errors.go:74:1 - Present: Method error_.empty (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.posn` @ /usr/local/go1.27rc1/src/go/types/errors.go:78:1 - Present: Method error_.posn (gojr/src/go/types/errors.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.msg` @ /usr/local/go1.27rc1/src/go/types/errors.go:86:1 - Present: Method error_.msg (gojr/src/go/types/errors.ts) - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=9, binary=1, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*error_.report` @ /usr/local/go1.27rc1/src/go/types/errors.go:106:1 - Present: Method error_.report (gojr/src/go/types/errors.ts) - control-flow shape: if=7, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=7, return=1, defer=0, go=0, call=14, binary=5, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.handleError` @ /usr/local/go1.27rc1/src/go/types/errors.go:156:1 - Present: Method Checker.handleError interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=8, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=15, return=0, defer=0, go=0, call=11, binary=15, unary=0, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.error` @ /usr/local/go1.27rc1/src/go/types/errors.go:234:1 - Present: Method Checker.error interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.errorf` @ /usr/local/go1.27rc1/src/go/types/errors.go:240:1 - Present: Method Checker.errorf interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.softErrorf` @ /usr/local/go1.27rc1/src/go/types/errors.go:246:1 - Present: Method Checker.softErrorf interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.versionErrorf` @ /usr/local/go1.27rc1/src/go/types/errors.go:253:1 - Present: Method Checker.versionErrorf interface method (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=4, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `atPos.Pos` @ /usr/local/go1.27rc1/src/go/types/errors.go:263:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `posSpan.Pos` @ /usr/local/go1.27rc1/src/go/types/errors.go:276:1 - Present: Method posSpan.Pos (gojr/src/go/types/errors.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### errsupport.go

#### Functions
- `tail` @ /usr/local/go1.27rc1/src/go/types/errsupport.go:109:1 - Present: FunctionDeclaration tail (gojr/src/go/types/errsupport.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.lookupError` @ /usr/local/go1.27rc1/src/go/types/errsupport.go:15:1 - Present: Method Checker.lookupError interface method (gojr/src/go/types/errsupport.ts) - control-flow shape: if=9, for=0, range=0, switch=2, typeSwitch=1, select=0, branch=0, assign=10, return=8, defer=0, go=0, call=19, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### eval.go

#### Functions
- `Eval` @ /usr/local/go1.27rc1/src/go/types/eval.go:24:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=3, binary=1, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `CheckExpr` @ /usr/local/go1.27rc1/src/go/types/eval.go:56:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=9, return=2, defer=1, go=0, call=9, binary=8, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### expr.go

#### Constants
- `conversion` - Present: EnumMember conversion in exprKind (gojr/src/go/types/universe.ts) - logic: declaration-only
- `expression` - Present: EnumMember expression in exprKind (gojr/src/go/types/universe.ts) - logic: declaration-only
- `statement` - Present: EnumMember statement in exprKind (gojr/src/go/types/universe.ts) - logic: declaration-only

#### Variables
- `unaryOpPredicates` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `op2str1` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `op2str2` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `binaryOpPredicates` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Types
- `opPredicates` (*ast.MapType) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `exprKind` (*ast.Ident) - Present: EnumDeclaration exprKind (gojr/src/go/types/universe.ts) - structure: not yet 1:1-audited
- `target` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `init` @ /usr/local/go1.27rc1/src/go/types/expr.go:63:1 - Present: FunctionDeclaration init (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `opPos` @ /usr/local/go1.27rc1/src/go/types/expr.go:88:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `opName` @ /usr/local/go1.27rc1/src/go/types/expr.go:101:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isShift` @ /usr/local/go1.27rc1/src/go/types/expr.go:239:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isComparison` @ /usr/local/go1.27rc1/src/go/types/expr.go:243:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `init` @ /usr/local/go1.27rc1/src/go/types/expr.go:759:1 - Present: FunctionDeclaration init (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `newTarget` @ /usr/local/go1.27rc1/src/go/types/expr.go:957:1 - Present: FunctionDeclaration newTarget (gojr/src/go/types/decl.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=3, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `keyVal` @ /usr/local/go1.27rc1/src/go/types/expr.go:1212:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=2, assign=9, return=7, defer=0, go=0, call=15, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `nth` @ /usr/local/go1.27rc1/src/go/types/expr.go:1323:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.op` @ /usr/local/go1.27rc1/src/go/types/expr.go:73:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=4, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.unary` @ /usr/local/go1.27rc1/src/go/types/expr.go:129:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=8, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=13, return=9, defer=0, go=0, call=27, binary=8, unary=5, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.chanElem` @ /usr/local/go1.27rc1/src/go/types/expr.go:198:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=8, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=7, defer=0, go=0, call=15, binary=11, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.updateExprType` @ /usr/local/go1.27rc1/src/go/types/expr.go:261:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=2, assign=5, return=5, defer=0, go=0, call=19, binary=5, unary=5, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.updateExprVal` @ /usr/local/go1.27rc1/src/go/types/expr.go:375:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.implicitTypeAndValue` @ /usr/local/go1.27rc1/src/go/types/expr.go:388:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=16, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=1, assign=4, return=22, defer=0, go=0, call=25, binary=7, unary=9, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.comparison` @ /usr/local/go1.27rc1/src/go/types/expr.go:483:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=12, for=0, range=0, switch=3, typeSwitch=0, select=0, branch=6, assign=26, return=2, defer=0, go=0, call=57, binary=9, unary=10, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.incomparableCause` @ /usr/local/go1.27rc1/src/go/types/expr.go:608:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=4, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.shift` @ /usr/local/go1.27rc1/src/go/types/expr.go:618:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=17, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=13, return=11, defer=0, go=0, call=55, binary=26, unary=8, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.binary` @ /usr/local/go1.27rc1/src/go/types/expr.go:780:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=17, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=11, defer=0, go=0, call=55, binary=31, unary=11, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.matchTypes` @ /usr/local/go1.27rc1/src/go/types/expr.go:876:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=11, defer=0, go=0, call=38, binary=6, unary=2, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.rawExpr` @ /usr/local/go1.27rc1/src/go/types/expr.go:975:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=1, go=0, call=8, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.nonGeneric` @ /usr/local/go1.27rc1/src/go/types/expr.go:999:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=4, return=2, defer=0, go=0, call=8, binary=6, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.exprInternal` @ /usr/local/go1.27rc1/src/go/types/expr.go:1028:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=20, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=19, assign=20, return=8, defer=0, go=0, call=54, binary=9, unary=16, composite=1, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.typeAssertion` @ /usr/local/go1.27rc1/src/go/types/expr.go:1247:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=5, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.expr` @ /usr/local/go1.27rc1/src/go/types/expr.go:1266:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.genericExpr` @ /usr/local/go1.27rc1/src/go/types/expr.go:1273:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.multiExpr` @ /usr/local/go1.27rc1/src/go/types/expr.go:1284:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=13, return=2, defer=0, go=0, call=16, binary=14, unary=5, composite=3, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.exprOrType` @ /usr/local/go1.27rc1/src/go/types/expr.go:1343:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.exclude` @ /usr/local/go1.27rc1/src/go/types/expr.go:1351:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=7, return=0, defer=0, go=0, call=5, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.singleValue` @ /usr/local/go1.27rc1/src/go/types/expr.go:1378:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=6, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### exprstring.go

#### Functions
- `ExprString` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:18:1 - Present: FunctionDeclaration ExprString (gojr/src/go/types/exprstring.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `WriteExpr` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:27:1 - Present: FunctionDeclaration WriteExpr (gojr/src/go/types/exprstring.ts) - control-flow shape: if=8, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=71, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `writeSigExpr` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:170:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=10, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `writeFieldList` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:195:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=1, return=0, defer=0, go=0, call=6, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `writeIdentList` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:221:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `writeExprList` @ /usr/local/go1.27rc1/src/go/types/exprstring.go:230:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### format.go

#### Functions
- `sprintf` @ /usr/local/go1.27rc1/src/go/types/format.go:18:1 - Present: FunctionDeclaration sprintf (gojr/src/go/types/format.ts) - control-flow shape: if=4, for=0, range=4, switch=0, typeSwitch=1, select=0, branch=0, assign=18, return=1, defer=0, go=0, call=32, binary=4, unary=4, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `ndigits` @ /usr/local/go1.27rc1/src/go/types/format.go:119:1 - Present: FunctionDeclaration ndigits (gojr/src/go/types/format.ts) - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=0, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `stripAnnotations` @ /usr/local/go1.27rc1/src/go/types/format.go:173:1 - Present: FunctionDeclaration stripAnnotations (gojr/src/go/types/format.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=4, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.sprintf` @ /usr/local/go1.27rc1/src/go/types/format.go:91:1 - Present: Method Checker.sprintf interface method (gojr/src/go/types/format.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.trace` @ /usr/local/go1.27rc1/src/go/types/format.go:101:1 - Present: Method Checker.trace interface method (gojr/src/go/types/format.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=7, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.dump` @ /usr/local/go1.27rc1/src/go/types/format.go:131:1 - Present: Method Checker.dump interface method (gojr/src/go/types/format.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.qualifier` @ /usr/local/go1.27rc1/src/go/types/format.go:135:1 - Present: Method Checker.qualifier interface method (gojr/src/go/types/format.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=3, defer=0, go=0, call=5, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.markImports` @ /usr/local/go1.27rc1/src/go/types/format.go:154:1 - Present: Method Checker.markImports interface method (gojr/src/go/types/format.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### gccgosizes.go

#### Variables
- `gccgoArchSizes` - Present: Variable gccgoArchSizes (gojr/src/go/types/sizes.ts) - logic: declaration-only

### gcsizes.go

#### Types
- `gcSizes` (struct) - Present: ClassDeclaration gcSizes (gojr/src/go/types/sizes.ts) - structure: not yet 1:1-audited

#### Functions
- `gcSizesFor` @ /usr/local/go1.27rc1/src/go/types/gcsizes.go:168:1 - Present: FunctionDeclaration gcSizesFor (gojr/src/go/types/sizes.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*gcSizes.Alignof` @ /usr/local/go1.27rc1/src/go/types/gcsizes.go:15:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=6, return=8, defer=1, go=0, call=14, binary=8, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*gcSizes.Offsetsof` @ /usr/local/go1.27rc1/src/go/types/gcsizes.go:79:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=8, return=1, defer=0, go=0, call=5, binary=4, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*gcSizes.Sizeof` @ /usr/local/go1.27rc1/src/go/types/gcsizes.go:101:1 - Present: Method gcSizes.Sizeof (gojr/src/go/types/sizes.ts) - control-flow shape: if=10, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=10, return=13, defer=0, go=0, call=15, binary=24, unary=4, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### generate.go

### gotype.go

#### Constants
- `usageString` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Variables
- `testFiles` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `xtestFiles` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `allErrors` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `verbose` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `compiler` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `printAST` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `printTrace` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `parseComments` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `panicOnError` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `fset` - Present: Property fset on Checker (gojr/src/go/types/check.ts); Variable fset (gojr/src/go/types/format.ts); Variable fset (gojr/src/go/types/methodset.ts); Variable fset (gojr/src/go/types/struct.ts) - logic: declaration-only
- `errorCount` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `sequential` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `parserMode` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Functions
- `initParserMode` @ /usr/local/go1.27rc1/src/go/types/gotype.go:122:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=0, defer=0, go=0, call=0, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `usage` @ /usr/local/go1.27rc1/src/go/types/gotype.go:166:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `report` @ /usr/local/go1.27rc1/src/go/types/gotype.go:172:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `parse` @ /usr/local/go1.27rc1/src/go/types/gotype.go:185:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `parseStdin` @ /usr/local/go1.27rc1/src/go/types/gotype.go:196:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `parseFiles` @ /usr/local/go1.27rc1/src/go/types/gotype.go:204:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=3, switch=0, typeSwitch=0, select=0, branch=1, assign=7, return=1, defer=1, go=1, call=11, binary=2, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `parseDir` @ /usr/local/go1.27rc1/src/go/types/gotype.go:247:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=3, defer=0, go=0, call=5, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `getPkgFiles` @ /usr/local/go1.27rc1/src/go/types/gotype.go:265:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=5, defer=0, go=0, call=7, binary=4, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `checkPkgFiles` @ /usr/local/go1.27rc1/src/go/types/gotype.go:291:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=2, return=0, defer=1, go=0, call=8, binary=2, unary=1, composite=2, funcLiteral=2 - logic: not yet 1:1-audited
- `printStats` @ /usr/local/go1.27rc1/src/go/types/gotype.go:321:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=6, binary=1, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `main` @ /usr/local/go1.27rc1/src/go/types/gotype.go:336:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=10, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### hash.go

#### Variables
- `_` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `_` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Types
- `Hasher` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `HasherIgnoreTags` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `hasher` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `btoi` @ /usr/local/go1.27rc1/src/go/types/hash.go:303:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `Hasher.Hash` @ /usr/local/go1.27rc1/src/go/types/hash.go:30:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `HasherIgnoreTags.Hash` @ /usr/local/go1.27rc1/src/go/types/hash.go:38:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `Hasher.Equal` @ /usr/local/go1.27rc1/src/go/types/hash.go:43:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `HasherIgnoreTags.Equal` @ /usr/local/go1.27rc1/src/go/types/hash.go:44:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasher.hash` @ /usr/local/go1.27rc1/src/go/types/hash.go:51:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=4, switch=0, typeSwitch=1, select=0, branch=0, assign=7, return=0, defer=0, go=0, call=76, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasher.hashTuple` @ /usr/local/go1.27rc1/src/go/types/hash.go:171:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=7, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasher.hashTypeParam` @ /usr/local/go1.27rc1/src/go/types/hash.go:189:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=6, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasher.hashTypeName` @ /usr/local/go1.27rc1/src/go/types/hash.go:210:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasher.shallowHash` @ /usr/local/go1.27rc1/src/go/types/hash.go:231:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=39, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### index.go

#### Types
- `indexedExpr` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `unpackIndexedExpr` @ /usr/local/go1.27rc1/src/go/types/index.go:504:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=0, binary=0, unary=2, composite=3, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.indexExpr` @ /usr/local/go1.27rc1/src/go/types/index.go:19:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=26, for=0, range=0, switch=1, typeSwitch=3, select=0, branch=1, assign=53, return=19, defer=0, go=0, call=55, binary=18, unary=14, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.sliceExpr` @ /usr/local/go1.27rc1/src/go/types/index.go:233:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=21, for=0, range=3, switch=1, typeSwitch=1, select=0, branch=1, assign=34, return=11, defer=0, go=0, call=43, binary=26, unary=10, composite=4, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.singleIndex` @ /usr/local/go1.27rc1/src/go/types/index.go:404:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=4, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.index` @ /usr/local/go1.27rc1/src/go/types/index.go:420:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=5, defer=0, go=0, call=10, binary=6, unary=6, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isValidIndex` @ /usr/local/go1.27rc1/src/go/types/index.go:449:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=6, defer=0, go=0, call=11, binary=6, unary=6, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*indexedExpr.Pos` @ /usr/local/go1.27rc1/src/go/types/index.go:500:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### infer.go

#### Constants
- `enableReverseTypeInference` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Types
- `tpWalker` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `cycleFinder` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `typeParamsString` @ /usr/local/go1.27rc1/src/go/types/infer.go:520:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=1, typeSwitch=0, select=0, branch=0, assign=1, return=4, defer=0, go=0, call=6, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isParameterized` @ /usr/local/go1.27rc1/src/go/types/infer.go:548:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `coreTerm` @ /usr/local/go1.27rc1/src/go/types/infer.go:655:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=5, defer=0, go=0, call=6, binary=5, unary=1, composite=1, funcLiteral=1 - logic: not yet 1:1-audited
- `killCycles` @ /usr/local/go1.27rc1/src/go/types/infer.go:695:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.infer` @ /usr/local/go1.27rc1/src/go/types/infer.go:35:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=42, for=2, range=10, switch=1, typeSwitch=0, select=0, branch=3, assign=43, return=10, defer=2, go=0, call=104, binary=45, unary=12, composite=0, funcLiteral=4 - logic: not yet 1:1-audited
- `*Checker.renameTParams` @ /usr/local/go1.27rc1/src/go/types/infer.go:471:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=2, defer=0, go=0, call=16, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*tpWalker.isParameterized` @ /usr/local/go1.27rc1/src/go/types/infer.go:561:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=2, switch=0, typeSwitch=1, select=0, branch=0, assign=5, return=16, defer=1, go=0, call=23, binary=11, unary=0, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*tpWalker.varList` @ /usr/local/go1.27rc1/src/go/types/infer.go:642:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*cycleFinder.typ` @ /usr/local/go1.27rc1/src/go/types/infer.go:708:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=4, switch=0, typeSwitch=1, select=0, branch=0, assign=7, return=1, defer=1, go=0, call=22, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*cycleFinder.varList` @ /usr/local/go1.27rc1/src/go/types/infer.go:793:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### initorder.go

#### Types
- `dependency` (interface) - Present: TypeAliasDeclaration dependency (gojr/src/go/types/initorder.ts) - structure: not yet 1:1-audited
- `graphNode` (struct) - Present: ClassDeclaration graphNode (gojr/src/go/types/initorder.ts) - structure: not yet 1:1-audited
- `nodeSet` (*ast.MapType) - Present: ClassDeclaration nodeSet (gojr/src/go/types/initorder.ts) - structure: not yet 1:1-audited
- `nodeQueue` (*ast.ArrayType) - Present: ClassDeclaration nodeQueue (gojr/src/go/types/initorder.ts) - structure: not yet 1:1-audited

#### Functions
- `findPath` @ /usr/local/go1.27rc1/src/go/types/initorder.go:140:1 - Present: FunctionDeclaration findPath (gojr/src/go/types/initorder.ts) - control-flow shape: if=3, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=5, defer=0, go=0, call=6, binary=3, unary=0, composite=1, funcLiteral=1 - logic: not yet 1:1-audited
- `dependencyGraph` @ /usr/local/go1.27rc1/src/go/types/initorder.go:229:1 - Present: FunctionDeclaration dependencyGraph (gojr/src/go/types/initorder.ts) - control-flow shape: if=5, for=0, range=9, switch=0, typeSwitch=0, select=0, branch=0, assign=10, return=2, defer=0, go=0, call=14, binary=4, unary=1, composite=1, funcLiteral=1 - logic: not yet 1:1-audited

#### Methods
- `*Checker.initOrder` @ /usr/local/go1.27rc1/src/go/types/initorder.go:20:1 - Present: Method Checker.initOrder interface method (gojr/src/go/types/initorder.ts) - control-flow shape: if=10, for=1, range=6, switch=0, typeSwitch=0, select=0, branch=2, assign=13, return=0, defer=0, go=0, call=36, binary=8, unary=5, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.reportCycle` @ /usr/local/go1.27rc1/src/go/types/initorder.go:168:1 - Present: Method Checker.reportCycle interface method (gojr/src/go/types/initorder.ts) - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=11, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*graphNode.cost` @ /usr/local/go1.27rc1/src/go/types/initorder.go:213:1 - Present: Method graphNode.cost (gojr/src/go/types/initorder.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*nodeSet.add` @ /usr/local/go1.27rc1/src/go/types/initorder.go:219:1 - Present: Method nodeSet.add (gojr/src/go/types/initorder.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `nodeQueue.Len` @ /usr/local/go1.27rc1/src/go/types/initorder.go:315:1 - Present: Method nodeQueue.Len (gojr/src/go/types/initorder.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `nodeQueue.Swap` @ /usr/local/go1.27rc1/src/go/types/initorder.go:317:1 - Present: Method nodeQueue.Swap (gojr/src/go/types/initorder.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `nodeQueue.Less` @ /usr/local/go1.27rc1/src/go/types/initorder.go:323:1 - Present: Method nodeQueue.Less (gojr/src/go/types/initorder.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=2, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*nodeQueue.Push` @ /usr/local/go1.27rc1/src/go/types/initorder.go:338:1 - Present: Method nodeQueue.Push (gojr/src/go/types/initorder.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*nodeQueue.Pop` @ /usr/local/go1.27rc1/src/go/types/initorder.go:342:1 - Present: Method nodeQueue.Pop (gojr/src/go/types/initorder.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=1, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### instantiate.go

#### Types
- `genericType` (interface) - Present: InterfaceDeclaration genericType (gojr/src/go/types/instantiate.ts) - structure: not yet 1:1-audited

#### Functions
- `Instantiate` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:54:1 - Present: FunctionDeclaration Instantiate (gojr/src/go/types/instantiate.ts) - control-flow shape: if=7, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=4, defer=0, go=0, call=19, binary=5, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `mentions` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:388:1 - Present: FunctionDeclaration mentions (gojr/src/go/types/instantiate.ts) - control-flow shape: if=3, for=0, range=2, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=4, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.instance` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:97:1 - Present: Method Checker.instance interface method (gojr/src/go/types/instantiate.ts) - control-flow shape: if=8, for=1, range=2, switch=0, typeSwitch=1, select=0, branch=0, assign=18, return=7, defer=0, go=0, call=36, binary=12, unary=3, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.validateTArgLen` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:198:1 - Present: Method Checker.validateTArgLen interface method (gojr/src/go/types/instantiate.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=5, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.verify` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:219:1 - Present: Method Checker.verify interface method (gojr/src/go/types/instantiate.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=5, binary=0, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.implements` @ /usr/local/go1.27rc1/src/go/types/instantiate.go:243:1 - Present: Method Checker.implements interface method (gojr/src/go/types/instantiate.ts) - control-flow shape: if=27, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=23, return=19, defer=0, go=0, call=45, binary=23, unary=9, composite=0, funcLiteral=2 - logic: not yet 1:1-audited

### interface.go

#### Variables
- `emptyInterface` - Present: Variable emptyInterface (gojr/src/go/types/interface.ts) - logic: declaration-only

#### Types
- `Interface` (struct) - Present: ClassDeclaration Interface (gojr/src/go/types/interface.ts) - structure: not yet 1:1-audited

#### Functions
- `NewInterface` @ /usr/local/go1.27rc1/src/go/types/interface.go:39:1 - Present: FunctionDeclaration NewInterface (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewInterfaceType` @ /usr/local/go1.27rc1/src/go/types/interface.go:53:1 - Present: FunctionDeclaration NewInterfaceType (gojr/src/go/types/interface.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=2, defer=0, go=0, call=6, binary=4, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Interface.typeSet` @ /usr/local/go1.27rc1/src/go/types/interface.go:29:1 - Present: Method Interface.typeSet (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newInterface` @ /usr/local/go1.27rc1/src/go/types/interface.go:77:1 - Present: Method Checker.newInterface (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.MarkImplicit` @ /usr/local/go1.27rc1/src/go/types/interface.go:89:1 - Present: Method Interface.MarkImplicit (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.NumExplicitMethods` @ /usr/local/go1.27rc1/src/go/types/interface.go:94:1 - Present: Method Interface.NumExplicitMethods (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.ExplicitMethod` @ /usr/local/go1.27rc1/src/go/types/interface.go:98:1 - Present: Method Interface.ExplicitMethod (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.NumEmbeddeds` @ /usr/local/go1.27rc1/src/go/types/interface.go:101:1 - Present: Method Interface.NumEmbeddeds (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.Embedded` @ /usr/local/go1.27rc1/src/go/types/interface.go:107:1 - Present: Method Interface.Embedded (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.EmbeddedType` @ /usr/local/go1.27rc1/src/go/types/interface.go:110:1 - Present: Method Interface.EmbeddedType (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.NumMethods` @ /usr/local/go1.27rc1/src/go/types/interface.go:113:1 - Present: Method Interface.NumMethods (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.Method` @ /usr/local/go1.27rc1/src/go/types/interface.go:117:1 - Present: Method Interface.Method (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.Empty` @ /usr/local/go1.27rc1/src/go/types/interface.go:120:1 - Present: Method Interface.Empty (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.IsComparable` @ /usr/local/go1.27rc1/src/go/types/interface.go:123:1 - Present: Method Interface.IsComparable (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.IsMethodSet` @ /usr/local/go1.27rc1/src/go/types/interface.go:127:1 - Present: Method Interface.IsMethodSet (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.IsImplicit` @ /usr/local/go1.27rc1/src/go/types/interface.go:130:1 - Present: Method Interface.IsImplicit (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.Complete` @ /usr/local/go1.27rc1/src/go/types/interface.go:139:1 - Present: Method Interface.Complete (gojr/src/go/types/interface.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.Underlying` @ /usr/local/go1.27rc1/src/go/types/interface.go:147:1 - Present: Method Interface.Underlying (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.String` @ /usr/local/go1.27rc1/src/go/types/interface.go:148:1 - Present: Method Interface.String (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Interface.cleanup` @ /usr/local/go1.27rc1/src/go/types/interface.go:153:1 - Present: Method Interface.cleanup (gojr/src/go/types/interface.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.interfaceType` @ /usr/local/go1.27rc1/src/go/types/interface.go:159:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=10, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=3, assign=16, return=1, defer=0, go=0, call=26, binary=13, unary=1, composite=0, funcLiteral=2 - logic: not yet 1:1-audited

### iter.go

#### Methods
- `*Interface.Methods` @ /usr/local/go1.27rc1/src/go/types/iter.go:20:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Interface.ExplicitMethods` @ /usr/local/go1.27rc1/src/go/types/iter.go:34:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Interface.EmbeddedTypes` @ /usr/local/go1.27rc1/src/go/types/iter.go:47:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Named.Methods` @ /usr/local/go1.27rc1/src/go/types/iter.go:60:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Scope.Children` @ /usr/local/go1.27rc1/src/go/types/iter.go:73:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Struct.Fields` @ /usr/local/go1.27rc1/src/go/types/iter.go:86:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Tuple.Variables` @ /usr/local/go1.27rc1/src/go/types/iter.go:99:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*MethodSet.Methods` @ /usr/local/go1.27rc1/src/go/types/iter.go:112:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Union.Terms` @ /usr/local/go1.27rc1/src/go/types/iter.go:125:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*TypeParamList.TypeParams` @ /usr/local/go1.27rc1/src/go/types/iter.go:138:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*TypeList.Types` @ /usr/local/go1.27rc1/src/go/types/iter.go:151:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

### labels.go

#### Types
- `block` (struct) - Present: ClassDeclaration block (gojr/src/go/types/labels.ts) - structure: not yet 1:1-audited

#### Methods
- `*Checker.labels` @ /usr/local/go1.27rc1/src/go/types/labels.go:15:1 - Present: Method Checker.labels interface method (gojr/src/go/types/labels.ts) - control-flow shape: if=2, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=11, return=0, defer=0, go=0, call=8, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*block.insert` @ /usr/local/go1.27rc1/src/go/types/labels.go:58:1 - Present: Method block.insert (gojr/src/go/types/labels.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=3, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*block.gotoTarget` @ /usr/local/go1.27rc1/src/go/types/labels.go:73:1 - Present: Method block.gotoTarget (gojr/src/go/types/labels.ts) - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=0, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*block.enclosingTarget` @ /usr/local/go1.27rc1/src/go/types/labels.go:84:1 - Present: Method block.enclosingTarget (gojr/src/go/types/labels.ts) - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=0, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.blockBranches` @ /usr/local/go1.27rc1/src/go/types/labels.go:96:1 - Present: Method Checker.blockBranches interface method (gojr/src/go/types/labels.ts) - control-flow shape: if=13, for=0, range=2, switch=1, typeSwitch=3, select=0, branch=0, assign=30, return=7, defer=0, go=0, call=43, binary=13, unary=3, composite=1, funcLiteral=4 - logic: not yet 1:1-audited

### literals.go

#### Methods
- `*Checker.langCompat` @ /usr/local/go1.27rc1/src/go/types/literals.go:21:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=5, defer=0, go=0, call=7, binary=14, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.basicLit` @ /usr/local/go1.27rc1/src/go/types/literals.go:48:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=11, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.funcLit` @ /usr/local/go1.27rc1/src/go/types/literals.go:83:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=7, return=0, defer=0, go=0, call=8, binary=2, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.compositeLit` @ /usr/local/go1.27rc1/src/go/types/literals.go:111:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=31, for=0, range=6, switch=1, typeSwitch=1, select=0, branch=16, assign=47, return=2, defer=0, go=0, call=65, binary=27, unary=8, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.indexedElts` @ /usr/local/go1.27rc1/src/go/types/literals.go:349:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=7, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=11, return=1, defer=0, go=0, call=9, binary=6, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### lookup.go

#### Types
- `embeddedType` (struct) - Present: ClassDeclaration embeddedType (gojr/src/go/types/lookup.ts) - structure: not yet 1:1-audited
- `instanceLookup` (struct) - Present: ClassDeclaration instanceLookup (gojr/src/go/types/lookup.ts) - structure: not yet 1:1-audited

#### Functions
- `LookupSelection` @ /usr/local/go1.27rc1/src/go/types/lookup.go:40:1 - Present: FunctionDeclaration LookupSelection (gojr/src/go/types/lookup.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=2, binary=0, unary=0, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `LookupFieldOrMethod` @ /usr/local/go1.27rc1/src/go/types/lookup.go:88:1 - Present: FunctionDeclaration LookupFieldOrMethod (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `lookupFieldOrMethod` @ /usr/local/go1.27rc1/src/go/types/lookup.go:97:1 - Present: FunctionDeclaration lookupFieldOrMethod (gojr/src/go/types/lookup.ts) - control-flow shape: if=6, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=3, defer=0, go=0, call=7, binary=6, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `lookupFieldOrMethodImpl` @ /usr/local/go1.27rc1/src/go/types/lookup.go:147:1 - Present: FunctionDeclaration lookupFieldOrMethodImpl (gojr/src/go/types/lookup.ts) - control-flow shape: if=15, for=1, range=2, switch=0, typeSwitch=1, select=0, branch=3, assign=22, return=8, defer=0, go=0, call=21, binary=21, unary=2, composite=3, funcLiteral=0 - logic: not yet 1:1-audited
- `consolidateMultiples` @ /usr/local/go1.27rc1/src/go/types/lookup.go:289:1 - Present: FunctionDeclaration consolidateMultiples (gojr/src/go/types/lookup.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=2, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `lookupType` @ /usr/local/go1.27rc1/src/go/types/lookup.go:309:1 - Present: FunctionDeclaration lookupType (gojr/src/go/types/lookup.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `MissingMethod` @ /usr/local/go1.27rc1/src/go/types/lookup.go:368:1 - Present: FunctionDeclaration MissingMethod (gojr/src/go/types/lookup.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasInvalidEmbeddedFields` @ /usr/local/go1.27rc1/src/go/types/lookup.go:550:1 - Present: FunctionDeclaration hasInvalidEmbeddedFields (gojr/src/go/types/lookup.ts) - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=5, binary=5, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isInterfacePtr` @ /usr/local/go1.27rc1/src/go/types/lookup.go:565:1 - Present: FunctionDeclaration isInterfacePtr (gojr/src/go/types/lookup.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `deref` @ /usr/local/go1.27rc1/src/go/types/lookup.go:629:1 - Present: FunctionDeclaration deref (gojr/src/go/types/lookup.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `derefStructPtr` @ /usr/local/go1.27rc1/src/go/types/lookup.go:645:1 - Present: FunctionDeclaration derefStructPtr (gojr/src/go/types/lookup.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `concat` @ /usr/local/go1.27rc1/src/go/types/lookup.go:656:1 - Present: FunctionDeclaration concat (gojr/src/go/types/lookup.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `methodIndex` @ /usr/local/go1.27rc1/src/go/types/lookup.go:664:1 - Present: FunctionDeclaration methodIndex (gojr/src/go/types/lookup.ts); FunctionDeclaration methodIndex (gojr/src/go/types/typeset.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `fieldPath` @ /usr/local/go1.27rc1/src/go/types/lookup.go:679:1 - Present: FunctionDeclaration fieldPath (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=4, return=1, defer=0, go=0, call=5, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*instanceLookup.lookup` @ /usr/local/go1.27rc1/src/go/types/lookup.go:331:1 - Present: Method instanceLookup.lookup (gojr/src/go/types/lookup.ts) - control-flow shape: if=2, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=3, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*instanceLookup.add` @ /usr/local/go1.27rc1/src/go/types/lookup.go:345:1 - Present: Method instanceLookup.add (gojr/src/go/types/lookup.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.missingMethod` @ /usr/local/go1.27rc1/src/go/types/lookup.go:381:1 - Present: Method Checker.missingMethod interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=18, for=0, range=2, switch=3, typeSwitch=0, select=0, branch=8, assign=33, return=3, defer=0, go=0, call=41, binary=22, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.hasAllMethods` @ /usr/local/go1.27rc1/src/go/types/lookup.go:540:1 - Present: Method Checker.hasAllMethods interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=3, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.interfacePtrError` @ /usr/local/go1.27rc1/src/go/types/lookup.go:571:1 - Present: Method Checker.interfacePtrError interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=6, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.funcString` @ /usr/local/go1.27rc1/src/go/types/lookup.go:581:1 - Present: Method Checker.funcString interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=4, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.assertableTo` @ /usr/local/go1.27rc1/src/go/types/lookup.go:600:1 - Present: Method Checker.assertableTo interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newAssertableTo` @ /usr/local/go1.27rc1/src/go/types/lookup.go:616:1 - Present: Method Checker.newAssertableTo interface method (gojr/src/go/types/lookup.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### map.go

#### Types
- `Map` (struct) - Present: ClassDeclaration Map (gojr/src/go/types/map.ts) - structure: not yet 1:1-audited

#### Functions
- `NewMap` @ /usr/local/go1.27rc1/src/go/types/map.go:16:1 - Present: FunctionDeclaration NewMap (gojr/src/go/types/map.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Map.Key` @ /usr/local/go1.27rc1/src/go/types/map.go:21:1 - Present: Method Map.Key (gojr/src/go/types/map.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Map.Elem` @ /usr/local/go1.27rc1/src/go/types/map.go:24:1 - Present: Method Map.Elem (gojr/src/go/types/map.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Map.Underlying` @ /usr/local/go1.27rc1/src/go/types/map.go:26:1 - Present: Method Map.Underlying (gojr/src/go/types/map.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Map.String` @ /usr/local/go1.27rc1/src/go/types/map.go:27:1 - Present: Method Map.String (gojr/src/go/types/map.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### methodset.go

#### Variables
- `emptyMethodSet` - Present: Variable emptyMethodSet (gojr/src/go/types/methodset.ts) - logic: declaration-only

#### Types
- `MethodSet` (struct) - Present: ClassDeclaration MethodSet (gojr/src/go/types/methodset.ts) - structure: not yet 1:1-audited
- `methodSet` (*ast.MapType) - Present: ClassDeclaration methodSet (gojr/src/go/types/methodset.ts) - structure: not yet 1:1-audited

#### Functions
- `NewMethodSet` @ /usr/local/go1.27rc1/src/go/types/methodset.go:72:1 - Present: FunctionDeclaration NewMethodSet (gojr/src/go/types/methodset.ts) - control-flow shape: if=13, for=2, range=5, switch=0, typeSwitch=1, select=0, branch=1, assign=24, return=5, defer=0, go=0, call=28, binary=14, unary=6, composite=4, funcLiteral=1 - logic: not yet 1:1-audited

#### Methods
- `*MethodSet.String` @ /usr/local/go1.27rc1/src/go/types/methodset.go:22:1 - Present: Method MethodSet.String (gojr/src/go/types/methodset.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=5, binary=1, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*MethodSet.Len` @ /usr/local/go1.27rc1/src/go/types/methodset.go:37:1 - Present: Method MethodSet.Len (gojr/src/go/types/methodset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*MethodSet.At` @ /usr/local/go1.27rc1/src/go/types/methodset.go:40:1 - Present: Method MethodSet.At (gojr/src/go/types/methodset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*MethodSet.Lookup` @ /usr/local/go1.27rc1/src/go/types/methodset.go:43:1 - Present: Method MethodSet.Lookup (gojr/src/go/types/methodset.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=4, defer=0, go=0, call=7, binary=4, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `methodSet.add` @ /usr/local/go1.27rc1/src/go/types/methodset.go:218:1 - Present: Method methodSet.add (gojr/src/go/types/methodset.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `methodSet.addOne` @ /usr/local/go1.27rc1/src/go/types/methodset.go:228:1 - Present: Method methodSet.addOne (gojr/src/go/types/methodset.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=2, defer=0, go=0, call=3, binary=3, unary=4, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

### mono.go

#### Types
- `monoGraph` (struct) - Present: ClassDeclaration monoGraph (gojr/src/go/types/mono.ts) - structure: not yet 1:1-audited
- `monoVertex` (struct) - Present: ClassDeclaration monoVertex (gojr/src/go/types/mono.ts) - structure: not yet 1:1-audited
- `monoEdge` (struct) - Present: ClassDeclaration monoEdge (gojr/src/go/types/mono.ts) - structure: not yet 1:1-audited

#### Methods
- `*Checker.monomorph` @ /usr/local/go1.27rc1/src/go/types/mono.go:86:1 - Present: Method Checker.monomorph interface method (gojr/src/go/types/mono.ts) - control-flow shape: if=2, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=9, return=1, defer=0, go=0, call=2, binary=4, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.reportInstanceLoop` @ /usr/local/go1.27rc1/src/go/types/mono.go:121:1 - Present: Method Checker.reportInstanceLoop interface method (gojr/src/go/types/mono.ts) - control-flow shape: if=0, for=2, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=10, return=0, defer=0, go=0, call=17, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*monoGraph.recordCanon` @ /usr/local/go1.27rc1/src/go/types/mono.go:167:1 - Present: Method monoGraph.recordCanon (gojr/src/go/types/mono.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*monoGraph.recordInstance` @ /usr/local/go1.27rc1/src/go/types/mono.go:176:1 - Present: Method monoGraph.recordInstance (gojr/src/go/types/mono.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*monoGraph.assign` @ /usr/local/go1.27rc1/src/go/types/mono.go:187:1 - Present: Method monoGraph.assign (gojr/src/go/types/mono.ts) - control-flow shape: if=3, for=4, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=12, return=1, defer=0, go=0, call=47, binary=8, unary=0, composite=0, funcLiteral=3 - logic: not yet 1:1-audited
- `*monoGraph.localNamedVertex` @ /usr/local/go1.27rc1/src/go/types/mono.go:269:1 - Present: Method monoGraph.localNamedVertex (gojr/src/go/types/mono.ts) - control-flow shape: if=7, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=12, return=4, defer=0, go=0, call=17, binary=8, unary=4, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*monoGraph.typeParamVertex` @ /usr/local/go1.27rc1/src/go/types/mono.go:311:1 - Present: Method monoGraph.typeParamVertex (gojr/src/go/types/mono.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=8, return=2, defer=0, go=0, call=4, binary=1, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*monoGraph.addEdge` @ /usr/local/go1.27rc1/src/go/types/mono.go:332:1 - Present: Method monoGraph.addEdge (gojr/src/go/types/mono.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

### named.go

#### Constants
- `lazyLoaded` - Present: Variable lazyLoaded (gojr/src/go/types/named.ts) - logic: declaration-only
- `unpacked` - Present: Variable unpacked (gojr/src/go/types/named.ts) - logic: declaration-only
- `hasMethods` - Present: Variable hasMethods (gojr/src/go/types/named.ts); Variable hasMethods (gojr/src/go/types/typeset.ts) - logic: declaration-only
- `hasUnder` - Present: Variable hasUnder (gojr/src/go/types/named.ts) - logic: declaration-only
- `hasVarSize` - Present: Variable hasVarSize (gojr/src/go/types/named.ts) - logic: declaration-only

#### Types
- `Named` (struct) - Present: ClassDeclaration Named (gojr/src/go/types/named.ts) - structure: not yet 1:1-audited
- `instance` (struct) - Present: ClassDeclaration instance (gojr/src/go/types/named.ts) - structure: not yet 1:1-audited
- `stateMask` (*ast.Ident) - Present: TypeAliasDeclaration stateMask (gojr/src/go/types/named.ts) - structure: not yet 1:1-audited

#### Functions
- `NewNamed` @ /usr/local/go1.27rc1/src/go/types/named.go:189:1 - Present: FunctionDeclaration NewNamed (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=5, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `safeUnderlying` @ /usr/local/go1.27rc1/src/go/types/named.go:816:1 - Present: FunctionDeclaration safeUnderlying (gojr/src/go/types/named.ts); FunctionDeclaration safeUnderlying (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Named.unpack` @ /usr/local/go1.27rc1/src/go/types/named.go:222:1 - Present: Method Named.unpack (gojr/src/go/types/named.ts) - control-flow shape: if=5, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=9, return=4, defer=2, go=0, call=21, binary=14, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Named.stateHas` @ /usr/local/go1.27rc1/src/go/types/named.go:288:1 - Present: Method Named.stateHas (gojr/src/go/types/named.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.setState` @ /usr/local/go1.27rc1/src/go/types/named.go:294:1 - Present: Method Named.setState (gojr/src/go/types/named.ts) - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=8, binary=10, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newNamed` @ /usr/local/go1.27rc1/src/go/types/named.go:320:1 - Present: Method Checker.newNamed (gojr/src/go/types/check.ts); Method Checker.newNamed interface method (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=2, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.newNamedInstance` @ /usr/local/go1.27rc1/src/go/types/named.go:338:1 - Present: Method Checker.newNamedInstance interface method (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=6, binary=5, unary=2, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.cleanup` @ /usr/local/go1.27rc1/src/go/types/named.go:361:1 - Present: Method Named.cleanup (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.Obj` @ /usr/local/go1.27rc1/src/go/types/named.go:372:1 - Present: Method Named.Obj (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.Origin` @ /usr/local/go1.27rc1/src/go/types/named.go:381:1 - Present: Method Named.Origin (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.TypeParams` @ /usr/local/go1.27rc1/src/go/types/named.go:390:1 - Present: Method Named.TypeParams (gojr/src/go/types/named.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.SetTypeParams` @ /usr/local/go1.27rc1/src/go/types/named.go:394:1 - Present: Method Named.SetTypeParams (gojr/src/go/types/named.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.TypeArgs` @ /usr/local/go1.27rc1/src/go/types/named.go:400:1 - Present: Method Named.TypeArgs (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.NumMethods` @ /usr/local/go1.27rc1/src/go/types/named.go:408:1 - Present: Method Named.NumMethods (gojr/src/go/types/named.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.Method` @ /usr/local/go1.27rc1/src/go/types/named.go:427:1 - Present: Method Named.Method (gojr/src/go/types/named.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=2, defer=1, go=0, call=15, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.expandMethod` @ /usr/local/go1.27rc1/src/go/types/named.go:464:1 - Present: Method Named.expandMethod (gojr/src/go/types/named.ts) - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=15, return=1, defer=0, go=0, call=14, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.SetUnderlying` @ /usr/local/go1.27rc1/src/go/types/named.go:525:1 - Present: Method Named.SetUnderlying (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=1, go=0, call=8, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.AddMethod` @ /usr/local/go1.27rc1/src/go/types/named.go:548:1 - Present: Method Named.AddMethod (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=6, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.methodIndex` @ /usr/local/go1.27rc1/src/go/types/named.go:560:1 - Present: Method Named.methodIndex (gojr/src/go/types/named.ts) - control-flow shape: if=4, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=4, defer=0, go=0, call=1, binary=2, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.rhs` @ /usr/local/go1.27rc1/src/go/types/named.go:583:1 - Present: Method Named.rhs (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.Underlying` @ /usr/local/go1.27rc1/src/go/types/named.go:595:1 - Present: Method Named.Underlying (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=5, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.String` @ /usr/local/go1.27rc1/src/go/types/named.go:613:1 - Present: Method Named.String (gojr/src/go/types/named.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.resolveUnderlying` @ /usr/local/go1.27rc1/src/go/types/named.go:632:1 - Present: Method Named.resolveUnderlying (gojr/src/go/types/named.ts) - control-flow shape: if=5, for=1, range=1, switch=0, typeSwitch=1, select=0, branch=1, assign=11, return=0, defer=1, go=0, call=16, binary=3, unary=2, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Named.lookupMethod` @ /usr/local/go1.27rc1/src/go/types/named.go:688:1 - Present: Method Named.lookupMethod (gojr/src/go/types/named.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=6, binary=3, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.context` @ /usr/local/go1.27rc1/src/go/types/named.go:703:1 - Present: Method Checker.context interface method (gojr/src/go/types/named.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Named.expandRHS` @ /usr/local/go1.27rc1/src/go/types/named.go:741:1 - Present: Method Named.expandRHS (gojr/src/go/types/named.ts) - control-flow shape: if=8, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=20, return=3, defer=1, go=0, call=26, binary=10, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

### object.go

#### Constants
- `_` - Present: EnumMember _ in VarKind (gojr/src/go/types/object.ts) - logic: declaration-only
- `PackageVar` - Present: EnumMember PackageVar in VarKind (gojr/src/go/types/object.ts); Variable PackageVar (gojr/src/go/types/object.ts) - logic: declaration-only
- `LocalVar` - Present: EnumMember LocalVar in VarKind (gojr/src/go/types/object.ts); Variable LocalVar (gojr/src/go/types/object.ts) - logic: declaration-only
- `RecvVar` - Present: EnumMember RecvVar in VarKind (gojr/src/go/types/object.ts); Variable RecvVar (gojr/src/go/types/object.ts) - logic: declaration-only
- `ParamVar` - Present: EnumMember ParamVar in VarKind (gojr/src/go/types/object.ts); Variable ParamVar (gojr/src/go/types/object.ts) - logic: declaration-only
- `ResultVar` - Present: EnumMember ResultVar in VarKind (gojr/src/go/types/object.ts); Variable ResultVar (gojr/src/go/types/object.ts) - logic: declaration-only
- `FieldVar` - Present: EnumMember FieldVar in VarKind (gojr/src/go/types/object.ts); Variable FieldVar (gojr/src/go/types/object.ts) - logic: declaration-only

#### Variables
- `varKindNames` - Present: Variable varKindNames (gojr/src/go/types/object.ts) - logic: declaration-only

#### Types
- `Object` (interface) - Present: InterfaceDeclaration Object (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `object` (struct) - Present: ClassDeclaration object_ (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `PkgName` (struct) - Present: ClassDeclaration PkgName (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Const` (struct) - Present: ClassDeclaration Const (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `TypeName` (struct) - Present: ClassDeclaration TypeName (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Var` (struct) - Present: ClassDeclaration Var (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `VarKind` (*ast.Ident) - Present: EnumDeclaration VarKind (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Func` (struct) - Present: ClassDeclaration Func (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Label` (struct) - Present: ClassDeclaration Label (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Builtin` (struct) - Present: ClassDeclaration Builtin (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited
- `Nil` (struct) - Present: ClassDeclaration Nil (gojr/src/go/types/object.ts) - structure: not yet 1:1-audited

#### Functions
- `isExported` @ /usr/local/go1.27rc1/src/go/types/object.go:69:1 - Present: FunctionDeclaration isExported (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Id` @ /usr/local/go1.27rc1/src/go/types/object.go:76:1 - Present: FunctionDeclaration Id (gojr/src/go/types/object.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewPkgName` @ /usr/local/go1.27rc1/src/go/types/object.go:212:1 - Present: FunctionDeclaration NewPkgName (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `NewConst` @ /usr/local/go1.27rc1/src/go/types/object.go:228:1 - Present: FunctionDeclaration NewConst (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `NewTypeName` @ /usr/local/go1.27rc1/src/go/types/object.go:253:1 - Present: FunctionDeclaration NewTypeName (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `_NewTypeNameLazy` @ /usr/local/go1.27rc1/src/go/types/object.go:259:1 - Present: FunctionDeclaration _NewTypeNameLazy (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewVar` @ /usr/local/go1.27rc1/src/go/types/object.go:344:1 - Present: FunctionDeclaration NewVar (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewParam` @ /usr/local/go1.27rc1/src/go/types/object.go:352:1 - Present: FunctionDeclaration NewParam (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewField` @ /usr/local/go1.27rc1/src/go/types/object.go:359:1 - Present: FunctionDeclaration NewField (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `newVar` @ /usr/local/go1.27rc1/src/go/types/object.go:367:1 - Present: FunctionDeclaration newVar (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `NewFunc` @ /usr/local/go1.27rc1/src/go/types/object.go:409:1 - Present: FunctionDeclaration NewFunc (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=0, binary=1, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `NewLabel` @ /usr/local/go1.27rc1/src/go/types/object.go:500:1 - Present: FunctionDeclaration NewLabel (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `newBuiltin` @ /usr/local/go1.27rc1/src/go/types/object.go:511:1 - Present: FunctionDeclaration newBuiltin (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `writeObject` @ /usr/local/go1.27rc1/src/go/types/object.go:520:1 - Present: FunctionDeclaration writeObject (gojr/src/go/types/object.ts) - control-flow shape: if=12, for=0, range=0, switch=0, typeSwitch=2, select=0, branch=0, assign=13, return=5, defer=0, go=0, call=44, binary=14, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `packagePrefix` @ /usr/local/go1.27rc1/src/go/types/object.go:622:1 - Present: FunctionDeclaration packagePrefix (gojr/src/go/types/object.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `ObjectString` @ /usr/local/go1.27rc1/src/go/types/object.go:641:1 - Present: FunctionDeclaration ObjectString (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `writeFuncName` @ /usr/local/go1.27rc1/src/go/types/object.go:656:1 - Present: FunctionDeclaration writeFuncName (gojr/src/go/types/object.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=11, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `objectKind` @ /usr/local/go1.27rc1/src/go/types/object.go:680:1 - Present: FunctionDeclaration objectKind (gojr/src/go/types/object.ts) - control-flow shape: if=4, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=0, assign=2, return=17, defer=0, go=0, call=7, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*object.Parent` @ /usr/local/go1.27rc1/src/go/types/object.go:107:1 - Present: Method object_.Parent (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Pos` @ /usr/local/go1.27rc1/src/go/types/object.go:110:1 - Present: Method object_.Pos (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Pkg` @ /usr/local/go1.27rc1/src/go/types/object.go:114:1 - Present: Method object_.Pkg (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Name` @ /usr/local/go1.27rc1/src/go/types/object.go:117:1 - Present: Method object_.Name (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Type` @ /usr/local/go1.27rc1/src/go/types/object.go:120:1 - Present: Method object_.Type (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Exported` @ /usr/local/go1.27rc1/src/go/types/object.go:125:1 - Present: Method object_.Exported (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.Id` @ /usr/local/go1.27rc1/src/go/types/object.go:128:1 - Present: Method object_.Id (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.String` @ /usr/local/go1.27rc1/src/go/types/object.go:130:1 - Present: Method object_.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.order` @ /usr/local/go1.27rc1/src/go/types/object.go:131:1 - Present: Method object_.order (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.scopePos` @ /usr/local/go1.27rc1/src/go/types/object.go:132:1 - Present: Method object_.scopePos (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.setParent` @ /usr/local/go1.27rc1/src/go/types/object.go:134:1 - Present: Method object_.setParent (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.setType` @ /usr/local/go1.27rc1/src/go/types/object.go:135:1 - Present: Method object_.setType (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.setOrder` @ /usr/local/go1.27rc1/src/go/types/object.go:136:1 - Present: Method object_.setOrder (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.setScopePos` @ /usr/local/go1.27rc1/src/go/types/object.go:137:1 - Present: Method object_.setScopePos (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.sameId` @ /usr/local/go1.27rc1/src/go/types/object.go:139:1 - Present: Method object_.sameId (gojr/src/go/types/object.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=4, defer=0, go=0, call=3, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*object.cmp` @ /usr/local/go1.27rc1/src/go/types/object.go:169:1 - Present: Method object_.cmp (gojr/src/go/types/object.ts) - control-flow shape: if=7, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=8, defer=0, go=0, call=4, binary=5, unary=5, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*PkgName.Imported` @ /usr/local/go1.27rc1/src/go/types/object.go:218:1 - Present: Method PkgName.Imported (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Const.Val` @ /usr/local/go1.27rc1/src/go/types/object.go:233:1 - Present: Method Const.Val (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Const.isDependency` @ /usr/local/go1.27rc1/src/go/types/object.go:235:1 - Present: Method Const.isDependency (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeName.IsAlias` @ /usr/local/go1.27rc1/src/go/types/object.go:267:1 - Present: Method TypeName.IsAlias (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=6, defer=0, go=0, call=0, binary=10, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `VarKind.String` @ /usr/local/go1.27rc1/src/go/types/object.go:325:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=3, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.Kind` @ /usr/local/go1.27rc1/src/go/types/object.go:333:1 - Present: Method Var.Kind (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.SetKind` @ /usr/local/go1.27rc1/src/go/types/object.go:337:1 - Present: Method Var.SetKind (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.Anonymous` @ /usr/local/go1.27rc1/src/go/types/object.go:373:1 - Present: Method Var.Anonymous (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.Embedded` @ /usr/local/go1.27rc1/src/go/types/object.go:376:1 - Present: Method Var.Embedded (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.IsField` @ /usr/local/go1.27rc1/src/go/types/object.go:379:1 - Present: Method Var.IsField (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.Origin` @ /usr/local/go1.27rc1/src/go/types/object.go:388:1 - Present: Method Var.Origin (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.isDependency` @ /usr/local/go1.27rc1/src/go/types/object.go:395:1 - Present: Method Var.isDependency (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.Signature` @ /usr/local/go1.27rc1/src/go/types/object.go:423:1 - Present: Method Func.Signature (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.FullName` @ /usr/local/go1.27rc1/src/go/types/object.go:440:1 - Present: Method Func.FullName (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.Scope` @ /usr/local/go1.27rc1/src/go/types/object.go:449:1 - Present: Method Func.Scope (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.Origin` @ /usr/local/go1.27rc1/src/go/types/object.go:458:1 - Present: Method Func.Origin (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.Pkg` @ /usr/local/go1.27rc1/src/go/types/object.go:469:1 - Present: Method Func.Pkg (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.hasPtrRecv` @ /usr/local/go1.27rc1/src/go/types/object.go:472:1 - Present: Method Func.hasPtrRecv (gojr/src/go/types/object.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.isDependency` @ /usr/local/go1.27rc1/src/go/types/object.go:490:1 - Present: Method Func.isDependency (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*PkgName.String` @ /usr/local/go1.27rc1/src/go/types/object.go:647:1 - Present: Method PkgName.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Const.String` @ /usr/local/go1.27rc1/src/go/types/object.go:648:1 - Present: Method Const.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeName.String` @ /usr/local/go1.27rc1/src/go/types/object.go:649:1 - Present: Method TypeName.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Var.String` @ /usr/local/go1.27rc1/src/go/types/object.go:650:1 - Present: Method Var.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Func.String` @ /usr/local/go1.27rc1/src/go/types/object.go:651:1 - Present: Method Func.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Label.String` @ /usr/local/go1.27rc1/src/go/types/object.go:652:1 - Present: Method Label.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Builtin.String` @ /usr/local/go1.27rc1/src/go/types/object.go:653:1 - Present: Method Builtin.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Nil.String` @ /usr/local/go1.27rc1/src/go/types/object.go:654:1 - Present: Method Nil.String (gojr/src/go/types/object.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### objset.go

#### Types
- `objset` (*ast.MapType) - Present: ClassDeclaration objset (gojr/src/go/types/objset.ts) - structure: not yet 1:1-audited

#### Methods
- `*objset.insert` @ /usr/local/go1.27rc1/src/go/types/objset.go:24:1 - Present: Method objset.insert (gojr/src/go/types/objset.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=2, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### operand.go

#### Constants
- `invalid` - Present: EnumMember invalid in operandMode (gojr/src/go/types/operand.ts); Variable invalid (gojr/src/go/types/operand.ts) - logic: declaration-only
- `novalue` - Present: EnumMember novalue in operandMode (gojr/src/go/types/operand.ts); Variable novalue (gojr/src/go/types/operand.ts) - logic: declaration-only
- `builtin` - Present: EnumMember builtin in operandMode (gojr/src/go/types/operand.ts); Variable builtin (gojr/src/go/types/operand.ts) - logic: declaration-only
- `typexpr` - Present: EnumMember typexpr in operandMode (gojr/src/go/types/operand.ts); Variable typexpr (gojr/src/go/types/operand.ts) - logic: declaration-only
- `constant_` - Present: EnumMember constant_ in operandMode (gojr/src/go/types/operand.ts); Variable constant_ (gojr/src/go/types/operand.ts) - logic: declaration-only
- `variable` - Present: EnumMember variable in operandMode (gojr/src/go/types/operand.ts); Variable variable (gojr/src/go/types/operand.ts) - logic: declaration-only
- `mapindex` - Present: EnumMember mapindex in operandMode (gojr/src/go/types/operand.ts); Variable mapindex (gojr/src/go/types/operand.ts) - logic: declaration-only
- `value` - Present: EnumMember value in operandMode (gojr/src/go/types/operand.ts); Variable value (gojr/src/go/types/operand.ts); Variable value (gojr/src/go/types/typestring.ts) - logic: declaration-only
- `nilvalue` - Present: EnumMember nilvalue in operandMode (gojr/src/go/types/operand.ts); Variable nilvalue (gojr/src/go/types/operand.ts) - logic: declaration-only
- `commaok` - Present: EnumMember commaok in operandMode (gojr/src/go/types/operand.ts); Variable commaok (gojr/src/go/types/operand.ts) - logic: declaration-only
- `commaerr` - Present: EnumMember commaerr in operandMode (gojr/src/go/types/operand.ts); Variable commaerr (gojr/src/go/types/operand.ts) - logic: declaration-only
- `cgofunc` - Present: EnumMember cgofunc in operandMode (gojr/src/go/types/operand.ts); Variable cgofunc (gojr/src/go/types/operand.ts) - logic: declaration-only

#### Variables
- `operandModeString` - Present: Variable operandModeString (gojr/src/go/types/operand.ts) - logic: declaration-only

#### Types
- `operandMode` (*ast.Ident) - Present: EnumDeclaration operandMode (gojr/src/go/types/operand.ts) - structure: not yet 1:1-audited
- `operand` (struct) - Present: ClassDeclaration operand (gojr/src/go/types/operand.ts) - structure: not yet 1:1-audited

#### Functions
- `operandString` @ /usr/local/go1.27rc1/src/go/types/operand.go:129:1 - Present: FunctionDeclaration operandString (gojr/src/go/types/format.ts) - control-flow shape: if=17, for=0, range=0, switch=3, typeSwitch=1, select=0, branch=1, assign=12, return=5, defer=0, go=0, call=48, binary=16, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `compositeKind` @ /usr/local/go1.27rc1/src/go/types/operand.go:251:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=0, return=11, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*operand.mode` @ /usr/local/go1.27rc1/src/go/types/operand.go:67:1 - Present: Method operand.mode (gojr/src/go/types/operand.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.typ` @ /usr/local/go1.27rc1/src/go/types/operand.go:71:1 - Present: Method operand.typ (gojr/src/go/types/operand.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.isValid` @ /usr/local/go1.27rc1/src/go/types/operand.go:75:1 - Present: Method operand.isValid (gojr/src/go/types/operand.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.invalidate` @ /usr/local/go1.27rc1/src/go/types/operand.go:79:1 - Present: Method operand.invalidate (gojr/src/go/types/operand.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.Pos` @ /usr/local/go1.27rc1/src/go/types/operand.go:85:1 - Present: Method operand.Pos (gojr/src/go/types/operand.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.String` @ /usr/local/go1.27rc1/src/go/types/operand.go:280:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.setConst` @ /usr/local/go1.27rc1/src/go/types/operand.go:285:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=10, return=1, defer=0, go=0, call=4, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.isNil` @ /usr/local/go1.27rc1/src/go/types/operand.go:314:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=3, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*operand.assignableTo` @ /usr/local/go1.27rc1/src/go/types/operand.go:328:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=24, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=27, return=21, defer=0, go=0, call=32, binary=39, unary=10, composite=0, funcLiteral=4 - logic: not yet 1:1-audited

### package.go

#### Types
- `Package` (struct) - Present: ClassDeclaration Package (gojr/src/go/types/package.ts) - structure: not yet 1:1-audited

#### Functions
- `NewPackage` @ /usr/local/go1.27rc1/src/go/types/package.go:28:1 - Present: FunctionDeclaration NewPackage (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Package.Path` @ /usr/local/go1.27rc1/src/go/types/package.go:34:1 - Present: Method Package.Path (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.Name` @ /usr/local/go1.27rc1/src/go/types/package.go:37:1 - Present: Method Package.Name (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.SetName` @ /usr/local/go1.27rc1/src/go/types/package.go:40:1 - Present: Method Package.SetName (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.GoVersion` @ /usr/local/go1.27rc1/src/go/types/package.go:46:1 - Present: Method Package.GoVersion (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.Scope` @ /usr/local/go1.27rc1/src/go/types/package.go:52:1 - Present: Method Package.Scope (gojr/src/go/types/package.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.Complete` @ /usr/local/go1.27rc1/src/go/types/package.go:61:1 - Present: Method Package.Complete (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.MarkComplete` @ /usr/local/go1.27rc1/src/go/types/package.go:64:1 - Present: Method Package.MarkComplete (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.Imports` @ /usr/local/go1.27rc1/src/go/types/package.go:75:1 - Present: Method Package.Imports (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.SetImports` @ /usr/local/go1.27rc1/src/go/types/package.go:79:1 - Present: Method Package.SetImports (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Package.String` @ /usr/local/go1.27rc1/src/go/types/package.go:81:1 - Present: Method Package.String (gojr/src/go/types/package.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### pointer.go

#### Types
- `Pointer` (struct) - Present: ClassDeclaration Pointer (gojr/src/go/types/pointer.ts) - structure: not yet 1:1-audited

#### Functions
- `NewPointer` @ /usr/local/go1.27rc1/src/go/types/pointer.go:16:1 - Present: FunctionDeclaration NewPointer (gojr/src/go/types/pointer.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Pointer.Elem` @ /usr/local/go1.27rc1/src/go/types/pointer.go:19:1 - Present: Method Pointer.Elem (gojr/src/go/types/pointer.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Pointer.Underlying` @ /usr/local/go1.27rc1/src/go/types/pointer.go:21:1 - Present: Method Pointer.Underlying (gojr/src/go/types/pointer.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Pointer.String` @ /usr/local/go1.27rc1/src/go/types/pointer.go:22:1 - Present: Method Pointer.String (gojr/src/go/types/pointer.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### predicates.go

#### Types
- `ifacePair` (struct) - Present: ClassDeclaration ifacePair (gojr/src/go/types/predicates.ts) - structure: not yet 1:1-audited
- `comparer` (struct) - Present: ClassDeclaration comparer (gojr/src/go/types/predicates.ts) - structure: not yet 1:1-audited

#### Functions
- `isValid` @ /usr/local/go1.27rc1/src/go/types/predicates.go:18:1 - Present: FunctionDeclaration isValid (gojr/src/go/types/decl.ts); FunctionDeclaration isValid (gojr/src/go/types/predicates.ts); FunctionDeclaration isValid (gojr/src/go/types/recording.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isBoolean` @ /usr/local/go1.27rc1/src/go/types/predicates.go:24:1 - Present: FunctionDeclaration isBoolean (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isInteger` @ /usr/local/go1.27rc1/src/go/types/predicates.go:25:1 - Present: FunctionDeclaration isInteger (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isUnsigned` @ /usr/local/go1.27rc1/src/go/types/predicates.go:26:1 - Present: FunctionDeclaration isUnsigned (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isFloat` @ /usr/local/go1.27rc1/src/go/types/predicates.go:27:1 - Present: FunctionDeclaration isFloat (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isComplex` @ /usr/local/go1.27rc1/src/go/types/predicates.go:28:1 - Present: FunctionDeclaration isComplex (gojr/src/go/types/predicates.ts); FunctionDeclaration isComplex (gojr/src/go/types/sizes.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isNumeric` @ /usr/local/go1.27rc1/src/go/types/predicates.go:29:1 - Present: FunctionDeclaration isNumeric (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isString` @ /usr/local/go1.27rc1/src/go/types/predicates.go:30:1 - Present: FunctionDeclaration isString (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isIntegerOrFloat` @ /usr/local/go1.27rc1/src/go/types/predicates.go:31:1 - Present: FunctionDeclaration isIntegerOrFloat (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isConstType` @ /usr/local/go1.27rc1/src/go/types/predicates.go:32:1 - Present: FunctionDeclaration isConstType (gojr/src/go/types/decl.ts); FunctionDeclaration isConstType (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isBasic` @ /usr/local/go1.27rc1/src/go/types/predicates.go:37:1 - Present: FunctionDeclaration isBasic (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allBoolean` @ /usr/local/go1.27rc1/src/go/types/predicates.go:46:1 - Present: FunctionDeclaration allBoolean (gojr/src/go/types/predicates.ts); FunctionDeclaration allBoolean (gojr/src/go/types/recording.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allInteger` @ /usr/local/go1.27rc1/src/go/types/predicates.go:47:1 - Present: FunctionDeclaration allInteger (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allUnsigned` @ /usr/local/go1.27rc1/src/go/types/predicates.go:48:1 - Present: FunctionDeclaration allUnsigned (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allNumeric` @ /usr/local/go1.27rc1/src/go/types/predicates.go:49:1 - Present: FunctionDeclaration allNumeric (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allString` @ /usr/local/go1.27rc1/src/go/types/predicates.go:50:1 - Present: FunctionDeclaration allString (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allOrdered` @ /usr/local/go1.27rc1/src/go/types/predicates.go:51:1 - Present: FunctionDeclaration allOrdered (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allNumericOrString` @ /usr/local/go1.27rc1/src/go/types/predicates.go:52:1 - Present: FunctionDeclaration allNumericOrString (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `allBasic` @ /usr/local/go1.27rc1/src/go/types/predicates.go:57:1 - Present: FunctionDeclaration allBasic (gojr/src/go/types/predicates.ts); FunctionDeclaration allBasic (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=3, defer=0, go=0, call=4, binary=3, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `hasName` @ /usr/local/go1.27rc1/src/go/types/predicates.go:67:1 - Present: FunctionDeclaration hasName (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isTypeLit` @ /usr/local/go1.27rc1/src/go/types/predicates.go:78:1 - Present: FunctionDeclaration isTypeLit (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isTyped` @ /usr/local/go1.27rc1/src/go/types/predicates.go:89:1 - Present: FunctionDeclaration isTyped (gojr/src/go/types/predicates.ts); FunctionDeclaration isTyped (gojr/src/go/types/recording.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=0, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isUntyped` @ /usr/local/go1.27rc1/src/go/types/predicates.go:98:1 - Present: FunctionDeclaration isUntyped (gojr/src/go/types/predicates.ts); FunctionDeclaration isUntyped (gojr/src/go/types/recording.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isUntypedNumeric` @ /usr/local/go1.27rc1/src/go/types/predicates.go:104:1 - Present: FunctionDeclaration isUntypedNumeric (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=0, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `IsInterface` @ /usr/local/go1.27rc1/src/go/types/predicates.go:112:1 - Present: FunctionDeclaration IsInterface (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isNonTypeParamInterface` @ /usr/local/go1.27rc1/src/go/types/predicates.go:118:1 - Present: FunctionDeclaration isNonTypeParamInterface (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isTypeParam` @ /usr/local/go1.27rc1/src/go/types/predicates.go:123:1 - Present: FunctionDeclaration isTypeParam (gojr/src/go/types/decl.ts); FunctionDeclaration isTypeParam (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasEmptyTypeset` @ /usr/local/go1.27rc1/src/go/types/predicates.go:132:1 - Present: FunctionDeclaration hasEmptyTypeset (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=3, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isGeneric` @ /usr/local/go1.27rc1/src/go/types/predicates.go:143:1 - Present: FunctionDeclaration isGeneric (gojr/src/go/types/decl.ts); FunctionDeclaration isGeneric (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=3, binary=12, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Comparable` @ /usr/local/go1.27rc1/src/go/types/predicates.go:153:1 - Present: FunctionDeclaration Comparable (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `comparableType` @ /usr/local/go1.27rc1/src/go/types/predicates.go:160:1 - Present: FunctionDeclaration comparableType (gojr/src/go/types/predicates.ts) - control-flow shape: if=7, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=5, return=8, defer=0, go=0, call=14, binary=6, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasNil` @ /usr/local/go1.27rc1/src/go/types/predicates.go:211:1 - Present: FunctionDeclaration hasNil (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=5, defer=0, go=0, call=4, binary=4, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `samePkg` @ /usr/local/go1.27rc1/src/go/types/predicates.go:226:1 - Present: FunctionDeclaration samePkg (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `identicalOrigin` @ /usr/local/go1.27rc1/src/go/types/predicates.go:504:1 - Present: FunctionDeclaration identicalOrigin (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `identicalInstance` @ /usr/local/go1.27rc1/src/go/types/predicates.go:512:1 - Present: FunctionDeclaration identicalInstance (gojr/src/go/types/context.ts); FunctionDeclaration identicalInstance (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `Default` @ /usr/local/go1.27rc1/src/go/types/predicates.go:523:1 - Present: FunctionDeclaration Default (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=1, return=7, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `maxType` @ /usr/local/go1.27rc1/src/go/types/predicates.go:549:1 - Present: FunctionDeclaration maxType (gojr/src/go/types/predicates.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=4, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `clone` @ /usr/local/go1.27rc1/src/go/types/predicates.go:566:1 - Present: FunctionDeclaration clone (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isValidName` @ /usr/local/go1.27rc1/src/go/types/predicates.go:572:1 - Present: FunctionDeclaration isValidName (gojr/src/go/types/predicates.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*ifacePair.identical` @ /usr/local/go1.27rc1/src/go/types/predicates.go:241:1 - Present: Method ifacePair.identical (gojr/src/go/types/predicates.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*comparer.identical` @ /usr/local/go1.27rc1/src/go/types/predicates.go:252:1 - Present: Method comparer.identical (gojr/src/go/types/predicates.ts) - control-flow shape: if=30, for=1, range=6, switch=0, typeSwitch=1, select=0, branch=0, assign=39, return=26, defer=0, go=0, call=65, binary=36, unary=11, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

### range.go

#### Functions
- `rangeKeyVal` @ /usr/local/go1.27rc1/src/go/types/range.go:205:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=10, for=0, range=0, switch=2, typeSwitch=1, select=0, branch=0, assign=8, return=22, defer=0, go=0, call=60, binary=22, unary=3, composite=0, funcLiteral=2 - logic: not yet 1:1-audited

#### Methods
- `*Checker.rangeStmt` @ /usr/local/go1.27rc1/src/go/types/range.go:26:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=17, for=1, range=3, switch=1, typeSwitch=0, select=0, branch=5, assign=25, return=1, defer=1, go=0, call=46, binary=30, unary=18, composite=3, funcLiteral=1 - logic: not yet 1:1-audited

### recording.go

#### Methods
- `*Checker.record` @ /usr/local/go1.27rc1/src/go/types/recording.go:18:1 - Present: Method Checker.record interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=10, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordUntyped` @ /usr/local/go1.27rc1/src/go/types/recording.go:45:1 - Present: Method Checker.recordUntyped interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=6, binary=2, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordTypeAndValue` @ /usr/local/go1.27rc1/src/go/types/recording.go:59:1 - Present: Method Checker.recordTypeAndValue interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=7, binary=7, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordBuiltinType` @ /usr/local/go1.27rc1/src/go/types/recording.go:77:1 - Present: Method Checker.recordBuiltinType interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=0, for=1, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordCommaOkTypes` @ /usr/local/go1.27rc1/src/go/types/recording.go:97:1 - Present: Method Checker.recordCommaOkTypes interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=3, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=1, assign=8, return=1, defer=0, go=0, call=16, binary=10, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordInstance` @ /usr/local/go1.27rc1/src/go/types/recording.go:132:1 - Present: Method Checker.recordInstance interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=13, return=2, defer=0, go=0, call=6, binary=7, unary=3, composite=3, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordDef` @ /usr/local/go1.27rc1/src/go/types/recording.go:158:1 - Present: Method Checker.recordDef interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordUse` @ /usr/local/go1.27rc1/src/go/types/recording.go:165:1 - Present: Method Checker.recordUse interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordImplicit` @ /usr/local/go1.27rc1/src/go/types/recording.go:173:1 - Present: Method Checker.recordImplicit interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordSelection` @ /usr/local/go1.27rc1/src/go/types/recording.go:181:1 - Present: Method Checker.recordSelection interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=3, binary=6, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.recordScope` @ /usr/local/go1.27rc1/src/go/types/recording.go:189:1 - Present: Method Checker.recordScope interface method (gojr/src/go/types/recording.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### resolver.go

#### Types
- `declInfo` (struct) - Present: ClassDeclaration declInfo (gojr/src/go/types/resolver.ts) - structure: not yet 1:1-audited

#### Functions
- `validatedImportPath` @ /usr/local/go1.27rc1/src/go/types/resolver.go:86:1 - Present: FunctionDeclaration validatedImportPath (gojr/src/go/types/resolver.ts) - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=4, defer=0, go=0, call=6, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `dir` @ /usr/local/go1.27rc1/src/go/types/resolver.go:740:1 - Present: FunctionDeclaration dir (gojr/src/go/types/resolver.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*declInfo.hasInitializer` @ /usr/local/go1.27rc1/src/go/types/resolver.go:37:1 - Present: Method declInfo.hasInitializer (gojr/src/go/types/resolver.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*declInfo.addDep` @ /usr/local/go1.27rc1/src/go/types/resolver.go:42:1 - Present: Method declInfo.addDep (gojr/src/go/types/resolver.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.arityMatch` @ /usr/local/go1.27rc1/src/go/types/resolver.go:55:1 - Present: Method Checker.arityMatch interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=3, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=10, binary=12, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.declarePkgObj` @ /usr/local/go1.27rc1/src/go/types/resolver.go:105:1 - Present: Method Checker.declarePkgObj interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=8, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.filename` @ /usr/local/go1.27rc1/src/go/types/resolver.go:128:1 - Present: Method Checker.filename interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=5, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.importPackage` @ /usr/local/go1.27rc1/src/go/types/resolver.go:136:1 - Present: Method Checker.importPackage interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=14, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=22, return=3, defer=0, go=0, call=13, binary=30, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.collectObjects` @ /usr/local/go1.27rc1/src/go/types/resolver.go:216:1 - Present: Method Checker.collectObjects interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=32, for=0, range=8, switch=0, typeSwitch=1, select=0, branch=0, assign=58, return=6, defer=0, go=0, call=90, binary=40, unary=7, composite=7, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.sortObjects` @ /usr/local/go1.27rc1/src/go/types/resolver.go:502:1 - Present: Method Checker.sortObjects interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=6, binary=0, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.unpackRecv` @ /usr/local/go1.27rc1/src/go/types/resolver.go:523:1 - Present: Method Checker.unpackRecv interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=2, select=0, branch=0, assign=10, return=1, defer=0, go=0, call=7, binary=2, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.resolveBaseTypeName` @ /usr/local/go1.27rc1/src/go/types/resolver.go:565:1 - Present: Method Checker.resolveBaseTypeName interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=9, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=6, assign=10, return=2, defer=0, go=0, call=5, binary=7, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.packageObjects` @ /usr/local/go1.27rc1/src/go/types/resolver.go:635:1 - Present: Method Checker.packageObjects interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=4, for=0, range=5, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=8, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.unusedImports` @ /usr/local/go1.27rc1/src/go/types/resolver.go:700:1 - Present: Method Checker.unusedImports interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.errorUnusedPkg` @ /usr/local/go1.27rc1/src/go/types/resolver.go:717:1 - Present: Method Checker.errorUnusedPkg interface method (gojr/src/go/types/resolver.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=3, binary=7, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### return.go

#### Functions
- `hasBreak` @ /usr/local/go1.27rc1/src/go/types/return.go:110:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=9, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=1, return=13, defer=0, go=0, call=12, binary=16, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasBreakList` @ /usr/local/go1.27rc1/src/go/types/return.go:177:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.isTerminating` @ /usr/local/go1.27rc1/src/go/types/return.go:17:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=3, return=12, defer=0, go=0, call=11, binary=10, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isTerminatingList` @ /usr/local/go1.27rc1/src/go/types/return.go:79:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=2, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isTerminatingSwitch` @ /usr/local/go1.27rc1/src/go/types/return.go:89:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=0, go=0, call=2, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### scope.go

#### Types
- `Scope` (struct) - Present: ClassDeclaration Scope (gojr/src/go/types/scope.ts) - structure: not yet 1:1-audited
- `lazyObject` (struct) - Present: ClassDeclaration lazyObject (gojr/src/go/types/scope.ts) - structure: not yet 1:1-audited

#### Functions
- `NewScope` @ /usr/local/go1.27rc1/src/go/types/scope.go:37:1 - Present: FunctionDeclaration NewScope (gojr/src/go/types/scope.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=2, binary=3, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `resolve` @ /usr/local/go1.27rc1/src/go/types/scope.go:175:1 - Present: FunctionDeclaration resolve (gojr/src/go/types/scope.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=7, binary=2, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

#### Methods
- `*Scope.Parent` @ /usr/local/go1.27rc1/src/go/types/scope.go:48:1 - Present: Method Scope.Parent (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Len` @ /usr/local/go1.27rc1/src/go/types/scope.go:51:1 - Present: Method Scope.Len (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Names` @ /usr/local/go1.27rc1/src/go/types/scope.go:54:1 - Present: Method Scope.Names (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.NumChildren` @ /usr/local/go1.27rc1/src/go/types/scope.go:66:1 - Present: Method Scope.NumChildren (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Child` @ /usr/local/go1.27rc1/src/go/types/scope.go:69:1 - Present: Method Scope.Child (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Lookup` @ /usr/local/go1.27rc1/src/go/types/scope.go:73:1 - Present: Method Scope.Lookup (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.lookupIgnoringCase` @ /usr/local/go1.27rc1/src/go/types/scope.go:78:1 - Present: Method Scope.lookupIgnoringCase (gojr/src/go/types/scope.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=5, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Insert` @ /usr/local/go1.27rc1/src/go/types/scope.go:93:1 - Present: Method Scope.Insert (gojr/src/go/types/scope.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=5, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope._InsertLazy` @ /usr/local/go1.27rc1/src/go/types/scope.go:117:1 - Present: Method Scope._InsertLazy (gojr/src/go/types/scope.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.insert` @ /usr/local/go1.27rc1/src/go/types/scope.go:125:1 - Present: Method Scope.insert (gojr/src/go/types/scope.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.WriteTo` @ /usr/local/go1.27rc1/src/go/types/scope.go:137:1 - Present: Method Scope.WriteTo (gojr/src/go/types/scope.ts) - control-flow shape: if=1, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=7, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.String` @ /usr/local/go1.27rc1/src/go/types/scope.go:158:1 - Present: Method Scope.String (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Parent` @ /usr/local/go1.27rc1/src/go/types/scope.go:200:1 - Present: Method lazyObject.Parent (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Pos` @ /usr/local/go1.27rc1/src/go/types/scope.go:201:1 - Present: Method lazyObject.Pos (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Pkg` @ /usr/local/go1.27rc1/src/go/types/scope.go:202:1 - Present: Method lazyObject.Pkg (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Name` @ /usr/local/go1.27rc1/src/go/types/scope.go:203:1 - Present: Method lazyObject.Name (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Type` @ /usr/local/go1.27rc1/src/go/types/scope.go:204:1 - Present: Method lazyObject.Type (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Exported` @ /usr/local/go1.27rc1/src/go/types/scope.go:205:1 - Present: Method lazyObject.Exported (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.Id` @ /usr/local/go1.27rc1/src/go/types/scope.go:206:1 - Present: Method lazyObject.Id (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.String` @ /usr/local/go1.27rc1/src/go/types/scope.go:207:1 - Present: Method lazyObject.String (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.order` @ /usr/local/go1.27rc1/src/go/types/scope.go:208:1 - Present: Method lazyObject.order (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.setType` @ /usr/local/go1.27rc1/src/go/types/scope.go:209:1 - Present: Method lazyObject.setType (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.setOrder` @ /usr/local/go1.27rc1/src/go/types/scope.go:210:1 - Present: Method lazyObject.setOrder (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.setParent` @ /usr/local/go1.27rc1/src/go/types/scope.go:211:1 - Present: Method lazyObject.setParent (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.sameId` @ /usr/local/go1.27rc1/src/go/types/scope.go:212:1 - Present: Method lazyObject.sameId (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.scopePos` @ /usr/local/go1.27rc1/src/go/types/scope.go:213:1 - Present: Method lazyObject.scopePos (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*lazyObject.setScopePos` @ /usr/local/go1.27rc1/src/go/types/scope.go:214:1 - Present: Method lazyObject.setScopePos (gojr/src/go/types/scope.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### scope2.go

#### Methods
- `*Scope.LookupParent` @ /usr/local/go1.27rc1/src/go/types/scope2.go:24:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=4, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Pos` @ /usr/local/go1.27rc1/src/go/types/scope2.go:37:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.End` @ /usr/local/go1.27rc1/src/go/types/scope2.go:38:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Contains` @ /usr/local/go1.27rc1/src/go/types/scope2.go:43:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Scope.Innermost` @ /usr/local/go1.27rc1/src/go/types/scope2.go:52:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=4, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### selection.go

#### Constants
- `FieldVal` - Present: EnumMember FieldVal in SelectionKind (gojr/src/go/types/selection.ts); Variable FieldVal (gojr/src/go/types/selection.ts) - logic: declaration-only
- `MethodVal` - Present: EnumMember MethodVal in SelectionKind (gojr/src/go/types/selection.ts); Variable MethodVal (gojr/src/go/types/selection.ts) - logic: declaration-only
- `MethodExpr` - Present: EnumMember MethodExpr in SelectionKind (gojr/src/go/types/selection.ts); Variable MethodExpr (gojr/src/go/types/selection.ts) - logic: declaration-only

#### Types
- `SelectionKind` (*ast.Ident) - Present: EnumDeclaration SelectionKind (gojr/src/go/types/selection.ts) - structure: not yet 1:1-audited
- `Selection` (struct) - Present: ClassDeclaration Selection (gojr/src/go/types/selection.ts) - structure: not yet 1:1-audited

#### Functions
- `SelectionString` @ /usr/local/go1.27rc1/src/go/types/selection.go:161:1 - Present: FunctionDeclaration SelectionString (gojr/src/go/types/selection.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=12, binary=1, unary=4, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Selection.Kind` @ /usr/local/go1.27rc1/src/go/types/selection.go:84:1 - Present: Method Selection.Kind (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.Recv` @ /usr/local/go1.27rc1/src/go/types/selection.go:87:1 - Present: Method Selection.Recv (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.Obj` @ /usr/local/go1.27rc1/src/go/types/selection.go:91:1 - Present: Method Selection.Obj (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.Type` @ /usr/local/go1.27rc1/src/go/types/selection.go:95:1 - Present: Method Selection.Type (gojr/src/go/types/selection.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=11, return=3, defer=0, go=0, call=3, binary=1, unary=4, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.Index` @ /usr/local/go1.27rc1/src/go/types/selection.go:139:1 - Present: Method Selection.Index (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.Indirect` @ /usr/local/go1.27rc1/src/go/types/selection.go:148:1 - Present: Method Selection.Indirect (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Selection.String` @ /usr/local/go1.27rc1/src/go/types/selection.go:150:1 - Present: Method Selection.String (gojr/src/go/types/selection.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### signature.go

#### Variables
- `methodExprSentinel` - Present: Variable methodExprSentinel (gojr/src/go/types/signature.ts) - logic: declaration-only

#### Types
- `Signature` (struct) - Present: ClassDeclaration Signature (gojr/src/go/types/signature.ts) - structure: not yet 1:1-audited

#### Functions
- `NewSignature` @ /usr/local/go1.27rc1/src/go/types/signature.go:58:1 - Present: FunctionDeclaration NewSignature (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewSignatureType` @ /usr/local/go1.27rc1/src/go/types/signature.go:76:1 - Present: FunctionDeclaration NewSignatureType (gojr/src/go/types/signature.ts) - control-flow shape: if=10, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=2, assign=9, return=1, defer=0, go=0, call=15, binary=10, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `unpointer` @ /usr/local/go1.27rc1/src/go/types/signature.go:357:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isCGoTypeObj` @ /usr/local/go1.27rc1/src/go/types/signature.go:514:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=5, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Signature.Recv` @ /usr/local/go1.27rc1/src/go/types/signature.go:136:1 - Present: Method Signature.Recv (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.TypeParams` @ /usr/local/go1.27rc1/src/go/types/signature.go:139:1 - Present: Method Signature.TypeParams (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.RecvTypeParams` @ /usr/local/go1.27rc1/src/go/types/signature.go:142:1 - Present: Method Signature.RecvTypeParams (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.Params` @ /usr/local/go1.27rc1/src/go/types/signature.go:146:1 - Present: Method Signature.Params (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.Results` @ /usr/local/go1.27rc1/src/go/types/signature.go:149:1 - Present: Method Signature.Results (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.Variadic` @ /usr/local/go1.27rc1/src/go/types/signature.go:152:1 - Present: Method Signature.Variadic (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.Underlying` @ /usr/local/go1.27rc1/src/go/types/signature.go:154:1 - Present: Method Signature.Underlying (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Signature.String` @ /usr/local/go1.27rc1/src/go/types/signature.go:155:1 - Present: Method Signature.String (gojr/src/go/types/signature.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.funcType` @ /usr/local/go1.27rc1/src/go/types/signature.go:161:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=13, return=0, defer=1, go=0, call=17, binary=9, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.collectRecv` @ /usr/local/go1.27rc1/src/go/types/signature.go:209:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=1, range=3, switch=0, typeSwitch=1, select=0, branch=1, assign=27, return=1, defer=0, go=0, call=50, binary=15, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.recordParenthesizedRecvTypes` @ /usr/local/go1.27rc1/src/go/types/signature.go:377:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=5, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.collectParams` @ /usr/local/go1.27rc1/src/go/types/signature.go:402:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=7, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=15, return=2, defer=0, go=0, call=19, binary=13, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.declareParams` @ /usr/local/go1.27rc1/src/go/types/signature.go:463:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.validRecv` @ /usr/local/go1.27rc1/src/go/types/signature.go:473:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=2, select=0, branch=1, assign=6, return=1, defer=0, go=0, call=10, binary=4, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### sizes.go

#### Variables
- `basicSizes` - Present: Variable basicSizes (gojr/src/go/types/sizes.ts) - logic: declaration-only
- `gcArchSizes` - Present: Variable gcArchSizes (gojr/src/go/types/sizes.ts) - logic: declaration-only
- `stdSizes` - Present: Variable stdSizes (gojr/src/go/types/sizes.ts) - logic: declaration-only

#### Types
- `Sizes` (interface) - Present: InterfaceDeclaration Sizes (gojr/src/go/types/sizes.ts) - structure: not yet 1:1-audited
- `StdSizes` (struct) - Present: ClassDeclaration StdSizes (gojr/src/go/types/sizes.ts) - structure: not yet 1:1-audited

#### Functions
- `_IsSyncAtomicAlign64` @ /usr/local/go1.27rc1/src/go/types/sizes.go:117:1 - Present: FunctionDeclaration _IsSyncAtomicAlign64 (gojr/src/go/types/sizes.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=8, binary=8, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `SizesFor` @ /usr/local/go1.27rc1/src/go/types/sizes.go:260:1 - Present: FunctionDeclaration SizesFor (gojr/src/go/types/sizes.ts) - control-flow shape: if=2, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=2, return=3, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `align` @ /usr/local/go1.27rc1/src/go/types/sizes.go:340:1 - Present: FunctionDeclaration align (gojr/src/go/types/sizes.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=13, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*StdSizes.Alignof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:53:1 - Present: Method StdSizes.Alignof (gojr/src/go/types/sizes.ts) - control-flow shape: if=6, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=6, return=8, defer=1, go=0, call=14, binary=8, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*StdSizes.Offsetsof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:129:1 - Present: Method StdSizes.Offsetsof (gojr/src/go/types/sizes.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=8, return=1, defer=0, go=0, call=5, binary=4, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*StdSizes.Sizeof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:167:1 - Present: Method StdSizes.Sizeof (gojr/src/go/types/sizes.ts) - control-flow shape: if=10, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=12, return=14, defer=0, go=0, call=15, binary=26, unary=5, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Config.alignof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:277:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Config.offsetsof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:288:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=5, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Config.offsetof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:310:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=3, defer=0, go=0, call=2, binary=2, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Config.sizeof` @ /usr/local/go1.27rc1/src/go/types/sizes.go:329:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### slice.go

#### Types
- `Slice` (struct) - Present: ClassDeclaration Slice (gojr/src/go/types/slice.ts) - structure: not yet 1:1-audited

#### Functions
- `NewSlice` @ /usr/local/go1.27rc1/src/go/types/slice.go:16:1 - Present: FunctionDeclaration NewSlice (gojr/src/go/types/slice.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Slice.Elem` @ /usr/local/go1.27rc1/src/go/types/slice.go:19:1 - Present: Method Slice.Elem (gojr/src/go/types/slice.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Slice.Underlying` @ /usr/local/go1.27rc1/src/go/types/slice.go:21:1 - Present: Method Slice.Underlying (gojr/src/go/types/slice.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Slice.String` @ /usr/local/go1.27rc1/src/go/types/slice.go:22:1 - Present: Method Slice.String (gojr/src/go/types/slice.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### stmt.go

#### Constants
- `breakOk` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `continueOk` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `fallthroughOk` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `finalSwitchCase` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `inTypeSwitch` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Types
- `stmtContext` (*ast.Ident) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `valueMap` (*ast.MapType) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `valueType` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `trimTrailingEmptyStmts` @ /usr/local/go1.27rc1/src/go/types/stmt.go:107:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `assignOp` @ /usr/local/go1.27rc1/src/go/types/stmt.go:165:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `goVal` @ /usr/local/go1.27rc1/src/go/types/stmt.go:196:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=3, return=6, defer=0, go=0, call=5, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.funcBody` @ /usr/local/go1.27rc1/src/go/types/stmt.go:18:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=0, defer=1, go=0, call=11, binary=2, unary=1, composite=1, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.usage` @ /usr/local/go1.27rc1/src/go/types/stmt.go:57:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=3, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=2, defer=0, go=0, call=7, binary=8, unary=3, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*Checker.simpleStmt` @ /usr/local/go1.27rc1/src/go/types/stmt.go:101:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.stmtList` @ /usr/local/go1.27rc1/src/go/types/stmt.go:116:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=5, return=0, defer=0, go=0, call=3, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.multipleDefaults` @ /usr/local/go1.27rc1/src/go/types/stmt.go:129:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=1, switch=0, typeSwitch=1, select=0, branch=0, assign=4, return=0, defer=0, go=0, call=5, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.openScope` @ /usr/local/go1.27rc1/src/go/types/stmt.go:155:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=4, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.closeScope` @ /usr/local/go1.27rc1/src/go/types/stmt.go:161:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.suspendedCall` @ /usr/local/go1.27rc1/src/go/types/stmt.go:173:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=5, return=1, defer=0, go=0, call=3, binary=1, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.caseValues` @ /usr/local/go1.27rc1/src/go/types/stmt.go:237:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=6, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=5, assign=4, return=0, defer=0, go=0, call=20, binary=3, unary=9, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.isNil` @ /usr/local/go1.27rc1/src/go/types/stmt.go:277:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.caseTypes` @ /usr/local/go1.27rc1/src/go/types/stmt.go:307:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=7, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=2, assign=8, return=1, defer=0, go=0, call=14, binary=17, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.caseTypes_currently_unused` @ /usr/local/go1.27rc1/src/go/types/stmt.go:359:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=7, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=2, assign=10, return=1, defer=0, go=0, call=14, binary=8, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.stmt` @ /usr/local/go1.27rc1/src/go/types/stmt.go:411:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=47, for=0, range=5, switch=5, typeSwitch=4, select=0, branch=4, assign=63, return=15, defer=7, go=0, call=152, binary=71, unary=40, composite=2, funcLiteral=1 - logic: not yet 1:1-audited

### struct.go

#### Types
- `Struct` (struct) - Present: ClassDeclaration Struct (gojr/src/go/types/struct.ts) - structure: not yet 1:1-audited

#### Functions
- `NewStruct` @ /usr/local/go1.27rc1/src/go/types/struct.go:27:1 - Present: FunctionDeclaration NewStruct (gojr/src/go/types/struct.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=6, binary=4, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `embeddedFieldIdent` @ /usr/local/go1.27rc1/src/go/types/struct.go:181:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=2, return=6, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Struct.NumFields` @ /usr/local/go1.27rc1/src/go/types/struct.go:43:1 - Present: Method Struct.NumFields (gojr/src/go/types/struct.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Struct.Field` @ /usr/local/go1.27rc1/src/go/types/struct.go:46:1 - Present: Method Struct.Field (gojr/src/go/types/struct.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Struct.Tag` @ /usr/local/go1.27rc1/src/go/types/struct.go:49:1 - Present: Method Struct.Tag (gojr/src/go/types/struct.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Struct.Underlying` @ /usr/local/go1.27rc1/src/go/types/struct.go:56:1 - Present: Method Struct.Underlying (gojr/src/go/types/struct.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Struct.String` @ /usr/local/go1.27rc1/src/go/types/struct.go:57:1 - Present: Method Struct.String (gojr/src/go/types/struct.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Struct.markComplete` @ /usr/local/go1.27rc1/src/go/types/struct.go:62:1 - Present: Method Struct.markComplete (gojr/src/go/types/struct.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.structType` @ /usr/local/go1.27rc1/src/go/types/struct.go:68:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=10, for=0, range=2, switch=0, typeSwitch=1, select=0, branch=2, assign=23, return=2, defer=0, go=0, call=31, binary=10, unary=2, composite=0, funcLiteral=3 - logic: not yet 1:1-audited
- `*Checker.declareInSet` @ /usr/local/go1.27rc1/src/go/types/struct.go:200:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=7, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.tag` @ /usr/local/go1.27rc1/src/go/types/struct.go:211:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### subst.go

#### Types
- `substMap` (*ast.MapType) - Present: ClassDeclaration substMap (gojr/src/go/types/subst.ts) - structure: not yet 1:1-audited
- `subster` (struct) - Present: ClassDeclaration subster (gojr/src/go/types/subst.ts) - structure: not yet 1:1-audited

#### Functions
- `makeSubstMap` @ /usr/local/go1.27rc1/src/go/types/subst.go:20:1 - Present: FunctionDeclaration makeSubstMap (gojr/src/go/types/subst.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=5, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `makeRenameMap` @ /usr/local/go1.27rc1/src/go/types/subst.go:31:1 - Present: FunctionDeclaration makeRenameMap (gojr/src/go/types/subst.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=5, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `cloneVar` @ /usr/local/go1.27rc1/src/go/types/subst.go:375:1 - Present: FunctionDeclaration cloneVar (gojr/src/go/types/selection.ts); FunctionDeclaration cloneVar (gojr/src/go/types/subst.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `substList` @ /usr/local/go1.27rc1/src/go/types/subst.go:395:1 - Present: FunctionDeclaration substList (gojr/src/go/types/subst.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=4, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `cloneFunc` @ /usr/local/go1.27rc1/src/go/types/subst.go:418:1 - Present: FunctionDeclaration cloneFunc (gojr/src/go/types/subst.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `replaceRecvType` @ /usr/local/go1.27rc1/src/go/types/subst.go:438:1 - Present: FunctionDeclaration replaceRecvType (gojr/src/go/types/subst.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=7, return=1, defer=0, go=0, call=7, binary=3, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `substMap.empty` @ /usr/local/go1.27rc1/src/go/types/subst.go:40:1 - Present: Method substMap.empty (gojr/src/go/types/subst.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `substMap.lookup` @ /usr/local/go1.27rc1/src/go/types/subst.go:44:1 - Present: Method substMap.lookup (gojr/src/go/types/subst.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.subst` @ /usr/local/go1.27rc1/src/go/types/subst.go:58:1 - Present: Method Checker.subst interface method (gojr/src/go/types/subst.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=2, return=4, defer=0, go=0, call=4, binary=3, unary=0, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.typ` @ /usr/local/go1.27rc1/src/go/types/subst.go:92:1 - Present: Method subster.typ (gojr/src/go/types/subst.ts) - control-flow shape: if=22, for=0, range=4, switch=0, typeSwitch=1, select=0, branch=0, assign=45, return=18, defer=0, go=0, call=58, binary=30, unary=11, composite=10, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.typOrNil` @ /usr/local/go1.27rc1/src/go/types/subst.go:359:1 - Present: Method subster.typOrNil (gojr/src/go/types/subst.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.var_` @ /usr/local/go1.27rc1/src/go/types/subst.go:366:1 - Present: Method subster.var_ (gojr/src/go/types/subst.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.tuple` @ /usr/local/go1.27rc1/src/go/types/subst.go:382:1 - Present: Method subster.tuple (gojr/src/go/types/subst.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=1, binary=2, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.func_` @ /usr/local/go1.27rc1/src/go/types/subst.go:409:1 - Present: Method subster.func_ (gojr/src/go/types/subst.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*subster.term` @ /usr/local/go1.27rc1/src/go/types/subst.go:425:1 - Present: Method subster.term (gojr/src/go/types/subst.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### termlist.go

#### Constants
- `termSep` - Present: Variable termSep (gojr/src/go/types/termlist.ts) - logic: declaration-only

#### Variables
- `allTermlist` - Present: Variable allTermlist (gojr/src/go/types/termlist.ts) - logic: declaration-only

#### Types
- `termlist` (*ast.ArrayType) - Present: ClassDeclaration termlist (gojr/src/go/types/termlist.ts) - structure: not yet 1:1-audited

#### Methods
- `termlist.String` @ /usr/local/go1.27rc1/src/go/types/termlist.go:27:1 - Present: Method termlist.String (gojr/src/go/types/termlist.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=5, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.isEmpty` @ /usr/local/go1.27rc1/src/go/types/termlist.go:42:1 - Present: Method termlist.isEmpty (gojr/src/go/types/termlist.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.isAll` @ /usr/local/go1.27rc1/src/go/types/termlist.go:55:1 - Present: Method termlist.isAll (gojr/src/go/types/termlist.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.norm` @ /usr/local/go1.27rc1/src/go/types/termlist.go:68:1 - Present: Method termlist.norm (gojr/src/go/types/termlist.ts) - control-flow shape: if=4, for=1, range=1, switch=0, typeSwitch=0, select=0, branch=2, assign=7, return=2, defer=0, go=0, call=5, binary=8, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.union` @ /usr/local/go1.27rc1/src/go/types/termlist.go:102:1 - Present: Method termlist.union (gojr/src/go/types/termlist.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.intersect` @ /usr/local/go1.27rc1/src/go/types/termlist.go:107:1 - Present: Method termlist.intersect (gojr/src/go/types/termlist.ts) - control-flow shape: if=2, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=5, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.equal` @ /usr/local/go1.27rc1/src/go/types/termlist.go:126:1 - Present: Method termlist.equal (gojr/src/go/types/termlist.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.includes` @ /usr/local/go1.27rc1/src/go/types/termlist.go:132:1 - Present: Method termlist.includes (gojr/src/go/types/termlist.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.supersetOf` @ /usr/local/go1.27rc1/src/go/types/termlist.go:142:1 - Present: Method termlist.supersetOf (gojr/src/go/types/termlist.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `termlist.subsetOf` @ /usr/local/go1.27rc1/src/go/types/termlist.go:152:1 - Present: Method termlist.subsetOf (gojr/src/go/types/termlist.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=3, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### trie.go

#### Types
- `trie` (*ast.MapType) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Methods
- `trie.insert` @ /usr/local/go1.27rc1/src/go/types/trie.go:24:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=1, range=1, switch=0, typeSwitch=2, select=0, branch=0, assign=9, return=5, defer=0, go=0, call=11, binary=6, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `trie.pickValue` @ /usr/local/go1.27rc1/src/go/types/trie.go:86:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### tuple.go

#### Types
- `Tuple` (struct) - Present: ClassDeclaration Tuple (gojr/src/go/types/tuple.ts) - structure: not yet 1:1-audited

#### Functions
- `NewTuple` @ /usr/local/go1.27rc1/src/go/types/tuple.go:18:1 - Present: FunctionDeclaration NewTuple (gojr/src/go/types/tuple.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Tuple.Len` @ /usr/local/go1.27rc1/src/go/types/tuple.go:26:1 - Present: Method Tuple.Len (gojr/src/go/types/tuple.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Tuple.At` @ /usr/local/go1.27rc1/src/go/types/tuple.go:34:1 - Present: Method Tuple.At (gojr/src/go/types/tuple.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Tuple.Underlying` @ /usr/local/go1.27rc1/src/go/types/tuple.go:36:1 - Present: Method Tuple.Underlying (gojr/src/go/types/tuple.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Tuple.String` @ /usr/local/go1.27rc1/src/go/types/tuple.go:37:1 - Present: Method Tuple.String (gojr/src/go/types/tuple.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### type.go

#### Types
- `Type` (interface) - Present: InterfaceDeclaration Type (gojr/src/go/types/type.ts) - structure: not yet 1:1-audited

### typelists.go

#### Types
- `TypeParamList` (struct) - Present: ClassDeclaration TypeParamList (gojr/src/go/types/typelists.ts) - structure: not yet 1:1-audited
- `TypeList` (struct) - Present: ClassDeclaration TypeList (gojr/src/go/types/typelists.ts) - structure: not yet 1:1-audited

#### Functions
- `newTypeList` @ /usr/local/go1.27rc1/src/go/types/typelists.go:49:1 - Present: FunctionDeclaration newTypeList (gojr/src/go/types/typelists.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=1, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `bindTParams` @ /usr/local/go1.27rc1/src/go/types/typelists.go:89:1 - Present: FunctionDeclaration bindTParams (gojr/src/go/types/typelists.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=2, binary=2, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*TypeParamList.Len` @ /usr/local/go1.27rc1/src/go/types/typelists.go:17:1 - Present: Method TypeParamList.Len (gojr/src/go/types/typelists.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParamList.At` @ /usr/local/go1.27rc1/src/go/types/typelists.go:20:1 - Present: Method TypeParamList.At (gojr/src/go/types/typelists.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParamList.list` @ /usr/local/go1.27rc1/src/go/types/typelists.go:25:1 - Present: Method TypeParamList.list (gojr/src/go/types/typelists.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParamList.String` @ /usr/local/go1.27rc1/src/go/types/typelists.go:32:1 - Present: Method TypeParamList.String (gojr/src/go/types/typelists.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=5, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeList.Len` @ /usr/local/go1.27rc1/src/go/types/typelists.go:58:1 - Present: Method TypeList.Len (gojr/src/go/types/typelists.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeList.At` @ /usr/local/go1.27rc1/src/go/types/typelists.go:61:1 - Present: Method TypeList.At (gojr/src/go/types/typelists.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeList.list` @ /usr/local/go1.27rc1/src/go/types/typelists.go:66:1 - Present: Method TypeList.list (gojr/src/go/types/typelists.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeList.String` @ /usr/local/go1.27rc1/src/go/types/typelists.go:73:1 - Present: Method TypeList.String (gojr/src/go/types/typelists.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=5, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### typeparam.go

#### Variables
- `lastID` - Present: Variable lastID (gojr/src/go/types/typeparam.ts) - logic: declaration-only

#### Types
- `TypeParam` (struct) - Present: ClassDeclaration TypeParam (gojr/src/go/types/typeparam.ts) - structure: not yet 1:1-audited

#### Functions
- `nextID` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:19:1 - Present: FunctionDeclaration nextID (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `NewTypeParam` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:39:1 - Present: FunctionDeclaration NewTypeParam (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.newTypeParam` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:44:1 - Present: Method Checker.newTypeParam (gojr/src/go/types/check.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=3, binary=4, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.Obj` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:66:1 - Present: Method TypeParam.Obj (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.Index` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:70:1 - Present: Method TypeParam.Index (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.Constraint` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:75:1 - Present: Method TypeParam.Constraint (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.SetConstraint` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:85:1 - Present: Method TypeParam.SetConstraint (gojr/src/go/types/typeparam.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.Underlying` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:99:1 - Present: Method TypeParam.Underlying (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.String` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:103:1 - Present: Method TypeParam.String (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.cleanup` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:108:1 - Present: Method TypeParam.cleanup (gojr/src/go/types/typeparam.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.iface` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:114:1 - Present: Method TypeParam.iface (gojr/src/go/types/typeparam.ts) - control-flow shape: if=5, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=9, return=3, defer=0, go=0, call=6, binary=3, unary=3, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.is` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:157:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*TypeParam.typeset` @ /usr/local/go1.27rc1/src/go/types/typeparam.go:165:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### typeset.go

#### Variables
- `topTypeSet` - Present: Variable topTypeSet (gojr/src/go/types/typeset.ts) - logic: declaration-only
- `invalidTypeSet` - Present: Variable invalidTypeSet (gojr/src/go/types/typeset.ts) - logic: declaration-only

#### Types
- `_TypeSet` (struct) - Present: ClassDeclaration _TypeSet (gojr/src/go/types/typeset.ts) - structure: not yet 1:1-audited

#### Functions
- `computeInterfaceTypeSet` @ /usr/local/go1.27rc1/src/go/types/typeset.go:155:1 - Present: FunctionDeclaration computeInterfaceTypeSet (gojr/src/go/types/typeset.ts) - control-flow shape: if=16, for=0, range=3, switch=1, typeSwitch=1, select=0, branch=5, assign=26, return=3, defer=1, go=0, call=54, binary=26, unary=13, composite=3, funcLiteral=3 - logic: not yet 1:1-audited
- `intersectTermLists` @ /usr/local/go1.27rc1/src/go/types/typeset.go:327:1 - Present: FunctionDeclaration intersectTermLists (gojr/src/go/types/typeset.ts) - control-flow shape: if=3, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=6, return=1, defer=0, go=0, call=7, binary=5, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `compareFunc` @ /usr/local/go1.27rc1/src/go/types/typeset.go:351:1 - Present: FunctionDeclaration compareFunc (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `sortMethods` @ /usr/local/go1.27rc1/src/go/types/typeset.go:355:1 - Present: FunctionDeclaration sortMethods (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `assertSortedMethods` @ /usr/local/go1.27rc1/src/go/types/typeset.go:359:1 - Present: FunctionDeclaration assertSortedMethods (gojr/src/go/types/typeset.ts) - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=0, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `computeUnionTypeSet` @ /usr/local/go1.27rc1/src/go/types/typeset.go:375:1 - Present: FunctionDeclaration computeUnionTypeSet (gojr/src/go/types/typeset.ts) - control-flow shape: if=6, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=10, return=3, defer=0, go=0, call=12, binary=5, unary=4, composite=1, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*_TypeSet.IsEmpty` @ /usr/local/go1.27rc1/src/go/types/typeset.go:36:1 - Present: Method _TypeSet.IsEmpty (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.IsAll` @ /usr/local/go1.27rc1/src/go/types/typeset.go:39:1 - Present: Method _TypeSet.IsAll (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.IsMethodSet` @ /usr/local/go1.27rc1/src/go/types/typeset.go:42:1 - Present: Method _TypeSet.IsMethodSet (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.IsComparable` @ /usr/local/go1.27rc1/src/go/types/typeset.go:45:1 - Present: Method _TypeSet.IsComparable (gojr/src/go/types/typeset.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=3, binary=3, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*_TypeSet.NumMethods` @ /usr/local/go1.27rc1/src/go/types/typeset.go:55:1 - Present: Method _TypeSet.NumMethods (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.Method` @ /usr/local/go1.27rc1/src/go/types/typeset.go:59:1 - Present: Method _TypeSet.Method (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.LookupMethod` @ /usr/local/go1.27rc1/src/go/types/typeset.go:62:1 - Present: Method _TypeSet.LookupMethod (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.String` @ /usr/local/go1.27rc1/src/go/types/typeset.go:66:1 - Present: Method _TypeSet.String (gojr/src/go/types/typeset.ts) - control-flow shape: if=5, for=0, range=1, switch=1, typeSwitch=0, select=0, branch=0, assign=2, return=3, defer=0, go=0, call=15, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.hasTerms` @ /usr/local/go1.27rc1/src/go/types/typeset.go:105:1 - Present: Method _TypeSet.hasTerms (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.subsetOf` @ /usr/local/go1.27rc1/src/go/types/typeset.go:108:1 - Present: Method _TypeSet.subsetOf (gojr/src/go/types/typeset.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.all` @ /usr/local/go1.27rc1/src/go/types/typeset.go:113:1 - Present: Method _TypeSet.all (gojr/src/go/types/typeset.ts) - control-flow shape: if=4, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=3, defer=0, go=0, call=9, binary=1, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*_TypeSet.is` @ /usr/local/go1.27rc1/src/go/types/typeset.go:138:1 - Present: Method _TypeSet.is (gojr/src/go/types/typeset.ts) - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=4, binary=1, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### typestring.go

#### Types
- `Qualifier` (func) - Present: TypeAliasDeclaration Qualifier (gojr/src/go/types/typestring.ts) - structure: not yet 1:1-audited
- `typeWriter` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `RelativeTo` @ /usr/local/go1.27rc1/src/go/types/typestring.go:35:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=4, defer=0, go=0, call=1, binary=2, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `TypeString` @ /usr/local/go1.27rc1/src/go/types/typestring.go:50:1 - Present: FunctionDeclaration TypeString (gojr/src/go/types/typestring.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `WriteType` @ /usr/local/go1.27rc1/src/go/types/typestring.go:59:1 - Present: FunctionDeclaration WriteType (gojr/src/go/types/typestring.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `WriteSignature` @ /usr/local/go1.27rc1/src/go/types/typestring.go:66:1 - Present: FunctionDeclaration WriteSignature (gojr/src/go/types/typestring.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `newTypeWriter` @ /usr/local/go1.27rc1/src/go/types/typestring.go:81:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `newTypeHasher` @ /usr/local/go1.27rc1/src/go/types/typestring.go:85:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `subscript` @ /usr/local/go1.27rc1/src/go/types/typestring.go:503:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=1, range=0, switch=0, typeSwitch=0, select=0, branch=1, assign=3, return=1, defer=0, go=0, call=5, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*typeWriter.byte` @ /usr/local/go1.27rc1/src/go/types/typestring.go:90:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=3, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.string` @ /usr/local/go1.27rc1/src/go/types/typestring.go:104:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.error` @ /usr/local/go1.27rc1/src/go/types/typestring.go:108:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.typ` @ /usr/local/go1.27rc1/src/go/types/typestring.go:115:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=32, for=0, range=4, switch=1, typeSwitch=1, select=0, branch=5, assign=17, return=1, defer=1, go=0, call=102, binary=35, unary=4, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.typeSet` @ /usr/local/go1.27rc1/src/go/types/typestring.go:353:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=3, for=0, range=2, switch=1, typeSwitch=0, select=0, branch=0, assign=3, return=0, defer=0, go=0, call=17, binary=1, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.typeList` @ /usr/local/go1.27rc1/src/go/types/typestring.go:388:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=4, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.tParamList` @ /usr/local/go1.27rc1/src/go/types/typestring.go:399:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=1, return=0, defer=0, go=0, call=9, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.typeName` @ /usr/local/go1.27rc1/src/go/types/typestring.go:428:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.tuple` @ /usr/local/go1.27rc1/src/go/types/typestring.go:433:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=5, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=11, binary=9, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*typeWriter.signature` @ /usr/local/go1.27rc1/src/go/types/typestring.go:471:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=2, defer=1, go=0, call=13, binary=9, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

### typeterm.go

#### Types
- `term` (struct) - Present: ClassDeclaration term (gojr/src/go/types/typeterm.ts) - structure: not yet 1:1-audited

#### Methods
- `*term.String` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:21:1 - Present: Method term.String (gojr/src/go/types/typeterm.ts) - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=4, defer=0, go=0, call=2, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.equal` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:35:1 - Present: Method term.equal (gojr/src/go/types/typeterm.ts) - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=3, defer=0, go=0, call=1, binary=10, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.union` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:49:1 - Present: Method term.union (gojr/src/go/types/typeterm.ts) - control-flow shape: if=2, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=8, defer=0, go=0, call=1, binary=8, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.intersect` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:81:1 - Present: Method term.intersect (gojr/src/go/types/typeterm.ts) - control-flow shape: if=2, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=6, defer=0, go=0, call=1, binary=6, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.includes` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:109:1 - Present: Method term.includes (gojr/src/go/types/typeterm.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=2, return=3, defer=0, go=0, call=2, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.subsetOf` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:127:1 - Present: Method term.subsetOf (gojr/src/go/types/typeterm.ts) - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=6, defer=0, go=0, call=1, binary=5, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*term.disjoint` @ /usr/local/go1.27rc1/src/go/types/typeterm.go:155:1 - Present: Method term.disjoint (gojr/src/go/types/typeterm.ts) - control-flow shape: if=3, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=4, binary=4, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### typexpr.go

#### Functions
- `goTypeName` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:215:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.ident` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:20:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=12, for=0, range=0, switch=1, typeSwitch=1, select=0, branch=0, assign=21, return=7, defer=0, go=0, call=25, binary=16, unary=4, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.typ` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:139:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.varType` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:146:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.validVarType` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:154:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=9, binary=1, unary=1, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `*Checker.declaredType` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:181:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=6, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.genericType` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:199:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=7, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.typInternal` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:221:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=9, for=0, range=0, switch=3, typeSwitch=1, select=0, branch=0, assign=27, return=14, defer=1, go=0, call=54, binary=5, unary=13, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*Checker.instantiatedType` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:391:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=9, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=11, return=4, defer=1, go=0, call=35, binary=6, unary=2, composite=0, funcLiteral=2 - logic: not yet 1:1-audited
- `*Checker.arrayLength` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:462:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=10, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=7, return=5, defer=0, go=0, call=18, binary=7, unary=10, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.typeList` @ /usr/local/go1.27rc1/src/go/types/typexpr.go:510:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=4, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### under.go

#### Variables
- `emptyTypeError` - Present: Variable emptyTypeError (gojr/src/go/types/under.ts) - logic: declaration-only

#### Types
- `typeError` (struct) - Present: ClassDeclaration typeError (gojr/src/go/types/under.ts) - structure: not yet 1:1-audited

#### Functions
- `underIs` @ /usr/local/go1.27rc1/src/go/types/under.go:14:1 - Present: FunctionDeclaration underIs (gojr/src/go/types/under.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `all` @ /usr/local/go1.27rc1/src/go/types/under.go:22:1 - Present: FunctionDeclaration all (gojr/src/go/types/under.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=4, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `typeset` @ /usr/local/go1.27rc1/src/go/types/under.go:35:1 - Present: FunctionDeclaration typeset (gojr/src/go/types/under.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=1 - logic: not yet 1:1-audited
- `typeErrorf` @ /usr/local/go1.27rc1/src/go/types/under.go:49:1 - Present: FunctionDeclaration typeErrorf (gojr/src/go/types/under.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=0, binary=1, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `commonUnder` @ /usr/local/go1.27rc1/src/go/types/under.go:79:1 - Present: FunctionDeclaration commonUnder (gojr/src/go/types/under.ts) - control-flow shape: if=8, for=0, range=1, switch=1, typeSwitch=0, select=0, branch=2, assign=5, return=6, defer=0, go=0, call=8, binary=9, unary=2, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*typeError.format` @ /usr/local/go1.27rc1/src/go/types/under.go:58:1 - Present: Method typeError.format (gojr/src/go/types/under.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### unify.go

#### Constants
- `unificationDepthLimit` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `panicAtUnificationDepthLimit` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `enableCoreTypeUnification` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `traceInference` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `assign` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only
- `exact` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Types
- `unifier` (struct) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `unifyMode` (*ast.Ident) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited
- `typeParamsById` (*ast.ArrayType) - Missing from current checker same-kind exact-symbol inventory - structure: not yet 1:1-audited

#### Functions
- `newUnifier` @ /usr/local/go1.27rc1/src/go/types/unify.go:92:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=3, return=1, defer=0, go=0, call=6, binary=2, unary=2, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `asInterface` @ /usr/local/go1.27rc1/src/go/types/unify.go:276:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `unifyMode.String` @ /usr/local/go1.27rc1/src/go/types/unify.go:127:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=0, return=5, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.unify` @ /usr/local/go1.27rc1/src/go/types/unify.go:144:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.tracef` @ /usr/local/go1.27rc1/src/go/types/unify.go:148:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=3, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.String` @ /usr/local/go1.27rc1/src/go/types/unify.go:155:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=12, binary=1, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `typeParamsById.Len` @ /usr/local/go1.27rc1/src/go/types/unify.go:182:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `typeParamsById.Less` @ /usr/local/go1.27rc1/src/go/types/unify.go:183:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `typeParamsById.Swap` @ /usr/local/go1.27rc1/src/go/types/unify.go:184:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.join` @ /usr/local/go1.27rc1/src/go/types/unify.go:189:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=1, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=3, binary=5, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.asBoundTypeParam` @ /usr/local/go1.27rc1/src/go/types/unify.go:215:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=2, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.setHandle` @ /usr/local/go1.27rc1/src/go/types/unify.go:226:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=0, defer=0, go=0, call=1, binary=2, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.at` @ /usr/local/go1.27rc1/src/go/types/unify.go:237:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.set` @ /usr/local/go1.27rc1/src/go/types/unify.go:243:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=0, defer=0, go=0, call=2, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.unknowns` @ /usr/local/go1.27rc1/src/go/types/unify.go:252:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.inferred` @ /usr/local/go1.27rc1/src/go/types/unify.go:266:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*unifier.nify` @ /usr/local/go1.27rc1/src/go/types/unify.go:287:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=62, for=2, range=7, switch=2, typeSwitch=1, select=0, branch=0, assign=56, return=40, defer=1, go=0, call=111, binary=95, unary=15, composite=2, funcLiteral=1 - logic: not yet 1:1-audited

### union.go

#### Constants
- `maxTermCount` - Present: Variable maxTermCount (gojr/src/go/types/union.ts) - logic: declaration-only

#### Types
- `Union` (struct) - Present: ClassDeclaration Union (gojr/src/go/types/union.ts) - structure: not yet 1:1-audited
- `Term` (*ast.Ident) - Present: ClassDeclaration Term (gojr/src/go/types/union.ts) - structure: not yet 1:1-audited

#### Functions
- `NewUnion` @ /usr/local/go1.27rc1/src/go/types/union.go:23:1 - Present: FunctionDeclaration NewUnion (gojr/src/go/types/union.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=1, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `NewTerm` @ /usr/local/go1.27rc1/src/go/types/union.go:40:1 - Present: FunctionDeclaration NewTerm (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=1, composite=1, funcLiteral=0 - logic: not yet 1:1-audited
- `parseUnion` @ /usr/local/go1.27rc1/src/go/types/union.go:54:1 - Present: FunctionDeclaration parseUnion (gojr/src/go/types/union.ts) - control-flow shape: if=11, for=0, range=2, switch=1, typeSwitch=0, select=0, branch=4, assign=9, return=3, defer=0, go=0, call=27, binary=12, unary=5, composite=1, funcLiteral=1 - logic: not yet 1:1-audited
- `parseTilde` @ /usr/local/go1.27rc1/src/go/types/union.go:139:1 - Present: FunctionDeclaration parseTilde (gojr/src/go/types/union.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=7, return=1, defer=0, go=0, call=6, binary=3, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `overlappingTerm` @ /usr/local/go1.27rc1/src/go/types/union.go:171:1 - Present: FunctionDeclaration overlappingTerm (gojr/src/go/types/union.ts) - control-flow shape: if=4, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=0, return=2, defer=0, go=0, call=7, binary=7, unary=3, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `flattenUnion` @ /usr/local/go1.27rc1/src/go/types/union.go:193:1 - Present: FunctionDeclaration flattenUnion (gojr/src/go/types/union.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=4, return=1, defer=0, go=0, call=3, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Union.Len` @ /usr/local/go1.27rc1/src/go/types/union.go:30:1 - Present: Method Union.Len (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Union.Term` @ /usr/local/go1.27rc1/src/go/types/union.go:31:1 - Present: Method Union.Term (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Union.Underlying` @ /usr/local/go1.27rc1/src/go/types/union.go:33:1 - Present: Method Union.Underlying (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Union.String` @ /usr/local/go1.27rc1/src/go/types/union.go:34:1 - Present: Method Union.String (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Term.Tilde` @ /usr/local/go1.27rc1/src/go/types/union.go:42:1 - Present: Method Term.Tilde (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Term.Type` @ /usr/local/go1.27rc1/src/go/types/union.go:43:1 - Present: Method Term.Type (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Term.String` @ /usr/local/go1.27rc1/src/go/types/union.go:44:1 - Present: Method Term.String (gojr/src/go/types/union.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### universe.go

#### Constants
- `_Append` - Present: EnumMember _Append in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Cap` - Present: EnumMember _Cap in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Clear` - Present: EnumMember _Clear in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Close` - Present: EnumMember _Close in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Complex` - Present: EnumMember _Complex in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Copy` - Present: EnumMember _Copy in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Delete` - Present: EnumMember _Delete in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Imag` - Present: EnumMember _Imag in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Len` - Present: EnumMember _Len in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Make` - Present: EnumMember _Make in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Max` - Present: EnumMember _Max in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Min` - Present: EnumMember _Min in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_New` - Present: EnumMember _New in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Panic` - Present: EnumMember _Panic in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Print` - Present: EnumMember _Print in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Println` - Present: EnumMember _Println in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Real` - Present: EnumMember _Real in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Recover` - Present: EnumMember _Recover in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Add` - Present: EnumMember _Add in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Alignof` - Present: EnumMember _Alignof in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Offsetof` - Present: EnumMember _Offsetof in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Sizeof` - Present: EnumMember _Sizeof in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Slice` - Present: EnumMember _Slice in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_SliceData` - Present: EnumMember _SliceData in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_String` - Present: EnumMember _String in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_StringData` - Present: EnumMember _StringData in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Assert` - Present: EnumMember _Assert in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only
- `_Trace` - Present: EnumMember _Trace in builtinId (gojr/src/go/types/universe.ts) - logic: declaration-only

#### Variables
- `Universe` - Present: Variable Universe (gojr/src/go/types/universe.ts) - logic: declaration-only
- `Unsafe` - Present: Variable Unsafe (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeIota` - Present: Variable universeIota (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeBool` - Present: Variable universeBool (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeByte` - Present: Variable universeByte (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeRune` - Present: Variable universeRune (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeError` - Present: Variable universeError (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeAny` - Present: Variable universeAny (gojr/src/go/types/universe.ts) - logic: declaration-only
- `universeComparable` - Present: Variable universeComparable (gojr/src/go/types/universe.ts) - logic: declaration-only
- `Typ` - Present: Variable Typ (gojr/src/go/types/universe.ts) - logic: declaration-only
- `basicAliases` - Present: Variable basicAliases (gojr/src/go/types/universe.ts) - logic: declaration-only
- `predeclaredConsts` - Present: Variable predeclaredConsts (gojr/src/go/types/universe.ts) - logic: declaration-only
- `predeclaredFuncs` - Present: Variable predeclaredFuncs (gojr/src/go/types/universe.ts) - logic: declaration-only

#### Types
- `builtinId` (*ast.Ident) - Present: EnumDeclaration builtinId (gojr/src/go/types/universe.ts) - structure: not yet 1:1-audited

#### Functions
- `defPredeclaredTypes` @ /usr/local/go1.27rc1/src/go/types/universe.go:77:1 - Present: FunctionDeclaration defPredeclaredTypes (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=2, switch=0, typeSwitch=0, select=0, branch=0, assign=10, return=0, defer=0, go=0, call=20, binary=0, unary=4, composite=4, funcLiteral=0 - logic: not yet 1:1-audited
- `defPredeclaredConsts` @ /usr/local/go1.27rc1/src/go/types/universe.go:131:1 - Present: FunctionDeclaration defPredeclaredConsts (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `defPredeclaredNil` @ /usr/local/go1.27rc1/src/go/types/universe.go:137:1 - Present: FunctionDeclaration defPredeclaredNil (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=1, composite=2, funcLiteral=0 - logic: not yet 1:1-audited
- `defPredeclaredFuncs` @ /usr/local/go1.27rc1/src/go/types/universe.go:219:1 - Present: FunctionDeclaration defPredeclaredFuncs (gojr/src/go/types/universe.ts) - control-flow shape: if=1, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=1, assign=1, return=0, defer=0, go=0, call=3, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `DefPredeclaredTestFuncs` @ /usr/local/go1.27rc1/src/go/types/universe.go:232:1 - Present: FunctionDeclaration DefPredeclaredTestFuncs (gojr/src/go/types/universe.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=5, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `init` @ /usr/local/go1.27rc1/src/go/types/universe.go:240:1 - Present: FunctionDeclaration init (gojr/src/go/types/universe.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=10, return=0, defer=0, go=0, call=17, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `def` @ /usr/local/go1.27rc1/src/go/types/universe.go:262:1 - Present: FunctionDeclaration def (gojr/src/go/types/universe.ts) - control-flow shape: if=4, for=0, range=0, switch=0, typeSwitch=1, select=0, branch=0, assign=8, return=1, defer=0, go=0, call=10, binary=3, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### util.go

#### Constants
- `isTypes2` - Missing from current checker same-kind exact-symbol inventory - logic: declaration-only

#### Functions
- `cmpPos` @ /usr/local/go1.27rc1/src/go/types/util.go:28:1 - Present: FunctionDeclaration cmpPos (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `hasDots` @ /usr/local/go1.27rc1/src/go/types/util.go:31:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `dddErrPos` @ /usr/local/go1.27rc1/src/go/types/util.go:34:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `isdddArray` @ /usr/local/go1.27rc1/src/go/types/util.go:37:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=2, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=1, return=2, defer=0, go=0, call=0, binary=4, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `argErrPos` @ /usr/local/go1.27rc1/src/go/types/util.go:47:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `startPos` @ /usr/local/go1.27rc1/src/go/types/util.go:50:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `endPos` @ /usr/local/go1.27rc1/src/go/types/util.go:53:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `makeFromLiteral` @ /usr/local/go1.27rc1/src/go/types/util.go:56:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

### validtype.go

#### Functions
- `makeObjList` @ /usr/local/go1.27rc1/src/go/types/validtype.go:190:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=1, switch=0, typeSwitch=0, select=0, branch=0, assign=2, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `*Checker.validType` @ /usr/local/go1.27rc1/src/go/types/validtype.go:16:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=0, defer=0, go=0, call=1, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.validType0` @ /usr/local/go1.27rc1/src/go/types/validtype.go:30:1 - Missing from current checker same-kind exact-symbol inventory - control-flow shape: if=11, for=0, range=6, switch=0, typeSwitch=1, select=0, branch=0, assign=10, return=8, defer=1, go=0, call=36, binary=10, unary=4, composite=0, funcLiteral=1 - logic: not yet 1:1-audited

### version.go

#### Variables
- `go1_9` - Present: Variable go1_9 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_13` - Present: Variable go1_13 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_14` - Present: Variable go1_14 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_17` - Present: Variable go1_17 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_18` - Present: Variable go1_18 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_20` - Present: Variable go1_20 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_21` - Present: Variable go1_21 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_22` - Present: Variable go1_22 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_23` - Present: Variable go1_23 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_26` - Present: Variable go1_26 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go1_27` - Present: Variable go1_27 (gojr/src/go/types/version.ts) - logic: declaration-only
- `go_current` - Present: Variable go_current (gojr/src/go/types/version.ts) - logic: declaration-only

#### Types
- `goVersion` (*ast.Ident) - Present: ClassDeclaration goVersion (gojr/src/go/types/version.ts) - structure: not yet 1:1-audited

#### Functions
- `asGoVersion` @ /usr/local/go1.27rc1/src/go/types/version.go:23:1 - Present: FunctionDeclaration asGoVersion (gojr/src/go/types/version.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited

#### Methods
- `goVersion.isValid` @ /usr/local/go1.27rc1/src/go/types/version.go:28:1 - Present: Method goVersion.isValid (gojr/src/go/types/version.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=0, binary=1, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `goVersion.cmp` @ /usr/local/go1.27rc1/src/go/types/version.go:34:1 - Present: Method goVersion.cmp (gojr/src/go/types/version.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=3, binary=0, unary=0, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.allowVersion` @ /usr/local/go1.27rc1/src/go/types/version.go:59:1 - Present: Method Checker.allowVersion (gojr/src/go/types/check.ts) - control-flow shape: if=0, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=1, defer=0, go=0, call=2, binary=2, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
- `*Checker.verifyVersionf` @ /usr/local/go1.27rc1/src/go/types/version.go:65:1 - Present: Method Checker.verifyVersionf (gojr/src/go/types/check.ts) - control-flow shape: if=1, for=0, range=0, switch=0, typeSwitch=0, select=0, branch=0, assign=0, return=2, defer=0, go=0, call=2, binary=0, unary=1, composite=0, funcLiteral=0 - logic: not yet 1:1-audited
