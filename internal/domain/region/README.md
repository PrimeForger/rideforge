# Regional & Legal Compliance Domain (`internal/domain/region`)

## Purpose
Encapsulates regulatory boundaries, state transport permit rules, and cross-border ride validation.

## Public Interface
`Region` enum (`Kerala`, `Karnataka`), `BorderRule`, `IsRideAllowed(from, to Region) bool`.

## Dependencies
None. Pure domain logic.

## Invariants & Gotchas
Cross-border rides between Kerala and Karnataka are strictly disallowed due to interstate permit regulations. Geographic proximity must NEVER bypass this legal check.

## Human-Approval Required
Yes. Modifying region boundaries or rules carries direct legal and regulatory compliance implications.

## Related ADRs
ADR-0018 (Regional and Legal Ride Validation).

## Known Issues / Tech Debt
Rules are currently hardcoded; moving to a dynamic database-backed regional policy engine is planned.
