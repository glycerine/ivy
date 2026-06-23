## transliteration from Go into Typescript guidance

The ivy/gojr (Go-junior) directory implements a fully faithful
mechanical transliteration of the /usr/local/go1.27rc1/src/go
Go standard library from Go into Typescript.

This is not a rewrite. This is a transliteration. Preserve 
the structure, the ordering, the decomposition, and the naming. 
Your creative judgment is not needed and not wanted here.

Before translating, list every symbol, function, method, production rule 
in the source. Then translate each one in order. After translating, 
confirm that every item in your inventory appears in the output.

Your job is not to understand or explain the source Go code. Your 
job is to produce a complete mechanical translation into Typescript with 
no omissions. Treat this like a human translator translating 
a legal contract -- every clause must appear in the output, 
even redundant or awkward ones.

If the user says "port" this should be taken as the transliteration
contract-like mechanical converstion described above. There no
room for deviation from the original source logic or naming
in this "port". Exact behavior and symbol level naming must
be preserved.

## on completeness

Go wide, finish an entire layer when translating. Do not defer work or bugs.
When encountering bugs, fix them immediately. When encounting work
that needs doing as a dependency, pause the current task and implement
the dependency, then return to the current task. Otherwise we risk
difficult to detect heisenbugs from quietly missing functionality.
Do not create new TODOs, rather work on them immediately.
