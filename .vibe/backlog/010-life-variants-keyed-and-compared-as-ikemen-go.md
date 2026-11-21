---
status: todo
---
# Life bar variants keyed and compared as Ikemen GO

## Description
Found by the audit of 2026-11-21 (medium, to re-check against fightscreen.go). Introduced by v0.3.0:
- For the life bar, `p1.bg150.*` is read as a bg1 variant from 50 % (`parse.go:385`). Ikemen GO's `readMultipleValuesF` (fightscreen.go:378-396) takes the second run of digits ("150"), so the variant is keyed 150 and never shows. bg0 and the power bar are not affected.
- `Pick` compares a float64 percentage (`model.go:143-149`); Ikemen GO compares the float32 ratio `life >= k/100` (fightscreen.go:607, 651). Life 290/1000 passed as `float64(float32(0.29))*100` is 28.999…, so `front29` is missed where Ikemen GO shows it.

🔒 MUST: Ikemen GO behaviour (roadmap decision `038`).

## Acceptance Criteria
- [ ] Life variant keys are read as `readMultipleValuesF` reads them, with a test on `bg1` and `bg2` keys
- [ ] A life exactly on a threshold picks the same variant as Ikemen GO; `game`'s caller passes what `Pick` needs
