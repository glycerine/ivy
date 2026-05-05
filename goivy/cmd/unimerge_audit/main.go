// Command unimerge_audit reports the mechanical blockers for flattening the
// layered Go Ivy packages into the root goivy package.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var defaultSeeds = []string{
	"ast",
	"lexer",
	"ivyutils",
	"logic",
	"parser",
	"ivylogic",
	"module",
	"z3bridge",
	"logicutil",
	"logicparser",
	"actions",
	"isolate",
	"interp",
	"compiler",
	"art",
	"proof",
	"trace",
	"bmc",
	"temporal",
	"tactics",
	"webui",
	"iupdr",
	"check",
}

var defaultExcludes = []string{
	"xtracer",
}

type goPackage struct {
	ImportPath     string
	Name           string
	Dir            string
	GoFiles        []string
	CgoFiles       []string
	TestGoFiles    []string
	XTestGoFiles   []string
	IgnoredGoFiles []string
	Imports        []string
	TestImports    []string
	XTestImports   []string
}

type fileInfo struct {
	Rel        string
	Dir        string
	Base       string
	Package    string
	IsTest     bool
	IsXTest    bool
	IsIgnored  bool
	BuildExprs []constraint.Expr
	Directives []directive
}

type directive struct {
	Line int
	Text string
}

type declaration struct {
	Name string
	Kind string
	File fileInfo
}

type collision struct {
	Name  string
	Decls []declaration
}

func main() {
	var includeVprint bool
	var details bool
	var strict bool
	var maxDetails int
	var seedCSV string
	var excludeCSV string

	flag.BoolVar(&includeVprint, "include-vprint", false, "include copied vprint debug helper files in audits")
	flag.BoolVar(&details, "details", false, "print detailed collision groups")
	flag.BoolVar(&strict, "strict", false, "exit non-zero when blocking collisions remain")
	flag.IntVar(&maxDetails, "max", 40, "maximum detailed collision groups to print per section")
	flag.StringVar(&seedCSV, "seeds", strings.Join(defaultSeeds, ","), "comma-separated seed package dirs under the module root")
	flag.StringVar(&excludeCSV, "exclude", strings.Join(defaultExcludes, ","), "comma-separated local package dirs to leave outside the flat package")
	flag.Parse()

	modulePath, err := goListModule()
	must(err)

	seeds := parseSeeds(seedCSV)
	pkgs, err := goListDeps(seeds)
	must(err)
	closure := localClosure(modulePath, pkgs)
	excluded := excludeLocalPackages(modulePath, closure, parsePackageDirs(excludeCSV))
	files, err := scanFiles(closure, includeVprint)
	must(err)

	flatCollisions := basenameCollisions(files)
	prodDecls, xtestDecls, err := collectDecls(files)
	must(err)
	prodCollisions := declarationCollisions(prodDecls)
	xtestCollisions := declarationCollisions(xtestDecls)
	clients, err := outsideClients(modulePath, closure)
	must(err)
	directives := collectDirectives(files)

	fmt.Printf("module=%s\n", modulePath)
	fmt.Printf("seed_packages=%d\n", len(seeds))
	fmt.Printf("closure_packages=%d\n", len(closure))
	fmt.Printf("excluded_packages=%d\n", len(excluded))
	fmt.Printf("go_files=%d\n", len(files))
	fmt.Printf("vprint_ignored=%v\n", !includeVprint)
	fmt.Printf("flat_basename_collisions=%d\n", len(flatCollisions))
	fmt.Printf("top_level_collisions=%d\n", len(prodCollisions))
	fmt.Printf("external_test_collisions=%d\n", len(xtestCollisions))
	fmt.Printf("outside_client_packages=%d\n", len(clients))
	fmt.Printf("go_directives=%d\n", len(directives))

	if details {
		printBasenameCollisions(flatCollisions, maxDetails)
		printDeclCollisions("top-level declaration collisions", prodCollisions, maxDetails)
		printDeclCollisions("external test declaration collisions", xtestCollisions, maxDetails)
		printClients(clients)
		printDirectives(directives)
	}

	if strict && (len(flatCollisions) > 0 || len(prodCollisions) > 0 || len(xtestCollisions) > 0) {
		os.Exit(1)
	}
}

func parseSeeds(seedCSV string) []string {
	var seeds []string
	for _, seed := range strings.Split(seedCSV, ",") {
		seed = strings.TrimSpace(seed)
		if seed == "" {
			continue
		}
		if strings.HasPrefix(seed, "./") {
			seeds = append(seeds, seed)
		} else {
			seeds = append(seeds, "./"+seed)
		}
	}
	return seeds
}

func parsePackageDirs(csv string) []string {
	var dirs []string
	for _, dir := range strings.Split(csv, ",") {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dir = strings.TrimPrefix(dir, "./")
		dir = strings.Trim(dir, "/")
		if dir != "" {
			dirs = append(dirs, dir)
		}
	}
	return dirs
}

func goListModule() (string, error) {
	out, err := exec.Command("go", "list", "-m").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func goListDeps(seeds []string) ([]goPackage, error) {
	args := append([]string{"list", "-deps", "-json"}, seeds...)
	out, err := exec.Command("go", args...).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, ee.Stderr)
		}
		return nil, err
	}
	return parseGoListJSON(out)
}

func goListAll() ([]goPackage, error) {
	out, err := exec.Command("go", "list", "-json", "./...").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("go list -json ./...: %w\n%s", err, ee.Stderr)
		}
		return nil, err
	}
	return parseGoListJSON(out)
}

func parseGoListJSON(out []byte) ([]goPackage, error) {
	dec := json.NewDecoder(bytes.NewReader(out))
	var pkgs []goPackage
	for dec.More() {
		var pkg goPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, err
		}
		pkgs = append(pkgs, pkg)
	}
	return pkgs, nil
}

func localClosure(modulePath string, pkgs []goPackage) map[string]goPackage {
	closure := map[string]goPackage{}
	localPrefix := modulePath + "/"
	for _, pkg := range pkgs {
		if pkg.ImportPath != modulePath && !strings.HasPrefix(pkg.ImportPath, localPrefix) {
			continue
		}
		if !strings.Contains(filepath.ToSlash(pkg.Dir), "/goivy/") {
			continue
		}
		closure[pkg.ImportPath] = pkg
	}
	return closure
}

func excludeLocalPackages(modulePath string, closure map[string]goPackage, dirs []string) []string {
	var excluded []string
	for _, dir := range dirs {
		importPath := modulePath
		if dir != "." && dir != "" {
			importPath += "/" + dir
		}
		if _, ok := closure[importPath]; ok {
			delete(closure, importPath)
			excluded = append(excluded, dir)
		}
	}
	sort.Strings(excluded)
	return excluded
}

func scanFiles(closure map[string]goPackage, includeVprint bool) ([]fileInfo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var importPaths []string
	for importPath := range closure {
		importPaths = append(importPaths, importPath)
	}
	sort.Strings(importPaths)

	ignoredByDir := map[string]map[string]bool{}
	xtestByDir := map[string]map[string]bool{}
	for _, pkg := range closure {
		ignoredByDir[pkg.Dir] = toSet(pkg.IgnoredGoFiles)
		xtestByDir[pkg.Dir] = toSet(pkg.XTestGoFiles)
	}

	var files []fileInfo
	for _, importPath := range importPaths {
		pkg := closure[importPath]
		entries, err := os.ReadDir(pkg.Dir)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
				continue
			}
			base := entry.Name()
			if !includeVprint && isVprintFile(base) {
				continue
			}
			full := filepath.Join(pkg.Dir, base)
			rel, err := filepath.Rel(cwd, full)
			if err != nil {
				return nil, err
			}
			info := fileInfo{
				Rel:       filepath.ToSlash(rel),
				Dir:       filepath.ToSlash(filepath.Dir(rel)),
				Base:      base,
				IsTest:    strings.HasSuffix(base, "_test.go"),
				IsXTest:   xtestByDir[pkg.Dir][base],
				IsIgnored: ignoredByDir[pkg.Dir][base] || hasBuildIgnore(full),
			}
			info.Package, info.BuildExprs, info.Directives, err = parseFileHeader(full)
			if err != nil {
				return nil, err
			}
			files = append(files, info)
		}
	}
	return files, nil
}

func isVprintFile(base string) bool {
	return base == "vprint.go" || strings.HasSuffix(base, "_vprint.go")
}

func hasBuildIgnore(path string) bool {
	_, exprs, _, err := parseFileHeader(path)
	if err != nil {
		return false
	}
	for _, expr := range exprs {
		if expr.String() == "ignore" {
			return true
		}
	}
	return false
}

func parseFileHeader(path string) (string, []constraint.Expr, []directive, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return "", nil, nil, err
	}
	directives := scanDirectives(src)
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, src, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return "", nil, nil, err
	}
	var exprs []constraint.Expr
	for _, group := range f.Comments {
		for _, comment := range group.List {
			text := comment.Text
			if strings.HasPrefix(text, "//go:build ") || strings.HasPrefix(text, "// +build ") {
				parsed, err := constraint.Parse(text)
				if err != nil {
					return "", nil, nil, fmt.Errorf("%s: %w", path, err)
				}
				exprs = append(exprs, parsed)
			}
		}
	}
	return f.Name.Name, exprs, directives, nil
}

func scanDirectives(src []byte) []directive {
	var directives []directive
	for i, line := range strings.Split(string(src), "\n") {
		text := strings.TrimSpace(line)
		if strings.HasPrefix(text, "//go:embed ") || strings.HasPrefix(text, "//go:generate ") {
			directives = append(directives, directive{
				Line: i + 1,
				Text: text,
			})
		}
	}
	return directives
}

func basenameCollisions(files []fileInfo) map[string][]fileInfo {
	byBase := map[string][]fileInfo{}
	for _, file := range files {
		byBase[file.Base] = append(byBase[file.Base], file)
	}
	for base, group := range byBase {
		if len(group) < 2 {
			delete(byBase, base)
		}
	}
	return byBase
}

func collectDecls(files []fileInfo) ([]declaration, []declaration, error) {
	var prod []declaration
	var xtest []declaration
	fset := token.NewFileSet()
	for _, file := range files {
		if file.IsIgnored {
			continue
		}
		f, err := parser.ParseFile(fset, file.Rel, nil, 0)
		if err != nil {
			return nil, nil, err
		}
		target := &prod
		if file.IsXTest {
			target = &xtest
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				// Multiple init functions are legal in one Go package, so they
				// are not flattening blockers.
				if d.Recv == nil && d.Name.Name != "_" && d.Name.Name != "init" {
					*target = append(*target, declaration{Name: d.Name.Name, Kind: "func", File: file})
				}
			case *ast.GenDecl:
				kind := d.Tok.String()
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						*target = append(*target, declaration{Name: s.Name.Name, Kind: "type", File: file})
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if name.Name == "_" {
								continue
							}
							*target = append(*target, declaration{Name: name.Name, Kind: kind, File: file})
						}
					}
				}
			}
		}
	}
	return prod, xtest, nil
}

func declarationCollisions(decls []declaration) []collision {
	byName := map[string][]declaration{}
	for _, decl := range decls {
		byName[decl.Name] = append(byName[decl.Name], decl)
	}

	var collisions []collision
	for name, group := range byName {
		if len(group) < 2 {
			continue
		}
		if hasCompatiblePair(group) {
			collisions = append(collisions, collision{Name: name, Decls: group})
		}
	}
	sort.Slice(collisions, func(i, j int) bool {
		return collisions[i].Name < collisions[j].Name
	})
	return collisions
}

func hasCompatiblePair(group []declaration) bool {
	for i := 0; i < len(group); i++ {
		for j := i + 1; j < len(group); j++ {
			if compatibleBuilds(group[i].File.BuildExprs, group[j].File.BuildExprs) {
				return true
			}
		}
	}
	return false
}

func compatibleBuilds(a, b []constraint.Expr) bool {
	tags := map[string]bool{}
	collectTags(a, tags)
	collectTags(b, tags)
	var names []string
	for tag := range tags {
		names = append(names, tag)
	}
	sort.Strings(names)
	if len(names) > 20 {
		return true
	}
	limit := 1 << len(names)
	for mask := 0; mask < limit; mask++ {
		ok := func(tag string) bool {
			for i, name := range names {
				if name == tag {
					return mask&(1<<i) != 0
				}
			}
			return false
		}
		if evalBuilds(a, ok) && evalBuilds(b, ok) {
			return true
		}
	}
	return false
}

func collectTags(exprs []constraint.Expr, tags map[string]bool) {
	for _, expr := range exprs {
		collectTagsExpr(expr, tags)
	}
}

func collectTagsExpr(expr constraint.Expr, tags map[string]bool) {
	switch e := expr.(type) {
	case nil:
	case *constraint.TagExpr:
		tags[e.Tag] = true
	case *constraint.NotExpr:
		collectTagsExpr(e.X, tags)
	case *constraint.AndExpr:
		collectTagsExpr(e.X, tags)
		collectTagsExpr(e.Y, tags)
	case *constraint.OrExpr:
		collectTagsExpr(e.X, tags)
		collectTagsExpr(e.Y, tags)
	}
}

func evalBuilds(exprs []constraint.Expr, ok func(tag string) bool) bool {
	for _, expr := range exprs {
		if !expr.Eval(ok) {
			return false
		}
	}
	return true
}

func outsideClients(modulePath string, closure map[string]goPackage) (map[string][]string, error) {
	all, err := goListAll()
	if err != nil {
		return nil, err
	}
	clients := map[string][]string{}
	for _, pkg := range all {
		if _, inClosure := closure[pkg.ImportPath]; inClosure {
			continue
		}
		if pkg.ImportPath == modulePath || !strings.HasPrefix(pkg.ImportPath, modulePath+"/") {
			continue
		}
		var imports []string
		seen := map[string]bool{}
		for _, importPath := range append(append(pkg.Imports, pkg.TestImports...), pkg.XTestImports...) {
			if _, ok := closure[importPath]; ok && !seen[importPath] {
				seen[importPath] = true
				imports = append(imports, strings.TrimPrefix(importPath, modulePath+"/"))
			}
		}
		if len(imports) > 0 {
			sort.Strings(imports)
			clients[strings.TrimPrefix(pkg.ImportPath, modulePath+"/")] = imports
		}
	}
	return clients, nil
}

func collectDirectives(files []fileInfo) []string {
	var out []string
	for _, file := range files {
		for _, directive := range file.Directives {
			out = append(out, fmt.Sprintf("%s:%d: %s", file.Rel, directive.Line, directive.Text))
		}
	}
	sort.Strings(out)
	return out
}

func printBasenameCollisions(collisions map[string][]fileInfo, max int) {
	fmt.Println()
	fmt.Println("flat basename collisions:")
	if len(collisions) == 0 {
		fmt.Println("  none")
		return
	}
	var bases []string
	for base := range collisions {
		bases = append(bases, base)
	}
	sort.Strings(bases)
	for i, base := range bases {
		if i >= max {
			fmt.Printf("  ... %d more\n", len(bases)-max)
			return
		}
		fmt.Printf("  %s\n", base)
		for _, file := range collisions[base] {
			fmt.Printf("    %s\n", file.Rel)
		}
	}
}

func printDeclCollisions(title string, collisions []collision, max int) {
	fmt.Println()
	fmt.Println(title + ":")
	if len(collisions) == 0 {
		fmt.Println("  none")
		return
	}
	for i, collision := range collisions {
		if i >= max {
			fmt.Printf("  ... %d more\n", len(collisions)-max)
			return
		}
		fmt.Printf("  %s\n", collision.Name)
		for _, decl := range collision.Decls {
			fmt.Printf("    %-5s %s\n", decl.Kind, decl.File.Rel)
		}
	}
}

func printClients(clients map[string][]string) {
	fmt.Println()
	fmt.Println("outside clients importing closure packages:")
	if len(clients) == 0 {
		fmt.Println("  none")
		return
	}
	var names []string
	for name := range clients {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Printf("  %s: %s\n", name, strings.Join(clients[name], ", "))
	}
}

func printDirectives(directives []string) {
	fmt.Println()
	fmt.Println("go directives in closure:")
	if len(directives) == 0 {
		fmt.Println("  none")
		return
	}
	for _, directive := range directives {
		fmt.Printf("  %s\n", directive)
	}
}

func toSet(values []string) map[string]bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	return set
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
