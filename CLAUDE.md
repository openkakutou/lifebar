# CLAUDE.md — lifebar

> Edit freely. Add `<!-- keep -->` on a section heading to preserve it on regeneration by /vibe:init.

## Project overview

OpenKakutou `lifebar` is a Go library (reading today, writing planned in backlog item 006) for MUGEN/Ikemen GO lifebar files (`fight.def` and the files its `[Files]` section references), decided in roadmap decision `046`, which replaces the separate parsers of `lifebar-viewer-web`, `lifebar-editor` and `game`. It also owns the lifebar behavior those consumers must share: element layout, bar fill spans and the trailing life meter, as Ikemen GO's `fightscreen.go` does them. Consumers: `game` (native and web, through its shared core, roadmap decision `045`), `lifebar-viewer-web`, `lifebar-editor`. Depends on `character` (its `air` package, for the `[Begin Action]` animations); `sff` (sprites), `fnt` (fonts) and `snd` (sounds) are planned with backlog item 005. It compiles to WASM and has no rendering dependency, like `stage`.

**Stack:** Go 1.26, compiled to WASM (`GOOS=js GOARCH=wasm`) for web consumers, imported as a Go module by native ones.
**Type:** library

## Project language

English — all documentation, backlog items, code comments, and other generated content for this project must be written in English.

## Compatibility target (MUST) <!-- keep -->

Every feature and fix must be compatible with **both MUGEN 1.0/1.1 and Ikemen GO**, extensions included, and validated against real community files — never MUGEN-only, never from memory. Where the two engines diverge, record the divergence and the chosen behavior in an ADR, and ground engine-behavior claims in a source (fixture, doc, engine source), marking anything unverified. This is an org-wide rule (`roadmap/repos.md`), already settled: do not ask the user whether it applies, and carry it as a 🔒 MUST in every need framing, implementer prompt and review.

**Reference for verification:** Ikemen GO, which itself targets MUGEN compatibility. Behavior is checked against Ikemen GO (its source, and end-to-end differential tests running the same content in both, roadmap decision `038`), never by a separate MUGEN check. This applies to content that works in Ikemen GO (characters, stages, lifebars); OpenKakutou-only extensions (e.g. 3D characters, decision `028`) have no reference and are exempt.

## Release propagation <!-- keep -->

Consumers pin an exact tag (WASM download or `go.mod`), propagated by this repo's own release step, as in `sff`/`character` (roadmap decisions `016`, `025`). After `/vibe:release`, bump every consumer listed in `roadmap/repos.md`, run its tests, commit and push; stop and flag on red.

## Development workflow (Vibe Coding)

The user is the **Product Owner only**. They describe requirements and evaluate outcomes — they do NOT write or manually test code.

### Definition of Done

A task is complete ONLY when ALL of the following are true:

- [ ] Implementation code is written
- [ ] Tests covering the nominal path exist and pass
- [ ] Tests covering edge cases and error paths exist and pass
- [ ] Lint exits with code 0 — `gofmt -l .` reports no files, `go vet ./...` is clean
- [ ] All tests pass — `go test ./...`
- [ ] No debug artifacts left in code

**Never present a result to the user if tests are failing.**

### TDD workflow

1. Write tests first, 2. confirm they fail, 3. write the minimum implementation, 4. confirm they pass, 5. refactor, 6. run `gofmt -l .` and `go vet ./...`, 7. present the result with the test output.

### Self-correction loop

If tests or lint fail: diagnose, fix, re-run, up to 3 attempts; escalate only on an ambiguous requirement.

## Testing conventions

- Tests co-located with source, `*_test.go`, one test file per source file, built-in `go test`
- Test names describe behavior: `TestLoad_ReturnsError_OnMissingFile`
- Always cover: happy path + at least 2 edge cases + error/invalid input
- Round-trip tests for any write-path code

## Constraints

- Never leave dead code or unused imports
- No rendering dependency; must compile cleanly to WASM (`GOOS=js GOARCH=wasm`)
- Read and write paths separated: write-only concerns never leak into read-path types

## Review agents

| Agent | Active | Reason |
|---|---|---|
| `vibe:review-tests` | ✅ | always active |
| `vibe:review-naming` | ✅ | always active |
| `vibe:review-security` | ✅ | always active |
| `vibe:review-dependencies` | ✅ | always active |
| `vibe:review-robustness` | ✅ | always active |
| `vibe:review-hygiene` | ✅ | always active |
| `vibe:review-antipatterns` | ✅ | always active |
| `vibe:review-simplicity` | ✅ | always active |
| `vibe:review-overengineering` | ✅ | always active |
| `vibe:review-correctness` | ✅ | always active |
| `vibe:review-comments` | ✅ | always active |
| `vibe:review-signatures` | ✅ | always active |
| `vibe:review-coordination` | ✅ | always active |
| `vibe:review-docs` | ✅ | active: Markdown documentation beyond the README (`.vibe/`, ADRs) |
| `vibe:review-ddd` | ❌ | inactive: no explicit domain layer |
| `vibe:review-performance` | ❌ | inactive: library |
| `vibe:review-web-security` | ❌ | inactive: no HTTP surface |
| `vibe:review-hexagonal` | ❌ | inactive: no declared hexagonal architecture |
