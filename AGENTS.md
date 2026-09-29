# AGENTS.md — plugin-quickshell

Standalone plugin repo for the `quickshell` check verb (`verb:quickshell`). The
plugin is a Go module at `candy/plugin-quickshell/` (module path
`github.com/opencharly/plugin-quickshell/candy/plugin-quickshell`); the root
`charly.yml` declares `discover: candy` so the repo is a project, and the candy
manifest carries the `quickshell-skill` `skill:` entity (the corpus source for
`/charly-check:quickshell`).

Canonical files:

- `charly.yml` — the root project manifest (`discover: candy` only).
- `candy/plugin-quickshell/charly.yml` — the `plugin-quickshell:` candy entity
  (`plugin:` block, `primary:`, `plan:` check) + the `quickshell-skill` skill
  entity.
- `candy/plugin-quickshell/methods.go` — the `ping`/`call` methods.
- `candy/plugin-quickshell/provider.go` / `plugin.go` — the provider +
  `NewMeta()`.
- `candy/plugin-quickshell/schema/quickshell.cue` — the self-contained schema.
- `.github/workflows/tag-on-merge.yml` — CalVer tag + `CHANGELOG/` on merge.
- `README.md` — user overview only; never agent guidance.

## Load these skills first (R0)

- `/charly-internals:plugin` — the plugin authoring reference: the `plugin:`
  block, the unified Provider model, the out-of-process shape, the per-plugin
  CUE-schema contract, placement. Load before touching the provider or schema.
- `/charly-check:quickshell` — the `quickshell:` verb this candy serves.
- `/charly-check:check` — the check orchestrator, beds and the R10 sequence.
- `/charly-internals:git-workflow` — before any git/PR action.

## Build / validate / test

- `go build ./...` in `candy/plugin-quickshell/` — compile the plugin module.
- `go test ./...` in `candy/plugin-quickshell/` — the plugin's Go tests
  (`methods_test.go`).
- `charly box validate` at the repo root — the structural check (the candy +
  `plugin:` block, CUE schema, the skill entity).
- The merge gate is the **org-wide** `charly/pr-validator` (required check
  `validate / validate`, defined in `opencharly/.github`); this repo has **no**
  per-repo candy gate.
- R10 consumer: a Quickshell-desktop bed whose check composes this plugin.

## Modify this repo

- Edit the `plugin-quickshell:` candy entity, the Go source, and
  `schema/quickshell.cue` **together** — the schema is the single source for the
  `params/` struct, so a field change not mirrored in the schema desyncs the
  generated types.
- The `primary:` declaration is load-bearing: charly's pre-parse reads it off the
  raw YAML, so the scalar form (`quickshell: ping`) only parses when the candy
  manifest declares it.
- The `skill:` entity is the corpus source for `/charly-check:quickshell`; a
  change to the verb surface belongs in BOTH the candy and the skill body.

## Landing

- PR-only. Every change lands through a pull request; the org-required
  `charly/pr-validator` validates the diff and body and arms native auto-merge on
  PASS. Direct pushes to `main` are blocked.
- History lives in `CHANGELOG/` (written by `tag-on-merge` at merge time); the PR
  body IS the changelog.
- The authoritative rulebook is the umbrella `AGENTS.md` in
  `opencharly/opencharly` and `charly/AGENTS.md` in the charly repo. Do not
  restate its rules here.
