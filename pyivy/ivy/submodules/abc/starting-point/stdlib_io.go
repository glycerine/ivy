package startingpoint

import (
	"strconv"
	"strings"
)

func (p *libraryPass) convertIO(c libraryCall) (string, string, bool) {
	a := c.args
	switch c.name {
	case "fopen":
		if len(a) != 2 {
			return "", arity(c, 2), false
		}
		mode, err := strconv.Unquote(strings.TrimSpace(a[1]))
		if err != nil {
			return "", "Dynamic fopen mode retained; choose OpenFile flags explicitly.", false
		}
		flags, ok := fopenFlags(mode)
		if !ok {
			return "", "Unsupported fopen mode retained (extensions/text translation need review).", false
		}
		os := p.imports.use("os")
		operation := os + "Open(spPath)"
		if flags != "" {
			operation = os + "OpenFile(spPath, " + strings.ReplaceAll(flags, "os.", os) + ", 0666)"
		}
		filename := a[0]
		if text, ok := p.text(filename); ok {
			filename = text
		}
		return strings.ReplaceAll(operation, "spPath", filename), "Literal mode mapped; handle Go file/error results downstream.", true
	case "fclose":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		if c.standalone {
			return "_ = (" + a[0] + ").Close()", "C ignored close status retained; consider handling the Go error.", true
		}
		os := p.imports.use("os")
		return iife("spFile *"+os+"File", "int32", "if spErr := spFile.Close(); spErr != nil { return -1 }; return 0", a[0]), "Close error maps to -1; errno not reproduced.", true
	case "remove", "rename":
		n := 1
		if c.name == "rename" {
			n = 2
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		for i := range a {
			if text, ok := p.text(a[i]); ok {
				a[i] = text
			}
		}
		os := p.imports.use("os")
		params, call := "spPath string", os+"Remove(spPath)"
		if n == 2 {
			params = "spOld, spNew string"
			call = os + "Rename(spOld, spNew)"
		}
		return iife(params, "int32", "if spErr := "+call+"; spErr != nil { return -1 }; return 0", a...), "Filesystem error converted to C status; errno and cross-platform rename differences require review.", true
	case "fgets":
		if len(a) != 3 {
			return "", arity(c, 3), false
		}
		buffer, ok := p.bytes(c.rawArgs[0])
		if !ok {
			return "", "fgets needs a known byte array/slice; string mutation or raw-pointer capacity requires a representation decision.", false
		}
		os, io := p.imports.use("os"), p.imports.use("io")
		body := `if spLimit <= 0 || spLimit > len(spBuffer) { return nil }
if spLimit == 1 { spBuffer[0] = 0; return spBuffer[:0] }
spN := 0
for spN < spLimit-1 {
 spRead, spErr := spFile.Read(spBuffer[spN:spN+1])
 if spRead > 0 {
  spN++
  if spBuffer[spN-1] == '\n' { break }
 }
 if spErr != nil {
  if spErr != $IOEOF { return nil } // Review explicit I/O error handling.
  break
 }
 if spRead == 0 { break }
}
if spN == 0 { return nil }
spBuffer[spN] = 0
return spBuffer[:spN]`
		body = strings.ReplaceAll(body, "$IOEOF", io+"EOF")
		return iife("spBuffer []byte, spLimit int, spFile *"+os+"File", "[]byte", body, buffer, "int("+a[1]+")", a[2]), "Bounded byte read preserves newline, NUL, partial final line and no read-ahead. Review errno/ferror; replace byte-at-a-time loop with coordinated buffering later.", true
	case "fgetc", "getc", "getchar":
		n := 1
		if c.name == "getchar" {
			n = 0
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		os := p.imports.use("os")
		file := os + "Stdin"
		if n == 1 {
			file = a[0]
		}
		return iife("spFile *"+os+"File", "int32", "var spByte [1]byte\nspN, _ := spFile.Read(spByte[:])\nif spN == 1 { return int32(spByte[0]) }; return -1", file), "Byte value or EOF (-1); inspect error separately if the C code uses ferror.", true
	case "fputc", "putc", "putchar":
		n := 2
		if c.name == "putchar" {
			n = 1
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		os := p.imports.use("os")
		file := os + "Stdout"
		if n == 2 {
			file = a[1]
		}
		return iife("spChar int32, spFile *"+os+"File", "int32", "spByte := byte(spChar)\nspN, spErr := spFile.Write([]byte{spByte})\nif spErr != nil || spN != 1 { return -1 }; return int32(spByte)", "int32("+a[0]+")", file), "Character converted to unsigned byte; -1 on write failure.", true
	case "puts", "fputs":
		n := 1
		if c.name == "fputs" {
			n = 2
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		text := a[0]
		if v, ok := p.text(text); ok {
			text = v
		}
		if c.name == "puts" {
			fm := p.imports.use("fmt")
			if c.standalone {
				return fm + "Println(" + text + ")", "Newline appended; ignored output status retained.", true
			}
			return iife("spText string", "int32", "_, spErr := "+fm+"Println(spText)\nif spErr != nil { return -1 }; return 0", text), "Nonnegative success / negative failure retained.", true
		}
		io, os := p.imports.use("io"), p.imports.use("os")
		return iife("spText string, spFile *"+os+"File", "int32", "_, spErr := "+io+"WriteString(spFile, spText)\nif spErr != nil { return -1 }; return 0", text, a[1]), "No newline appended; return value preserves success/failure sign.", true
	case "fflush":
		if len(a) != 1 {
			return "", arity(c, 1), false
		}
		file := c.rawArgs[0]
		if p.kind(file) != "file" && file != "stdout" && file != "stderr" && file != "stdin" && !strings.HasSuffix(file, ".Stdout") && !strings.HasSuffix(file, ".Stderr") {
			return "", "fflush retained: establish whether the stream is buffered, input, or flush-all before choosing behavior (Sync is not fflush).", false
		}
		os := p.imports.use("os")
		return iife("_ *"+os+"File", "int32", "// os.File has no user-space output buffer. Do not substitute Sync.\nreturn 0", a[0]), "No-op for the chosen unbuffered os.File representation; revise if bufio.Writer is introduced.", true
	case "fseek", "ftell", "rewind":
		n := 1
		if c.name == "fseek" {
			n = 3
		}
		if len(a) != n {
			return "", arity(c, n), false
		}
		os, io := p.imports.use("os"), p.imports.use("io")
		if c.name == "ftell" {
			return iife("spFile *"+os+"File", "int64", "spPos, spErr := spFile.Seek(0, "+io+"SeekCurrent)\nif spErr != nil { return -1 }; return spPos", a[0]), "LP64 position result; errno not reproduced.", true
		}
		whence := io + "SeekStart"
		offset := "0"
		if n == 3 {
			suffix, ok := map[string]string{"SEEK_SET": "SeekStart", "SEEK_CUR": "SeekCurrent", "SEEK_END": "SeekEnd"}[strings.TrimSpace(c.rawArgs[2])]
			if !ok {
				return "", "fseek origin must be a known SEEK_* constant.", false
			}
			whence = io + suffix
			offset = "int64(" + a[1] + ")"
		}
		if c.name == "rewind" {
			return iife("spFile *"+os+"File", "", "_, _ = spFile.Seek(0, "+whence+")", a[0]), "Reposition only; C stream error flags do not exist in this draft representation.", true
		}
		return iife("spFile *"+os+"File, spOffset int64", "int32", "_, spErr := spFile.Seek(spOffset, "+whence+")\nif spErr != nil { return -1 }; return 0", a[0], offset), "Seek status retained; buffering/error indicators need review.", true
	case "fread", "fwrite":
		if len(a) != 4 {
			return "", arity(c, 4), false
		}
		buffer, ok := p.bytes(c.rawArgs[0])
		if !ok {
			return "", "Binary I/O requires a known byte buffer; raw structures and element representations remain for review.", false
		}
		os := p.imports.use("os")
		operation := "spFile.Write(spBuffer[:spSize*spCount])"
		if c.name == "fread" {
			operation = p.imports.use("io") + "ReadFull(spFile, spBuffer[:spSize*spCount])"
		}
		body := "if spSize <= 0 || spCount <= 0 { return 0 }\nif spCount > len(spBuffer)/spSize { panic(\"C binary I/O exceeds known buffer\") }\nspN, _ := " + operation + "\nreturn uint64(spN / spSize)"
		return iife("spBuffer []byte, spSize, spCount int, spFile *"+os+"File", "uint64", body, buffer, "int("+a[1]+")", "int("+a[2]+")", a[3]), "Complete element count retained; ignored error/EOF must be integrated if status is inspected elsewhere.", true
	}
	return "", "", false
}
func fopenFlags(mode string) (string, bool) {
	switch mode {
	case "r", "rb":
		return "", true
	case "r+", "rb+", "r+b":
		return "os.O_RDWR", true
	case "w", "wb":
		return "os.O_WRONLY | os.O_CREATE | os.O_TRUNC", true
	case "w+", "wb+", "w+b":
		return "os.O_RDWR | os.O_CREATE | os.O_TRUNC", true
	case "a", "ab":
		return "os.O_WRONLY | os.O_CREATE | os.O_APPEND", true
	case "a+", "ab+", "a+b":
		return "os.O_RDWR | os.O_CREATE | os.O_APPEND", true
	case "wx", "wbx":
		return "os.O_WRONLY | os.O_CREATE | os.O_TRUNC | os.O_EXCL", true
	case "w+x", "wb+x", "w+bx":
		return "os.O_RDWR | os.O_CREATE | os.O_TRUNC | os.O_EXCL", true
	}
	return "", false
}
