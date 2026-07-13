## 1. Model

- [x] 1.1 Add `ErrAliasTaken` sentinel error to `internal/link/model/link.go`

## 2. Storage

- [x] 2.1 Add `UpdateAlias(ctx, userID int, currentCode, newCode string) error` to `LinkStorage` interface
- [x] 2.2 Implement `UpdateAlias` in `internal/link/storage/link.go` — SELECT check + UPDATE with UNIQUE fallback

## 3. Service

- [x] 3.1 Add `UpdateAlias(ctx, userID int, currentCode, newCode string) error` to `LinkService` interface
- [x] 3.2 Implement validation logic in service: length (3-32), charset `[a-zA-Z0-9_-]`, reserved words (`health`, `auth`, `link`, `rps`, `static`)
- [x] 3.3 Delegate to storage and wrap errors

## 4. Handler

- [x] 4.1 Add `GetEditForm` handler — returns `CodeEditField` templ component
- [x] 4.2 Add `UpdateAlias` handler — parses form, calls service, returns `CodeDisplay` or error
- [x] 4.3 Register routes: `GET /link/:code/edit-form`, `PATCH /link/:code/alias` under auth group

## 5. View

- [x] 5.1 Create `CodeDisplay` templ component — shows code + edit button
- [x] 5.2 Create `CodeEditField` templ component — input + Save/Cancel with HTMX attributes
- [x] 5.3 Create `AliasError` templ component — inline error display
- [x] 5.4 Update `Link` template to use `CodeDisplay` instead of raw code text
- [x] 5.5 Update `CreateLinkSuccess` template to use `CodeDisplay`
- [x] 5.6 Run `templ generate`

## 6. Tests

- [x] 6.1 Add service tests: valid update, too short, too long, invalid chars, reserved word, alias taken, not found
- [x] 6.2 Add storage test for `UpdateAlias`
