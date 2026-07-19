# Link Click Statistics

## Purpose

Record every redirect click with referrer information and display per-link statistics with daily charts and top referrers.

## Requirements

### Requirement: Record click on redirect

The system SHALL record each redirect on a short link in the `link_click` table with referrer information.

#### Scenario: Successful click recording
- **WHEN** user follows a short link `/CODE`
- **THEN** the system records link id, referrer (from HTTP header), and timestamp in `link_click`
- **AND** `link_link.clicks` is incremented
- **AND** user receives HTTP 302 redirect to the original URL

#### Scenario: Click without referrer (direct visit)
- **WHEN** user follows a short link directly (without HTTP Referer)
- **THEN** `referrer` is saved as an empty string
- **AND** redirect proceeds normally

#### Scenario: Click on deleted link
- **WHEN** user follows a code for a non-existent link
- **THEN** the click is NOT recorded
- **AND** user receives 404

### Requirement: View click statistics

The system SHALL display a statistics page for each link with a daily click chart and referrer list.

#### Scenario: Daily click chart
- **WHEN** the link owner opens the statistics page `/link/:code/stats`
- **THEN** the system displays a bar chart (inline SVG) of clicks for the last 30 days
- **AND** each column corresponds to one day
- **AND** column height is proportional to the number of clicks on that day

#### Scenario: Top referrers list
- **WHEN** the link owner opens the statistics page
- **THEN** the system displays a top-10 referrers table
- **AND** the table contains columns: source, click count
- **AND** referrers are sorted descending by click count

#### Scenario: Access denied for non-owner
- **WHEN** a user who is not the link owner opens `/link/:code/stats`
- **THEN** the system returns 403

#### Scenario: Stats for non-existent link
- **WHEN** user opens statistics for a non-existent code
- **THEN** the system returns 404

### Requirement: Navigate to statistics

The system SHALL provide a link to the statistics page from the user's link list.

#### Scenario: Stats link in link table
- **WHEN** user views their link list
- **THEN** each table row contains a "Statistics" button/link
- **AND** clicking navigates to `/link/:code/stats`
