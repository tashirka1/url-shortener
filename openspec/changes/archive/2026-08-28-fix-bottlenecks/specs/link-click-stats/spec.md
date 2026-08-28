## MODIFIED Requirements

### Requirement: Record click on redirect
The system SHALL record each redirect on a short link in the `link_click` table with referrer and user-agent information.

#### Scenario: Successful click recording
- **WHEN** user follows a short link `/CODE` with `Referer: https://t.co` and `User-Agent: Mozilla/5.0`
- **THEN** the system records link id, referrer `https://t.co`, user_agent `Mozilla/5.0` (truncated to 512 chars if longer), and timestamp in `link_click`
- **AND** `link_link.clicks` is incremented
- **AND** user receives HTTP 303 redirect to the original URL

#### Scenario: Click without referrer (direct visit)
- **WHEN** user follows a short link directly (without HTTP Referer)
- **THEN** `referrer` is saved as an empty string and `user_agent` is still recorded
- **AND** redirect proceeds normally

#### Scenario: User-Agent truncation
- **WHEN** user follows a short link with User-Agent longer than 512 bytes
- **THEN** the system stores only the first 512 bytes
