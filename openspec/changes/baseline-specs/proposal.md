# Baseline Specs

## Why

The project has been through several rewrites (split logic, ledger redesign, UI overhaul) but has no spec describing what it actually does. Every future change currently starts from reading code. This change captures the system's current behavior as the project's first capability specs, so future changes can be written as deltas against a known baseline.

## What Changes

Documentation-only change — no code, no behavior changes. Delta specs are created for five new capabilities describing behavior as it exists today:

- `authentication` — OIDC login/logout flow, group-to-role mapping, session cookies, dev-mode bypass
- `user-management` — admin CRUD of users, aid classes, and Actual Budget payee mapping
- `expense-splitting` — manual split creation and clearing, even and aid-weighted distribution, auto-credit of payments
- `user-ledger` — per-user ledger view with running balance, credit/debit sign conventions, access rules
- `actual-budget-integration` — tagged transaction and payee fetching, 5-minute cache, cache refresh, error surfacing

Known gaps observed while baselining (aid-weighting lives only in browser JS, auto-split runs on admin page load, no tests) are intentionally **not** specified as requirements here — this change documents current behavior only. Fixing them is future work that can now be proposed as deltas.

## Capabilities

### New Capabilities

- `authentication`: OIDC authentication via Authelia, role mapping from groups, signed session cookies, dev-mode bypass
- `user-management`: admin-managed users with aid classes and Actual Budget payee mapping
- `expense-splitting`: splitting tagged expenses among users, including auto-created credits
- `user-ledger`: read-only per-user ledger of splits with running balance
- `actual-budget-integration`: Actual Budget API access for payees and tagged transactions with caching

### Modified Capabilities

None — this is the first spec in the project.

## Impact

- No code changes; no runtime impact.
- Adds five spec files under `openspec/specs/` once the change is synced/archived.
- Establishes the baseline against which future changes will be diffed.
