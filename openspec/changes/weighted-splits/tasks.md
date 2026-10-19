## 1. Weight state and persistence

- [x] 1.1 Add weight resolution to the split-dialog Alpine component in `templates/admin.html`: `weights[p.id]` per participant resolved from localStorage override (`wom_weight_override_<userID>`) or the aid-class default (0.5 / 1.0 / 1.5), with try/catch localStorage helpers per design D1/D2/D5. Verify by opening the dialog: a participant with no override shows 0.5 for `needs_help`, 1.0 for `regular`, 1.5 for `will_help`.
- [x] 1.2 Add an editable weight field and a reset control to each participant row. Verify: editing a weight and reloading the page (then reopening the dialog) shows the edited value; resetting shows the class default again.

## 2. Weighted allocation helper

- [x] 2.1 Add a "By Weight" button next to "Evenly" and "Evenly w/ Aid" that fills amounts via `trunc(total × wᵢ / Σw)` with remainder cents distributed one at a time in participant order (design D3). Verify: $9.00 at weights 0.5/1.0/1.5 fills $1.50/$3.00/$4.50; $10.00 at the same weights fills amounts that sum to exactly $10.00.
- [x] 2.2 Handle edge weights: zero-weight participants get $0.00, and an all-zero weight sum falls back to `calculateEvenly()`. Verify: a zero-weight participant shows $0.00 and no NaN appears in any amount field; all-zero weights produce a correct even split.

## 3. Regression and validation

- [x] 3.1 Run `openspec validate weighted-splits` and confirm no errors, then in the local test environment (`docker-compose.test.yml`) create a split with "By Weight", submit, and verify the ledger and balances update exactly as a manually entered split would. Verify "Evenly" and "Evenly w/ Aid" still behave as before and payments/auto-credit are unaffected.
