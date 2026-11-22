---
status: done
---
# Each side's win call read as Ikemen GO's readATS

## Description
Found by the audit of 2026-11-21 (high, checked against fightscreen.go:3457-3468). `playerKeys` (`parse.go:618`) merges `win.*` with `p1.win.*` / `p2.win.*` key by key. Ikemen GO does not merge: when `pN.win.text`, `spr` or `anim` exists it reads the whole call from `pN.win.*`, otherwise from `win.*` alone. A side's call then inherits `displaytime`, `snd` or `font` from the shared keys where Ikemen GO gives the defaults, or takes a player key Ikemen GO ignores. `TestParse_ReadsEachSidesWinCall` (parse_test.go:855) pins the merge. Introduced by v0.2.0.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [x] A side's call comes wholly from `pN.win.*` when one of its text, spr or anim keys exists, else wholly from `win.*`
- [x] The test expectation follows Ikemen GO; the same rule is checked for the other `readATS` / `readLayout` users

## Done
2026-11-22 - v0.3.1: `announcementPrefix` picks the player or shared prefix as readATS does (`TestParse_ReadsEachSidesWinCall`, `TestParse_APlayerKeyWithoutTextSprOrAnimIsIgnored`, `TestParse_APlayerSpriteSelectsThePlayerKeys`). The other readATS / readLayout users (`win.top`, `win.bgN`, `win2`-`win4`, `ai.win`, `ai.lose`, fightscreen.go:3483-3518) are not read by the library yet; when they are, they go through the same helper.
