package ivy2cpp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/ivy/goivy"
)

func CompileAndGenerate(filename string, params map[string]string, cfg Config) (*Output, error) {
	cfg, ivyParams, err := mergeParams(params, cfg)
	if err != nil {
		return nil, err
	}
	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	if isolate := ivyParams["isolate"]; isolate != "" {
		mod.Cfg.Isolate = isolate
	}
	sig := goivy.NewSigOn(mod.Cfg.IuCfg)
	if err := goivy.SourceFile(filename, mod, sig, map[string]interface{}{"create_isolate": false}); err != nil {
		return nil, err
	}
	return Generate(mod, cfg)
}

func mergeParams(params map[string]string, cfg Config) (Config, map[string]string, error) {
	ivyParams := map[string]string{}
	for k, v := range params {
		switch k {
		case "target":
			if cfg.Target == "" {
				cfg.Target = v
			}
		case "classname":
			if cfg.ClassName == "" {
				cfg.ClassName = v
			}
		case "main":
			if cfg.MainName == "" {
				cfg.MainName = v
			}
		case "outdir":
			if cfg.OutDir == "" {
				cfg.OutDir = v
			}
		case "trace":
			cfg.Trace = parseBool(v)
		case "stdafx":
			cfg.Stdafx = parseBool(v)
		case "build":
			cfg.Build = parseBool(v)
		case "isolate":
			ivyParams[k] = v
		default:
			return cfg, nil, fmt.Errorf("ivy2cpp: unknown parameter %q", k)
		}
	}
	return cfg, ivyParams, nil
}

func BuildRequested(params map[string]string) bool {
	if params == nil {
		return false
	}
	return parseBool(params["build"])
}

func ParseArgs(args []string) (map[string]string, string, error) {
	params := map[string]string{}
	rest := args
	for len(rest) > 0 && strings.Contains(rest[0], "=") {
		parts := strings.SplitN(rest[0], "=", 2)
		if parts[0] == "" {
			return nil, "", fmt.Errorf("bad parameter %q", rest[0])
		}
		params[parts[0]] = parts[1]
		rest = rest[1:]
	}
	if len(rest) != 1 || !strings.HasSuffix(rest[0], ".ivy") {
		return nil, "", fmt.Errorf("usage: ivy2cpp [key=value ...] file.ivy")
	}
	return params, rest[0], nil
}

func WriteOutput(out *Output, outDir string) error {
	if out == nil {
		return fmt.Errorf("ivy2cpp: nil output")
	}
	dir := outDir
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, out.BaseName+".h"), []byte(out.Header), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, out.BaseName+".cpp"), []byte(out.Impl), 0o644); err != nil {
		return err
	}
	return nil
}

func parseBool(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
