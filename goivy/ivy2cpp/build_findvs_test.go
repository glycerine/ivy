package ivy2cpp

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindVsParsesVswhereOutput(t *testing.T) {
	payloads := []string{
		`[{"installationPath":"C:\\Program Files\\Microsoft Visual Studio\\2022\\Community"}]`,
		`{"installationPath":"C:\\Program Files\\Microsoft Visual Studio\\2022\\BuildTools"}`,
		`"C:\\Program Files\\Microsoft Visual Studio\\2022\\Enterprise"`,
	}
	for _, payload := range payloads {
		got, err := parseVSWhereInstallationPath([]byte(payload))
		if err != nil {
			t.Fatalf("parseVSWhereInstallationPath(%s): %v", payload, err)
		}
		if !strings.Contains(got, "Visual Studio") {
			t.Fatalf("parsed installation path %q does not look like Visual Studio", got)
		}
	}
}

func TestFindVsAbsentReturnsClearError(t *testing.T) {
	_, err := parseVSWhereInstallationPath([]byte(`[]`))
	if err == nil || !strings.Contains(err.Error(), "installationPath") {
		t.Fatalf("empty vswhere payload error = %v, want installationPath diagnostic", err)
	}
}

func TestVSInfoFromInstallChoosesLatestMSVCToolset(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Microsoft Visual Studio", "2022", "Community")
	for _, version := range []string{"14.28.29910", "14.40.10000", "14.9.99999"} {
		for _, dir := range []string{
			filepath.Join(root, "VC", "Tools", "MSVC", version, "bin", "Hostx64", "x64"),
			filepath.Join(root, "VC", "Tools", "MSVC", version, "include"),
			filepath.Join(root, "VC", "Tools", "MSVC", version, "lib", "x64"),
		} {
			if err := mkdirAll(dir); err != nil {
				t.Fatalf("mkdir %s: %v", dir, err)
			}
		}
	}
	info, err := vsInfoFromInstall(root)
	if err != nil {
		t.Fatalf("vsInfoFromInstall: %v", err)
	}
	if info.ToolsVersion != "14.40.10000" {
		t.Fatalf("tool version = %q, want latest 14.40.10000", info.ToolsVersion)
	}
	if !strings.HasSuffix(info.BinDir, filepath.Join("14.40.10000", "bin", "Hostx64", "x64")) {
		t.Fatalf("unexpected bin dir %q", info.BinDir)
	}
}

func TestMsvcBuildPlanIncludesToolchainPaths(t *testing.T) {
	old := findVSFunc
	t.Cleanup(func() { findVSFunc = old })
	findVSFunc = func() (vsInfo, error) {
		return vsInfo{
			InstallDir:   `C:\VS`,
			ToolsVersion: "14.40.10000",
			BinDir:       `C:\VS\VC\Tools\MSVC\14.40.10000\bin\Hostx64\x64`,
			IncludeDirs:  []string{`C:\VS\VC\Tools\MSVC\14.40.10000\include`, `C:\SDK\Include\ucrt`},
			LibDirs:      []string{`C:\VS\VC\Tools\MSVC\14.40.10000\lib\x64`, `C:\SDK\Lib\ucrt\x64`},
			Env:          []string{`PATH=C:\Windows\System32`, `INCLUDE=C:\OldInclude`, `LIB=C:\OldLib`},
		}, nil
	}
	plan, err := msvcBuildPlan(&Output{BaseName: "tiny", EmitMain: true}, "cl", `C:\tmp\tiny.cpp`, `C:\tmp\tiny.exe`, false)
	if err != nil {
		t.Fatalf("msvcBuildPlan: %v", err)
	}
	for _, want := range []string{
		`C:\VS\VC\Tools\MSVC\14.40.10000\include`,
		`C:\SDK\Include\ucrt`,
		`/LIBPATH:C:\VS\VC\Tools\MSVC\14.40.10000\lib\x64`,
		`/LIBPATH:C:\SDK\Lib\ucrt\x64`,
	} {
		if !sliceContains(plan.Args, want) {
			t.Fatalf("MSVC plan missing %q in args: %#v", want, plan.Args)
		}
	}
	path := envValue(plan.Env, "PATH")
	if !strings.HasPrefix(path, `C:\VS\VC\Tools\MSVC\14.40.10000\bin\Hostx64\x64;`) {
		t.Fatalf("PATH was not prefixed with MSVC bin dir: %q", path)
	}
	if !strings.HasPrefix(envValue(plan.Env, "INCLUDE"), `C:\VS\VC\Tools\MSVC\14.40.10000\include;C:\SDK\Include\ucrt;`) {
		t.Fatalf("INCLUDE was not prefixed with VS include dirs: %q", envValue(plan.Env, "INCLUDE"))
	}
}

func TestMsvcBuildPlanWithoutFindVSStillBuilds(t *testing.T) {
	old := findVSFunc
	t.Cleanup(func() { findVSFunc = old })
	findVSFunc = func() (vsInfo, error) {
		return vsInfo{}, errors.New("vswhere unavailable in test")
	}
	plan, err := msvcBuildPlan(&Output{BaseName: "tiny", EmitMain: true}, "cl", `C:\tmp\tiny.cpp`, `C:\tmp\tiny.exe`, false)
	if err != nil {
		t.Fatalf("msvcBuildPlan without VS discovery should still return a plan: %v", err)
	}
	if plan.Compiler != "cl" || !sliceContains(plan.Args, "/EHsc") || !sliceContains(plan.Args, `C:\tmp\tiny.cpp`) {
		t.Fatalf("fallback MSVC plan is not usable: %#v", plan)
	}
	if len(plan.Env) != 0 {
		t.Fatalf("fallback plan should not synthesize an env when findVS fails: %#v", plan.Env)
	}
}

func mkdirAll(path string) error {
	return os.MkdirAll(path, 0o755)
}
