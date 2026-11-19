---
status: todo
depends_on: [001]
---
# Write fight.def, preserving the original format

## Description
Write path for `lifebar-editor`: serialize the model back, keeping comments, key order and unknown keys (port of `lifebar-editor/src/lifebar/serialize.ts`).

## Acceptance Criteria
- [ ] Read → write → read round-trips every fixture with identical model and unchanged unrelated lines
