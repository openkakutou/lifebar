---
status: todo
depends_on: [001, 002]
---
# WASM entrypoint with a release smoke gate

## Description
Roadmap decision `046`: `cmd/wasm` exposing parse, layout and meter to JavaScript (contract style of `sff`/`fnt`), with `cmd/wasm/smoke.mjs` run by the release workflow.

## Acceptance Criteria
- [ ] `lifebar-viewer-web` can load a lifebar through the module
- [ ] The release workflow fails if the smoke test fails
