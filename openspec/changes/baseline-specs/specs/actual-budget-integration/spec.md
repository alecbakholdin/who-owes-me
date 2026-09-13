## Purpose

Reads payees and transactions from an Actual Budget server via its HTTP API, with a short-lived cache to avoid repeated API calls.

## ADDED Requirements

### Requirement: Actual Budget connection

The system SHALL connect to the Actual Budget HTTP API using the `ACTUAL_SERVER_URL`, `ACTUAL_API_KEY`, and `ACTUAL_BUDGET_ID` environment variables. API calls SHALL fail with an error when the budget ID is not configured.

#### Scenario: Budget ID missing

- **WHEN** a request to the Actual API is made without `ACTUAL_BUDGET_ID` configured
- **THEN** the call returns an error instead of making a network request

### Requirement: Payee fetching

The system SHALL fetch the budget's payee list, each with an ID and name, for display and for payee-to-user mapping.

#### Scenario: Payees listed

- **WHEN** payees are fetched successfully
- **THEN** the admin dashboard shows each payee's name

### Requirement: Tagged transaction fetching

The system SHALL fetch all transactions whose notes contain the split tag, along with their date, amount in cents, payee, and notes.

#### Scenario: Tagged transactions fetched

- **WHEN** transactions are fetched for the configured split tag
- **THEN** only transactions whose notes contain the tag are returned

### Requirement: Response caching

Payee lists and tagged transaction lists SHALL be cached in memory for 5 minutes. Subsequent requests within that window SHALL be served from the cache without calling the Actual API.

#### Scenario: Cache hit

- **WHEN** payees or tagged transactions are requested twice within 5 minutes
- **THEN** the Actual API is called only once

#### Scenario: Cache expiry

- **WHEN** more than 5 minutes have passed since a list was cached
- **THEN** the next request fetches fresh data from the Actual API

### Requirement: Cache refresh

Admins SHALL be able to clear the cache, forcing the next load to fetch fresh data from Actual Budget.

#### Scenario: Manual refresh

- **WHEN** an admin triggers the cache refresh action
- **THEN** all cached Actual API data is discarded and the admin dashboard reloads with fresh data

### Requirement: API error surfacing

When fetching payees or transactions fails, the admin dashboard SHALL still render and display a message describing the failure. The user ledger SHALL remain available using stored split data.

#### Scenario: Actual API unreachable

- **WHEN** the Actual API cannot be reached while loading the admin dashboard
- **THEN** the dashboard renders with an error message about the failed fetch

### Requirement: Transaction data enrichment

Stored splits SHALL be enriched with their source transaction's data when the transaction is still available from Actual Budget, including overriding credit/debit determination by the transaction's amount.

#### Scenario: Split enriched from transaction

- **WHEN** a split's transaction is present in the fetched tagged transactions
- **THEN** the ledger row and balance calculations use the transaction's amount sign to determine credit or debit
