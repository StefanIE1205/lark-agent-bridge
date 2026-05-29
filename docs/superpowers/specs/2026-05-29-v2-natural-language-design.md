# V2 Natural Language Optimization Design

Based on `docs/v2-natural-language-optimization.md`.

## Approach

Incremental vertical slices (方案 B): implement each task, test immediately, move to next.

## Task Order

### P0: Frontend Fixes
- F001: Fix WorkDir passing
- F002: Wire progress_interval_ms
- F003: Precise bot mention detection
- F004: Tighten admin default policy
- F005: Wire ApprovalManager
- F006: Deduplicate final progress message
- F007: Thread-scoped chat defaults
- F008: Message deduplication

### P1: Natural Language Entry
- T101: Intent Router
- T102: Message Entry refactor
- T103: Natural language ACK

### P2: Context and Clarification
- T201: Conversation Memory
- T202: Project Resolver
- T203: Clarification Manager
- T204: Agent Selector

### P3: Smoother Task Execution
- T301: Session Broker
- T302: Agent Runtime Capability
- T303: Output Summarizer

### P4: Lark Experience
- T401: Markdown/Post replies
- T402: Message update merging
- T403: Approval text upgrade
- T404: Attachment entry

## Architecture

Message flow:
```
Lark Adapter → Engine.HandleMessage
  ├── slash command → legacy command router
  ├── pending clarification → clarification resolver
  └── normal text → Conversation Orchestrator
        ├── Intent Router
        ├── Context Resolver
        ├── Clarification Manager
        └── Session Broker → Agent Runtime
```
