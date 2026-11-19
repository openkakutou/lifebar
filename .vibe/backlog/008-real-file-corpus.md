---
status: todo
depends_on: [001]
---
# Real-file corpus test

## Description
Parse every lifebar of the real corpus already used by `character`/`stage` (local, not redistributed, roadmap decision `044` for committed fixtures) and report what fails.

🔒 MUST: MUGEN 1.0/1.1 and Ikemen GO, validated against real lifebar packs; every rule grounded in Ikemen GO's source (`fightscreen.go`, cited by file:line and commit).

## Acceptance Criteria
- [ ] A test or command parses the whole corpus and lists failures by file
