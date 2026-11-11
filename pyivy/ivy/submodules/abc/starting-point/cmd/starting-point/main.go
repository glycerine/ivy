package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glycerine/abc/starting-point"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string) error {
	flags := flag.NewFlagSet("starting-point", flag.ContinueOnError)
	out := flags.String("out", "", "new output directory (must not exist)")
	pkg := flags.String("package", "", "Go package override (default: first source file directory name, shared by all outputs)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() == 1 {
		info, err := os.Stat(flags.Arg(0))
		if err == nil && info.IsDir() {
			destination := *out
			if destination == "" {
				destination, err = startingpoint.DefaultTreeOutput(flags.Arg(0))
				if err != nil {
					return err
				}
			}
			report, err := startingpoint.TranslateTree(flags.Arg(0), destination, startingpoint.Options{Package: *pkg})
			if err != nil {
				return err
			}
			fmt.Printf("Translated %d files into %s (%d source-preserving fallbacks)\n", len(report.Files), report.Output, report.Fallbacks)
			return nil
		}
	}
	if *out == "" || flags.NArg() == 0 {
		return fmt.Errorf("usage: starting-point [-package NAME] [-out NEW_DIRECTORY] SOURCE_DIRECTORY; or starting-point -out NEW_DIRECTORY source.h ...")
	}
	packageName, err := startingpoint.ResolvePackage(flags.Arg(0), *pkg)
	if err != nil {
		return err
	}
	// Refuse overwrite, including repaired output supplied by mistake.
	if err := os.Mkdir(*out, 0755); err != nil {
		return err
	}
	start := time.Now()
	var results []startingpoint.Result
	for i, path := range flags.Args() {
		s, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		r, err := startingpoint.AnalyzeWithOptions(path, string(s), startingpoint.Options{Package: packageName})
		if err != nil {
			return err
		}
		name := fmt.Sprintf("%02d_%s", i, filepath.Base(path))
		if err = os.WriteFile(filepath.Join(*out, name+".original"), s, 0644); err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(*out, name+".go"), []byte(r.Draft), 0644); err != nil {
			return err
		}
		results = append(results, r)
	}
	data, err := json.MarshalIndent(struct {
		Version    string
		Package    string
		DurationNS int64
		Files      []startingpoint.Result
	}{"prototype-19", packageName, time.Since(start).Nanoseconds(), results}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*out, "run.json"), append(data, '\n'), 0644)
}
