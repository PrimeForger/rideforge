# Module README Sync Rule

## Mandatory Instruction
Any time you add, remove, or change the public interface, invariants, dependencies, or configuration of a module, update that module's `README.md` in the same change, before marking the task complete. This applies even if not explicitly asked.

## Ground Rules for README Updates
1. **Ground Every Claim in Real Code**:
   - Every entry under `Public Interface` must match actual exported types, functions, methods, or constants in that directory.
   - Every entry under `Dependencies` must reflect real internal/external imports from that package.
   - Do not hallucinate or speculate on hypothetical capabilities.

2. **Standard 7-Section Template**:
   Every module `README.md` must strictly contain these 7 sections:
   - `## Purpose`: 1-2 concise paragraphs stating the single responsibility of this package.
   - `## Public Interface`: List of key exported types, structs, interfaces, and entrypoint functions.
   - `## Dependencies`: Internal packages imported and external third-party dependencies used.
   - `## Invariants & Gotchas`: Concurrency rules, ordering guarantees, memory ownership, zero-value semantics, error wrapping.
   - `## Human-Approval Required`: Explicit flags (`YES` / `NO`). If `YES`, clearly document what requires human sign-off (e.g., legal boundary transitions, state machine modifications, ranking/matching algorithms, Redis Lua scripts, database schema migrations).
   - `## Related ADRs`: References to `docs/01-adr/` or `N/A`.
   - `## Known Issues / Tech Debt`: Existing limitations, known bugs, or TODOs.

3. **Human Approval Constraints**:
   - Never modify or remove `Human-Approval Required: YES` constraints without explicit confirmation from the human engineer.
   - Packages governing legal/regulatory rules (`internal/domain/region`), core state transitions (`internal/domain/ride`, `internal/domain/driver`), financial/outbox guarantees (`internal/domain/outbox`, `internal/infrastructure/outbox`), matching/ranking weights (`internal/application/matching`, `internal/application/dispatch/...`), or database/cache scripts (`internal/infrastructure/postgres`, `internal/infrastructure/redis/scripts`) must retain their human-approval guards.

4. **Co-Staging Requirement**:
   - The pre-commit hook and CI pipeline will reject any commit that stages or modifies `.go` files within a directory without also staging changes to that directory's `README.md` (unless only non-functional tests/comments are altered and interfaces remain unchanged).
