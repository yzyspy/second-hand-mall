<!--
  SYNC IMPACT REPORT
  ==================
  Version change: [template] → 1.0.0 (initial ratification)
  Modified principles: N/A (initial population from blank template)
  Added sections:
    - Core Principles (5 principles)
    - Technology Stack
    - Development Workflow
    - Governance
  Removed sections: N/A
  Templates requiring updates:
    ✅ .specify/templates/plan-template.md — Constitution Check gates aligned
    ✅ .specify/templates/spec-template.md — no structural changes required
    ✅ .specify/templates/tasks-template.md — no structural changes required
  Follow-up TODOs:
    - None; all fields resolved from project context.
-->

# Second-Hand Mall Admin Constitution

## Core Principles

### I. Admin-First Design

Every feature in mall-admin MUST serve a platform administrator use case. No
end-user-facing features belong in this tool. Each screen or capability MUST
map to a concrete admin operation (e.g., product moderation, user management,
order review). Features without a clear admin purpose MUST be rejected.

**Rationale**: The admin tool has a narrowly defined audience and mission. Scope
creep toward end-user features dilutes the tool and duplicates the mini-program.

### II. API-Driven Architecture

All data reads and mutations MUST go through the mall-server RESTful API.
The admin frontend MUST NOT access the database directly. Every new screen
MUST be backed by a documented API contract before implementation begins.
If a required endpoint does not yet exist, it MUST be specified in
`contracts/` and the backend work tracked as a prerequisite task.

**Rationale**: Keeping the frontend stateless against a single authoritative
backend API ensures consistency with the mini-program and avoids divergent
data access paths.

### III. Component-Based UI

UI MUST be composed of independently testable, single-responsibility
components. No component may exceed 300 lines. Shared UI elements
(tables, forms, modals) MUST live in a common `components/` directory and
MUST NOT be duplicated across pages. Each page component MUST remain a
thin coordinator of sub-components.

**Rationale**: A component-first approach enables parallel development,
isolated testing, and consistent visual language across the admin tool.

### IV. Security and Access Control

All admin routes MUST require authentication via JWT (matching the
mall-server `Authorization: Bearer <token>` scheme). Destructive operations
(delete, ban, force-close) MUST require a confirmation step. Sensitive data
fields (phone numbers, personal info) MUST be masked in list views.
No admin credentials or tokens may be committed to version control.

**Rationale**: Admin tools have elevated privileges; a single compromised
session can affect all platform data. Defense-in-depth at the UI layer is
mandatory even when the backend enforces its own auth.

### V. Simplicity and YAGNI

Features MUST solve a stated admin problem before being built. Abstractions,
helper utilities, and shared services MUST only be introduced when the same
pattern appears three or more times. Configuration and feature flags MUST NOT
be added for hypothetical future requirements. The admin tool MUST ship
working features, not frameworks.

**Rationale**: Admin tools tend toward over-engineering. Strict YAGNI keeps
the tool lean, maintainable, and focused on real operator needs.

## Technology Stack

- **Framework**: Vue 3 with TypeScript (Composition API)
- **Build Tool**: Vite
- **UI Component Library**: Element Plus
- **HTTP Client**: Axios with centralized request interceptor for JWT injection
- **State Management**: Pinia (only when local component state is insufficient)
- **Routing**: Vue Router 4 with route-level auth guards
- **Backend API**: mall-server at `http://localhost:8080` (dev); production URL via env var
- **Linting**: ESLint + Prettier enforced in CI

All technology choices MUST remain consistent with this list. Introducing
a new dependency requires a documented rationale and amendment to this section.

## Development Workflow

- All features MUST start with a spec (`/speckit.specify`) and pass
  Constitution Check before implementation.
- Feature branches follow the pattern `###-feature-name` off `main`.
- UI changes MUST be manually verified in a browser on the happy path and
  key edge cases before a feature is declared complete.
- API contracts (`contracts/`) MUST be authored before implementing any page
  that requires a new or modified endpoint.
- Tasks are organized by user story; each story MUST be independently
  deliverable as an MVP increment.

## Governance

This constitution supersedes all other development practices for the
mall-admin project. Amendments require:

1. A documented rationale (written in the PR description).
2. Version bump per semantic versioning (MAJOR/MINOR/PATCH).
3. Updated `LAST_AMENDED_DATE`.
4. Propagation checks across plan, spec, and task templates.

All PRs and code reviews MUST verify compliance with the five Core Principles.
Complexity violations MUST be justified in the Implementation Plan's
Complexity Tracking table. The primary runtime development guidance file is
`.claude/CLAUDE.md` (project root).

**Version**: 1.0.0 | **Ratified**: 2026-05-19 | **Last Amended**: 2026-05-19
