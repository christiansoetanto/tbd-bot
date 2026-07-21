## Gates

### Task
- `gofmt -l .`
- `go vet ./...`
- `go test ./...`

### Gate
- `gofmt -l .`
- `go vet ./...`
- `go test ./...`

## Invariants

## Fast path

## Review sources
- README.md
- docs/agents/roles-and-journeys.md

## Models
- mechanical: claude-haiku-4-5
- standard: claude-sonnet-5
- top: session model

## Worktree notes
- Run `go mod download` after worktree creation.

## Smoke
- launch command: `go run main.go`
- base URL: N/A (Discord bot service)
- output dir: `docs/smoke/`

## Briefs
- dispatch briefs under `docs/briefs/` are gitignored
