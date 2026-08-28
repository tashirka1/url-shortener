# Session Cookie Handling

## Purpose

Manage session cookie attributes and error handling for authentication flows to support http on localhost and offline single-binary operation.

## Requirements

### Requirement: Session cookie is not Secure
The system SHALL set session cookies with `Secure=false`, `HttpOnly=true`, `SameSite=Lax`, `Path=/`, `MaxAge=604800` (7 days) for `SetUserId`, and `MaxAge=-1` for `ClearSession`, to support http on localhost and offline single-binary operation.

#### Scenario: Login sets cookie without Secure
- **WHEN** user successfully logs in via `POST /auth/login`
- **THEN** the session cookie is set with `Secure=false` and is sent over http

#### Scenario: Auth middleware clears invalid session
- **WHEN** session is missing or invalid
- **THEN** middleware returns `HX-Redirect: /auth/login` and clears cookie with same `Path` and `SameSite` attributes

### Requirement: ClearSession returns error
The system SHALL make `ClearSession` return `error` and the `Logout` handler SHALL handle it: on error return 500, on success redirect 303 to `/auth/login`.

#### Scenario: Successful logout
- **WHEN** user sends `GET /auth/logout` and session save succeeds
- **THEN** the system clears the cookie and returns 303 to `/auth/login`

#### Scenario: Logout save failure
- **WHEN** `ClearSession` fails to save the cookie
- **THEN** the system logs the error and returns 500 instead of redirecting

#### Scenario: ClearSession uses consistent options
- **WHEN** `ClearSession` is called
- **THEN** it uses `Path=/`, `HttpOnly=true`, `SameSite=Lax`, `Secure=false`, `MaxAge=-1` matching `SetUserId` options except `MaxAge`
