---
status: todo
---
# Each side's win call read as Ikemen GO's readATS

## Description
Found by the audit of 2026-11-21 (high, checked against fightscreen.go:3457-3468). `playerKeys` (`parse.go:618`) merges `win.*` with `p1.win.*` / `p2.win.*` key by key. Ikemen GO does not merge: when `pN.win.text`, `spr` or `anim` exists it reads the whole call from `pN.win.*`, otherwise from `win.*` alone. A side's call then inherits `displaytime`, `snd` or `font` from the shared keys where Ikemen GO gives the defaults, or takes a player key Ikemen GO ignores. `TestParse_ReadsEachSidesWinCall` (parse_test.go:855) pins the merge. Introduced by v0.2.0.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] A side's call comes wholly from `pN.win.*` when one of its text, spr or anim keys exists, else wholly from `win.*`
- [ ] The test expectation follows Ikemen GO; the same rule is checked for the other `readATS` / `readLayout` users
