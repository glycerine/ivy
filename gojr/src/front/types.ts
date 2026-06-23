export enum TypeKind {
  Basic = "Basic",
  Array = "Array",
  Slice = "Slice",
  Struct = "Struct",
  Pointer = "Pointer",
  Tuple = "Tuple",
  Signature = "Signature",
  Interface = "Interface",
  Map = "Map",
  Named = "Named"
}

export enum BasicKind {
  Invalid = "invalid",
  Bool = "bool",
  Int = "int",
  Int8 = "int8",
  Int16 = "int16",
  Int32 = "int32",
  Int64 = "int64",
  Uint = "uint",
  Uint8 = "uint8",
  Uint16 = "uint16",
  Uint32 = "uint32",
  Uint64 = "uint64",
  Uintptr = "uintptr",
  Byte = "byte",
  Rune = "rune",
  Float32 = "float32",
  Float64 = "float64",
  Complex64 = "complex64",
  Complex128 = "complex128",
  String = "string",
  Error = "error",
  Any = "any",
  UntypedBool = "untyped bool",
  UntypedInt = "untyped int",
  UntypedFloat = "untyped float",
  UntypedComplex = "untyped complex",
  UntypedString = "untyped string",
  UntypedNil = "untyped nil"
}

export interface Type {
  readonly kind: TypeKind;
  underlying(): Type;
  typeString(): string;
}

export class BasicType implements Type {
  public readonly kind = TypeKind.Basic;

  public constructor(
    public readonly basicKind: BasicKind,
    public readonly info: ReadonlySet<BasicInfo> = new Set()
  ) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return this.basicKind;
  }
}

export type BasicInfo = "bool" | "integer" | "float" | "complex" | "string" | "numeric" | "untyped" | "nil";

export class ArrayType implements Type {
  public readonly kind = TypeKind.Array;

  public constructor(
    public readonly length: number,
    public readonly element: Type
  ) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return `[${this.length}]${this.element.typeString()}`;
  }
}

export class SliceType implements Type {
  public readonly kind = TypeKind.Slice;

  public constructor(public readonly element: Type) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return `[]${this.element.typeString()}`;
  }
}

export interface StructField {
  name: string;
  type: Type;
  embedded: boolean;
  tag?: string;
}

export class StructType implements Type {
  public readonly kind = TypeKind.Struct;

  public constructor(public readonly fields: StructField[]) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    const fields = this.fields
      .map((field) => `${field.name} ${field.type.typeString()}`)
      .join("; ");
    return `struct{${fields}}`;
  }
}

export class PointerType implements Type {
  public readonly kind = TypeKind.Pointer;

  public constructor(public readonly base: Type) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return `*${this.base.typeString()}`;
  }
}

export class TupleType implements Type {
  public readonly kind = TypeKind.Tuple;

  public constructor(public readonly variables: VarObject[]) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return this.variables.map((variable) => variable.type.typeString()).join(", ");
  }
}

export class SignatureType implements Type {
  public readonly kind = TypeKind.Signature;

  public constructor(
    public readonly receiver: VarObject | undefined,
    public readonly params: TupleType,
    public readonly results: TupleType,
    public readonly variadic = false
  ) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    const params = this.params.variables.map((param, index) => {
      const prefix = param.name ? `${param.name} ` : "";
      const isVariadic = this.variadic && index === this.params.variables.length - 1;
      const type = isVariadic && param.type instanceof SliceType
        ? `...${param.type.element.typeString()}`
        : param.type.typeString();
      return `${prefix}${type}`;
    }).join(", ");
    const results = this.results.variables.map((result) => {
      const prefix = result.name ? `${result.name} ` : "";
      return `${prefix}${result.type.typeString()}`;
    });
    if (results.length === 0) return `func(${params})`;
    if (results.length === 1 && !this.results.variables[0]?.name) return `func(${params}) ${results[0]}`;
    return `func(${params}) (${results.join(", ")})`;
  }
}

export class InterfaceType implements Type {
  public readonly kind = TypeKind.Interface;
  private completed = false;

  public constructor(public readonly methods: FuncObject[] = []) {}

  public complete(): this {
    this.methods.sort((left, right) => left.name.localeCompare(right.name));
    this.completed = true;
    return this;
  }

  public isComplete(): boolean {
    return this.completed;
  }

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    if (this.methods.length === 0) return "interface{}";
    return `interface{${this.methods.map((method) => method.name + method.type.typeString().replace(/^func/, "")).join("; ")}}`;
  }
}

export class MapType implements Type {
  public readonly kind = TypeKind.Map;

  public constructor(
    public readonly key: Type,
    public readonly value: Type
  ) {}

  public underlying(): Type {
    return this;
  }

  public typeString(): string {
    return `map[${this.key.typeString()}]${this.value.typeString()}`;
  }
}

export class NamedType implements Type {
  public readonly kind = TypeKind.Named;
  private readonly methods: FuncObject[] = [];

  public constructor(
    public readonly object: TypeNameObject,
    private currentUnderlying: Type
  ) {}

  public underlying(): Type {
    return this.currentUnderlying.underlying();
  }

  public setUnderlying(type: Type): void {
    this.currentUnderlying = type;
  }

  public addMethod(method: FuncObject): void {
    this.methods.push(method);
    this.methods.sort((left, right) => left.name.localeCompare(right.name));
  }

  public methodSet(): FuncObject[] {
    return [...this.methods];
  }

  public typeString(): string {
    const pkg = this.object.pkg ? `${this.object.pkg.name}.` : "";
    return `${pkg}${this.object.name}`;
  }
}

export enum ObjectKind {
  Const = "Const",
  TypeName = "TypeName",
  Var = "Var",
  Func = "Func",
  PackageName = "PackageName",
  Builtin = "Builtin",
  Nil = "Nil"
}

export interface TypeObject {
  readonly kind: ObjectKind;
  readonly name: string;
  type: Type;
  readonly parent: Scope | undefined;
  readonly pkg: PackageInfo | undefined;
  exported(): boolean;
}

abstract class ObjectBase implements TypeObject {
  public constructor(
    public readonly kind: ObjectKind,
    public readonly name: string,
    public type: Type,
    public readonly parent: Scope | undefined = undefined,
    public readonly pkg: PackageInfo | undefined = undefined
  ) {}

  public exported(): boolean {
    return /^[A-Z]/.test(this.name);
  }
}

export class ConstObject extends ObjectBase {
  public constructor(name: string, type: Type, public readonly value: unknown, parent?: Scope, pkg?: PackageInfo) {
    super(ObjectKind.Const, name, type, parent, pkg);
  }
}

export class TypeNameObject extends ObjectBase {
  public constructor(name: string, type: Type, parent?: Scope, pkg?: PackageInfo) {
    super(ObjectKind.TypeName, name, type, parent, pkg);
  }

  public setType(type: Type): void {
    this.type = type;
  }
}

export class VarObject extends ObjectBase {
  public constructor(name: string, type: Type, public readonly embedded = false, parent?: Scope, pkg?: PackageInfo) {
    super(ObjectKind.Var, name, type, parent, pkg);
  }
}

export class FuncObject extends ObjectBase {
  public constructor(name: string, public readonly signature: SignatureType, parent?: Scope, pkg?: PackageInfo) {
    super(ObjectKind.Func, name, signature, parent, pkg);
  }
}

export class PackageNameObject extends ObjectBase {
  public constructor(name: string, type: Type, public readonly imported: PackageInfo, parent?: Scope, pkg?: PackageInfo) {
    super(ObjectKind.PackageName, name, type, parent, pkg);
  }
}

export class BuiltinObject extends ObjectBase {
  public constructor(name: string, type: Type, parent?: Scope) {
    super(ObjectKind.Builtin, name, type, parent);
  }
}

export class NilObject extends ObjectBase {
  public constructor(type: Type, parent?: Scope) {
    super(ObjectKind.Nil, "nil", type, parent);
  }

  public override exported(): boolean {
    return false;
  }
}

export class Scope {
  private readonly objects = new Map<string, TypeObject>();

  public constructor(
    public readonly parent?: Scope,
    public readonly comment = ""
  ) {}

  public insert(object: TypeObject): TypeObject | undefined {
    const existing = this.objects.get(object.name);
    if (existing) return existing;
    this.objects.set(object.name, object);
    return undefined;
  }

  public lookup(name: string): TypeObject | undefined {
    return this.objects.get(name);
  }

  public lookupParent(name: string): { scope: Scope; object: TypeObject } | undefined {
    const object = this.lookup(name);
    if (object) return { scope: this, object };
    return this.parent?.lookupParent(name);
  }

  public names(): string[] {
    return [...this.objects.keys()].sort();
  }

  public children(): TypeObject[] {
    return this.names().flatMap((name) => {
      const object = this.objects.get(name);
      return object ? [object] : [];
    });
  }
}

export class PackageInfo {
  public readonly scope: Scope;

  public constructor(
    public readonly path: string,
    public readonly name: string,
    parent?: Scope
  ) {
    this.scope = new Scope(parent, "package");
  }
}

export interface Universe {
  scope: Scope;
  basic: {
    invalid: BasicType;
    bool: BasicType;
    int64: BasicType;
    float64: BasicType;
    complex128: BasicType;
    string: BasicType;
    error: BasicType;
    any: BasicType;
    untypedBool: BasicType;
    untypedInt: BasicType;
    untypedFloat: BasicType;
    untypedComplex: BasicType;
    untypedString: BasicType;
    untypedNil: BasicType;
  };
}

export function newUniverse(): Universe {
  const scope = new Scope(undefined, "universe");
  const basic = {
    invalid: new BasicType(BasicKind.Invalid),
    bool: new BasicType(BasicKind.Bool, new Set(["bool"])),
    int64: new BasicType(BasicKind.Int64, new Set(["integer", "numeric"])),
    float64: new BasicType(BasicKind.Float64, new Set(["float", "numeric"])),
    complex128: new BasicType(BasicKind.Complex128, new Set(["complex", "numeric"])),
    string: new BasicType(BasicKind.String, new Set(["string"])),
    error: new BasicType(BasicKind.Error),
    any: new BasicType(BasicKind.Any),
    untypedBool: new BasicType(BasicKind.UntypedBool, new Set(["bool", "untyped"])),
    untypedInt: new BasicType(BasicKind.UntypedInt, new Set(["integer", "numeric", "untyped"])),
    untypedFloat: new BasicType(BasicKind.UntypedFloat, new Set(["float", "numeric", "untyped"])),
    untypedComplex: new BasicType(BasicKind.UntypedComplex, new Set(["complex", "numeric", "untyped"])),
    untypedString: new BasicType(BasicKind.UntypedString, new Set(["string", "untyped"])),
    untypedNil: new BasicType(BasicKind.UntypedNil, new Set(["nil", "untyped"]))
  };

  const integerInfo = new Set<BasicInfo>(["integer", "numeric"]);
  const floatInfo = new Set<BasicInfo>(["float", "numeric"]);
  const complexInfo = new Set<BasicInfo>(["complex", "numeric"]);
  for (const [name, type] of [
    ["bool", basic.bool],
    ["int", new BasicType(BasicKind.Int, integerInfo)],
    ["int8", new BasicType(BasicKind.Int8, integerInfo)],
    ["int16", new BasicType(BasicKind.Int16, integerInfo)],
    ["int32", new BasicType(BasicKind.Int32, integerInfo)],
    ["int64", basic.int64],
    ["uint", new BasicType(BasicKind.Uint, integerInfo)],
    ["uint8", new BasicType(BasicKind.Uint8, integerInfo)],
    ["uint16", new BasicType(BasicKind.Uint16, integerInfo)],
    ["uint32", new BasicType(BasicKind.Uint32, integerInfo)],
    ["uint64", new BasicType(BasicKind.Uint64, integerInfo)],
    ["uintptr", new BasicType(BasicKind.Uintptr, integerInfo)],
    ["byte", new BasicType(BasicKind.Byte, integerInfo)],
    ["rune", new BasicType(BasicKind.Rune, integerInfo)],
    ["float32", new BasicType(BasicKind.Float32, floatInfo)],
    ["float64", basic.float64],
    ["complex64", new BasicType(BasicKind.Complex64, complexInfo)],
    ["complex128", basic.complex128],
    ["string", basic.string],
    ["error", basic.error],
    ["any", basic.any]
  ] as const) {
    scope.insert(new TypeNameObject(name, type, scope));
  }

  scope.insert(new ConstObject("true", basic.untypedBool, true, scope));
  scope.insert(new ConstObject("false", basic.untypedBool, false, scope));
  scope.insert(new NilObject(basic.untypedNil, scope));

  const empty = new TupleType([]);
  scope.insert(new BuiltinObject("panic", new SignatureType(undefined, tuple(varOf("", basic.any)), empty, false), scope));
  scope.insert(new BuiltinObject("panicOn", new SignatureType(undefined, tuple(varOf("err", basic.error)), empty, false), scope));
  scope.insert(new BuiltinObject("len", new SignatureType(undefined, tuple(varOf("", basic.any)), tuple(varOf("", basic.int64)), false), scope));
  scope.insert(new BuiltinObject("cap", new SignatureType(undefined, tuple(varOf("", basic.any)), tuple(varOf("", basic.int64)), false), scope));
  scope.insert(new BuiltinObject("append", new SignatureType(
    undefined,
    tuple(varOf("slice", new SliceType(basic.any)), varOf("values", new SliceType(basic.any))),
    tuple(varOf("", new SliceType(basic.any))),
    true
  ), scope));
  scope.insert(new BuiltinObject("make", new SignatureType(
    undefined,
    tuple(varOf("type", basic.any), varOf("size", new SliceType(basic.any))),
    tuple(varOf("", basic.any)),
    true
  ), scope));
  for (const name of ["new", "delete", "clear", "copy", "min", "max", "complex", "real", "imag", "print", "println"]) {
    scope.insert(new BuiltinObject(name, new SignatureType(undefined, tuple(varOf("args", new SliceType(basic.any))), tuple(varOf("", basic.any)), true), scope));
  }

  return { scope, basic };
}

export function tuple(...variables: VarObject[]): TupleType {
  return new TupleType(variables);
}

export function varOf(name: string, type: Type): VarObject {
  return new VarObject(name, type);
}

export function sameType(left: Type, right: Type): boolean {
  if (left === right) return true;
  if (left instanceof NamedType || right instanceof NamedType) {
    return left instanceof NamedType && right instanceof NamedType && left.object === right.object;
  }
  if (left.kind !== right.kind) return false;
  if (left instanceof BasicType && right instanceof BasicType) return left.basicKind === right.basicKind;
  if (left instanceof ArrayType && right instanceof ArrayType) {
    return left.length === right.length && sameType(left.element, right.element);
  }
  if (left instanceof SliceType && right instanceof SliceType) return sameType(left.element, right.element);
  if (left instanceof PointerType && right instanceof PointerType) return sameType(left.base, right.base);
  if (left instanceof MapType && right instanceof MapType) return sameType(left.key, right.key) && sameType(left.value, right.value);
  if (left instanceof StructType && right instanceof StructType) return sameStruct(left, right);
  if (left instanceof SignatureType && right instanceof SignatureType) return sameSignature(left, right);
  if (left instanceof InterfaceType && right instanceof InterfaceType) return sameInterface(left, right);
  return false;
}

export function assignableTo(source: Type, target: Type): boolean {
  if (sameType(source, target)) return true;
  if (target instanceof BasicType && target.basicKind === BasicKind.Any) return true;
  if (target instanceof InterfaceType) return implementsInterface(source, target);
  if (source instanceof BasicType && source.basicKind === BasicKind.UntypedNil) return isNilAssignable(target);
  if (source instanceof BasicType && target instanceof BasicType) {
    if (source.info.has("integer") && target.info.has("integer")) return true;
    if (source.info.has("untyped") && target.info.has("numeric") && source.info.has("numeric")) return true;
    if (source.info.has("untyped") && target.info.has("string") && source.info.has("string")) return true;
    if (source.info.has("untyped") && target.info.has("bool") && source.info.has("bool")) return true;
  }
  const sourceUnderlying = source.underlying();
  const targetUnderlying = target.underlying();
  if (source !== sourceUnderlying || target !== targetUnderlying) {
    return sameType(sourceUnderlying, targetUnderlying);
  }
  return false;
}

export function implementsInterface(type: Type, iface: InterfaceType): boolean {
  if (iface.methods.length === 0) return true;
  const methods = methodSet(type);
  return iface.methods.every((required) => {
    const found = methods.find((method) => method.name === required.name);
    return Boolean(found && sameSignature(found.signature, required.signature));
  });
}

export function methodSet(type: Type): FuncObject[] {
  if (type instanceof NamedType) {
    return type.methodSet().filter((method) => {
      const receiver = method.signature.receiver?.type;
      return !(receiver instanceof PointerType);
    });
  }
  if (type instanceof PointerType && type.base instanceof NamedType) return type.base.methodSet();
  return [];
}

export function isNilAssignable(type: Type): boolean {
  const actual = type.underlying();
  return actual instanceof PointerType ||
    actual instanceof SliceType ||
    actual instanceof MapType ||
    actual instanceof SignatureType ||
    actual instanceof InterfaceType ||
    (actual instanceof BasicType && actual.basicKind === BasicKind.Any);
}

function sameStruct(left: StructType, right: StructType): boolean {
  return left.fields.length === right.fields.length &&
    left.fields.every((field, index) => {
      const other = right.fields[index];
      if (!other) return false;
      return field.name === other.name &&
        field.embedded === other.embedded &&
        field.tag === other.tag &&
        sameType(field.type, other.type);
    });
}

function sameSignature(left: SignatureType, right: SignatureType): boolean {
  return left.variadic === right.variadic &&
    sameTuple(left.params, right.params) &&
    sameTuple(left.results, right.results);
}

function sameTuple(left: TupleType, right: TupleType): boolean {
  return left.variables.length === right.variables.length &&
    left.variables.every((variable, index) => {
      const other = right.variables[index];
      return other ? sameType(variable.type, other.type) : false;
    });
}

function sameInterface(left: InterfaceType, right: InterfaceType): boolean {
  return left.methods.length === right.methods.length &&
    left.methods.every((method, index) => {
      const other = right.methods[index];
      return other ? method.name === other.name && sameSignature(method.signature, other.signature) : false;
    });
}
