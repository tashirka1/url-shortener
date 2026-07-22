# REST API

## Purpose

Provide a JSON-based HTTP API for programmatic access to URL shortening functionality, authenticated via Bearer tokens.

## Requirements

### Requirement: API is authorized via Bearer token
The system SHALL authorize API requests via the `Authorization: Bearer <token>` header. Unauthorized requests SHALL receive a 401 response.

#### Scenario: Valid token
- **WHEN** request contains a valid `Authorization: Bearer <token>` header
- **THEN** the request is authorized as the token's owner user and the token's `last_used_at` is updated

#### Scenario: Missing token
- **WHEN** request does not contain an `Authorization` header
- **THEN** the system returns `401 Unauthorized` with JSON body `{"error": "missing authorization header"}`

#### Scenario: Invalid token
- **WHEN** request contains an `Authorization` header with a token whose hash does not match any active (non-revoked) token
- **THEN** the system returns `401 Unauthorized` with JSON body `{"error": "invalid token"}`

#### Scenario: Revoked token
- **WHEN** request contains a revoked token
- **THEN** the system returns `401 Unauthorized` with JSON body `{"error": "token revoked"}`

### Requirement: POST /api/v1/link — Create short link
The system SHALL allow creating short links via the API.

#### Scenario: Successful creation
- **WHEN** user sends `POST /api/v1/link` with JSON body `{"url": "https://example.com"}`
- **THEN** the system creates a short link and returns `201 Created` with JSON `{"code": "abc1234", "url": "https://example.com", "short_url": "http://localhost:8000/abc1234"}`

#### Scenario: Missing URL field
- **WHEN** user sends `POST /api/v1/link` without `url` field or with an empty URL
- **THEN** the system returns `400 Bad Request` with JSON body `{"error": "url is required"}`

#### Scenario: Invalid URL
- **WHEN** user sends `POST /api/v1/link` with an invalid URL
- **THEN** the system returns `400 Bad Request` with JSON body `{"error": "invalid url"}`

#### Scenario: Duplicate URL
- **WHEN** user sends `POST /api/v1/link` with a URL that already exists for this user
- **THEN** the system returns `409 Conflict` with JSON body `{"error": "link already exists", "code": "abc1234"}`

### Requirement: GET /api/v1/link — List links
The system SHALL return the user's short links via the API with cursor pagination.

#### Scenario: Successful list
- **WHEN** user sends `GET /api/v1/link`
- **THEN** the system returns `200 OK` with JSON `{"links": [...], "next_cursor": 42}`

#### Scenario: Pagination
- **WHEN** user sends `GET /api/v1/link?cursor=42`
- **THEN** the system returns the next page of links starting from cursor `42`, `next_cursor` contains the ID of the last returned link

### Requirement: DELETE /api/v1/link/:code — Delete link
The system SHALL allow deleting short links via the API.

#### Scenario: Successful deletion
- **WHEN** user sends `DELETE /api/v1/link/abc1234`
- **THEN** the system deletes the link and returns `200 OK` with JSON `{"deleted": "abc1234"}`

#### Scenario: Link not found
- **WHEN** user sends `DELETE /api/v1/link/nonexistent`
- **THEN** the system returns `404 Not Found` with JSON body `{"error": "link not found"}`

### Requirement: GET /api/v1/link/:code/stats — Click statistics
The system SHALL return click statistics for a link via the API.

#### Scenario: Successful statistics
- **WHEN** user sends `GET /api/v1/link/abc1234/stats`
- **THEN** the system returns `200 OK` with JSON `{"code": "abc1234", "total_clicks": 42, "daily_clicks": [{"date": "2026-07-01", "clicks": 5}, ...], "top_referrers": [{"referrer": "https://t.co", "clicks": 10}, ...]}`

#### Scenario: Link not found
- **WHEN** user sends `GET /api/v1/link/nonexistent/stats`
- **THEN** the system returns `404 Not Found` with JSON body `{"error": "link not found"}`
