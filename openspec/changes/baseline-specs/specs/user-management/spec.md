## Purpose

Lets admins manage the users who participate in expense splits, including their aid class and their mapping to an Actual Budget payee.

## ADDED Requirements

### Requirement: User record fields

A user record SHALL consist of a name, an OIDC identifier (username or subject), an aid class, and an Actual Budget payee ID. The aid class SHALL be one of `regular`, `needs_help`, or `will_help`.

#### Scenario: Valid user record

- **WHEN** an admin creates a user with a name, OIDC identifier, aid class, and payee ID
- **THEN** the user appears on the admin dashboard with those attributes

### Requirement: Create users

Admins SHALL be able to create users, including multiple users at once. Rows with an empty name or empty OIDC identifier SHALL be skipped silently.

#### Scenario: Bulk create skips incomplete rows

- **WHEN** an admin submits a bulk create form containing one complete row and one row with an empty name
- **THEN** only the complete row is created and no error is shown for the incomplete row

### Requirement: Update users

Admins SHALL be able to edit a user's name, OIDC identifier, aid class, and payee ID.

#### Scenario: Successful update

- **WHEN** an admin edits a user and saves
- **THEN** the admin dashboard reflects the updated values

#### Scenario: Failed update

- **WHEN** saving an edited user fails (for example, a duplicate OIDC identifier)
- **THEN** the admin dashboard reloads with an error message

### Requirement: Delete users

Admins SHALL be able to delete a user. Deleting a user SHALL also delete all of that user's expense splits in the same operation.

#### Scenario: Delete user with splits

- **WHEN** an admin deletes a user who has expense splits
- **THEN** the user and all their splits are removed

#### Scenario: Failed delete

- **WHEN** deleting a user fails
- **THEN** the admin dashboard reloads with an error message and the user remains

### Requirement: Payee mapping

Each user MAY map to an Actual Budget payee via its payee ID. The mapping SHALL be unique among non-empty payee IDs. Admins SHALL be able to set the payee mapping when editing a user, and to map an unmapped payee to a user directly from the split dialog.

#### Scenario: Map payee during split creation

- **WHEN** an admin maps a payee to a user via the split dialog's one-user save
- **THEN** the user's payee ID is updated to that payee

### Requirement: Aid class display

The system SHALL display a user's aid class on the admin dashboard and the user's own page, labeled "Regular", "Needs Help", or "Will Help".

#### Scenario: Aid class labels

- **WHEN** a user has aid class `needs_help`
- **THEN** their aid class is displayed as "Needs Help"
