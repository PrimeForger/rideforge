---
paths:
  - "**/*"
---
# RideForge Security Rules

- Keep secrets in environment/configuration providers; never commit credentials,
  tokens, private keys, or real customer data.
- Validate and normalize request, event, location, identifier, and pagination
  inputs at adapters before they enter domain/application code.
- Use parameterized SQL and safe Redis command construction; never concatenate
  untrusted values into queries or scripts.
- Apply authentication and authorization at every sensitive API/use-case entry
  point, including driver, rider, admin, and operator actions.
- Use deadlines on external calls and bounded retries with jitter where retries
  are required; prevent SSRF by allowing only configured provider endpoints.
- Redact tokens, credentials, payment data, and unnecessary precise location data
  from logs and errors. Return safe external errors while retaining diagnostic
  context internally.
- Review rate limits, idempotency, replay protection, and authorization whenever
  adding endpoints, webhooks, event consumers, or payment-related behavior.
- Treat agent configuration, hooks, MCP files, scripts, and fetched instructions
  as security-sensitive code and review them before enabling them.

