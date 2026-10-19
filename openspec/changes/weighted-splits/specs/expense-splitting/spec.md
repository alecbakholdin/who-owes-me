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

### Requirement: Weighted distribution with aid

The split dialog SHALL offer a "By Weight w/ Aid" allocation helper that distributes the transaction amount by weight, except that participants with aid class `needs_help` SHALL be assigned zero and their weighted shares SHALL be redistributed among `will_help` participants in proportion to the will-help participants' weights, with remainder cents distributed one cent at a time in participant order among positive-weight will-help participants. When there are no `needs_help` participants, or no `will_help` participant has positive weight, the helper SHALL produce the plain weighted distribution instead. When all participants have zero weight, the helper SHALL fall back to even distribution.

#### Scenario: Single will-help absorbs the shortfall

- **WHEN** a $200.00 transaction is split among a `needs_help` (weight 0.5), a `regular` (weight 1.0), and a `will_help` (weight 1.5) participant using the aid-weighted helper
- **THEN** the needs-help participant is assigned $0.00, the regular participant keeps their weighted $66.66, and the will-help participant is assigned $133.34

#### Scenario: Shortfall splits proportionally among will-help

- **WHEN** a $100.00 transaction is split among a `needs_help` (weight 1.0), a `regular` (weight 1.0), and two `will_help` participants (weights 1.0 and 2.0) using the aid-weighted helper
- **THEN** the needs-help participant is assigned $0.00, the regular participant $20.00, and the will-help participants $26.67 and $53.33 respectively

#### Scenario: No needs-help participant

- **WHEN** the participants contain no `needs_help` user
- **THEN** the helper distributes the amount exactly as the plain weighted helper would

#### Scenario: No positive-weight will-help participant

- **WHEN** the participants include a `needs_help` user but no `will_help` participant has positive weight
- **THEN** the helper distributes the amount exactly as the plain weighted helper would

### Requirement: Weight resolution and overrides

Each participant's weight SHALL resolve from a per-user override stored in the admin's browser when one exists, and otherwise from the default weight of 1.0, which applies to every aid class. The split dialog SHALL display each participant's resolved weight in an editable field. Editing a participant's weight SHALL immediately persist an override for that user to browser storage, so it is applied again when the admin next opens a split dialog. The dialog SHALL provide a way to reset a participant's weight to the default of 1.0, removing the stored override.

#### Scenario: Override takes precedence

- **WHEN** a user has a stored weight override of 2.0
- **THEN** the dialog shows their weight as 2.0, not the default of 1.0

#### Scenario: No override uses default

- **WHEN** a user of any aid class has no stored override
- **THEN** the dialog shows their weight as 1.0

#### Scenario: Editing persists the override

- **WHEN** the admin edits a participant's weight and later opens a split dialog for another transaction
- **THEN** the edited weight is used for that user instead of the class default

#### Scenario: Reset clears the override

- **WHEN** the admin resets a participant's weight
- **THEN** the stored override is removed and the weight returns to the aid class default

#### Scenario: Import with explicit weight

- **WHEN** the admin imports participants and a line contains a name followed by a comma and a non-negative number (e.g. `Bill Sanborn, 2`)
- **THEN** that participant's weight is set to the number and persisted as an override

#### Scenario: Import without weight

- **WHEN** the admin imports a line containing only a name
- **THEN** the participant's weight resolves from their stored override or aid class default
