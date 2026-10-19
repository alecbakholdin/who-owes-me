## ADDED Requirements

### Requirement: Weighted distribution

The split dialog SHALL offer a "By Weight" allocation helper that fills each participant's amount in proportion to that participant's weight. Each participant's amount SHALL be the transaction total multiplied by the participant's weight divided by the sum of all participants' weights, truncated to whole cents. Any remainder cents SHALL be distributed one cent at a time to participants in participant order, matching the even distribution helper. A participant with weight zero SHALL receive zero. If all participants have weight zero, the helper SHALL fall back to even distribution instead of producing invalid amounts.

#### Scenario: Weights divide evenly

- **WHEN** a $9.00 transaction is split by weight among participants with weights 0.5, 1.0, and 1.5
- **THEN** the participants are assigned $1.50, $3.00, and $4.50 respectively

#### Scenario: Remainder cents

- **WHEN** a $10.00 transaction is split by weight among the same three participants
- **THEN** the assigned amounts are $1.67, $3.33, and $5.00, summing to exactly $10.00

#### Scenario: Zero-weight participant

- **WHEN** a participant has weight zero
- **THEN** that participant is assigned $0.00 and the full amount is distributed among the remaining participants

#### Scenario: All participants zero weight

- **WHEN** every participant has weight zero and the weighted helper is used
- **THEN** the amount is distributed evenly among the participants

### Requirement: Weight resolution and overrides

Each participant's weight SHALL resolve from a per-user override stored in the admin's browser when one exists, and otherwise from the aid class defaults: `needs_help` = 0.5, `regular` = 1.0, `will_help` = 1.5. The split dialog SHALL display each participant's resolved weight in an editable field. Editing a participant's weight SHALL immediately persist an override for that user to browser storage, so it is applied again when the admin next opens a split dialog. The dialog SHALL provide a way to reset a participant's weight to the aid class default, removing the stored override.

#### Scenario: Override takes precedence

- **WHEN** a `regular` user has a stored weight override of 2.0
- **THEN** the dialog shows their weight as 2.0, not the regular default of 1.0

#### Scenario: No override uses class default

- **WHEN** a `will_help` user has no stored override
- **THEN** the dialog shows their weight as 1.5

#### Scenario: Editing persists the override

- **WHEN** the admin edits a participant's weight and later opens a split dialog for another transaction
- **THEN** the edited weight is used for that user instead of the class default

#### Scenario: Reset clears the override

- **WHEN** the admin resets a participant's weight
- **THEN** the stored override is removed and the weight returns to the aid class default
