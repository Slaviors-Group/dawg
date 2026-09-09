# Contributing

Thank you for your interest in contributing to DAWG! This guide covers development setup, coding standards, and the PR workflow.

## Development Setup

### Prerequisites

| Tool | Version | Purpose |
|---|---|---|
| **Go** | 1.22+ | Engine compilation |
| **Node.js** | 20+ | Desktop app & Playwright |
| **Rust** | stable | Tauri desktop shell |
| **Python** | 3.10+ | mitmproxy |
| **Biome** | 1.9.4 | Formatting & linting |

### Clone & Build

```bash
# Clone the repository
git clone https://github.com/Slaviors-Group/dawg.git
cd dawg

# Build the Go engine
cd engine
go build -o ../bin/dawg ./cmd/dawg
npm install  # Playwright dependencies
go test ./...

# Build the desktop app
cd ../desktop
npm install
npm run dev
```

### Pre-commit Hooks

DAWG uses pre-commit hooks for code quality:

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gofmt
        name: gofmt
        entry: gofmt -w
        language: system
        files: \.go$
      - id: go-vet
        name: go vet
        entry: go vet
        language: system
        files: \.go$
      - id: biome
        name: biome
        entry: npx biome check --write
        language: system
        files: \.(ts|tsx|css|json)$
```

Install pre-commit:

```bash
pip install pre-commit
pre-commit install
```

---

## Project Structure

```
dawg/
├── engine/          # Go Engine CLI (core logic)
│   ├── cmd/dawg/    # CLI entry point (Cobra)
│   └── internal/    # Pipeline components
├── desktop/         # Tauri v2 Desktop Shell
│   ├── src/         # React 19 frontend
│   └── src-tauri/   # Rust backend
├── schema/          # OCI Manifest + OPA policies
├── web/             # VitePress documentation
└── test/            # E2E smoke tests
```

### Language-per-Concern Rule

| Concern | Language | Location |
|---|---|---|
| Engine CLI | Go | `engine/` |
| Desktop shell | TypeScript/React | `desktop/src/` |
| Desktop backend | Rust | `desktop/src-tauri/` |
| Schema | JSON Schema + Rego | `schema/` |

Do not mix languages across concerns. Each component has a clear boundary.

---

## Coding Standards

### Go (Engine)

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use `slog` for structured logging with `component` + `artifact_id` fields
- Write table-driven tests
- Keep packages focused — one responsibility per package
- No external dependencies unless approved (check `DECISIONS.md`)

### TypeScript/React (Desktop)

- Use Biome for formatting (100 char line width, 2-space indent)
- Functional components with hooks
- Props interfaces defined inline or co-located
- Tailwind CSS for styling (no inline styles)
- Framer Motion for animations

### Rust (Tauri Backend)

- Standard `cargo fmt` and `cargo clippy`
- Keep IPC commands thin — delegate to CLI or Rust libraries
- Error handling with `Result` types

---

## Testing

### Engine Tests

```bash
cd engine
go test ./...
```

All Go packages have corresponding `_test.go` files. Aim for:

- Unit tests for pure functions
- Integration tests for pipeline components
- Table-driven tests where applicable

### Desktop Tests

```bash
cd desktop
npm run lint
```

### End-to-End Smoke Test

```powershell
.\test\smoke_test.ps1
```

### Writing New Tests

- Place tests in the same package as the code
- Use descriptive test names: `TestSanitize_FieldNameMatch`
- Use table-driven patterns for multiple cases
- Mock external dependencies (Docker, filesystem) where possible

---

## Pull Request Workflow

### 1. Create a Branch

```bash
git checkout -b feature/my-feature staging
```

Branch naming:
- `feature/description` — new features
- `fix/description` — bug fixes
- `docs/description` — documentation changes

### 2. Make Changes

- Follow the coding standards above
- Add tests for new functionality
- Update documentation if behavior changes

### 3. Verify Locally

```bash
# Engine
cd engine && go test ./...

# Desktop
cd desktop && npm run lint

# Full smoke test
.\test\smoke_test.ps1
```

### 4. Commit

Write clear, concise commit messages:

```
feat: add cassette replay for third-party APIs

- Implement mitmproxy cassette mode for HTTP replay
- Add canonical body fingerprint matching
- Block outbound requests during replay

Closes #42
```

### 5. Open a PR

Target the `staging` branch. Include:

- Description of changes
- Related issue number
- Screenshots (for UI changes)
- Test results

### 6. Review & Merge

- At least 1 approval required
- All CI checks must pass
- Squash merge to keep history clean

---

## Architectural Decisions

All architectural decisions are recorded in `DECISIONS.md` files:

- `agent/implement-p0/core-engine-abim/DECISIONS.md` — Engine decisions
- `agent/implement-p0/desktop-abim/DECISIONS.md` — Desktop decisions
- `agent/implement-integration+bundling/DECISIONS.md` — Integration decisions

If you make an architectural decision, document it in the appropriate file following the format in `Instruction.md`.

---

## Reporting Issues

Found a bug or have a feature request? Open an issue on [GitHub Issues](https://github.com/Slaviors-Group/dawg/issues).

Include:

- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Environment details (OS, Go version, Node version)
- DAWG version (`dawg doctor` output)

---

## License

By contributing, you agree that your contributions will be licensed under the Apache-2.0 License.
