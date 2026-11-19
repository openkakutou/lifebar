---
date: 2026-11-19
status: accepted
---
# Typed model built from an ordered tokenizer; deterministic player selectors
**Context:** Item 001 reads `fight.def` as Ikemen GO does (`src/common.go`, `src/fightscreen.go` @ 54d4148a). Two shapes were possible, and Ikemen GO's `p*.`/`p1|p3.` selector resolution (`common.go:1027-1048`) iterates a Go map, so when a section holds both `p*.pos` and `p1.pos` the winner is random.
**Decision:** `Parse` returns a typed model (`Bar`, `Element`, `Text`, one per player) plus warnings, built from an unexported tokenizer that copies Ikemen GO's rules and keeps section order and line numbers — the `stage`/`character` pattern. For selectors, an explicit `pN.` key always wins over a selector key, and between selector keys the first in the file wins.
**Reason:** Mirrors Ikemen GO's own structures, gives the future WASM API one shape, and keeps the ordered data a format-preserving write (item 006) can build on. The selector rule is the only deterministic choice that matches what an author writing both keys means; Ikemen GO itself is nondeterministic there.
**Rejected alternatives:** An exported generic document plus typed views — a second public shape used only by item 006, which can export the tokenizer later at little cost. A table of key descriptors shared by read and write — a second style next to `stage`'s hand-written parsing, for extensibility items 003/004 do not need yet.
