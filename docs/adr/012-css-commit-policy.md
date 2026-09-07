# ADR 012 — Generated CSS committed; Tailwind binary not committed

## Status

Accepted

## Context

Developers and CI need CSS without requiring Node. Clone-and-run should work without downloading Tailwind first.

## Decision

- Commit minified `web/static/css/app.css`
- Download pinned Tailwind standalone CLI into gitignored `.tools/`
- CI runs `make css-check` (rebuild + `git diff --exit-code`)

## Consequences

PRs that change utility classes must also commit CSS. Binary size of the repo grows slightly with CSS, not with the ~100MB CLI.
