//go:build windows

package ivy2cpp

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func defaultFindVS() (vsInfo, error) {
	install, err := findVSInstallWithVSWhere()
	if err != nil {
		install, err = findVSInstallLegacy()
		if err != nil {
			return vsInfo{}, err
		}
	}
	info, err := vsInfoFromInstall(install)
	if err != nil {
		return vsInfo{}, err
	}
	if env, err := captureVCVarsEnv(info.InstallDir); err == nil && len(env) != 0 {
		info.Env = env
		info.IncludeDirs = append(info.IncludeDirs, splitWindowsPathList(envValue(env, "INCLUDE"))...)
		info.LibDirs = append(info.LibDirs, splitWindowsPathList(envValue(env, "LIB"))...)
	}
	return info, nil
}

func findVSInstallWithVSWhere() (string, error) {
	vswhere, err := locateVSWhere()
	if err != nil {
		return "", err
	}
	cmd := exec.Command(vswhere,
		"-latest",
		"-products", "*",
		"-requires", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
		"-format", "json",
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("vswhere failed: %w: %s", err, stderr.String())
	}
	return parseVSWhereInstallationPath(out)
}

func locateVSWhere() (string, error) {
	if path := strings.TrimSpace(os.Getenv("VSWHERE")); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	for _, root := range []string{os.Getenv("ProgramFiles(x86)"), os.Getenv("ProgramFiles")} {
		if root == "" {
			continue
		}
		path := filepath.Join(root, "Microsoft Visual Studio", "Installer", "vswhere.exe")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	if path, err := exec.LookPath("vswhere.exe"); err == nil {
		return path, nil
	}
	return "", fmt.Errorf("vswhere.exe not found")
}

func findVSInstallLegacy() (string, error) {
	drive := "C"
	if windir := strings.TrimSpace(os.Getenv("WINDIR")); len(windir) >= 2 && windir[1] == ':' {
		drive = windir[:1]
	}
	for version := 17; version >= 10; version-- {
		for _, suffix := range []string{"", " (x86)"} {
			path := fmt.Sprintf(`%s:\Program Files%s\Microsoft Visual Studio %d.0`, drive, suffix, version)
			if _, err := os.Stat(path); err == nil {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("vswhere.exe not found and no suitable Visual Studio installation found")
}

func captureVCVarsEnv(install string) ([]string, error) {
	candidates := []string{
		filepath.Join(install, "VC", "Auxiliary", "Build", "vcvarsall.bat"),
		filepath.Join(install, "VC", "vcvarsall.bat"),
	}
	var bat string
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			bat = candidate
			break
		}
	}
	if bat == "" {
		return nil, fmt.Errorf("vcvarsall.bat not found under %s", install)
	}
	cmdline := fmt.Sprintf(`call "%s" amd64 >nul && set`, bat)
	cmd := exec.Command("cmd.exe", "/s", "/c", cmdline)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("vcvarsall.bat failed: %w", err)
	}
	return parseSetOutput(string(out)), nil
}

func parseSetOutput(text string) []string {
	var env []string
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, "=") {
			continue
		}
		env = append(env, line)
	}
	return env
}

func splitWindowsPathList(text string) []string {
	var out []string
	seen := map[string]bool{}
	for _, part := range strings.Split(text, ";") {
		part = strings.TrimSpace(part)
		if part == "" || seen[strings.ToUpper(part)] {
			continue
		}
		seen[strings.ToUpper(part)] = true
		out = append(out, part)
	}
	return out
}
