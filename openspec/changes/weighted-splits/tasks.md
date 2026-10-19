## 1. Weight state and persistence

- [x] 1.1 Add weight resolution to the split-dialog Alpine component in `templates/admin.html`: `weights[p.id]` per participant resolved from localStorage override (`wom_weight_override_<userID>`) or the default weight 1.0 for every aid class, with try/catch localStorage helpers per design D1/D2/D5. Verify by opening the dialog: a participant with no override shows 1.0 regardless of aid class.
- [x] 1.2 Add an editable weight field and a reset control to each participant row. Verify: editing a weight and reloading the page (then reopening the dialog) shows the edited value; resetting shows the class default again.

## 2. Weighted allocation helper

- [x] 2.1 Add a "By Weight" button next to "Evenly" and "Evenly w/ Aid" that fills amounts via `trunc(total × wᵢ / Σw)` with remainder cents distributed one at a time in participant order (design D3). Verify: $9.00 at weights 0.5/1.0/1.5 fills $1.50/$3.00/$4.50; $10.00 at the same weights fills amounts that sum to exactly $10.00.
- [x] 2.2 Handle edge weights: zero-weight participants get $0.00, and an all-zero weight sum falls back to `calculateEvenly()`. Verify: a zero-weight participant shows $0.00 and no NaN appears in any amount field; all-zero weights produce a correct even split.

## 3. Regression and validation

- [x] 3.1 Run `openspec validate weighted-splits` and confirm no errors, then in the local test environment (`docker-compose.test.yml`) create a split with "By Weight", submit, and verify the ledger and balances update exactly as a manually entered split would. Verify "Evenly" and "Evenly w/ Aid" still behave as before and payments/auto-credit are unaffected.
- [x] 3.2 Extend participant import to accept an optional per-line weight (`Bill Sanborn, 2`), applied via the override path; a bare name still uses the override/class default. Verify in the test environment that both forms import correctly and a non-numeric second column still parses as the legacy comma-separated name list.
- [x] 3.3 Add a "By Weight w/ Aid" helper that zeroes needs_help shares and redistributes the shortfall among positive-weight will_help participants proportional to their weights, falling back to plain weighted distribution without needs_help or positive-weight will_help. Verify the spec scenarios ($200 → $0/$66.66/$133.34; $100 → $0/$20.00/$26.67/$53.33) and that "By Weight" and the other helpers are unaffected.
