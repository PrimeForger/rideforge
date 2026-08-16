# RideForge — AI Agent Context Initialization

Before doing any implementation, analysis, modification, or recommendation for this project, first locate and read:

`RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md`

This file is the authoritative instruction for how you must build your understanding of the RideForge project.

## Your First Responsibility

Read `RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md` completely and follow the instructions defined inside it.

Do not start working on the actual task yet.

The context-loading document will tell you:

- Which project documentation must be loaded.
- Which documents are authoritative for different types of decisions.
- The correct order in which documentation should be understood.
- Which documents are relevant to dispatch, architecture, AI, geospatial systems, configuration, reliability, etc.
- Which documents should only be loaded when relevant to the requested task.
- How to avoid consuming unnecessary context/tokens.
- How to resolve conflicts between documents.
- Which terminology and architectural principles must be preserved.
- How the existing architecture and business rules are expected to be handled.
- What constraints must not be violated when modifying the application.

## Important Instructions

Treat the documentation referenced by the context-loading file as the project's established architectural and business context.

Do not assume that a common industry implementation is automatically correct for RideForge.

Before making an architectural or implementation decision:

1. Check the relevant project documentation.
2. Follow the documented business rules.
3. Follow the applicable ADRs.
4. Follow the architecture/component documentation.
5. Preserve existing constraints and decisions unless the task explicitly asks to change them.
6. If multiple documents are relevant, load only the necessary ones according to the context-loading instructions.
7. Do not unnecessarily load the entire documentation tree.

### Dispatch-Specific Requirement

Pay particular attention to the documented distinction between:

- Smart Stand Dispatch
- Smart Dispatch
- AI-assisted optimization
- Hierarchical dispatch configuration
- Cross-location candidate discovery
- Candidate eligibility vs candidate preference
- Geographic discovery vs legal eligibility

Do not reinterpret these concepts using assumptions from Uber, Ola, Rapido, or other systems unless explicitly requested.

## Handling Conflicts or Ambiguity

If the documentation appears to contain conflicting information:

1. Follow the precedence rules defined in `RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md`.
2. Check the relevant ADR and authoritative business/architecture documentation.
3. Do not silently invent a resolution.
4. Clearly identify the conflict before making a consequential architectural change.
5. Preserve established behavior unless the requested task explicitly changes it.

## Do Not Start the Task Yet

At this stage, your only job is to initialize your understanding of the project.

After reading the context-loading instructions and the relevant documents required by them, briefly confirm that you have completed the context initialization.

Do **not** implement anything yet.

I will provide the actual question/task in my next message.
