# Original ABC hierarchy flatten oracle

Generated with the repository's original `./abc` executable on 2026-11-12:

    ./abc -c 'read_blif abcgo/base/abc/test_vectors/hierarchy_reused.blif; write_blif /tmp/abc-hierarchy-oracle.blif'

The source reports two flattened logic instances. `hierarchy_flat_oracle.blif`
is its output after removing only the timestamp comment. A Go regression compares
the exact local SOP and ordered fanin-net names for both module instances.
The original executable is not required to run Go tests.
