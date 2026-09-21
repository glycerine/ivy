package ivy2golang

import (
	"fmt"
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
}

func BuildOutput(out *Output, outDir string) (string, error) {
	plan, err := BuildPlanFor(out, outDir, out.Config)
	if err != nil {
		return "", err
	}
	if err := WriteOutput(out, outDir); err != nil {
		return "", err
	}
	cmd := exec.Command("go", plan.Args...)
	cmd.Env = append(os.Environ(), plan.Env...)
	data, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ivy2golang: go build failed: %w\n%s", err, string(data))
	}
	return plan.OutputPath, nil
}

func BuildPlanFor(out *Output, outDir string, cfg Config) (*BuildPlan, error) {
	if out == nil {
		return nil, fmt.Errorf("ivy2golang: nil output")
	}
	dir := outputDirectory(outDir)
	goFile := filepath.Join(dir, out.BaseName+".go")
	outputPath := filepath.Join(dir, out.BaseName)
	compileOnly := out.Target == "class" || (!out.EmitMain && out.Target != "")
	if compileOnly {
		outputPath += ".a"
	} else if runtime.GOOS == "windows" {
		outputPath += ".exe"
	}
	cacheDir := filepath.Join(outputBaseDirectory(outDir), ".ivy2golang-gocache")
	tmpDir := filepath.Join(outputBaseDirectory(outDir), ".ivy2golang-gotmp")
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
	}, nil
}
