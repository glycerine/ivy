#!/bin/sh
# Truth and codec integration is now included in the completed GIA core suite.
set -eu
exec "$(dirname "$0")/test_ported_core.sh" "$@"
