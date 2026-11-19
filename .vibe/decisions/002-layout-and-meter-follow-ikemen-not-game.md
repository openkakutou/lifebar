---
date: 2026-11-19
status: accepted
---
# Layout, frame and meter follow Ikemen GO, not the code ported from `game`
**Context:** Item 002 moved element placement, bar fill, animation frames, screen scale and the trailing life meter from `game/native/internal/lifebar` into this library. Checked against Ikemen GO (@54d4148a), that code diverged in several places.
**Decision:** This library follows Ikemen GO. Divergences from `game`'s code, each now covered by a test:
1. A mirrored element (facing -1) is shifted one unit before scaling (`common.go:1462-1467`; unit `fnt_scale`, `fightscreen.go:4846`). `game`'s mirrored test value 45.5 becomes 46.
2. A frame's X offset is multiplied by layout scale × facing × the frame's H flip (`anim.go:1072-1073`, `1091-1092`); `game` used the facing sign only.
3. vfacing and V/HV frame flips apply on Y (`anim.go:1073`, `1092`); `game` ignored them.
4. Sprites are shifted by `gameHeight-240` (`common.go:1472`) and centred by `(gameWidth-320)/2`: Ikemen passes it as the draw origin (`common.go:1475` → `anim.go:1146`); bar clips add it too (`fightscreen.go:715`).
5. Scale ports `FightScreen.setScale` (`fightscreen.go:5972`) and `getViewport` for any game size, with offsetX and portraitScale; `game` assumed a 4:3 game.
6. Bar clips use `calcBarFillRect` (`fightscreen.go:418`) with offset, mid position and screen scale (`fightscreen.go:717-730`).
7. Animation frames: a last frame of time -1 holds (`anim.go:445`), frames of time ≤ 0 elsewhere are skipped (`anim.go:753`), a LoopStart outside the frames is 0 (`anim.go:440`), a blank frame shows nothing (`anim.go:604`).
8. The meter uses the bar's own `mid.freeze`, `mid.delay`, `mid.mult`, `mid.steps` (`game` hard-coded 30/1/8), adds the `mid.freeze = 0` branch triggered on a new hit (`fightscreen.go:570-575`), and starts its front fill at 0 like `newLifeBar` (`game` started at 1). It computes in float32, as Ikemen GO does.
**Reason:** Roadmap decision `041`: Ikemen GO's observable behavior is the specification; where OpenKakutou differs, OpenKakutou is wrong.
**Rejected alternatives:** Porting `game`'s code as is and fixing it later — `game` 061 would then inherit the same divergences through the library.
