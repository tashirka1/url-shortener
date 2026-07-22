# QR Code

## Purpose

Generate QR code PNG images for short links and display them in a modal dialog from the link list.

## Requirements

### Requirement: Generate QR code for a link
The system SHALL generate a QR code PNG for any existing link belonging to the authenticated user.

#### Scenario: Successful QR generation
- **WHEN** authenticated user sends `GET /link/:code/qr` with a valid code belonging to them
- **THEN** the system returns an HTML fragment containing the QR code as a base64 data URI image

#### Scenario: Link not found
- **WHEN** authenticated user sends `GET /link/:code/qr` with a code that does not exist
- **THEN** the system returns 404 Not Found

#### Scenario: Link belongs to another user
- **WHEN** authenticated user sends `GET /link/:code/qr` with a code belonging to another user
- **THEN** the system returns 404 Not Found

#### Scenario: Unauthenticated request
- **WHEN** unauthenticated user sends `GET /link/:code/qr`
- **THEN** the system redirects to login page

### Requirement: QR button in link list
The system SHALL display a "QR" button for each link in the link list table.

#### Scenario: Clicking QR button opens modal
- **WHEN** user clicks "QR" button in a link row
- **THEN** the system displays a modal dialog with the QR code image and the short URL text

#### Scenario: Closing QR modal
- **WHEN** user clicks close button or backdrop in the QR modal
- **THEN** the modal dialog is closed

### Requirement: QR code format
The QR code SHALL use Medium error correction level and 256x256 pixel size.

#### Scenario: QR code parameters
- **WHEN** the system generates a QR code
- **THEN** it SHALL use `qrcode.Medium` recovery level
- **THEN** it SHALL generate a 256x256 pixel image
