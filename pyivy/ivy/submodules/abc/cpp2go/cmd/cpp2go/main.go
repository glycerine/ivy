package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glycerine/abc/cpp2go"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string) error {
	flags := flag.NewFlagSet("cpp2go", flag.ContinueOnError)
	out := flags.String("out", "", "new output directory (must not exist)")
	pkg := flags.String("package", "", "Go package override (default: first source file directory name, shared by all outputs)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *out == "" || flags.NArg() == 0 {
		return fmt.Errorf("usage: cpp2go -out NEW_DIRECTORY source.h ...")
	}
	packageName, err := cpp2go.ResolvePackage(flags.Arg(0), *pkg)
	if err != nil {
		return err
	}
	// Refuse overwrite, including repaired output supplied by mistake.
	if err := os.Mkdir(*out, 0755); err != nil {
		return err
	}
	start := time.Now()
	var results []cpp2go.Result
	for i, path := range flags.Args() {
		s, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		r, err := cpp2go.AnalyzeWithOptions(path, string(s), cpp2go.Options{Package: packageName})
		if err != nil {
			return err
		}
		name := fmt.Sprintf("%02d_%s", i, filepath.Base(path))
		if err = os.WriteFile(filepath.Join(*out, name+".original"), s, 0644); err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(*out, name+".go.txt"), []byte(r.Draft), 0644); err != nil {
			return err
		}
		results = append(results, r)
	}
	data, err := json.MarshalIndent(struct {
		Version    string
		Package    string
		DurationNS int64
		Files      []cpp2go.Result
	}{"prototype-10", packageName, time.Since(start).Nanoseconds(), results}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*out, "run.json"), append(data, '\n'), 0644)
}
