# authentication Specification

## Purpose

Authenticates users via OIDC (designed for Authelia), maps group membership to roles, and maintains session cookies.

## Requirements

### Requirement: OIDC login flow

The system SHALL authenticate users through the OIDC authorization code flow: unauthenticated requests to protected routes are redirected to `/login`, which redirects to the configured OIDC provider. The callback endpoint SHALL validate the OAuth state, exchange the authorization code, verify the returned ID token, and fetch user info from the provider.

#### Scenario: Successful login

- **WHEN** a user completes the OIDC flow at the configured provider
- **THEN** the system verifies the ID token, fetches user info, sets session cookies, and redirects them to the home page

#### Scenario: OAuth state mismatch

- **WHEN** the callback receives a `state` parameter that does not match the stored `oauth_state` cookie
- **THEN** the system responds with 400 Bad Request and does not create a session

#### Scenario: Unauthenticated access to a protected route

- **WHEN** a request without a valid session hits a protected route
- **THEN** the system redirects to `/login`

### Requirement: Role mapping from OIDC groups

The system SHALL determine a user's role from the `groups` claim of the provider's user info: membership in the `whoowesme_admin` group grants admin access; all other users are treated as regular users.

#### Scenario: Admin group membership

- **WHEN** the logged-in user's user info contains the `whoowesme_admin` group
- **THEN** the session is marked as admin and admin routes are accessible

#### Scenario: Non-admin user

- **WHEN** the logged-in user's user info does not contain the `whoowesme_admin` group
- **THEN** the session is marked as a regular user and admin routes return 403

### Requirement: Username resolution

The system SHALL identify the user by the `preferred_username` claim, falling back to the ID token subject when the claim is empty. This identifier is used to look up the user record.

#### Scenario: Preferred username available

- **WHEN** the user info contains a non-empty `preferred_username`
- **THEN** the user record is looked up by that username

#### Scenario: Missing preferred username

- **WHEN** `preferred_username` is empty
- **THEN** the ID token subject is used for the user lookup

### Requirement: Session cookies

The system SHALL maintain session state in HTTP-only cookies: an `auth_token` cookie holding the ID token and an `app_session` cookie holding an HMAC-signed `isAdmin|username` value. Cookies SHALL be `SameSite=Lax`, and marked `Secure` when `APP_ENV` is `production`.

#### Scenario: Session cookie signature is invalid

- **WHEN** the `app_session` cookie fails signature verification, including after a server restart (the signing secret is generated per process)
- **THEN** the user is redirected to `/login`

### Requirement: Development bypass

When OIDC is not configured (no `OIDC_ISSUER_URL` or `OIDC_CLIENT_ID`, or provider setup fails at startup), the system SHALL bypass authentication and treat every request as an admin "Dev User", auto-creating that user record if it does not exist.

#### Scenario: OIDC unconfigured

- **WHEN** the server starts without OIDC environment variables and a request hits a protected route
- **THEN** the request is served as the admin Dev User without any login

### Requirement: Unregistered users

A logged-in non-admin user with no matching user record SHALL be denied access with a 403 page. Admins without a user record SHALL be allowed to proceed so they can bootstrap the user list.

#### Scenario: Non-admin without user record

- **WHEN** a regular user logs in but no user record matches their username
- **THEN** they see a 403 "unregistered" page

#### Scenario: Admin without user record

- **WHEN** an admin logs in but no user record matches their username
- **THEN** they can still access the admin dashboard

### Requirement: Logout

The system SHALL clear both session cookies on logout and redirect to the provider's `end_session_endpoint` with the ID token as `id_token_hint` when the provider publishes one; otherwise it SHALL redirect to `/login`.

#### Scenario: Provider supports end-session

- **WHEN** a user logs out and the provider advertises an `end_session_endpoint`
- **THEN** cookies are cleared and the browser is redirected to that endpoint with `id_token_hint` set

#### Scenario: No end-session endpoint

- **WHEN** a user logs out and the provider has no `end_session_endpoint`
- **THEN** cookies are cleared and the browser is redirected to `/login`
