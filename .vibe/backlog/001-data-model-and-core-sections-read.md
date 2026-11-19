---
status: todo
---
# Data model and reading of the core fight.def sections

## Description
Roadmap decision `046`. A pure-data model and `Parse` for `fight.def`: `[Files]` (sprite, font, sound and fightfx references), `[Lifebar]`/`[Powerbar]` bars (bg0-2, mid, front, top, range, layerno, facing, scale), `[Face]`, `[Name]`, `[Time]`, `[WinIcon]`, `[Combo]`, `[Round]`, `[Action]`. Both naming conventions real packs use: classic MUGEN per-player sections (`P1 Life Bar`, …) and Ikemen GO generic ones (`Lifebar`, …). Tolerate malformed files as Ikemen GO does (decision `041`).

Start from what the existing readers already handle and their tests: `lifebar-viewer-web/src/lifebar/` (section families, Ikemen GO mode prefixes), `game/native/internal/lifebar/` (bars, elements, texts).

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO, validated against real lifebar packs; every rule grounded in Ikemen GO's source (`fightscreen.go`, cited by file:line and commit).

## Acceptance Criteria
- [ ] A real MUGEN lifebar and a real Ikemen GO lifebar (e.g. V-HD Bars) parse into the model
- [ ] Every fixture the existing readers test against parses with the same values
- [ ] Malformed input returns a descriptive error or is tolerated as in Ikemen GO, never a panic
