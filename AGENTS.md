# RideForge Codex Instructions

You are working in RideForge, a Go ride-hailing platform with real-time location,
occupancy, demand, and dispatch workloads. These instructions apply to the whole
repository.

## Before substantial work

1. Read `docs/STATUS.md` first.
2. Read `docs/RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md` and load the
   documentation required by that task.
3. Inspect relevant ADRs under `docs/adr/` before changing architecture,
   persistence, events, dispatch, security, or reliability behavior.
4. Inspect the current implementation before assuming the documented target
   architecture is already complete.
5. Check `git diff` and `git status`; preserve unrelated user changes.

## Architecture invariants

- `internal/domain`: pure business rules and state; no I/O or framework imports.
- `internal/application`: use cases and orchestration; depend on ports only.
- `internal/ports`: small inbound and outbound contracts.
- `internal/infrastructure`: concrete Redis, PostgreSQL, Kafka, HTTP, and
  observability adapters.
- `internal/bootstrap`: composition root and dependency wiring.
- Keep domain decisions independent from infrastructure details.
- Preserve documented ADR decisions and make architectural changes explicit.

## Go implementation rules

- Follow `gofmt`, idiomatic Go, small interfaces, constructor injection, and
  contextual `%w` error wrapping.
- Use `context.Context` for cancellation and timeouts at I/O boundaries.
- Bound goroutines, queues, retries, buffers, and cache growth.
- Treat Redis Lua scripts, outbox/event delivery, ride/driver state transitions,
  matching, ranking, and regional/legal rules as high-risk code.
- Do not log secrets, credentials, tokens, precise private location data, or
  unnecessary personally identifiable information.
- Validate external/user input at the adapter boundary and use parameterized
  database queries.

## Required workflow

For multi-file or architectural work: inspect → plan → test-first → implement →
review → verify → update project state. Do not silently broaden scope.

For Go changes, normally run:

```text
go test -race ./...
go vet ./...
go test -cover ./...
```

Use targeted package tests during iteration, then run the applicable full checks.
Update the affected module README when public interfaces, invariants,
dependencies, or configuration change, following `.agents/rules/module-readme-sync.md`.

At session end, update `docs/STATUS.md` only with verified results and record
session history under `docs/sessions/YYYY-MM-DD.md` when appropriate.

## Project-local guidance

- Rules live in `.agents/rules/`.
- Reusable skills live in `.agents/skills/`.
- Detailed workflow references live in `.agents/workflows/`.
- Prefer the RideForge-specific rules and skills over generic ECC guidance.
- Do not install or copy the complete ECC repository, global hooks, MCP
  configuration, unrelated agents, or unrelated language/framework skills.

