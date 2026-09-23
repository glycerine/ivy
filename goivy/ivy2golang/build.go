package ivy2golang

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type BuildPlan struct {
	GoFile     string
	OutputPath string
	Args       []string
	Env        []string
	WorkDir    string
}

func BuildOutput(out *Output, outDir string) (string, error) {
	return buildOutputWithEnv(out, outDir, nil)
}

func buildOutputWithEnv(out *Output, outDir string, env []string) (string, error) {
	plan, err := BuildPlanFor(out, outDir, out.Config)
	if err != nil {
		return "", err
	}
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	if err := formatGoOutputFile(plan.GoFile); err != nil {
		return "", err
	}
	if env == nil {
		env = plan.Env
	}
	cmd := exec.Command("go", plan.Args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Dir = plan.WorkDir
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ivy2golang: go build failed: %w\n%s", err, string(data))
	}
	return plan.OutputPath, nil
}

func formatGoOutputFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	formatted, err := format.Source(data)
	if err != nil {
		return fmt.Errorf("ivy2golang: gofmt generated Go failed for %s: %w", path, err)
	}
	if bytes.Equal(data, formatted) {
		return nil
	}
	return os.WriteFile(path, formatted, 0o644)
}

func BuildPlanFor(out *Output, outDir string, cfg Config) (*BuildPlan, error) {
	if out == nil {
		return nil, fmt.Errorf("ivy2golang: nil output")
	}
	dir, err := filepath.Abs(outputDirectory(outDir))
	if err != nil {
		return nil, err
	}
	baseDir, err := filepath.Abs(outputBaseDirectory(outDir))
	if err != nil {
		return nil, err
	}
	goFile := filepath.Join(dir, goSourceFileName(out.BaseName))
	outputPath := filepath.Join(dir, out.BaseName)
	compileOnly := out.Target == "class" || (!out.EmitMain && out.Target != "")
	if compileOnly {
		outputPath += ".a"
	} else if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cacheDir, err := generatedBuildDir("IVY2GOLANG_GOCACHE", filepath.Join(baseDir, ".ivy2golang-gocache"))
	if err != nil {
		return nil, err
	}
	tmpDir, err := generatedBuildDir("IVY2GOLANG_GOTMPDIR", filepath.Join(baseDir, ".ivy2golang-gotmp"))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, err
	}
	return &BuildPlan{
		GoFile:     goFile,
		OutputPath: outputPath,
		Args:       []string{"build", "-o", outputPath, goFile},
		Env:        []string{"GOCACHE=" + cacheDir, "GOTMPDIR=" + tmpDir},
		WorkDir:    moduleRootForGeneratedBuild(),
	}, nil
}

func generatedBuildDir(envName, fallback string) (string, error) {
	dir := os.Getenv(envName)
	if dir == "" {
		dir = fallback
	}
	return filepath.Abs(dir)
}

func moduleRootForGeneratedBuild() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok || !filepath.IsAbs(file) {
		return ""
	}
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
