# Agent Directives

## Development & Hot-Reloading (`air`)

- **Do NOT kill the `air` process.** The project uses `air` for hot-reloading. When you modify Go code (`*.go`) or templates (`*.html`), `air` will automatically detect the changes, rebuild the `tmp/main` binary, and restart the backend server.
- Killing `tmp/main` manually or trying to restart the server with `go build` is unnecessary and disrupts the automated development workflow. Just save your changes and allow `air` to do its job.

## Project Context

**Who Owes Me** is a lightweight, locally-hosted web application built to help treasurers split expenses directly from an [Actual Budget](https://actualbudget.org/) instance.

- **Source of Truth:** It relies entirely on Actual Budget for financial data (payees, transactions). The app does not manage complex financial state, it only manages the mapping of "who owes what" via SQLite.
- **Auto-Splits:** It reads transactions from Actual Budget using a specific tag (e.g., `#gsu2026`). It treats positive amounts as payments to credit users and allows admins to split negative amounts.
- **Roles:** Authelia (OIDC) handles authentication and role management (`whoowesme_admin` vs `whoowesme_user`).

## Tech Stack
- **Backend:** Golang (`net/http`, `go-chi/chi`)
- **Frontend:** HTML Templates, HTMX, Alpine.js, Bulma CSS (1.0+)
- **Database:** SQLite (`data.db`) for tracking users, mappings to Actual Budget payees, and split amounts.
- **Auth:** OIDC (designed for Authelia)

## Key Concepts
- **Users:** Represent individuals in the application. They have different Aid Classes (Regular, Needs Help, Will Help) determining how expenses are dynamically split among them.
- **Expense Splits:** Manual splits distribute debt for a transaction among participants. Auto-created splits automatically credit a user when they make a payment.
- **Actual Budget Payees:** Users map directly to "Payees" in Actual Budget to sync their deposits and transactions.

## Making UI Changes
- The app uses Bulma 1.0 which natively supports dark mode via the `data-theme="dark"` attribute on the `<html>` tag. Ensure styling and contrast works properly in both light and dark modes.
- Avoid introducing complex JavaScript frameworks. Stick to Alpine.js for interactivity and HTMX for async requests.