# ivy2go live audit / TODO log

Created: 2026-05-25 06:06:56 UTC

A living catalogue of every known gap between `ivy2go` and `ivy2cpp`.
Items are added as discovered; items are marked `DONE` when fixed.

Format mirrors `ivy2cpp/AUDIT2_*.md`:

> ## STATUS NNN — short title
> Created: YYYY-MM-DD HH:MM:SS UTC
>
> **Gap.**
> **Python references.** (file:line)
> **ivy2cpp references.** (file:line)
> **Go (ivy2go) plan.** (file:line "land here")
> **Verification.** (test name)

Statuses: `OPEN`, `IN PROGRESS`, `DONE`.

---

## OPEN 001 — every milestone M1–M10

Created: 2026-05-25 06:06:56 UTC

**Gap.** The entire ivy2cpp surface (32,913 LOC, 200+ tests). This
single umbrella item tracks the milestone roadmap in
`ARCHITECTURE_TODO.md` §4. Per-milestone gaps will be split out as
individual numbered items when they're spotted in implementation.

**ivy2cpp references.** All 38 production files + 18 test files.

**ivy2go plan.** Sequential M0 → M10 in `ARCHITECTURE_TODO.md`.

**Verification.** Acceptance gate in `ARCHITECTURE_TODO.md` §5.

---

(Additional items appear here as discovered.)
