package pytesthelper

import _ "embed"

//go:embed sidecar.py
var SidecarDotPy []byte

//go:embed ivy_ast_dump.py
var IvyAstDumpDotPy []byte
