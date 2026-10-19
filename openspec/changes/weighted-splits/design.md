# Design: Weighted Splits

## Context

The split dialog lives in `templates/admin.html` as an Alpine.js component with `participants` (id, name, aid_class), `splits[p.id]` (cents), `splitDollars[p.id]` (display), and `totalAmount` (cents). Two helpers exist: `calculateEvenly()` and `calculateEvenlyWithAid()`. Submitted amounts flow through the existing `POST /admin/splits` form unchanged. The constraint for this change: **no database schema changes** — all new state is browser-local.

## Goals / Non-Goals

**Goals:**
- A "By Weight" helper that fills amounts proportionally to a weight per participant.
- Weight = per-user override when set, else aid-class default (needs_help 0.5, regular 1.0, will_help 1.5).
- Overrides editable inline in the dialog and persisted in the admin's browser.

**Non-Goals:**
- No backend, DB, or API changes; no changes to payments/credits; no changes to Evenly / Evenly w/ Aid; no separate settings page; overrides are not shared between admins or browsers (accepted limitation of the constraint).

## Decisions

**D1. Override storage: localStorage, keyed by user ID.**
Key format: `wom_weight_override_<userID>`, value = decimal weight string (e.g. `"1.5"`). Keyed by user ID rather than name so renames don't lose overrides; the dialog already has user IDs. Namespaced to avoid collisions with other localStorage users of the domain.
*Alternative considered*: cookies or per-transaction-only weights. Rejected — cookies add size/expiry complexity for the same single-browser scope; per-transaction weights don't persist and miss the point of defaults+overrides.

**D2. Default weights: 0.5 / 1.0 / 1.5.**
Chosen so a needs-help participant pays half a regular share and a will-help participant pays one-and-a-half, giving a strictly proportional middle ground between "Evenly" (all 1×) and "Evenly w/ Aid" (needs-help 0×, will-help absorbs). Defaults are the spec'd contract; overrides tune per user.

**D3. Allocation math mirrors the even helper.**
`amount = trunc(total × wᵢ / Σw)` in cents; remainder cents distributed one at a time in participant order (same pattern as `calculateEvenly`'s `Math.sign(remainder)` loop), so all helpers behave consistently. Weights are non-negative decimals; a zero weight contributes zero. If Σw = 0 (all weights zero), fall back to `calculateEvenly()` to avoid division by zero.
*Alternative considered*: round-to-nearest with largest-remainder distribution. Rejected — inconsistent with the existing helper and harder to reason about by hand.

**D4. Override lifecycle: persist on edit, explicit reset.**
The weight input writes to localStorage on change (no save button — the dialog already works in immediate-feedback style). A reset control per row removes the key, reverting to the class default. Invalid/empty input is treated as "no override" (class default) rather than stored.
*Alternative considered*: persist only when "By Weight" is clicked. Rejected — an override you set but then submit with "Evenly" would silently vanish.

**D5. State shape in the Alpine component.**
Add `weights[p.id]` (resolved number) plus helper functions `resolveWeight(aidClass, userId)`, `saveWeightOverride(userId, value)`, `clearWeightOverride(userId)`, `calculateByWeight()`. All reads of localStorage go through a try/catch (private-mode browsers can throw on write) and fall back to defaults on any error.

## Risks / Trade-offs

- **[Overrides are per-browser]** → Accepted consequence of the no-schema-change constraint; documented in the proposal. If overrides need to be shared later, promoting them to a `users` column is a future schema change.
- **[Weights produce odd cents]** → Mitigated by remainder-in-participant-order; total always reconciles exactly, as with the even helper.
- **[Admin confusion: which helper does what]** → The weight field is visible next to each participant in the dialog, and the existing two buttons stay unchanged, so "By Weight" is purely additive.

## Migration Plan

None — UI-only. Browsers without stored keys simply see class defaults on first use.

## Open Questions

None.
