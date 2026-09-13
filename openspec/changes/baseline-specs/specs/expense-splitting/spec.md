## Purpose

Splits expenses from Actual Budget among users, including automatically crediting payments users make against their account.

## ADDED Requirements

### Requirement: Tagged transaction selection

The system SHALL treat only Actual Budget transactions whose notes contain the configured split tag (the `SPLIT_TAG` environment variable, defaulting to `#gsu2026`) as splittable. The tag SHALL be stripped from notes before display.

#### Scenario: Tagged and untagged transactions

- **WHEN** transactions are loaded
- **THEN** only transactions whose notes contain the split tag are shown and considered for splits

#### Scenario: Notes displayed without the tag

- **WHEN** a tagged transaction's note is displayed
- **THEN** the tag text is removed from the displayed note

### Requirement: Automatic credit for payments

When the admin dashboard loads, the system SHALL create a split for every tagged transaction that has no split yet, has a positive amount, and whose payee maps to a user, crediting the full transaction amount to that user. Such splits SHALL be marked auto-created.

#### Scenario: Positive payment from a mapped payee

- **WHEN** a tagged transaction of +$50.00 has a payee mapped to user A and no split exists
- **THEN** an auto-created split of $50.00 is created for user A

#### Scenario: Transaction already split

- **WHEN** a tagged transaction already has at least one split
- **THEN** no auto-credit is created for it

#### Scenario: Unmapped payee

- **WHEN** a tagged positive transaction's payee is not mapped to any user
- **THEN** no auto-credit is created

#### Scenario: Negative transaction

- **WHEN** a tagged transaction has a negative amount
- **THEN** no auto-credit is created from it

### Requirement: Manual split creation

Admins SHALL be able to split a transaction among any users by submitting an amount per participant. Submitting SHALL replace all existing splits for that transaction: every prior split is removed and a new split is recorded for each submitted participant. Submitting zero participants clears the transaction's splits.

#### Scenario: Replace splits

- **WHEN** an admin re-splits a transaction with a different participant list
- **THEN** the old splits are removed and only the new participants have splits

#### Scenario: Clear splits

- **WHEN** an admin submits a split form with no participant amounts
- **THEN** all splits for that transaction are removed

### Requirement: Split record contents

Each split SHALL record the transaction ID, the user, the amount owed in cents, whether it was auto-created, and a copy of the transaction's date and note taken at save time. A user SHALL have at most one split per transaction.

#### Scenario: Duplicate split prevention

- **WHEN** a split is saved for a transaction and user that already have one
- **THEN** the existing split's amount, date, and note are updated instead of creating a duplicate

### Requirement: Even distribution

When the admin uses the even distribution helper, the transaction amount SHALL be divided equally among participants, with any remainder cents assigned one cent at a time to participants.

#### Scenario: Amount not divisible by participant count

- **WHEN** a $10.00 transaction is split evenly among 3 participants
- **THEN** two participants are assigned $3.34 and one is assigned $3.32

### Requirement: Aid-weighted distribution

When the admin uses the aid-weighted distribution helper and the participants include at least one `needs_help` and at least one `will_help` user, participants with aid class `needs_help` SHALL be assigned zero, `regular` participants SHALL receive the even base share, and `will_help` participants SHALL cover the needs-help shortfall plus any remainder. When participants do not include both a `needs_help` and a `will_help` user, the helper SHALL fall back to even distribution.

#### Scenario: Full aid mix

- **WHEN** a $9.00 transaction is split among one `needs_help`, one `regular`, and one `will_help` participant using the aid-weighted helper
- **THEN** the needs-help participant is assigned $0.00, the regular participant $3.00, and the will-help participant $6.00

#### Scenario: No needs-help participant

- **WHEN** the participants contain no `needs_help` user
- **THEN** the helper distributes the amount evenly

### Requirement: Credit and debit sign convention

A split SHALL count as a credit (increasing the user's balance) when it was auto-created or when the source transaction amount is positive; otherwise it SHALL count as a debit. When transaction data is available for a split's transaction, the transaction amount SHALL take precedence over the auto-created flag.

#### Scenario: Positive transaction

- **WHEN** a split's transaction has a positive amount
- **THEN** the split is treated as a credit regardless of how it was created

#### Scenario: Negative transaction

- **WHEN** a split's transaction has a negative amount
- **THEN** the split is treated as a debit
