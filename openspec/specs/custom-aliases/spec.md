# Custom Aliases

## Purpose

Allow authenticated users to replace auto-generated short codes with custom aliases for their links.

## Requirements

### Requirement: Edit link alias after creation
The system SHALL allow authenticated users to replace the auto-generated short code of their link with a custom alias.

#### Scenario: Successful alias update from link list
- **WHEN** authenticated user clicks edit button on a link row and submits a valid new alias
- **THEN** the system updates the code and returns the updated code display in the table row

#### Scenario: Successful alias update from create result
- **WHEN** authenticated user creates a link and edits the generated code in the success fragment
- **THEN** the system updates the code and returns the updated code display

#### Scenario: Cancel edit restores original code
- **WHEN** authenticated user clicks edit and then clicks cancel
- **THEN** the system returns the original code display without changes

#### Scenario: Alias too short
- **WHEN** authenticated user submits an alias shorter than 3 characters
- **THEN** the system returns an inline validation error and does not update the code

#### Scenario: Alias too long
- **WHEN** authenticated user submits an alias longer than 32 characters
- **THEN** the system returns an inline validation error and does not update the code

#### Scenario: Alias with invalid characters
- **WHEN** authenticated user submits an alias containing characters outside `[a-zA-Z0-9_-]`
- **THEN** the system returns an inline validation error and does not update the code

#### Scenario: Alias is a reserved word
- **WHEN** authenticated user submits an alias matching `health`, `auth`, `link`, `rps`, or `static`
- **THEN** the system returns an inline validation error and does not update the code

#### Scenario: Alias update for non-existent code
- **WHEN** authenticated user sends an update for a code that does not exist
- **THEN** the system returns 404 Not Found

#### Scenario: Alias update for another user's link
- **WHEN** authenticated user sends an update for a code belonging to another user
- **THEN** the system returns 404 Not Found

#### Scenario: Unauthenticated alias update
- **WHEN** unauthenticated user sends an alias update request
- **THEN** the system redirects to login page

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
