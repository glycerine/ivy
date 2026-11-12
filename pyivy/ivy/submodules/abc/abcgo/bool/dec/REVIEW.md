# Decomposition review

Separate compactness/readability review: preserve shared graph operations,
use stable node indices to avoid quadratic printer lookup, and keep backend
conversion algorithms behind narrow interfaces to avoid import cycles.

Separate bug review: corrected default printer names to match ABC conventions
and reject malformed SOP input. Added printer and malformed-input regressions.
The original factoring/truth/NPN/network tests remain active. Main-thread
`go test -race ./bool/dec` passes after these fixes. Concrete consumers continue
porting; passing DEC tests does not establish full ABC completion.
