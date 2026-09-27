---
description: Verify that all Go modules modified in the current branch or staged commits have their corresponding module README.md updated.
---

# Module README Sync Check (`/doc-sync-check`)

Use `/doc-sync-check` to verify that any code changes to Go modules include corresponding documentation updates in each module's `README.md`.

---

## What This Command Does

1. **Detects Modified Modules**:
   - Inspects staged files (`git diff --cached --name-only`) or branch diff against base (`git diff --name-only origin/staging...HEAD` or `git status --porcelain`).
   - Identifies all directories containing modified `.go` files.
2. **Checks README Existence & Staging**:
   - For every directory containing touched `.go` files, verifies that `<module-dir>/README.md` is present and also modified/staged.
3. **Flags Blocking Issues**:
   - If any module code was changed without its `README.md` being updated, the workflow fails.
   - Updates `docs/STATUS.md` under `## 4. Active Blockers & Known Issues` listing the out-of-sync modules.
4. **Validates 7-Section Structure**:
   - Checks that any updated `README.md` retains the mandatory 7 sections (`Purpose`, `Public Interface`, `Dependencies`, `Invariants & Gotchas`, `Human-Approval Required`, `Related ADRs`, `Known Issues / Tech Debt`).

---

## Procedure Checklist

- [ ] Run git diff to find touched Go source files:
  ```bash
  # For staged commits:
  staged_files=$(git diff --cached --name-only)
  # For working tree changes:
  changed_files=$(git diff --name-only HEAD)
  ```
- [ ] Determine unique package directories:
  ```bash
  modules_touched=$(echo "$staged_files" | grep -E '\.go$' | xargs -I{} dirname {} | sort -u)
  ```
- [ ] Check each module for `<module-dir>/README.md`:
  - If missing from the diff: fail verification.
  - If `Human-Approval Required: YES` is altered: verify explicit human sign-off was received.
- [ ] If out-of-sync modules exist:
  - Add blocker entry in `docs/STATUS.md` under `## 4. Active Blockers & Known Issues`.
  - Block completion until `<module-dir>/README.md` is updated.
- [ ] Once all updated READMEs are verified and staged:
  - Clear blocker entry from `docs/STATUS.md`.
