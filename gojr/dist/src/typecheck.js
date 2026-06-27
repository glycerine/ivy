import { REPL_FILENAME, diagnosticFilename, withDiagnosticSourceContext } from "./diagnostics.js";
import { parseFrontSource, parseFrontSourceFiles } from "./front/parser.js";
import { TokenKind } from "./front/token.js";
import { syncIntrinsicSource } from "./stubPackages.js";
import { Config, Info, Int, Int8, Int16, Int32, Int64, UntypedInt, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr, Float32, Float64, Bool, NewArray, NewChecker, NewConst, NewField, NewFunc, NewInterfaceType, NewNamed, NewPackage, NewPointer, NewPkgName, NewSignatureType, NewSlice, NewStruct, NewTerm, NewTypeName, NewTypeParam, NewTuple, NewUnion, NewVar, NoPos, String as GoString, Typ, emptyInterface, ensureUniverseInitialized, UniverseLookup, Unsafe } from "./go/types/index.js";
import { isIntrinsicPackageImport } from "./intrinsicPackages.js";
export const GOJR_SYNTHETIC_CHECK_PREFIX = "__gojr_check_statements";
const standardTypePackageCache = new Map();
export function isGoJuniorSyntheticCheckName(name) {
    return name === GOJR_SYNTHETIC_CHECK_PREFIX || name.startsWith(`${GOJR_SYNTHETIC_CHECK_PREFIX}_`);
}
export function checkGoJuniorSource(source, filename, config = {}) {
    const sourceFile = { filename, source };
    const parsed = parseFrontSource(source, filename);
    const result = checkGoJuniorFiles(parsed.file ? [parsed.file] : [], parsed.statements, parsed.diagnostics, config, [sourceFile]);
    return {
        ...result,
        ...(parsed.file ? { file: parsed.file } : {})
    };
}
export function checkGoJuniorSourceFiles(sourceFiles, config = {}) {
    const parsed = parseFrontSourceFiles(sourceFiles);
    return checkGoJuniorFiles(parsed.files, parsed.statements, parsed.diagnostics, config, sourceFiles);
}
export function checkGoJuniorFiles(files, statements = [], parserDiagnostics = [], config = {}, sourceFiles = []) {
    ensureUniverseInitialized();
    const inferredPackageName = config.packageName ?? files.find((file) => file.name)?.name?.name ?? "main";
    const packagePath = config.packageInstance?.Path() ?? config.packagePath ?? inferredPackageName;
    const existingCodebasePackage = config.codebaseTxn?.PackageInfo(packagePath);
    const packageName = config.packageInstance?.Name() ?? existingCodebasePackage?.Name() ?? inferredPackageName;
    const diagnostics = [...parserDiagnostics];
    const fset = new FrontFileSet(files, sourceFiles);
    const info = new Info();
    info.Types = new Map();
    info.Defs = new Map();
    info.Uses = new Map();
    info.Scopes = new Map();
    info.Selections = new Map();
    const checkFiles = filesForChecking(files, statements, packageName, config.syntheticFunctionName);
    const pkg = config.packageInstance ?? packageForChecking(config, packagePath, packageName);
    seedPackageScope(pkg, config);
    const conf = new Config();
    conf.Importer = {
        Import(path) {
            const intrinsic = isIntrinsicPackageImport(path) ? standardTypePackage(path) : undefined;
            const imported = intrinsic ?? config.importer?.import(path) ?? config.codebaseTxn?.PackageInfo(path) ?? standardTypePackage(path);
            if (imported === undefined)
                return [null, new Error(`package ${path} is not available`)];
            return [imported, null];
        }
    };
    conf.DisableUnusedImportCheck = config.disableUnusedImportCheck ?? true;
    conf.GoJuniorSheetNamespaces = config.sheetNamespaces ?? null;
    conf.Error = (err) => diagnostics.push(diagnosticFromGoTypesError(err, fset));
    try {
        const err = NewChecker(conf, fset, pkg, info).Files(checkFiles);
        if (err !== null && err !== undefined && diagnostics.length === parserDiagnostics.length) {
            diagnostics.push(diagnosticFromGoTypesError(err, fset));
        }
    }
    catch (error) {
        diagnostics.push({
            filename: diagnosticFilename(undefined, firstFilename(files)),
            code: "GOJR_TYPE001",
            severity: "error",
            message: error instanceof Error ? error.message : String(error),
            ...(error instanceof Error && error.stack ? { stack: error.stack } : {})
        });
    }
    filterTopLevelExpressionConvenienceDiagnostics(diagnostics, statements, config);
    const resultDiagnostics = sourceFiles.length > 0
        ? withDiagnosticSourceContext(diagnostics, sourceFiles)
        : diagnostics;
    if (config.rollbackOnErrors !== false) {
        config.codebaseTxn?.Codebase().RollbackOnErrors(config.codebaseTxn, resultDiagnostics);
    }
    return {
        pkg,
        info,
        diagnostics: resultDiagnostics,
        files,
        statements
    };
}
function packageForChecking(config, packagePath, packageName) {
    if (config.codebaseTxn?.Kind() === "update") {
        return config.codebaseTxn.Codebase().BeginPackageUpdate(config.codebaseTxn, packagePath, packageName);
    }
    return NewPackage(packagePath, packageName);
}
function filterTopLevelExpressionConvenienceDiagnostics(diagnostics, statements, config) {
    const expressionOffsets = new Set(statements
        .filter((statement) => statement.kind === "ExprStmt" && statement.span !== undefined)
        .map((statement) => statement.span.offset));
    if (expressionOffsets.size === 0)
        return;
    const packageInspectionOffsets = new Set((config.allowPackageInspection ? statements : [])
        .filter((statement) => statement.kind === "ExprStmt" && statement.expr.kind === "Ident" && statement.span !== undefined)
        .map((statement) => statement.span.offset));
    const kept = diagnostics.filter((diagnostic) => !(diagnostic.code === "GOJR_TYPE001" &&
        / is not used$/.test(diagnostic.message) &&
        diagnostic.span !== undefined &&
        expressionOffsets.has(diagnostic.span.offset)) &&
        !(diagnostic.code === "GOJR_TYPE001" &&
            /^use of package .+ not in selector$/.test(diagnostic.message) &&
            diagnostic.span !== undefined &&
            packageInspectionOffsets.has(diagnostic.span.offset)));
    diagnostics.splice(0, diagnostics.length, ...kept);
}
class FrontFileSet {
    files;
    sourceFiles;
    constructor(files, sourceFiles = []) {
        this.files = files;
        this.sourceFiles = sourceFiles;
    }
    Position(pos) {
        const span = this.span(pos);
        return `${span.filename}:${span.line}:${span.column}`;
    }
    span(pos) {
        const offset = positionOffset(pos);
        const file = fileForOffset(this.files, offset);
        const filename = file?.span?.filename ?? firstFilename(this.files);
        const source = this.sourceFiles.find((candidate) => candidate.filename === filename)?.source;
        return spanFromOffset(filename, offset, source);
    }
}
function filesForChecking(files, statements, packageName, syntheticFunctionName = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const hasPackageClause = files.some((file) => file.name);
    if (statements.length === 0) {
        if (hasPackageClause)
            return files;
        return [syntheticPackageFile(files, statements, packageName, true)];
    }
    const promoted = promoteTopLevelShortDeclarations(statements);
    const synthetic = syntheticPackageFile(files, promoted.statements, packageName, !hasPackageClause, promoted.declarations, syntheticFunctionName);
    return hasPackageClause ? [...files, synthetic] : [synthetic];
}
function syntheticPackageFile(files, statements, packageName, includeDeclarations, extraDeclarations = [], syntheticFunctionName = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const first = files[0];
    const name = first?.name ?? ident(packageName, first?.span);
    const start = statements[0]?.span ?? first?.span;
    const declarations = includeDeclarations ? files.flatMap((file) => file.declarations) : importDeclarations(files);
    declarations.push(...extraDeclarations);
    if (statements.length > 0) {
        declarations.push(syntheticFunctionDecl(statements, start, syntheticFunctionName));
    }
    return {
        kind: "File",
        name,
        declarations,
        imports: declarations.flatMap((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import
            ? decl.specs.filter((spec) => spec.kind === "ImportSpec")
            : []),
        unresolved: [],
        comments: [],
        ...(start ? { span: start } : {})
    };
}
function promoteTopLevelShortDeclarations(statements) {
    const declarations = [];
    const remaining = [];
    for (const statement of statements) {
        if (statement.kind === "AssignStmt" && statement.token === TokenKind.Define && allIdentifiers(statement.lhs)) {
            declarations.push(shortDeclarationAsVar(statement, statement.lhs));
            continue;
        }
        remaining.push(statement);
    }
    return { declarations, statements: remaining };
}
function allIdentifiers(exprs) {
    return exprs.every((expr) => expr.kind === "Ident");
}
function shortDeclarationAsVar(statement, names) {
    return {
        kind: "GenDecl",
        token: TokenKind.Var,
        grouped: false,
        specs: [{
                kind: "ValueSpec",
                names,
                values: statement.rhs,
                ...(statement.span ? { span: statement.span } : {})
            }],
        ...(statement.span ? { span: statement.span } : {})
    };
}
function syntheticFunctionDecl(statements, span, name = GOJR_SYNTHETIC_CHECK_PREFIX) {
    const resultCount = maxReturnValueCount(statements);
    const type = {
        kind: "FuncType",
        params: { kind: "FieldList", fields: [], ...(span ? { span } : {}) },
        ...(resultCount > 0 ? { results: syntheticAnyResults(resultCount, span) } : {}),
        ...(span ? { span } : {})
    };
    return {
        kind: "FuncDecl",
        name: ident(name, span),
        type,
        body: {
            kind: "BlockStmt",
            statements: resultCount > 0 ? [...statements, syntheticNilReturn(resultCount, span)] : statements,
            ...(span ? { span } : {})
        },
        ...(span ? { span } : {})
    };
}
function syntheticNilReturn(count, span) {
    return {
        kind: "ReturnStmt",
        results: globalThis.Array.from({ length: count }, () => ident("nil", span)),
        ...(span ? { span } : {})
    };
}
function syntheticAnyResults(count, span) {
    return {
        kind: "FieldList",
        fields: globalThis.Array.from({ length: count }, () => ({
            kind: "Field",
            names: [],
            type: ident("any", span),
            ...(span ? { span } : {})
        })),
        ...(span ? { span } : {})
    };
}
function maxReturnValueCount(statements) {
    let max = 0;
    for (const statement of statements) {
        max = Math.max(max, returnValueCount(statement));
    }
    return max;
}
function returnValueCount(statement) {
    switch (statement.kind) {
        case "ReturnStmt":
            return statement.results.length;
        case "BlockStmt":
            return maxReturnValueCount(statement.statements);
        case "LabeledStmt":
            return returnValueCount(statement.stmt);
        case "IfStmt":
            return Math.max(statement.body ? returnValueCount(statement.body) : 0, statement.else ? returnValueCount(statement.else) : 0);
        case "SwitchStmt":
        case "TypeSwitchStmt":
            return Math.max(0, ...statement.body.map((clause) => maxReturnValueCount(clause.body)));
        case "SelectStmt":
            return Math.max(0, ...statement.body.map((clause) => maxReturnValueCount(clause.body)));
        case "ForStmt":
        case "RangeStmt":
            return returnValueCount(statement.body);
        default:
            return 0;
    }
}
function importDeclarations(files) {
    return files.flatMap((file) => file.declarations.filter((decl) => decl.kind === "GenDecl" && decl.token === TokenKind.Import));
}
function seedPackageScope(pkg, config) {
    if (config.autoImportFmt ?? true) {
        pkg.Scope().Insert(NewPkgName(NoPos, pkg, "fmt", standardTypePackage("fmt") ?? fmtPackage()));
    }
    for (const object of config.predeclaredPackageObjects ?? []) {
        if (object.Pkg() !== pkg)
            continue;
        if (pkg.Scope().Lookup(object.Name()) === null)
            pkg.Scope().Insert(object);
    }
}
export function standardTypePackage(path) {
    ensureUniverseInitialized();
    const cached = standardTypePackageCache.get(path);
    if (cached)
        return cached;
    let pkg;
    if (path === "cmp")
        pkg = cmpPackage();
    else if (path === "fmt")
        pkg = fmtPackage();
    else if (path === "internal/reflectlite")
        pkg = reflectlitePackage();
    else if (path === "iter")
        pkg = iterPackage();
    else if (path === "io/fs")
        pkg = ioFsPackage();
    else if (path === "math")
        pkg = mathPackage();
    else if (path === "os")
        pkg = osPackage();
    else if (path === "runtime")
        pkg = runtimePackage();
    else if (path === "sync")
        pkg = syncPackage();
    else if (path === "sync/atomic")
        pkg = syncAtomicPackage();
    else if (path === "syscall/js")
        pkg = syscallJSPackage();
    else if (path === "testing")
        pkg = testingPackage();
    else if (path === "time")
        pkg = timePackage();
    else if (path === "unsafe")
        pkg = Unsafe;
    else if (path === "weak")
        pkg = weakPackage();
    if (pkg)
        standardTypePackageCache.set(path, pkg);
    return pkg;
}
function weakPackage() {
    return checkedSyntheticStandardPackage("weak", `package weak
type Pointer[T any] struct{}
func Make[T any](ptr *T) Pointer[T] { return Pointer[T]{} }
func (p Pointer[T]) Value() *T { return nil }
`);
}
function ioFsPackage() {
    return checkedSyntheticStandardPackage("io/fs", `package fs

import "time"

type FileMode uint32

func (m FileMode) IsDir() bool { return false }
func (m FileMode) IsRegular() bool { return false }
func (m FileMode) Perm() FileMode { return m }
func (m FileMode) String() string { return "" }
func (m FileMode) Type() FileMode { return m }

const (
  ModeDir FileMode = 2147483648
  ModeAppend FileMode = 1073741824
  ModeExclusive FileMode = 536870912
  ModeTemporary FileMode = 268435456
  ModeSymlink FileMode = 134217728
  ModeDevice FileMode = 67108864
  ModeNamedPipe FileMode = 33554432
  ModeSocket FileMode = 16777216
  ModeSetuid FileMode = 8388608
  ModeSetgid FileMode = 4194304
  ModeCharDevice FileMode = 2097152
  ModeSticky FileMode = 1048576
  ModeIrregular FileMode = 524288
  ModeType FileMode = 2399666176
  ModePerm FileMode = 511
)

var ErrInvalid error
var ErrPermission error
var ErrExist error
var ErrNotExist error
var ErrClosed error
var SkipDir error
var SkipAll error

type FileInfo interface {
  Name() string
  Size() int64
  Mode() FileMode
  ModTime() time.Time
  IsDir() bool
  Sys() any
}

type DirEntry interface {
  Name() string
  IsDir() bool
  Type() FileMode
  Info() (FileInfo, error)
}

type FS interface {
  Open(name string) (File, error)
}

type File interface {
  Stat() (FileInfo, error)
  Read([]byte) (int, error)
  Close() error
}

type ReadDirFile interface {
  File
  ReadDir(n int) ([]DirEntry, error)
}

type ReadFileFS interface {
  FS
  ReadFile(name string) ([]byte, error)
}

type ReadDirFS interface {
  FS
  ReadDir(name string) ([]DirEntry, error)
}

type StatFS interface {
  FS
  Stat(name string) (FileInfo, error)
}

type GlobFS interface {
  FS
  Glob(pattern string) ([]string, error)
}

type SubFS interface {
  FS
  Sub(dir string) (FS, error)
}

type PathError struct {
  Op string
  Path string
  Err error
}

func (e *PathError) Error() string { return "" }
func (e *PathError) Unwrap() error { return e.Err }

type WalkDirFunc func(path string, d DirEntry, err error) error

func FormatDirEntry(dir DirEntry) string { return "" }
func FormatFileInfo(info FileInfo) string { return "" }
func Glob(fsys FS, pattern string) ([]string, error) { return nil, nil }
func ReadDir(fsys FS, name string) ([]DirEntry, error) { return nil, nil }
func ReadFile(fsys FS, name string) ([]byte, error) { return nil, nil }
func Stat(fsys FS, name string) (FileInfo, error) { return nil, nil }
func Sub(fsys FS, dir string) (FS, error) { return nil, nil }
func ValidPath(name string) bool { return true }
func WalkDir(fsys FS, root string, fn WalkDirFunc) error { return nil }
func FileInfoToDirEntry(info FileInfo) DirEntry { return nil }
`);
}
function checkedSyntheticStandardPackage(path, source) {
    const parsed = parseFrontSourceFiles([{ filename: `gojr:intrinsic/${path}`, source }]);
    const packageName = parsed.files.find((file) => file.name)?.name?.name ?? path.split("/").filter(Boolean).at(-1) ?? path;
    const result = checkGoJuniorFiles(parsed.files, parsed.statements, parsed.diagnostics, {
        packageName,
        packagePath: path,
        autoImportFmt: false,
        disableUnusedImportCheck: true
    }, [{ filename: `gojr:intrinsic/${path}`, source }]);
    const errors = result.diagnostics.filter((diagnostic) => diagnostic.severity === "error");
    if (errors.length > 0) {
        throw new Error(`could not initialize intrinsic ${path}: ${errors.map((diagnostic) => diagnostic.message).join("; ")}`);
    }
    return result.pkg;
}
function timePackage() {
    return checkedSyntheticStandardPackage("time", `package time

type Time struct{}
type Duration int64
type Month int
type Weekday int
type Location struct{}
type Timer struct{}
type Ticker struct{}

const (
  Nanosecond Duration = 1
  Microsecond Duration = 1000
  Millisecond Duration = 1000000
  Second Duration = 1000000000
  Minute Duration = 60000000000
  Hour Duration = 3600000000000
)

var UTC *Location
var Local *Location

func Now() Time { return Time{} }
func Unix(sec int64, nsec int64) Time { return Time{} }
func Since(t Time) Duration { return 0 }
func Until(t Time) Duration { return 0 }
func Sleep(d Duration) {}
func Parse(layout, value string) (Time, error) { return Time{}, nil }
func ParseInLocation(layout, value string, loc *Location) (Time, error) { return Time{}, nil }
func LoadLocation(name string) (*Location, error) { return nil, nil }
func FixedZone(name string, offset int) *Location { return nil }

func (t Time) Add(d Duration) Time { return t }
func (t Time) AddDate(years int, months int, days int) Time { return t }
func (t Time) After(u Time) bool { return false }
func (t Time) Before(u Time) bool { return false }
func (t Time) Equal(u Time) bool { return true }
func (t Time) IsZero() bool { return true }
func (t Time) String() string { return "" }
func (t Time) Format(layout string) string { return "" }
func (t Time) Sub(u Time) Duration { return 0 }
func (t Time) Unix() int64 { return 0 }
func (t Time) UnixNano() int64 { return 0 }
func (t Time) Location() *Location { return nil }
func (t Time) In(loc *Location) Time { return t }
func (t Time) Local() Time { return t }
func (t Time) UTC() Time { return t }
func (t Time) Date() (year int, month Month, day int) { return 0, 0, 0 }
func (t Time) Clock() (hour int, min int, sec int) { return 0, 0, 0 }
func (t Time) Year() int { return 0 }
func (t Time) Month() Month { return 0 }
func (t Time) Day() int { return 0 }
func (t Time) Hour() int { return 0 }
func (t Time) Minute() int { return 0 }
func (t Time) Second() int { return 0 }
func (t Time) Nanosecond() int { return 0 }
func (t Time) Weekday() Weekday { return 0 }
func (t Time) MarshalText() ([]byte, error) { return nil, nil }
func (t *Time) UnmarshalText(data []byte) error { return nil }

func (d Duration) String() string { return "" }
func (d Duration) Nanoseconds() int64 { return int64(d) }
func (d Duration) Microseconds() int64 { return int64(d) / 1000 }
func (d Duration) Milliseconds() int64 { return int64(d) / 1000000 }
func (d Duration) Seconds() float64 { return 0 }
func (d Duration) Minutes() float64 { return 0 }
func (d Duration) Hours() float64 { return 0 }
`);
}
function osPackage() {
    return checkedSyntheticStandardPackage("os", `package os

import "io/fs"

type File struct{}
type FileMode = fs.FileMode
type FileInfo = fs.FileInfo
type DirEntry = fs.DirEntry

const (
  O_RDONLY int = 0
  O_WRONLY int = 1
  O_RDWR int = 2
  O_APPEND int = 8
  O_CREATE int = 512
  O_EXCL int = 2048
  O_SYNC int = 128
  O_TRUNC int = 1024

  ModeDir FileMode = 2147483648
  ModeAppend FileMode = 1073741824
  ModeExclusive FileMode = 536870912
  ModeTemporary FileMode = 268435456
  ModeSymlink FileMode = 134217728
  ModeDevice FileMode = 67108864
  ModeNamedPipe FileMode = 33554432
  ModeSocket FileMode = 16777216
  ModeSetuid FileMode = 8388608
  ModeSetgid FileMode = 4194304
  ModeCharDevice FileMode = 2097152
  ModeSticky FileMode = 1048576
  ModeIrregular FileMode = 524288
  ModeType FileMode = 2399666176
  ModePerm FileMode = 511

  PathSeparator = '/'
  PathListSeparator = ':'
  DevNull = "/dev/null"
)

var Args []string
var Stdin *File
var Stdout *File
var Stderr *File
var ErrInvalid error
var ErrPermission error
var ErrExist error
var ErrNotExist error
var ErrClosed error
var ErrDeadlineExceeded error
var ErrNoDeadline error
var ErrProcessDone error

type PathError struct {
  Op string
  Path string
  Err error
}

func (e *PathError) Error() string { return "" }
func (e *PathError) Unwrap() error { return e.Err }

type LinkError struct {
  Op string
  Old string
  New string
  Err error
}

func (e *LinkError) Error() string { return "" }
func (e *LinkError) Unwrap() error { return e.Err }

type SyscallError struct {
  Syscall string
  Err error
}

func (e *SyscallError) Error() string { return "" }
func (e *SyscallError) Unwrap() error { return e.Err }
func NewSyscallError(syscall string, err error) error { return err }

type ProcAttr struct {
  Dir string
  Env []string
  Files []*File
  Sys any
}

type Process struct{}
type ProcessState struct{}
func (p *Process) Kill() error { return nil }
func (p *Process) Release() error { return nil }
func (p *Process) Signal(sig Signal) error { return nil }
func (p *Process) Wait() (*ProcessState, error) { return nil, nil }
func (p *ProcessState) Exited() bool { return false }
func (p *ProcessState) ExitCode() int { return 0 }
func (p *ProcessState) String() string { return "" }
func (p *ProcessState) Success() bool { return false }
func (p *ProcessState) Sys() any { return nil }
func (p *ProcessState) SysUsage() any { return nil }

type Signal interface {
  String() string
  Signal()
}

func (f *File) Chdir() error { return nil }
func (f *File) Chmod(mode FileMode) error { return nil }
func (f *File) Chown(uid, gid int) error { return nil }
func (f *File) Close() error { return nil }
func (f *File) Fd() uintptr { return 0 }
func (f *File) Name() string { return "" }
func (f *File) Read(b []byte) (n int, err error) { return 0, nil }
func (f *File) ReadAt(b []byte, off int64) (n int, err error) { return 0, nil }
func (f *File) ReadDir(n int) ([]DirEntry, error) { return nil, nil }
func (f *File) Readdir(n int) ([]FileInfo, error) { return nil, nil }
func (f *File) Readdirnames(n int) (names []string, err error) { return nil, nil }
func (f *File) Seek(offset int64, whence int) (ret int64, err error) { return 0, nil }
func (f *File) SetDeadline(t any) error { return nil }
func (f *File) SetReadDeadline(t any) error { return nil }
func (f *File) SetWriteDeadline(t any) error { return nil }
func (f *File) Stat() (FileInfo, error) { return nil, nil }
func (f *File) Sync() error { return nil }
func (f *File) SyscallConn() (any, error) { return nil, nil }
func (f *File) Truncate(size int64) error { return nil }
func (f *File) Write(b []byte) (n int, err error) { return len(b), nil }
func (f *File) WriteAt(b []byte, off int64) (n int, err error) { return len(b), nil }
func (f *File) WriteString(s string) (n int, err error) { return len(s), nil }

func Chdir(dir string) error { return nil }
func Chmod(name string, mode FileMode) error { return nil }
func Chown(name string, uid, gid int) error { return nil }
func Chtimes(name string, atime any, mtime any) error { return nil }
func Clearenv() {}
func Create(name string) (*File, error) { return nil, nil }
func CreateTemp(dir, pattern string) (*File, error) { return nil, nil }
func DirFS(dir string) any { return nil }
func Environ() []string { return nil }
func Executable() (string, error) { return "", nil }
func Exit(code int) {}
func Expand(s string, mapping func(string) string) string { return "" }
func ExpandEnv(s string) string { return "" }
func FindProcess(pid int) (*Process, error) { return nil, nil }
func Getegid() int { return 0 }
func Getenv(key string) string { return "" }
func Geteuid() int { return 0 }
func Getgid() int { return 0 }
func Getgroups() ([]int, error) { return nil, nil }
func Getpagesize() int { return 4096 }
func Getpid() int { return 1 }
func Getppid() int { return 0 }
func Getuid() int { return 0 }
func Getwd() (dir string, err error) { return "", nil }
func Hostname() (name string, err error) { return "", nil }
func IsExist(err error) bool { return false }
func IsNotExist(err error) bool { return false }
func IsPathSeparator(c uint8) bool { return false }
func IsPermission(err error) bool { return false }
func IsTimeout(err error) bool { return false }
func Lchown(name string, uid, gid int) error { return nil }
func Link(oldname, newname string) error { return nil }
func LookupEnv(key string) (string, bool) { return "", false }
func Lstat(name string) (FileInfo, error) { return nil, nil }
func Mkdir(name string, perm FileMode) error { return nil }
func MkdirAll(path string, perm FileMode) error { return nil }
func MkdirTemp(dir, pattern string) (string, error) { return "", nil }
func NewFile(fd uintptr, name string) *File { return nil }
func Open(name string) (*File, error) { return nil, nil }
func OpenFile(name string, flag int, perm FileMode) (*File, error) { return nil, nil }
func Pipe() (r *File, w *File, err error) { return nil, nil, nil }
func ReadDir(name string) ([]DirEntry, error) { return nil, nil }
func ReadFile(name string) ([]byte, error) { return nil, nil }
func Readlink(name string) (string, error) { return "", nil }
func Remove(name string) error { return nil }
func RemoveAll(path string) error { return nil }
func Rename(oldpath, newpath string) error { return nil }
func SameFile(fi1, fi2 FileInfo) bool { return false }
func Setenv(key, value string) error { return nil }
func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error) { return nil, nil }
func Stat(name string) (FileInfo, error) { return nil, nil }
func Symlink(oldname, newname string) error { return nil }
func TempDir() string { return "" }
func Truncate(name string, size int64) error { return nil }
func Unsetenv(key string) error { return nil }
func UserCacheDir() (string, error) { return "", nil }
func UserConfigDir() (string, error) { return "", nil }
func UserHomeDir() (string, error) { return "", nil }
func WriteFile(name string, data []byte, perm FileMode) error { return nil }
`);
}
function syncPackage() {
    return checkedSyntheticStandardPackage("sync", syncIntrinsicSource);
}
function syncAtomicPackage() {
    return checkedSyntheticStandardPackage("sync/atomic", `package atomic

import "unsafe"

type Bool struct{}
func (x *Bool) CompareAndSwap(old, next bool) bool { return false }
func (x *Bool) Load() bool { return false }
func (x *Bool) Store(val bool) {}
func (x *Bool) Swap(next bool) bool { return false }

type Int32 struct{}
func (x *Int32) Add(delta int32) int32 { return 0 }
func (x *Int32) And(mask int32) int32 { return 0 }
func (x *Int32) CompareAndSwap(old, next int32) bool { return false }
func (x *Int32) Load() int32 { return 0 }
func (x *Int32) Or(mask int32) int32 { return 0 }
func (x *Int32) Store(val int32) {}
func (x *Int32) Swap(next int32) int32 { return 0 }

type Int64 struct{}
func (x *Int64) Add(delta int64) int64 { return 0 }
func (x *Int64) And(mask int64) int64 { return 0 }
func (x *Int64) CompareAndSwap(old, next int64) bool { return false }
func (x *Int64) Load() int64 { return 0 }
func (x *Int64) Or(mask int64) int64 { return 0 }
func (x *Int64) Store(val int64) {}
func (x *Int64) Swap(next int64) int64 { return 0 }

type Uint32 struct{}
func (x *Uint32) Add(delta uint32) uint32 { return 0 }
func (x *Uint32) And(mask uint32) uint32 { return 0 }
func (x *Uint32) CompareAndSwap(old, next uint32) bool { return false }
func (x *Uint32) Load() uint32 { return 0 }
func (x *Uint32) Or(mask uint32) uint32 { return 0 }
func (x *Uint32) Store(val uint32) {}
func (x *Uint32) Swap(next uint32) uint32 { return 0 }

type Uint64 struct{}
func (x *Uint64) Add(delta uint64) uint64 { return 0 }
func (x *Uint64) And(mask uint64) uint64 { return 0 }
func (x *Uint64) CompareAndSwap(old, next uint64) bool { return false }
func (x *Uint64) Load() uint64 { return 0 }
func (x *Uint64) Or(mask uint64) uint64 { return 0 }
func (x *Uint64) Store(val uint64) {}
func (x *Uint64) Swap(next uint64) uint64 { return 0 }

type Uintptr struct{}
func (x *Uintptr) Add(delta uintptr) uintptr { return 0 }
func (x *Uintptr) And(mask uintptr) uintptr { return 0 }
func (x *Uintptr) CompareAndSwap(old, next uintptr) bool { return false }
func (x *Uintptr) Load() uintptr { return 0 }
func (x *Uintptr) Or(mask uintptr) uintptr { return 0 }
func (x *Uintptr) Store(val uintptr) {}
func (x *Uintptr) Swap(next uintptr) uintptr { return 0 }

type Pointer[T any] struct{}
func (x *Pointer[T]) CompareAndSwap(old, next *T) bool { return false }
func (x *Pointer[T]) Load() *T { return nil }
func (x *Pointer[T]) Store(val *T) {}
func (x *Pointer[T]) Swap(next *T) *T { return nil }

type Value struct{}
func (v *Value) CompareAndSwap(old, next any) bool { return false }
func (v *Value) Load() any { return nil }
func (v *Value) Store(val any) {}
func (v *Value) Swap(next any) any { return nil }

func LoadInt32(addr *int32) int32 { return 0 }
func LoadInt64(addr *int64) int64 { return 0 }
func LoadUint32(addr *uint32) uint32 { return 0 }
func LoadUint64(addr *uint64) uint64 { return 0 }
func LoadUintptr(addr *uintptr) uintptr { return 0 }
func StoreInt32(addr *int32, val int32) {}
func StoreInt64(addr *int64, val int64) {}
func StoreUint32(addr *uint32, val uint32) {}
func StoreUint64(addr *uint64, val uint64) {}
func StoreUintptr(addr *uintptr, val uintptr) {}
func SwapInt32(addr *int32, next int32) int32 { return 0 }
func SwapInt64(addr *int64, next int64) int64 { return 0 }
func SwapUint32(addr *uint32, next uint32) uint32 { return 0 }
func SwapUint64(addr *uint64, next uint64) uint64 { return 0 }
func SwapUintptr(addr *uintptr, next uintptr) uintptr { return 0 }
func CompareAndSwapInt32(addr *int32, old, next int32) bool { return false }
func CompareAndSwapInt64(addr *int64, old, next int64) bool { return false }
func CompareAndSwapUint32(addr *uint32, old, next uint32) bool { return false }
func CompareAndSwapUint64(addr *uint64, old, next uint64) bool { return false }
func CompareAndSwapUintptr(addr *uintptr, old, next uintptr) bool { return false }
func AddInt32(addr *int32, delta int32) int32 { return 0 }
func AddInt64(addr *int64, delta int64) int64 { return 0 }
func AddUint32(addr *uint32, delta uint32) uint32 { return 0 }
func AddUint64(addr *uint64, delta uint64) uint64 { return 0 }
func AddUintptr(addr *uintptr, delta uintptr) uintptr { return 0 }
func AndInt32(addr *int32, mask int32) int32 { return 0 }
func AndInt64(addr *int64, mask int64) int64 { return 0 }
func AndUint32(addr *uint32, mask uint32) uint32 { return 0 }
func AndUint64(addr *uint64, mask uint64) uint64 { return 0 }
func AndUintptr(addr *uintptr, mask uintptr) uintptr { return 0 }
func OrInt32(addr *int32, mask int32) int32 { return 0 }
func OrInt64(addr *int64, mask int64) int64 { return 0 }
func OrUint32(addr *uint32, mask uint32) uint32 { return 0 }
func OrUint64(addr *uint64, mask uint64) uint64 { return 0 }
func OrUintptr(addr *uintptr, mask uintptr) uintptr { return 0 }
func LoadPointer(addr *unsafe.Pointer) unsafe.Pointer { return nil }
func StorePointer(addr *unsafe.Pointer, val unsafe.Pointer) {}
func SwapPointer(addr *unsafe.Pointer, next unsafe.Pointer) unsafe.Pointer { return nil }
func CompareAndSwapPointer(addr *unsafe.Pointer, old, next unsafe.Pointer) bool { return false }
`);
}
function cmpPackage() {
    const pkg = NewPackage("cmp", "cmp");
    if (pkg.Scope().Lookup("Compare") !== null)
        return pkg;
    const intType = Typ[Int];
    const boolType = Typ[Bool];
    const orderedConstraint = NewInterfaceType(null, [NewUnion([
            NewTerm(true, Typ[Int]),
            NewTerm(true, Typ[Int8]),
            NewTerm(true, Typ[Int16]),
            NewTerm(true, Typ[Int32]),
            NewTerm(true, Typ[Int64]),
            NewTerm(true, Typ[Uint]),
            NewTerm(true, Typ[Uint8]),
            NewTerm(true, Typ[Uint16]),
            NewTerm(true, Typ[Uint32]),
            NewTerm(true, Typ[Uint64]),
            NewTerm(true, Typ[Uintptr]),
            NewTerm(true, Typ[Float32]),
            NewTerm(true, Typ[Float64]),
            NewTerm(true, Typ[GoString])
        ])]).Complete();
    const orderedName = NewTypeName(NoPos, pkg, "Ordered", null);
    const orderedType = NewNamed(orderedName, orderedConstraint, null);
    orderedName.setType(orderedType);
    pkg.Scope().Insert(orderedName);
    const orderedTypeParam = () => NewTypeParam(NewTypeName(NoPos, pkg, "T", null), orderedType);
    const comparableTypeParam = () => {
        const comparable = UniverseLookup("comparable");
        const comparableType = comparable?.Type?.() ?? null;
        if (comparableType === null) {
            throw new Error("go/types: predeclared comparable is not initialized");
        }
        return NewTypeParam(NewTypeName(NoPos, pkg, "T", null), comparableType);
    };
    const compareT = orderedTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Compare", NewSignatureType(null, null, [compareT], NewTuple(NewVar(NoPos, pkg, "x", compareT), NewVar(NoPos, pkg, "y", compareT)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    const lessT = orderedTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Less", NewSignatureType(null, null, [lessT], NewTuple(NewVar(NoPos, pkg, "x", lessT), NewVar(NoPos, pkg, "y", lessT)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    const orT = comparableTypeParam();
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Or", NewSignatureType(null, null, [orT], NewTuple(NewVar(NoPos, pkg, "vals", NewSlice(orT))), NewTuple(NewVar(NoPos, pkg, "", orT)), true)));
    pkg.MarkComplete();
    return pkg;
}
function iterPackage() {
    const pkg = NewPackage("iter", "iter");
    if (pkg.Scope().Lookup("Pull") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const seqV = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const seqYield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "value", seqV)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const seqName = NewTypeName(NoPos, pkg, "Seq", null);
    const seqType = NewNamed(seqName, NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", seqYield)), null, false), null);
    seqType.SetTypeParams([seqV]);
    seqName.setType(seqType);
    pkg.Scope().Insert(seqName);
    const seq2K = NewTypeParam(NewTypeName(NoPos, pkg, "K", null), emptyInterface);
    const seq2V = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const seq2Yield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "key", seq2K), NewVar(NoPos, pkg, "value", seq2V)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const seq2Name = NewTypeName(NoPos, pkg, "Seq2", null);
    const seq2Type = NewNamed(seq2Name, NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", seq2Yield)), null, false), null);
    seq2Type.SetTypeParams([seq2K, seq2V]);
    seq2Name.setType(seq2Type);
    pkg.Scope().Insert(seq2Name);
    const pullV = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const pullYield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "value", pullV)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const pullSeq = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", pullYield)), null, false);
    const pullNext = NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "value", pullV), NewVar(NoPos, pkg, "ok", boolType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Pull", NewSignatureType(null, null, [pullV], NewTuple(NewVar(NoPos, pkg, "seq", pullSeq)), NewTuple(NewVar(NoPos, pkg, "next", pullNext), NewVar(NoPos, pkg, "stop", NewSignatureType(null, null, null, null, null, false))), false)));
    const pull2K = NewTypeParam(NewTypeName(NoPos, pkg, "K", null), emptyInterface);
    const pull2V = NewTypeParam(NewTypeName(NoPos, pkg, "V", null), emptyInterface);
    const pull2Yield = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "key", pull2K), NewVar(NoPos, pkg, "value", pull2V)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false);
    const pull2Seq = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "yield", pull2Yield)), null, false);
    const pull2Next = NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "key", pull2K), NewVar(NoPos, pkg, "value", pull2V), NewVar(NoPos, pkg, "ok", boolType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Pull2", NewSignatureType(null, null, [pull2K, pull2V], NewTuple(NewVar(NoPos, pkg, "seq", pull2Seq)), NewTuple(NewVar(NoPos, pkg, "next", pull2Next), NewVar(NoPos, pkg, "stop", NewSignatureType(null, null, null, null, null, false))), false)));
    pkg.MarkComplete();
    return pkg;
}
function mathPackage() {
    const pkg = NewPackage("math", "math");
    if (pkg.Scope().Lookup("NaN") !== null)
        return pkg;
    const float64Type = Typ[Float64];
    const uint32Type = Typ[Uint32];
    const uint64Type = Typ[Uint64];
    const intType = Typ[Int];
    const untypedIntType = Typ[UntypedInt];
    const boolType = Typ[Bool];
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt", untypedIntType, 9223372036854775807n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt", untypedIntType, -9223372036854775808n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt8", untypedIntType, 127n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt8", untypedIntType, -128n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt16", untypedIntType, 32767n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt16", untypedIntType, -32768n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt32", untypedIntType, 2147483647n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt32", untypedIntType, -2147483648n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxInt64", untypedIntType, 9223372036854775807n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MinInt64", untypedIntType, -9223372036854775808n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint", untypedIntType, 18446744073709551615n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint8", untypedIntType, 255n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint16", untypedIntType, 65535n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint32", untypedIntType, 4294967295n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxUint64", untypedIntType, 18446744073709551615n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxFloat32", float64Type, 3.4028234663852886e38));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "MaxFloat64", float64Type, Number.MAX_VALUE));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NaN", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Inf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "sign", intType)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Exp", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Floor", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "IsNaN", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", float64Type)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Log", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", float64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float32bits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", Typ[Float32])), NewTuple(NewVar(NoPos, pkg, "", uint32Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float32frombits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", uint32Type)), NewTuple(NewVar(NoPos, pkg, "", Typ[Float32])), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float64bits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "f", float64Type)), NewTuple(NewVar(NoPos, pkg, "", uint64Type)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Float64frombits", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", uint64Type)), NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)));
    pkg.MarkComplete();
    return pkg;
}
function fmtPackage() {
    const pkg = NewPackage("fmt", "fmt");
    if (pkg.Scope().Lookup("Printf") !== null)
        return pkg;
    const stringType = Typ[GoString];
    const intType = Typ[Int];
    const byteSliceType = NewSlice(Typ[Uint8]);
    const argsType = NewSlice(emptyInterface);
    const errorObject = UniverseLookup("error");
    const errorType = errorObject?.Type?.() ?? null;
    if (errorType === null) {
        throw new Error("go/types: predeclared error is not initialized");
    }
    const writerType = NewInterfaceType([
        NewFunc(NoPos, pkg, "Write", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", byteSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), false))
    ], null).Complete();
    const printfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const sprintfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true);
    const fprintfSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "w", writerType), NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const printSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    const fprintSig = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "w", writerType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprint", fprintSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprintf", fprintfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Fprintln", fprintSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Print", printSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Printf", printfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprintf", sprintfSig));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Errorf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", errorType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprint", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Sprintln", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", stringType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Append", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Appendf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Appendln", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "b", byteSliceType), NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", byteSliceType)), true)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Println", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "args", argsType)), NewTuple(NewVar(NoPos, pkg, "", intType), NewVar(NoPos, pkg, "", errorType)), true)));
    pkg.MarkComplete();
    return pkg;
}
function reflectlitePackage() {
    const pkg = NewPackage("internal/reflectlite", "reflectlite");
    if (pkg.Scope().Lookup("TypeOf") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const intType = Typ[Int];
    const anyType = emptyInterface;
    const kindName = NewTypeName(NoPos, pkg, "Kind", null);
    const kindType = NewNamed(kindName, intType, null);
    kindName.setType(kindType);
    pkg.Scope().Insert(kindName);
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Invalid", kindType, 0n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Interface", kindType, 20n));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Ptr", kindType, 22n));
    const typeName = NewTypeName(NoPos, pkg, "Type", null);
    const typeType = NewNamed(typeName, NewInterfaceType(null, null).Complete(), null);
    typeName.setType(typeType);
    pkg.Scope().Insert(typeName);
    const typeMethods = [
        NewFunc(NoPos, pkg, "AssignableTo", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "u", typeType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Comparable", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Elem", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)),
        NewFunc(NoPos, pkg, "Implements", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "u", typeType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)),
        NewFunc(NoPos, pkg, "Kind", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", kindType)), false)),
        NewFunc(NoPos, pkg, "String", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", Typ[GoString])), false))
    ];
    typeType.SetUnderlying(NewInterfaceType(typeMethods, null).Complete());
    const valueName = NewTypeName(NoPos, pkg, "Value", null);
    const valueType = NewNamed(valueName, NewStruct([], null), null);
    valueName.setType(valueType);
    pkg.Scope().Insert(valueName);
    const valueRecv = NewVar(NoPos, pkg, "v", valueType);
    valueType.AddMethod(NewFunc(NoPos, pkg, "Elem", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "IsNil", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Kind", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", kindType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Len", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Set", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "x", valueType)), null, false)));
    valueType.AddMethod(NewFunc(NoPos, pkg, "Type", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "TypeOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", anyType)), NewTuple(NewVar(NoPos, pkg, "", typeType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Swapper", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "slice", anyType)), NewTuple(NewVar(NoPos, pkg, "", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType), NewVar(NoPos, pkg, "j", intType)), null, false))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ValueOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "i", anyType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function syscallJSPackage() {
    const pkg = NewPackage("syscall/js", "js");
    if (pkg.Scope().Lookup("Value") !== null)
        return pkg;
    const anyType = emptyInterface;
    const boolType = Typ[Bool];
    const float64Type = Typ[Float64];
    const intType = Typ[Int];
    const stringType = Typ[GoString];
    const byteSliceType = NewSlice(Typ[Uint8]);
    const typeName = NewTypeName(NoPos, pkg, "Type", null);
    const typeType = NewNamed(typeName, intType, null);
    typeName.setType(typeType);
    pkg.Scope().Insert(typeName);
    typeType.AddMethod(NewFunc(NoPos, pkg, "String", NewSignatureType(NewVar(NoPos, pkg, "t", typeType), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    [
        "TypeUndefined",
        "TypeNull",
        "TypeBoolean",
        "TypeNumber",
        "TypeString",
        "TypeSymbol",
        "TypeObject",
        "TypeFunction"
    ].forEach((name, index) => {
        pkg.Scope().Insert(NewConst(NoPos, pkg, name, typeType, BigInt(index)));
    });
    const valueName = NewTypeName(NoPos, pkg, "Value", null);
    const valueType = NewNamed(valueName, NewStruct([], null), null);
    valueName.setType(valueType);
    pkg.Scope().Insert(valueName);
    const valueRecv = NewVar(NoPos, pkg, "v", valueType);
    const valueMethods = [
        ["Bool", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Call", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "m", stringType), NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["Delete", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType)), null, false)],
        ["Equal", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "w", valueType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Float", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", float64Type)), false)],
        ["Get", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)],
        ["Index", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)],
        ["InstanceOf", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "t", valueType)), NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Int", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)],
        ["Invoke", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["IsNaN", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["IsNull", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["IsUndefined", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Length", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)],
        ["New", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "args", NewSlice(anyType))), NewTuple(NewVar(NoPos, pkg, "", valueType)), true)],
        ["Set", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "p", stringType), NewVar(NoPos, pkg, "x", anyType)), null, false)],
        ["SetIndex", NewSignatureType(valueRecv, null, null, NewTuple(NewVar(NoPos, pkg, "i", intType), NewVar(NoPos, pkg, "x", anyType)), null, false)],
        ["String", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)],
        ["Truthy", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)],
        ["Type", NewSignatureType(valueRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", typeType)), false)]
    ];
    for (const [name, signature] of valueMethods) {
        valueType.AddMethod(NewFunc(NoPos, pkg, name, signature));
    }
    const funcName = NewTypeName(NoPos, pkg, "Func", null);
    const funcType = NewNamed(funcName, NewStruct([
        NewField(NoPos, pkg, "Value", valueType, true)
    ], null), null);
    funcName.setType(funcType);
    pkg.Scope().Insert(funcName);
    funcType.AddMethod(NewFunc(NoPos, pkg, "Release", NewSignatureType(NewVar(NoPos, pkg, "c", funcType), null, null, null, null, false)));
    const errorName = NewTypeName(NoPos, pkg, "Error", null);
    const errorType = NewNamed(errorName, NewStruct([
        NewField(NoPos, pkg, "Value", valueType, true)
    ], null), null);
    errorName.setType(errorType);
    pkg.Scope().Insert(errorName);
    errorType.AddMethod(NewFunc(NoPos, pkg, "Error", NewSignatureType(NewVar(NoPos, pkg, "e", errorType), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const valueErrorName = NewTypeName(NoPos, pkg, "ValueError", null);
    const valueErrorType = NewNamed(valueErrorName, NewStruct([
        NewField(NoPos, pkg, "Method", stringType, false),
        NewField(NoPos, pkg, "Type", typeType, false)
    ], null), null);
    valueErrorName.setType(valueErrorType);
    pkg.Scope().Insert(valueErrorName);
    valueErrorType.AddMethod(NewFunc(NoPos, pkg, "Error", NewSignatureType(NewVar(NoPos, pkg, "e", NewPointer(valueErrorType)), null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const funcOfParam = NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "this", valueType), NewVar(NoPos, pkg, "args", NewSlice(valueType))), NewTuple(NewVar(NoPos, pkg, "", anyType)), false);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "FuncOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "fn", funcOfParam)), NewTuple(NewVar(NoPos, pkg, "", funcType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Global", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Null", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Undefined", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ValueOf", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", anyType)), NewTuple(NewVar(NoPos, pkg, "", valueType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CopyBytesToGo", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "dst", byteSliceType), NewVar(NoPos, pkg, "src", valueType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CopyBytesToJS", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "dst", valueType), NewVar(NoPos, pkg, "src", byteSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function runtimePackage() {
    const pkg = NewPackage("runtime", "runtime");
    if (pkg.Scope().Lookup("GOOS") !== null)
        return pkg;
    const intType = Typ[Int];
    const boolType = Typ[Bool];
    const int64Type = Typ[Int64];
    const stringType = Typ[GoString];
    const uintptrType = Typ[Uintptr];
    const uint32Type = Typ[Uint32];
    const uint64Type = Typ[Uint64];
    const float64Type = Typ[Float64];
    const byteSliceType = NewSlice(Typ[Uint8]);
    const uintptrSliceType = NewSlice(uintptrType);
    pkg.Scope().Insert(NewConst(NoPos, pkg, "GOOS", stringType, "js"));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "GOARCH", stringType, "gojr"));
    pkg.Scope().Insert(NewConst(NoPos, pkg, "Compiler", stringType, "gojr"));
    pkg.Scope().Insert(NewVar(NoPos, pkg, "MemProfileRate", intType));
    const funcName = NewTypeName(NoPos, pkg, "Func", null);
    const funcType = NewNamed(funcName, NewStruct([], null), null);
    funcName.setType(funcType);
    pkg.Scope().Insert(funcName);
    const funcRecv = NewVar(NoPos, pkg, "f", NewPointer(funcType));
    funcType.AddMethod(NewFunc(NoPos, pkg, "Entry", NewSignatureType(funcRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrType)), false)));
    funcType.AddMethod(NewFunc(NoPos, pkg, "FileLine", NewSignatureType(funcRecv, null, null, NewTuple(NewVar(NoPos, pkg, "pc", uintptrType)), NewTuple(NewVar(NoPos, pkg, "file", stringType), NewVar(NoPos, pkg, "line", intType)), false)));
    funcType.AddMethod(NewFunc(NoPos, pkg, "Name", NewSignatureType(funcRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    const frameName = NewTypeName(NoPos, pkg, "Frame", null);
    const frameType = NewNamed(frameName, NewStruct([
        NewField(NoPos, pkg, "PC", uintptrType, false),
        NewField(NoPos, pkg, "Func", NewPointer(funcType), false),
        NewField(NoPos, pkg, "Function", stringType, false),
        NewField(NoPos, pkg, "File", stringType, false),
        NewField(NoPos, pkg, "Line", intType, false),
        NewField(NoPos, pkg, "Entry", uintptrType, false)
    ], null), null);
    frameName.setType(frameType);
    pkg.Scope().Insert(frameName);
    const framesName = NewTypeName(NoPos, pkg, "Frames", null);
    const framesType = NewNamed(framesName, NewStruct([], null), null);
    framesName.setType(framesType);
    pkg.Scope().Insert(framesName);
    const framesRecv = NewVar(NoPos, pkg, "ci", NewPointer(framesType));
    framesType.AddMethod(NewFunc(NoPos, pkg, "Next", NewSignatureType(framesRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "frame", frameType), NewVar(NoPos, pkg, "more", boolType)), false)));
    const stackRecordName = NewTypeName(NoPos, pkg, "StackRecord", null);
    const stackRecordType = NewNamed(stackRecordName, NewStruct([
        NewField(NoPos, pkg, "Stack0", NewArray(uintptrType, 32), false)
    ], null), null);
    stackRecordName.setType(stackRecordType);
    pkg.Scope().Insert(stackRecordName);
    stackRecordType.AddMethod(NewFunc(NoPos, pkg, "Stack", NewSignatureType(NewVar(NoPos, pkg, "r", NewPointer(stackRecordType)), null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrSliceType)), false)));
    const memProfileRecordName = NewTypeName(NoPos, pkg, "MemProfileRecord", null);
    const memProfileRecordType = NewNamed(memProfileRecordName, NewStruct([
        NewField(NoPos, pkg, "AllocBytes", int64Type, false),
        NewField(NoPos, pkg, "FreeBytes", int64Type, false),
        NewField(NoPos, pkg, "AllocObjects", int64Type, false),
        NewField(NoPos, pkg, "FreeObjects", int64Type, false),
        NewField(NoPos, pkg, "Stack0", NewArray(uintptrType, 32), false)
    ], null), null);
    memProfileRecordName.setType(memProfileRecordType);
    pkg.Scope().Insert(memProfileRecordName);
    const memProfileRecordRecv = NewVar(NoPos, pkg, "r", NewPointer(memProfileRecordType));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "InUseBytes", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", int64Type)), false)));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "InUseObjects", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", int64Type)), false)));
    memProfileRecordType.AddMethod(NewFunc(NoPos, pkg, "Stack", NewSignatureType(memProfileRecordRecv, null, null, null, NewTuple(NewVar(NoPos, pkg, "", uintptrSliceType)), false)));
    const blockProfileRecordName = NewTypeName(NoPos, pkg, "BlockProfileRecord", null);
    const blockProfileRecordType = NewNamed(blockProfileRecordName, NewStruct([
        NewField(NoPos, pkg, "Count", int64Type, false),
        NewField(NoPos, pkg, "Cycles", int64Type, false),
        NewField(NoPos, pkg, "StackRecord", stackRecordType, true)
    ], null), null);
    blockProfileRecordName.setType(blockProfileRecordType);
    pkg.Scope().Insert(blockProfileRecordName);
    const errorObject = UniverseLookup("error");
    const errorType = errorObject?.Type?.() ?? null;
    if (errorType === null) {
        throw new Error("go/types: predeclared error is not initialized");
    }
    const runtimeErrorInterface = NewInterfaceType([
        NewFunc(NoPos, pkg, "RuntimeError", NewSignatureType(null, null, null, null, null, false))
    ], [errorType]).Complete();
    const runtimeErrorName = NewTypeName(NoPos, pkg, "Error", null);
    const runtimeErrorType = NewNamed(runtimeErrorName, runtimeErrorInterface, null);
    runtimeErrorName.setType(runtimeErrorType);
    pkg.Scope().Insert(runtimeErrorName);
    const cleanupName = NewTypeName(NoPos, pkg, "Cleanup", null);
    const cleanupType = NewNamed(cleanupName, NewStruct([], null), null);
    cleanupName.setType(cleanupType);
    pkg.Scope().Insert(cleanupName);
    const cleanupRecv = NewVar(NoPos, pkg, "c", cleanupType);
    cleanupType.AddMethod(NewFunc(NoPos, pkg, "Stop", NewSignatureType(cleanupRecv, null, null, null, null, false)));
    const memStatsName = NewTypeName(NoPos, pkg, "MemStats", null);
    const bySizeStruct = NewStruct([
        NewField(NoPos, pkg, "Size", uint32Type, false),
        NewField(NoPos, pkg, "Mallocs", uint64Type, false),
        NewField(NoPos, pkg, "Frees", uint64Type, false)
    ], null);
    const memStatsType = NewNamed(memStatsName, NewStruct([
        NewField(NoPos, pkg, "Alloc", uint64Type, false),
        NewField(NoPos, pkg, "TotalAlloc", uint64Type, false),
        NewField(NoPos, pkg, "Sys", uint64Type, false),
        NewField(NoPos, pkg, "Lookups", uint64Type, false),
        NewField(NoPos, pkg, "Mallocs", uint64Type, false),
        NewField(NoPos, pkg, "Frees", uint64Type, false),
        NewField(NoPos, pkg, "HeapAlloc", uint64Type, false),
        NewField(NoPos, pkg, "HeapSys", uint64Type, false),
        NewField(NoPos, pkg, "HeapIdle", uint64Type, false),
        NewField(NoPos, pkg, "HeapInuse", uint64Type, false),
        NewField(NoPos, pkg, "HeapReleased", uint64Type, false),
        NewField(NoPos, pkg, "HeapObjects", uint64Type, false),
        NewField(NoPos, pkg, "StackInuse", uint64Type, false),
        NewField(NoPos, pkg, "StackSys", uint64Type, false),
        NewField(NoPos, pkg, "MSpanInuse", uint64Type, false),
        NewField(NoPos, pkg, "MSpanSys", uint64Type, false),
        NewField(NoPos, pkg, "MCacheInuse", uint64Type, false),
        NewField(NoPos, pkg, "MCacheSys", uint64Type, false),
        NewField(NoPos, pkg, "BuckHashSys", uint64Type, false),
        NewField(NoPos, pkg, "GCSys", uint64Type, false),
        NewField(NoPos, pkg, "OtherSys", uint64Type, false),
        NewField(NoPos, pkg, "NextGC", uint64Type, false),
        NewField(NoPos, pkg, "LastGC", uint64Type, false),
        NewField(NoPos, pkg, "PauseTotalNs", uint64Type, false),
        NewField(NoPos, pkg, "PauseNs", NewArray(uint64Type, 256), false),
        NewField(NoPos, pkg, "PauseEnd", NewArray(uint64Type, 256), false),
        NewField(NoPos, pkg, "NumGC", uint32Type, false),
        NewField(NoPos, pkg, "NumForcedGC", uint32Type, false),
        NewField(NoPos, pkg, "GCCPUFraction", float64Type, false),
        NewField(NoPos, pkg, "EnableGC", boolType, false),
        NewField(NoPos, pkg, "DebugGC", boolType, false),
        NewField(NoPos, pkg, "BySize", NewArray(bySizeStruct, 61), false)
    ], null), null);
    memStatsName.setType(memStatsType);
    pkg.Scope().Insert(memStatsName);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Caller", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "skip", intType)), NewTuple(NewVar(NoPos, pkg, "pc", uintptrType), NewVar(NoPos, pkg, "file", stringType), NewVar(NoPos, pkg, "line", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Callers", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "skip", intType), NewVar(NoPos, pkg, "pc", uintptrSliceType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "CallersFrames", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "callers", uintptrSliceType)), NewTuple(NewVar(NoPos, pkg, "", NewPointer(framesType))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "FuncForPC", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "pc", uintptrType)), NewTuple(NewVar(NoPos, pkg, "", NewPointer(funcType))), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "BlockProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(blockProfileRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    const cleanupT = NewTypeParam(NewTypeName(NoPos, pkg, "T", null), emptyInterface);
    const cleanupS = NewTypeParam(NewTypeName(NoPos, pkg, "S", null), emptyInterface);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "AddCleanup", NewSignatureType(null, null, [cleanupT, cleanupS], NewTuple(NewVar(NoPos, pkg, "ptr", NewPointer(cleanupT)), NewVar(NoPos, pkg, "cleanup", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "", cleanupS)), null, false)), NewVar(NoPos, pkg, "arg", cleanupS)), NewTuple(NewVar(NoPos, pkg, "", cleanupType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GC", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GOMAXPROCS", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "n", intType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GOROOT", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Goexit", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "GoroutineProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(stackRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Gosched", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "KeepAlive", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "x", emptyInterface)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "MemProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(memProfileRecordType)), NewVar(NoPos, pkg, "inuseZero", boolType)), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "MutexProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(blockProfileRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NumCPU", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "NumGoroutine", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ReadMemStats", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "m", NewPointer(memStatsType))), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ReadTrace", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "buf", byteSliceType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetBlockProfileRate", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "rate", intType)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetCPUProfileRate", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "hz", intType)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetFinalizer", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "obj", emptyInterface), NewVar(NoPos, pkg, "finalizer", emptyInterface)), null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "SetMutexProfileFraction", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "rate", intType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "StartTrace", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", errorType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "StopTrace", NewSignatureType(null, null, null, null, null, false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "ThreadCreateProfile", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "p", NewSlice(stackRecordType))), NewTuple(NewVar(NoPos, pkg, "n", intType), NewVar(NoPos, pkg, "ok", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Stack", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "buf", byteSliceType), NewVar(NoPos, pkg, "all", boolType)), NewTuple(NewVar(NoPos, pkg, "", intType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Version", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", stringType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function testingPackage() {
    const pkg = NewPackage("testing", "testing");
    if (pkg.Scope().Lookup("T") !== null)
        return pkg;
    const boolType = Typ[Bool];
    const stringType = Typ[GoString];
    const argsType = NewSlice(emptyInterface);
    const tName = NewTypeName(NoPos, pkg, "T", null);
    const tType = NewNamed(tName, NewStruct([], null), null);
    tName.setType(tType);
    const tPtr = NewPointer(tType);
    const recv = NewVar(NoPos, pkg, "t", tPtr);
    pkg.Scope().Insert(tName);
    const addMethod = (targetType, receiver, name, result, parameters = [], variadic = false) => {
        const params = NewTuple(...parameters);
        const results = result === null ? null : NewTuple(NewVar(NoPos, pkg, "", result));
        targetType.AddMethod(NewFunc(NoPos, pkg, name, NewSignatureType(receiver, null, null, params, results, variadic)));
    };
    const addCommonMethods = (targetType, receiver, receiverPtrType) => {
        addMethod(targetType, receiver, "Fail", null);
        addMethod(targetType, receiver, "FailNow", null);
        addMethod(targetType, receiver, "Failed", boolType);
        addMethod(targetType, receiver, "Fatal", null, [NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Fatalf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Error", null, [NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Errorf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Log", null, [NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Logf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Name", stringType);
        addMethod(targetType, receiver, "Helper", null);
        addMethod(targetType, receiver, "Skip", null, [NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "Skipf", null, [NewVar(NoPos, pkg, "format", stringType), NewVar(NoPos, pkg, "args", argsType)], true);
        addMethod(targetType, receiver, "SkipNow", null);
        addMethod(targetType, receiver, "Skipped", boolType);
        addMethod(targetType, receiver, "Cleanup", null, [NewVar(NoPos, pkg, "f", NewSignatureType(null, null, null, null, null, false))]);
        addMethod(targetType, receiver, "Run", boolType, [
            NewVar(NoPos, pkg, "name", stringType),
            NewVar(NoPos, pkg, "f", NewSignatureType(null, null, null, NewTuple(NewVar(NoPos, pkg, "", receiverPtrType)), null, false))
        ]);
    };
    addCommonMethods(tType, recv, tPtr);
    const bName = NewTypeName(NoPos, pkg, "B", null);
    const bType = NewNamed(bName, NewStruct([
        NewField(NoPos, pkg, "N", Typ[Int], false)
    ], null), null);
    bName.setType(bType);
    const bPtr = NewPointer(bType);
    const bRecv = NewVar(NoPos, pkg, "b", bPtr);
    pkg.Scope().Insert(bName);
    addCommonMethods(bType, bRecv, bPtr);
    addMethod(bType, bRecv, "ReportAllocs", null);
    addMethod(bType, bRecv, "ResetTimer", null);
    addMethod(bType, bRecv, "StartTimer", null);
    addMethod(bType, bRecv, "StopTimer", null);
    addMethod(bType, bRecv, "SetBytes", null, [NewVar(NoPos, pkg, "n", Typ[Int64])]);
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Short", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.Scope().Insert(NewFunc(NoPos, pkg, "Verbose", NewSignatureType(null, null, null, null, NewTuple(NewVar(NoPos, pkg, "", boolType)), false)));
    pkg.MarkComplete();
    return pkg;
}
function diagnosticFromGoTypesError(error, fset) {
    const err = error;
    const span = fset.span(err.go116start ?? err.Pos);
    return {
        filename: diagnosticFilename(span),
        code: "GOJR_TYPE001",
        severity: "error",
        message: err.Msg ?? err.message ?? err.Error?.() ?? String(error),
        span
    };
}
function ident(name, span) {
    return {
        kind: "Ident",
        name,
        ...(span ? { span } : {})
    };
}
function positionOffset(pos) {
    const n = typeof pos === "number" ? pos : Number(pos);
    return Number.isFinite(n) && n >= 0 ? n : 0;
}
function fileForOffset(files, offset) {
    for (const file of files) {
        const start = file.span?.offset ?? 0;
        const end = start + (file.span?.length ?? 0);
        if (offset >= start && (file.span === undefined || offset <= end))
            return file;
    }
    return files[0];
}
function firstFilename(files) {
    return files[0]?.span?.filename ?? REPL_FILENAME;
}
function spanFromOffset(filename, offset, source) {
    const location = source === undefined
        ? { line: 1, column: offset + 1 }
        : lineColumnFromOffset(source, offset);
    return {
        filename,
        offset,
        length: 1,
        line: location.line,
        column: location.column
    };
}
function lineColumnFromOffset(source, offset) {
    const target = Math.max(0, Math.min(offset, source.length));
    let line = 1;
    let lineStart = 0;
    for (let index = 0; index < target; index += 1) {
        if (source.charCodeAt(index) === 10) {
            line += 1;
            lineStart = index + 1;
        }
    }
    return {
        line,
        column: target - lineStart + 1
    };
}
