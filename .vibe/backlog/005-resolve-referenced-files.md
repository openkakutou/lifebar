---
status: todo
depends_on: [001]
---
# Resolve the files referenced by [Files]

## Description
Load the sprite file through `sff`, fonts through `fnt`, the lifebar sound and fightfx through `snd`/`sff`, with Ikemen GO's search paths (relative to the def, then `data/`). Report a missing file by name.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO, validated against real lifebar packs; every rule grounded in Ikemen GO's source (`fightscreen.go`, cited by file:line and commit).

## Acceptance Criteria
- [ ] A real lifebar folder resolves every referenced file
- [ ] A missing file is reported with its exact name, never a panic
