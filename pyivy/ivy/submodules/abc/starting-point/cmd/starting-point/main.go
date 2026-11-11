package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		source, isDirectory, err := resolveSourceArgument(flags.Arg(0), cwd)
		if err != nil {
			return err
		}
		if isDirectory {
			destination := *out
			if destination == "" {
				destination, err = startingpoint.DefaultTreeOutput(source)
				if err != nil {
					return err
				}
			}
			report, err := startingpoint.TranslateTree(source, destination, startingpoint.Options{Package: *pkg})
			if err != nil {
				return err
			}
			printTreeReport(os.Stdout, report)
			return nil
		}
		if *out == "" {
			return fmt.Errorf("source %q is a file, not a directory; use a source directory, or -out for a file batch", flags.Arg(0))
		}
	}
	if *out == "" || flags.NArg() == 0 {
		return fmt.Errorf("usage: starting-point [-package NAME] [-out NEW_DIRECTORY] SOURCE_DIRECTORY; or starting-point -out NEW_DIRECTORY source.h ...")
	}
	packageName, err := startingpoint.ResolvePackage(flags.Arg(0), *pkg)
	if err != nil {
		return err
	}
	basenames := make([]string, flags.NArg())
	counts := map[string]int{}
	for i, path := range flags.Args() {
		basenames[i] = filepath.Base(path)
		counts[basenames[i]]++
	}
	for i, name := range basenames {
		if counts[name] > 1 {
			basenames[i] = fmt.Sprintf("%02d_%s", i, name)
		}
	}
	names, err := startingpoint.PlanGoFilenames(basenames)
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
		if err = os.WriteFile(filepath.Join(*out, names[i]), []byte(r.Draft), 0644); err != nil {
			return err
		}
		results = append(results, r)
	}
	data, err := json.MarshalIndent(struct {
		Version    string
		Package    string
		DurationNS int64
		Files      []startingpoint.Result
	}{"prototype-24", packageName, time.Since(start).Nanoseconds(), results}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(*out, "run.json"), append(data, '\n'), 0644)
}

// A bare project name also denotes the current checkout when its basename
// matches. Prefer an actual subdirectory, but do not mistake a same-named
// executable (ABC ships one) for the requested checkout.
func resolveSourceArgument(arg, cwd string) (string, bool, error) {
	path := arg
	if !filepath.IsAbs(path) {
		path = filepath.Join(cwd, path)
	}
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return path, true, nil
	}
	if arg == filepath.Base(cwd) {
		return cwd, true, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("cannot access source %q: %w", arg, err)
	}
	return path, false, nil
}

func printTreeReport(w io.Writer, report startingpoint.TreeReport) {
	fmt.Fprintf(w, "Translated %d files into %s (%d source-preserving fallbacks)\n", len(report.Files), report.Output, report.Fallbacks)
	if report.Fallbacks == 0 {
		return
	}
	fmt.Fprintln(w, "Fallbacks (commented original source; analysis failed):")
	for _, file := range report.Files {
		if !file.Fallback {
			continue
		}
		path, err := filepath.Rel(report.Source, file.Source)
		if err != nil {
			path = file.Source
		}
		fmt.Fprintf(w, "  %s -> %s: %s\n", filepath.ToSlash(path), filepath.ToSlash(file.Output), file.FallbackReason)
	}
}
