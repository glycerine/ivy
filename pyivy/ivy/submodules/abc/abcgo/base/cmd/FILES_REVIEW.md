# Optional directory command API group

## Readability pass

Ported scandir, numeric rename, listing and script generation with explicit
paths and os.ReadDir/filepath instead of process-wide chdir and Windows find
handles. The original Windows-only registration is retained; functions are
available for testing and explicit use on every platform. The bzip2 extraction
helper uses the standard library's streaming decompressor.

## Separate bug and fidelity pass

Preserved filename-number sorting, output-number width, scan-token priority,
.mv filtering, script/txt exclusion, default script name, and current/target
script location. Generated commands quote filenames so spaces remain one argv.
Directory entries are omitted from generated scripts instead of being passed to
the network reader. Rename preflights duplicate and existing destinations;
regression verifies rejected collisions preserve both source files. Numeric
overflow and IO errors propagate instead of overflowing fixed source buffers.

Tests exercise actual directory scans, numeric renames and generated script
parsing in temporary directories. Full base/cmd race tests and vet pass.
