# API Token Management

## Purpose

Allow authenticated users to create, list, and revoke API tokens for programmatic access to the URL shortener service. Tokens are stored as SHA-256 hashes only.

## Requirements

### Requirement: User can create an API token
The system SHALL allow authenticated users to create a new API token with a human-readable name.

#### Scenario: Successful token creation
- **WHEN** user submits the token creation form with a name
- **THEN** the system generates a cryptographically secure random token, stores its SHA-256 hash in the `api_token` table, and displays the raw token exactly once

#### Scenario: Token name validation
- **WHEN** user submits an empty name or a name longer than 128 characters
- **THEN** the system returns a validation error and does not create the token

### Requirement: User can view API tokens
The system SHALL display all active API tokens of the user with name, prefix (first 8 characters), and last used date.

#### Scenario: Token list in dashboard
- **WHEN** user navigates to the API tokens page
- **THEN** the system renders a table of all non-revoked tokens, sorted by creation date descending

### Requirement: User can revoke an API token
The system SHALL allow the user to revoke an API token by ID, making it permanently invalid.

#### Scenario: Successful revocation
- **WHEN** user clicks "Revoke" on a token
- **THEN** the system sets the `revoked_at` timestamp and the token can no longer authorize API requests

#### Scenario: Revoke non-existent token
- **WHEN** user attempts to revoke a token with an ID that does not exist or belongs to another user
- **THEN** the system returns an error

### Requirement: Token is stored only as a hash
The system SHALL NOT store the raw token in the database. Only the SHA-256 hash.

#### Scenario: Token hash storage
- **WHEN** a token is created
- **THEN** only the SHA-256 hash (hex-encoded) of the full token is stored in `api_token.token_hash`
