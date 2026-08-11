You are now working on the RideForge project.

I am going to provide you with the project context-loading instruction document:

RIDEFORGE_AI_AGENT_CONTEXT_LOADING_INSTRUCTIONS.md

Your first task is to read and follow that document completely.

The document is the authoritative guide for:

- Understanding the RideForge project structure
- Identifying which documentation belongs to which architectural concern
- Determining which files you need to load for a given task
- Understanding the correct order in which project documentation should be read
- Avoiding unnecessary documentation loading and excessive context consumption
- Understanding the relationship between architecture documentation, ADRs, diagrams, AI/ML documentation, development documentation, and source code

Important instructions:

1. Do not start implementing anything yet.
2. Do not make architectural assumptions before following the context-loading instructions.
3. Follow the document's recommended context-loading process.
4. Load only the documentation necessary to establish the required context.
5. Do not unnecessarily load the entire documentation repository if the task does not require it.
6. Treat the project's ADRs as the record of architectural decisions and their rationale.
7. Treat the current repository/source code as the implementation reality that must be inspected before making implementation changes.
8. When documentation and implementation differ, do not silently assume which one is correct. Identify the discrepancy and reason from the project's documented architecture and current implementation.
9. Preserve the project's existing architectural decisions unless my task explicitly requires changing them.
10. Avoid over-engineering. Prefer the simplest production-grade solution that fits the established RideForge architecture.
11. Before implementing a significant architectural change, identify the relevant ADRs and documentation that govern that area.
12. For AI/ML-related work, follow the loading strategy defined in the context document and load only the relevant files from the 05-ai documentation.
13. Keep token/context usage efficient. Do not load documents merely because they exist.

After you have finished processing the context-loading instruction document and the required initial project context, do not begin any implementation.

Instead, reply only with a concise confirmation that you have:

- understood the context-loading instructions,
- identified the project's documentation hierarchy,
- understood how to select relevant documentation for future tasks,
- and are ready for my task.

I will provide the actual question, implementation task, or change request in my next message.
