export var TypeKind;
(function (TypeKind) {
    TypeKind["Basic"] = "Basic";
    TypeKind["Array"] = "Array";
    TypeKind["Slice"] = "Slice";
    TypeKind["Struct"] = "Struct";
    TypeKind["Pointer"] = "Pointer";
    TypeKind["Tuple"] = "Tuple";
    TypeKind["Signature"] = "Signature";
    TypeKind["Interface"] = "Interface";
    TypeKind["Map"] = "Map";
    TypeKind["Named"] = "Named";
})(TypeKind || (TypeKind = {}));
export var BasicKind;
(function (BasicKind) {
    BasicKind["Invalid"] = "invalid";
    BasicKind["Bool"] = "bool";
    BasicKind["Int"] = "int";
    BasicKind["Int8"] = "int8";
    BasicKind["Int16"] = "int16";
    BasicKind["Int32"] = "int32";
    BasicKind["Int64"] = "int64";
    BasicKind["Uint"] = "uint";
    BasicKind["Uint8"] = "uint8";
    BasicKind["Uint16"] = "uint16";
    BasicKind["Uint32"] = "uint32";
    BasicKind["Uint64"] = "uint64";
    BasicKind["Uintptr"] = "uintptr";
    BasicKind["Byte"] = "byte";
    BasicKind["Rune"] = "rune";
    BasicKind["Float32"] = "float32";
    BasicKind["Float64"] = "float64";
    BasicKind["Complex64"] = "complex64";
    BasicKind["Complex128"] = "complex128";
    BasicKind["String"] = "string";
    BasicKind["Error"] = "error";
    BasicKind["Any"] = "any";
    BasicKind["UntypedBool"] = "untyped bool";
    BasicKind["UntypedInt"] = "untyped int";
    BasicKind["UntypedFloat"] = "untyped float";
    BasicKind["UntypedComplex"] = "untyped complex";
    BasicKind["UntypedString"] = "untyped string";
    BasicKind["UntypedNil"] = "untyped nil";
})(BasicKind || (BasicKind = {}));
export class BasicType {
    basicKind;
    info;
    kind = TypeKind.Basic;
    constructor(basicKind, info = new Set()) {
        this.basicKind = basicKind;
        this.info = info;
    }
    underlying() {
        return this;
    }
    typeString() {
        return this.basicKind;
    }
}
export class ArrayType {
    length;
    element;
    kind = TypeKind.Array;
    constructor(length, element) {
        this.length = length;
        this.element = element;
    }
    underlying() {
        return this;
    }
    typeString() {
        return `[${this.length}]${this.element.typeString()}`;
    }
}
export class SliceType {
    element;
    kind = TypeKind.Slice;
    constructor(element) {
        this.element = element;
    }
    underlying() {
        return this;
    }
    typeString() {
        return `[]${this.element.typeString()}`;
    }
}
export class StructType {
    fields;
    kind = TypeKind.Struct;
    constructor(fields) {
        this.fields = fields;
    }
    underlying() {
        return this;
    }
    typeString() {
        const fields = this.fields
            .map((field) => `${field.name} ${field.type.typeString()}`)
            .join("; ");
        return `struct{${fields}}`;
    }
}
export class PointerType {
    base;
    kind = TypeKind.Pointer;
    constructor(base) {
        this.base = base;
    }
    underlying() {
        return this;
    }
    typeString() {
        return `*${this.base.typeString()}`;
    }
}
export class TupleType {
    variables;
    kind = TypeKind.Tuple;
    constructor(variables) {
        this.variables = variables;
    }
    underlying() {
        return this;
    }
    typeString() {
        return this.variables.map((variable) => variable.type.typeString()).join(", ");
    }
}
export class SignatureType {
    receiver;
    params;
    results;
    variadic;
    kind = TypeKind.Signature;
    constructor(receiver, params, results, variadic = false) {
        this.receiver = receiver;
        this.params = params;
        this.results = results;
        this.variadic = variadic;
    }
    underlying() {
        return this;
    }
    typeString() {
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
        if (results.length === 0)
            return `func(${params})`;
        if (results.length === 1 && !this.results.variables[0]?.name)
            return `func(${params}) ${results[0]}`;
        return `func(${params}) (${results.join(", ")})`;
    }
}
export class InterfaceType {
    methods;
    kind = TypeKind.Interface;
    completed = false;
    constructor(methods = []) {
        this.methods = methods;
    }
    complete() {
        this.methods.sort((left, right) => left.name.localeCompare(right.name));
        this.completed = true;
        return this;
    }
    isComplete() {
        return this.completed;
    }
    underlying() {
        return this;
    }
    typeString() {
        if (this.methods.length === 0)
            return "interface{}";
        return `interface{${this.methods.map((method) => method.name + method.type.typeString().replace(/^func/, "")).join("; ")}}`;
    }
}
export class MapType {
    key;
    value;
    kind = TypeKind.Map;
    constructor(key, value) {
        this.key = key;
        this.value = value;
    }
    underlying() {
        return this;
    }
    typeString() {
        return `map[${this.key.typeString()}]${this.value.typeString()}`;
    }
}
export class NamedType {
    object;
    currentUnderlying;
    kind = TypeKind.Named;
    methods = [];
    constructor(object, currentUnderlying) {
        this.object = object;
        this.currentUnderlying = currentUnderlying;
    }
    underlying() {
        return this.currentUnderlying.underlying();
    }
    setUnderlying(type) {
        this.currentUnderlying = type;
    }
    addMethod(method) {
        this.methods.push(method);
        this.methods.sort((left, right) => left.name.localeCompare(right.name));
    }
    methodSet() {
        return [...this.methods];
    }
    typeString() {
        const pkg = this.object.pkg ? `${this.object.pkg.name}.` : "";
        return `${pkg}${this.object.name}`;
    }
}
export var ObjectKind;
(function (ObjectKind) {
    ObjectKind["Const"] = "Const";
    ObjectKind["TypeName"] = "TypeName";
    ObjectKind["Var"] = "Var";
    ObjectKind["Func"] = "Func";
    ObjectKind["PackageName"] = "PackageName";
    ObjectKind["Builtin"] = "Builtin";
    ObjectKind["Nil"] = "Nil";
})(ObjectKind || (ObjectKind = {}));
class ObjectBase {
    kind;
    name;
    type;
    parent;
    pkg;
    constructor(kind, name, type, parent = undefined, pkg = undefined) {
        this.kind = kind;
        this.name = name;
        this.type = type;
        this.parent = parent;
        this.pkg = pkg;
    }
    exported() {
        return /^[A-Z]/.test(this.name);
    }
}
export class ConstObject extends ObjectBase {
    value;
    constructor(name, type, value, parent, pkg) {
        super(ObjectKind.Const, name, type, parent, pkg);
        this.value = value;
    }
}
export class TypeNameObject extends ObjectBase {
    constructor(name, type, parent, pkg) {
        super(ObjectKind.TypeName, name, type, parent, pkg);
    }
    setType(type) {
        this.type = type;
    }
}
export class VarObject extends ObjectBase {
    embedded;
    constructor(name, type, embedded = false, parent, pkg) {
        super(ObjectKind.Var, name, type, parent, pkg);
        this.embedded = embedded;
    }
}
export class FuncObject extends ObjectBase {
    signature;
    constructor(name, signature, parent, pkg) {
        super(ObjectKind.Func, name, signature, parent, pkg);
        this.signature = signature;
    }
}
export class PackageNameObject extends ObjectBase {
    imported;
    constructor(name, type, imported, parent, pkg) {
        super(ObjectKind.PackageName, name, type, parent, pkg);
        this.imported = imported;
    }
}
export class BuiltinObject extends ObjectBase {
    constructor(name, type, parent) {
        super(ObjectKind.Builtin, name, type, parent);
    }
}
export class NilObject extends ObjectBase {
    constructor(type, parent) {
        super(ObjectKind.Nil, "nil", type, parent);
    }
    exported() {
        return false;
    }
}
export class Scope {
    parent;
    comment;
    objects = new Map();
    constructor(parent, comment = "") {
        this.parent = parent;
        this.comment = comment;
    }
    insert(object) {
        const existing = this.objects.get(object.name);
        if (existing)
            return existing;
        this.objects.set(object.name, object);
        return undefined;
    }
    lookup(name) {
        return this.objects.get(name);
    }
    lookupParent(name) {
        const object = this.lookup(name);
        if (object)
            return { scope: this, object };
        return this.parent?.lookupParent(name);
    }
    names() {
        return [...this.objects.keys()].sort();
    }
    children() {
        return this.names().flatMap((name) => {
            const object = this.objects.get(name);
            return object ? [object] : [];
        });
    }
}
export class PackageInfo {
    path;
    name;
    scope;
    constructor(path, name, parent) {
        this.path = path;
        this.name = name;
        this.scope = new Scope(parent, "package");
    }
}
export function newUniverse() {
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
    const integerInfo = new Set(["integer", "numeric"]);
    const floatInfo = new Set(["float", "numeric"]);
    const complexInfo = new Set(["complex", "numeric"]);
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
    ]) {
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
    scope.insert(new BuiltinObject("append", new SignatureType(undefined, tuple(varOf("slice", new SliceType(basic.any)), varOf("values", new SliceType(basic.any))), tuple(varOf("", new SliceType(basic.any))), true), scope));
    scope.insert(new BuiltinObject("make", new SignatureType(undefined, tuple(varOf("type", basic.any), varOf("size", new SliceType(basic.any))), tuple(varOf("", basic.any)), true), scope));
    for (const name of ["new", "delete", "clear", "copy", "min", "max", "complex", "real", "imag", "print", "println"]) {
        scope.insert(new BuiltinObject(name, new SignatureType(undefined, tuple(varOf("args", new SliceType(basic.any))), tuple(varOf("", basic.any)), true), scope));
    }
    return { scope, basic };
}
export function tuple(...variables) {
    return new TupleType(variables);
}
export function varOf(name, type) {
    return new VarObject(name, type);
}
export function sameType(left, right) {
    if (left === right)
        return true;
    if (left instanceof NamedType || right instanceof NamedType) {
        return left instanceof NamedType && right instanceof NamedType && left.object === right.object;
    }
    if (left.kind !== right.kind)
        return false;
    if (left instanceof BasicType && right instanceof BasicType)
        return left.basicKind === right.basicKind;
    if (left instanceof ArrayType && right instanceof ArrayType) {
        return left.length === right.length && sameType(left.element, right.element);
    }
    if (left instanceof SliceType && right instanceof SliceType)
        return sameType(left.element, right.element);
    if (left instanceof PointerType && right instanceof PointerType)
        return sameType(left.base, right.base);
    if (left instanceof MapType && right instanceof MapType)
        return sameType(left.key, right.key) && sameType(left.value, right.value);
    if (left instanceof StructType && right instanceof StructType)
        return sameStruct(left, right);
    if (left instanceof SignatureType && right instanceof SignatureType)
        return sameSignature(left, right);
    if (left instanceof InterfaceType && right instanceof InterfaceType)
        return sameInterface(left, right);
    return false;
}
export function assignableTo(source, target) {
    if (sameType(source, target))
        return true;
    if (target instanceof BasicType && target.basicKind === BasicKind.Any)
        return true;
    if (target instanceof InterfaceType)
        return implementsInterface(source, target);
    if (source instanceof BasicType && source.basicKind === BasicKind.UntypedNil)
        return isNilAssignable(target);
    if (source instanceof BasicType && target instanceof BasicType) {
        if (source.info.has("integer") && target.info.has("integer"))
            return true;
        if (source.info.has("untyped") && target.info.has("numeric") && source.info.has("numeric"))
            return true;
        if (source.info.has("untyped") && target.info.has("string") && source.info.has("string"))
            return true;
        if (source.info.has("untyped") && target.info.has("bool") && source.info.has("bool"))
            return true;
    }
    const sourceUnderlying = source.underlying();
    const targetUnderlying = target.underlying();
    if (source !== sourceUnderlying || target !== targetUnderlying) {
        return sameType(sourceUnderlying, targetUnderlying);
    }
    return false;
}
export function implementsInterface(type, iface) {
    if (iface.methods.length === 0)
        return true;
    const methods = methodSet(type);
    return iface.methods.every((required) => {
        const found = methods.find((method) => method.name === required.name);
        return Boolean(found && sameSignature(found.signature, required.signature));
    });
}
export function methodSet(type) {
    if (type instanceof NamedType) {
        return type.methodSet().filter((method) => {
            const receiver = method.signature.receiver?.type;
            return !(receiver instanceof PointerType);
        });
    }
    if (type instanceof PointerType && type.base instanceof NamedType)
        return type.base.methodSet();
    return [];
}
export function isNilAssignable(type) {
    const actual = type.underlying();
    return actual instanceof PointerType ||
        actual instanceof SliceType ||
        actual instanceof MapType ||
        actual instanceof SignatureType ||
        actual instanceof InterfaceType ||
        (actual instanceof BasicType && actual.basicKind === BasicKind.Any);
}
function sameStruct(left, right) {
    return left.fields.length === right.fields.length &&
        left.fields.every((field, index) => {
            const other = right.fields[index];
            if (!other)
                return false;
            return field.name === other.name &&
                field.embedded === other.embedded &&
                field.tag === other.tag &&
                sameType(field.type, other.type);
        });
}
function sameSignature(left, right) {
    return left.variadic === right.variadic &&
        sameTuple(left.params, right.params) &&
        sameTuple(left.results, right.results);
}
function sameTuple(left, right) {
    return left.variables.length === right.variables.length &&
        left.variables.every((variable, index) => {
            const other = right.variables[index];
            return other ? sameType(variable.type, other.type) : false;
        });
}
function sameInterface(left, right) {
    return left.methods.length === right.methods.length &&
        left.methods.every((method, index) => {
            const other = right.methods[index];
            return other ? method.name === other.name && sameSignature(method.signature, other.signature) : false;
        });
}
