# Weighted Splits

## Why

Splitting an expense today means either accepting the fixed "Evenly" or "Evenly w/ Aid" schemes (needs-help pays $0, will-help covers everything) or typing every amount by hand. Treasurers need a middle ground: proportional weights per participant, with sane defaults per aid class and the ability to tweak individual users when needed.

## What Changes

UI-only change — no database schema changes, no backend or API changes. The existing split dialog gains a third allocation helper:

- A **"By Weight"** button in the split dialog, next to "Evenly" and "Evenly w/ Aid", that fills each participant's amount proportionally to a weight.
- **Default weight**: 1.0 for every user regardless of aid class, so untouched weights behave like an even split; aid classes only affect the aid helpers. (Assumption: neutral default, adjustable per user below.)
- **Per-user overrides**: each participant row in the dialog shows an editable weight (initialized from the user's override or their class default). Editing it persists an override in the browser's localStorage — no server or schema changes. A reset control clears the override back to the class default.
- Remainder cents from rounding are distributed one cent at a time in participant order, matching the existing even-distribution helper.
- **Non-goals**: payments/credits keep the existing auto-credit behavior; "Evenly" and "Evenly w/ Aid" remain unchanged; no new settings page — overrides live in the split dialog only.

Known limitation accepted for this change: overrides live in the admin's browser, so they are per-browser and not shared across admins. That is the direct consequence of the no-schema-change constraint.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `expense-splitting`: adds weighted distribution of a transaction amount among participants, weight resolution from aid class with per-user browser-stored overrides, and remainder handling.

## Impact

- `templates/admin.html` — split dialog UI and the Alpine.js component (weight inputs, "By Weight" button, localStorage read/write).
- No backend, database, or API changes. Submitted split amounts flow through the existing `POST /admin/splits` unchanged.
- Existing specs for `expense-splitting` gain two requirements; no existing requirements change.
