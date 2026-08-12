<!-- galdr:start -->
<galdr-config path="docs/agents/galdr.md" />
<!-- galdr:end -->

## Docker build context

`Dockerfile` copies the whole repo (`COPY . .`) and `deploy.yml` runs `docker compose up -d --build` on every push to master with no path filter, so anything not excluded in `.dockerignore` that changes will rebuild the image and restart the live bot.
Keep `.dockerignore` in sync when adding files the Go build does not need (see `TestDockerignoreExcludesDocsButKeepsGoBuildInputs` in `main_test.go`, which fails if a doc file is left in or a Go build input is excluded).

## Maintaining this file

Keep this file for knowledge useful to almost every future agent session in this project.
Do not repeat what the codebase already shows; point to the authoritative file or command instead.
Prefer rewriting or pruning existing entries over appending new ones.
When updating this file, preserve this bar for all agents and keep entries concise.
