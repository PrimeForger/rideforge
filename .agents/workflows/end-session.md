---
description: End coding session cleanly. Verifies tests/builds, updates living docs/STATUS.md, archives session notes to docs/sessions/, and stages changes.
---

# End Session Workflow (`/end-session`)

Use `/end-session` at the conclusion of a work session to verify deliverables and maintain persistent state for future agents.

---

## What This Command Does

1. **Verify State with Automated Signals**:
   - Runs `go test ./...` and `go vet ./...` (or static analysis).
   - Only marks items complete if automated signals actually pass. Never rely on agent self-report.
2. **Update Living `docs/STATUS.md`**:
   - Overwrites (does not append to) `docs/STATUS.md`.
   - Records:
     - Active Phase & Subsystem Focus
     - Completed Capabilities (Verified)
     - In-Progress Capabilities
     - Active Blockers & Known Issues
     - The next 1–3 concrete, actionable tasks
3. **Archive Session History (Don't Accumulate)**:
   - If session-by-session notes are needed, write to `docs/sessions/YYYY-MM-DD.md`.
   - Keeps `docs/STATUS.md` lightweight and fast to read for the next session.
4. **Git Commit Hygiene**:
   - Reminds or stages documentation together with the code it describes:
     ```bash
     git add docs/STATUS.md <modified_code_files>
     git commit -m "feat/fix: <description> and update STATUS.md"
     ```

---

## Procedure Checklist

- [ ] Execute tests: `go test -race ./...` (or targeted package test).
- [ ] Check git status: `git status -s`.
- [ ] Overwrite `docs/STATUS.md` with verified progress and next 1–3 tasks.
- [ ] If archiving notes, write `docs/sessions/YYYY-MM-DD.md`.
- [ ] Commit docs alongside the code changes they document.
