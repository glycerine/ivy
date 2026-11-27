// Package mapaudit guards the port against Go's randomized map iteration.
// ABC walks arrays and vectors in a fixed order; a Go loop over a map whose
// order can reach printed output, object or ID creation, tie-breaking, SAT
// variable numbering, written files, returned slice order or the error that
// is reported makes the port nondeterministic. Every range over a map (and
// every maps.Keys/Values/All not passed straight to slices.Sorted or
// slices.SortedFunc) in non-test code must be listed in allowlist_test.go with
// the reason its order cannot matter. Order-sensitive loops should instead
// walk a slice kept alongside the map, or slices.Sorted(maps.Keys(m)).
package mapaudit

import (
	"bufio"
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"
)

const modulePath = "github.com/glycerine/abcgo"

// uncheckedPackages lists packages (relative to the module root) that the
// audit skips entirely. Keep it empty unless a package cannot be checked.
var uncheckedPackages = map[string]bool{}

type listedPackage struct {
	importPath, dir string
	files           []string
}

// listModule runs one `go list -export -deps`: it yields the module's own
// packages and export data for every dependency, so each package can be
// type-checked from source without type-checking its imports.
func listModule(t *testing.T, root string) ([]listedPackage, map[string]string) {
	format := "{{if not .ForTest}}{{.ImportPath}}\t{{.Export}}\t{{.Dir}}\t{{with .Module}}{{.Path}}{{end}}\t{{join .GoFiles \" \"}}{{end}}"
	cmd := exec.Command("go", "list", "-pgo=off", "-export", "-deps", "-f", format, "./...")
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, stderr.String())
	}
	exports := map[string]string{}
	var pkgs []listedPackage
	scanner := bufio.NewScanner(bytes.NewReader(out))
	scanner.Buffer(nil, 1<<20)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "\t")
		if len(fields) != 5 || strings.Contains(fields[0], " ") {
			continue
		}
		exports[fields[0]] = fields[1]
		if fields[3] == modulePath {
			pkgs = append(pkgs, listedPackage{fields[0], fields[2], strings.Fields(fields[4])})
		}
	}
	return pkgs, exports
}

type site struct {
	key  string // file:function, the allowlist key
	line int
	what string
}

// findSites type-checks one package from source and reports its map
// iterations.
func findSites(fset *token.FileSet, imp types.Importer, root string, p listedPackage) ([]site, []error) {
	var files []*ast.File
	var errs []error
	for _, name := range p.files {
		f, err := parser.ParseFile(fset, filepath.Join(p.dir, name), nil, parser.SkipObjectResolution|parser.ParseComments)
		if err != nil {
			return nil, []error{err}
		}
		files = append(files, f)
	}
	info := &types.Info{Types: map[ast.Expr]types.TypeAndValue{}, Uses: map[*ast.Ident]types.Object{}}
	conf := types.Config{Importer: imp, Error: func(err error) { errs = append(errs, err) }}
	conf.Check(p.importPath, fset, files, info)
	if len(errs) > 0 {
		return nil, errs
	}
	var sites []site
	for _, f := range files {
		rel, _ := filepath.Rel(root, fset.Position(f.Pos()).Filename)
		rel = filepath.ToSlash(rel)
		for _, decl := range f.Decls {
			name := "<init>"
			if fd, ok := decl.(*ast.FuncDecl); ok {
				name = funcName(fd)
			}
			// Calls already wrapped in slices.Sorted*(...) are fine.
			sorted := map[ast.Expr]bool{}
			ast.Inspect(decl, func(n ast.Node) bool {
				switch n := n.(type) {
				case *ast.RangeStmt:
					if t := info.TypeOf(n.X); t != nil {
						if _, ok := t.Underlying().(*types.Map); ok {
							sites = append(sites, site{rel + ":" + name, fset.Position(n.Pos()).Line, "range over map"})
						}
					}
				case *ast.CallExpr:
					if pkg, fn := calledFunc(info, n); pkg == "slices" && strings.HasPrefix(fn, "Sorted") && len(n.Args) > 0 {
						sorted[n.Args[0]] = true
					} else if pkg == "maps" && (fn == "Keys" || fn == "Values" || fn == "All") && !sorted[n] {
						sites = append(sites, site{rel + ":" + name, fset.Position(n.Pos()).Line, "maps." + fn + " outside slices.Sorted"})
					}
				}
				return true
			})
		}
	}
	return sites, nil
}

func funcName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	t := fd.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if index, ok := t.(*ast.IndexExpr); ok {
		t = index.X
	}
	if index, ok := t.(*ast.IndexListExpr); ok {
		t = index.X
	}
	if id, ok := t.(*ast.Ident); ok {
		return id.Name + "." + fd.Name.Name
	}
	return fd.Name.Name
}

func calledFunc(info *types.Info, call *ast.CallExpr) (string, string) {
	fun := call.Fun
	if index, ok := fun.(*ast.IndexExpr); ok {
		fun = index.X
	}
	if index, ok := fun.(*ast.IndexListExpr); ok {
		fun = index.X
	}
	sel, ok := fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	obj, ok := info.Uses[sel.Sel].(*types.Func)
	if !ok || obj.Pkg() == nil {
		return "", ""
	}
	return obj.Pkg().Path(), obj.Name()
}

type importerFunc func(string) (*types.Package, error)

func (f importerFunc) Import(path string) (*types.Package, error) { return f(path) }

func TestNoUnreviewedMapIteration(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, exports := listModule(t, root)
	if len(pkgs) == 0 {
		t.Fatal("go list found no packages")
	}
	fset := token.NewFileSet()
	var importMu sync.Mutex
	gc := importer.ForCompiler(fset, "gc", func(path string) (io.ReadCloser, error) {
		file, ok := exports[path]
		if !ok || file == "" {
			return nil, fmt.Errorf("no export data for %s", path)
		}
		return os.Open(file)
	})
	imp := importerFunc(func(path string) (*types.Package, error) {
		importMu.Lock()
		defer importMu.Unlock()
		return gc.Import(path)
	})

	var mu sync.Mutex
	var all []site
	var failures []string
	var wg sync.WaitGroup
	limit := make(chan struct{}, 8)
	for _, p := range pkgs {
		rel := strings.TrimPrefix(strings.TrimPrefix(p.importPath, modulePath), "/")
		if uncheckedPackages[rel] {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			limit <- struct{}{}
			defer func() { <-limit }()
			sites, errs := findSites(fset, imp, root, p)
			mu.Lock()
			defer mu.Unlock()
			for _, err := range errs {
				failures = append(failures, fmt.Sprintf("%s: type-check: %v", p.importPath, err))
			}
			all = append(all, sites...)
		}()
	}
	wg.Wait()
	sort.Strings(failures)
	for _, f := range failures {
		t.Error(f)
	}

	found := map[string][]site{}
	for _, s := range all {
		found[s.key] = append(found[s.key], s)
	}
	for _, key := range slices.Sorted(maps.Keys(found)) {
		sites := found[key]
		entry, ok := allowlist[key]
		if !ok {
			for _, s := range sites {
				file := key[:strings.LastIndex(key, ":")]
				t.Errorf("%s:%d: %s in %s is not in internal/mapaudit/allowlist_test.go; iterate in a fixed order (a slice kept alongside, or slices.Sorted(maps.Keys(m))) or, if order cannot matter, allowlist it with the reason", file, s.line, s.what, key)
			}
			continue
		}
		if len(sites) != entry.n {
			lines := make([]string, len(sites))
			for i, s := range sites {
				lines[i] = fmt.Sprint(s.line)
			}
			t.Errorf("%s: %d map iterations (lines %s) but the allowlist expects %d; review the new or removed loop and update internal/mapaudit/allowlist_test.go", key, len(sites), strings.Join(lines, ","), entry.n)
		}
	}
	for _, key := range slices.Sorted(maps.Keys(allowlist)) {
		if _, ok := found[key]; !ok {
			t.Errorf("allowlist entry %s no longer matches any map iteration; remove it", key)
		}
	}
}
