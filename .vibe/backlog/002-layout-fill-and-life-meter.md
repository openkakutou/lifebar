---
status: todo
depends_on: [001]
---
# Element layout, bar fill spans and the trailing life meter

## Description
Roadmap decisions `045`, `046`: the lifebar behavior every consumer must share moves here, from `game/native/internal/lifebar` (`layout.go` `Place`/`Clips`, `lifebar.go` `FillSpan`/`Scale`/`Frame`, `meter.go` trailing fill). Pure computation: element rectangles, fill spans for a value, animation frame at a tick, the mid/trailing meter step. No drawing.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO, validated against real lifebar packs; every rule grounded in Ikemen GO's source (`fightscreen.go`, cited by file:line and commit).

## Acceptance Criteria
- [ ] `game` native tests for layout and meter pass unchanged against this package
- [ ] The trailing meter steps as in Ikemen GO for a scripted life sequence
