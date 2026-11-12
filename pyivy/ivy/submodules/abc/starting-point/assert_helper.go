package startingpoint

import (
	"fmt"
	"os"
	"path/filepath"
)

// NeedsAssertHelper reports whether a draft uses the generated package helper.
func (r Result) NeedsAssertHelper() bool {
	for _, change := range r.Library {
		if change.Name == "assert" && change.Status == "converted" {
			return true
		}
	}
	return false
}

// WriteAssertHelper is called once per output directory, after all translated
// files have been written, so source filenames cannot collide with the helper.
// The caller supplies the same resolved package name used in its other files.
func WriteAssertHelper(directory, packageName string) (string, error) {
	pkg, err := ResolvePackage("", packageName)
	if err != nil {
		return "", err
	}
	source := "package " + pkg + "\n\nfunc assert(x bool, message string) {\n\tif !x {\n\t\tpanic(\"assertion failed, we do not see: \" + message)\n\t}\n}\n"
	for n := 1; ; n++ {
		name := "assert.go"
		if n > 1 {
			name = fmt.Sprintf("assert%d.go", n)
		}
		path := filepath.Join(directory, name)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if os.IsExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		_, err = file.WriteString(source)
		closeErr := file.Close()
		if err != nil {
			return "", err
		}
		if closeErr != nil {
			return "", closeErr
		}
		return path, nil
	}
}
