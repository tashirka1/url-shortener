## Why

Current links get auto-generated 7-char base62 codes (e.g., `aB3xK9z`), which are hard to remember and share verbally. Users need the ability to replace the generated code with a meaningful custom alias (`my-link`, `blog-post`).

## What Changes

- Add `PATCH /link/:code/alias` endpoint to update a link's code
- Add `GET /link/:code/edit-form` endpoint returning an inline editing form
- Add inline code editing to both the create-success fragment and the link list rows
- Validation: 3-32 chars, `[a-zA-Z0-9_-]`, reserved words blocked, uniqueness enforced
- New sentinel error `ErrAliasTaken` for conflict detection

## Capabilities

### New Capabilities
- `custom-aliases`: Allow users to replace the auto-generated short code with a custom alias via inline HTMX editing in the create form result and the link list.

### Modified Capabilities

None.

## Impact

- `internal/link/model/` — new sentinel error
- `internal/link/storage/` — new `UpdateAlias` method in interface + SQL implementation
- `internal/link/service/` — new `UpdateAlias` method in interface + validation logic
- `internal/link/handler/` — two new Echo handlers, new route registration
- `internal/link/view/` — new templ components for inline editing, modifications to existing `Link` and `CreateLinkSuccess` templates
