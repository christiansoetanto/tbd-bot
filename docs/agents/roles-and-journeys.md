## Unverified Server Member

**Identity:** A Discord user joining the server who has not yet undergone religious vetting. [domain/domain.go:34-44](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L34-L44)
**Capabilities:** Can trigger vetting verification and answer vetting questions using `/sdverify`. [domain/domain.go:9](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L9)
**Lifecycle states touched:** Unverified -> Questioning -> Verified. [domain/domain.go:9-11](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L9-L11)
**Journey:** Joins server, executes `/sdverify` command, submits answers for secret code detector/vetting, receives assigned religion role. [README.md:5-8](file:///Users/chris/repos/christiansoetanto/tbd-bot/README.md#L5-L8)
**Boundaries:** Cannot post in restricted channels prior to role assignment. [README.md:6](file:///Users/chris/repos/christiansoetanto/tbd-bot/README.md#L6)
**Status:** Built [domain/domain.go:9-11](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L9-L11)

## Verified Server Member

**Identity:** A server member who has been assigned a religion role. [domain/domain.go:36-44](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L36-L44)
**Capabilities:** Can post Q&A questions, vote in polls, view Office of Readings and daily liturgical celebrations. [README.md:9-15](file:///Users/chris/repos/christiansoetanto/tbd-bot/README.md#L9-L15)
**Lifecycle states touched:** Verified. [domain/domain.go:34-44](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L34-L44)
**Journey:** Participates in Q&A discussions, views automated daily Catholic updates, interacts with bot poll components. [README.md:9-15](file:///Users/chris/repos/christiansoetanto/tbd-bot/README.md#L9-L15)
**Boundaries:** Subject to secret code vetting and detainment rules. [domain/domain.go:12](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L12)
**Status:** Built [domain/domain.go:34-44](file:///Users/chris/repos/christiansoetanto/tbd-bot/domain/domain.go#L34-L44)

## Changelog
- 2026-07-22 — doc-wide — initial setup draft
