# RPS Benchmark

## Purpose

Benchmark endpoints for measuring requests per second with different templating and database operation patterns.

## Requirements

### Requirement: Templ page update benchmark
The system SHALL support `GET /rps/templ-page-update` as a benchmark endpoint that increments `rps_log.duration` for `id=1` and returns 200 with an HTML fragment.

#### Scenario: Successful update
- **WHEN** client sends `GET /rps/templ-page-update`
- **THEN** the system executes `UPDATE rps_log SET duration = duration + 1 WHERE id = 1` and returns 200 with `UpdatePage` HTML

#### Scenario: Update SQL is valid
- **WHEN** `RPS.Update` is called
- **THEN** it SHALL NOT return `near "WHERE": syntax error` and SHALL affect 0 or 1 rows depending on existence of `id=1`

### Requirement: RPS update error handling
The system SHALL return 500 with JSON/HTML error if `RPS.Update` fails, and log the error.

#### Scenario: DB failure
- **WHEN** `RPS.Update` returns an error
- **THEN** handler returns HTTP 500 and does not render `UpdatePage`
