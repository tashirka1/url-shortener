## MODIFIED Requirements

### Requirement: Alias already taken
The system SHALL return an inline error when the requested alias is already taken, handling concurrent requests atomically via the `UNIQUE(code)` constraint.

#### Scenario: Alias already taken
- **WHEN** authenticated user submits an alias that another link already uses
- **THEN** the system returns an inline error indicating the alias is taken and does not update the code

#### Scenario: Concurrent alias race
- **WHEN** two concurrent requests try to claim the same free alias for different links
- **THEN** exactly one succeeds and the other receives `alias already taken` error (derived from `sqlite3.ErrConstraintUnique`)

#### Scenario: Alias update is atomic
- **WHEN** `UpdateAlias` is called
- **THEN** the storage SHALL perform a single `UPDATE link_link SET code=? WHERE code=? AND user_id=?` without a preceding `SELECT COUNT(*)`, and map `UNIQUE` violation to `ErrAliasTaken`
