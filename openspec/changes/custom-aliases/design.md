## Context

The URL shortener generates random 7-char base62 codes for links. Users want meaningful aliases (`my-link`) instead of random strings (`aB3xK9z`). The existing `link_link` table already has a UNIQUE constraint on `code`, and FTS5 syncs via triggers. This is a self-contained change within the `link` module.

## Goals / Non-Goals

**Goals:**
- Allow users to edit the short code after link creation
- Inline editing in both the create-success fragment and the link list table
- Validate: 3-32 chars, `[a-zA-Z0-9_-]`, reserved words blocked, uniqueness enforced
- Reuse a single templ component for the edit UI in both locations

**Non-Goals:**
- Editing other link fields (URL, etc.) — YAGNI
- Custom aliases at creation time (edit-only after creation)
- Admin override of alias conflicts

## Decisions

1. **Single PATCH endpoint** — One `PATCH /link/:code/alias` handles updates from all locations. No need for endpoint duplication.
2. **Check-first uniqueness** — SELECT before UPDATE to give clean `ErrAliasTaken` error (vs parsing SQLite constraint violation).
3. **Two-step HTMX edit flow** — GET loads the edit form, PATCH saves it. Simpler than contenteditable or JS-heavy approaches; pure HTMX.
4. **Reused templ component** — `CodeDisplay` + `CodeEditField` are shared between `Link` (list row) and `CreateLinkSuccess`. Avoids template duplication.

## Risks / Trade-offs

- [Uniqueness race] → Two concurrent edits to different links with the same new alias could both pass the SELECT check. Mitigation: UNIQUE constraint on `code` will catch the second UPDATE and return `ErrAliasTaken`.
- [FTS sync] → The existing `link_au` AFTER UPDATE trigger handles FTS reindexing automatically. No manual sync needed.
