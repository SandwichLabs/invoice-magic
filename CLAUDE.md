# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Runtime dependency

`invgen` shells out to the `typst` CLI for every render (`internal/render/renderer.go`). Without `typst` on PATH, both the `render` command and the `serve` web preview fail at runtime — not at build time. `render.CheckTypstInstalled()` is the preflight.

## Common commands

Build/test/lint are driven by Taskfile.yml:

- `task build` (alias `b`) — `go build -o invgen ./cmd/invgen`
- `task test` (alias `t`) — runs `go test -v` across all packages; depends on `task mod`
- `task test:coverage` — writes `coverage.out` + `coverage.html`
- `task lint` (alias `l`) / `task lint:fix` — golangci-lint
- `task run:render` — builds and renders `testdata/sample_invoice.json` to `output/test.pdf` (handy smoke test)
- `task default` — lint + test + build
- Single test: `go test -v ./internal/template -run TestManager_List`

Releases go through goreleaser (`task build:all`, `task release:<version>`); the release task enforces `main` branch + clean tree.

## Architecture

This is a Cobra CLI (`cmd/invgen` → `internal/cli`) with two distinct modes sharing the same render core:

1. **CLI render path**: `invgen render` reads JSON (file or stdin) → `internal/model.Invoice` for validation → `internal/render.Renderer` writes the JSON to a temp file and invokes `typst compile --root / --input data=<path> <template.typ> <output>`. Templates parse `sys.inputs.data` via Typst's `json()`. HTML output adds `--features html --format html`.

2. **Web/Sheets path**: `invgen serve` starts a chi HTTP server (`internal/web`) that pulls invoice rows from Google Sheets (`internal/sheets`), maps them to the same `model.Invoice` via configurable field mappings in `config.yaml` (`sheets.fields`), then renders through the same `render.Renderer`. The server also serves an HTMX-style preview UI from `web/templates` + `web/static`.

Google auth (`internal/auth`) is OAuth2 with token caching. `DefaultScopes()` is read-only; `--write` on `serve` (and the `provision` command) escalates to `AllScopes()` with automatic scope upgrade when an existing token is read-only. Credentials/token paths come from `google.credentials_file` / `google.token_file` in config. `invgen auth` must run before `serve`.

The `provision` command / `/provision` web route writes header rows back to a sheet so users can bootstrap a sheet to match the expected field mapping — this is why write scope exists.

`internal/template.Manager` is a thin filesystem wrapper around `templates/*.typ`; `template init` seeds new templates from the default. Template name resolution happens in the manager, not the renderer.

## Configuration

`config.yaml` is loaded via Viper from the working directory (or `--config`). It holds template/output dirs, Google credentials paths, and — importantly — the `sheets.fields` map that translates sheet column headers (or letters) into invoice fields. Changing this map is how users adapt `serve` to arbitrary sheet layouts without code changes.

Totals in `sheets.fields` are optional: if `total_net`/`total_tax`/`total_gross` are blank, they're computed from line items; otherwise the sheet values win. Keep that fallback in mind when editing the sheets source.
