# user-ledger Specification

## Purpose

Shows each user a read-only ledger of their splits with a running balance, and routes users to the correct dashboard after login.

## Requirements

### Requirement: Ledger contents

A user's ledger SHALL list every split recorded for that user. Each row SHALL show a date, a note, a signed amount, and a running balance. Credits SHALL be shown with a `+` sign and debits with a `−` sign.

#### Scenario: Mixed credits and debits

- **WHEN** a user has both a credit split and a debit split
- **THEN** both appear in their ledger with `+` and `−` amounts respectively

#### Scenario: No splits

- **WHEN** a user has no splits
- **THEN** the ledger shows an empty state ("No transactions yet")

### Requirement: Running balance

The running balance SHALL accumulate credits positively and debits negatively in chronological order, and SHALL be shown on each row. The final running balance SHALL be displayed as the user's balance, with distinct visual treatment for positive, negative, and zero balances.

#### Scenario: Balance accumulates chronologically

- **WHEN** a user has a $10.00 credit dated before a $4.00 debit
- **THEN** the credit row shows a running balance of $10.00 and the debit row shows $6.00, with $6.00 shown as the overall balance

### Requirement: Row metadata fallbacks

A ledger row SHALL prefer the split's stored date and note, falling back to the source transaction's date and note when the stored values are empty. When both are empty, the date SHALL display as "Unknown" and the note as "Expense Share".

#### Scenario: Split created before transaction metadata was stored

- **WHEN** a split has an empty date and its transaction exists
- **THEN** the transaction's date is shown

#### Scenario: No metadata available

- **WHEN** a split has an empty date and no source transaction can be found
- **THEN** the date is shown as "Unknown"

### Requirement: Ledger ordering

The ledger SHALL be ordered newest first, with running balances computed from the oldest split.

#### Scenario: Newest first display

- **WHEN** a user has splits on three different dates
- **THEN** the most recent split appears at the top of the ledger

### Requirement: Ledger access control

The ledger SHALL be accessible to the user it belongs to and to admins. Non-admin users SHALL be denied access to other users' ledgers with a 403 error.

#### Scenario: Admin views any ledger

- **WHEN** an admin requests another user's ledger page
- **THEN** the ledger is displayed

#### Scenario: Non-admin views someone else's ledger

- **WHEN** a regular user requests a ledger for a different user
- **THEN** a 403 error page is shown

#### Scenario: Unknown user ledger

- **WHEN** an admin requests a ledger for a user identifier that does not match any user record
- **THEN** a 404 error page is shown

### Requirement: Post-login routing

After authentication, admins SHALL be sent to the admin dashboard and regular users to their own ledger page when they visit the home route.

#### Scenario: Admin visits home

- **WHEN** an admin visits `/`
- **THEN** they are redirected to `/admin`

#### Scenario: Regular user visits home

- **WHEN** a regular user visits `/`
- **THEN** they are redirected to their own ledger page
