---
status: todo
depends_on: [001]
---
# Per-value element variants and remaining element keys

## Description
From `game` 059: per-value variants (`front.<n>`, `bg0.<n>`, `red`, `value`, `warn`, `shift`, `scalefill`, `palfx`), `vfacing`, `window`, `doubleres`, and the thresholds at which they switch.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO, validated against real lifebar packs; every rule grounded in Ikemen GO's source (`fightscreen.go`, cited by file:line and commit).

## Acceptance Criteria
- [ ] Variants are read and resolved for a value at the same thresholds as Ikemen GO
