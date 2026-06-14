# 🥏 Who Owes Me

A lightweight, locally-hosted web application designed to help frisbee team treasurers split expenses directly from an [Actual Budget](https://actualbudget.org/) instance.

This application acts as a pure "Split Manager" UI on top of Actual Budget. It relies entirely on Actual Budget as the source of truth for financial data, avoiding complex data synchronization problems, and utilizes OIDC (like Authelia) for authentication and role management.

## Tech Stack
* **Backend**: Golang (`net/http`, `chi` router)
* **Frontend**: HTML Templates, HTMX, Alpine.js, Bulma CSS
* **Database**: SQLite (strictly for user mappings and split percentages)
* **Auth**: OIDC (Authelia recommended)

---

## 🧪 Running the Local Test Environment

To verify functionality without impacting your live servers, this repository includes a complete, self-contained Docker Compose setup containing:
1. The **Who Owes Me** App
2. A local **Authelia** instance (with pre-configured test users)
3. A local **Actual Server** instance

### Step 1: Start the Environment

```bash
docker-compose -f docker-compose.test.yml up -d
```

### Step 2: Bootstrap Actual Budget

1. Open your browser to the local Actual server: **http://localhost:5006**
2. It will prompt you to set an initial server password. **You must set it to `testpassword`** (this is hardcoded in the test `docker-compose.test.yml` for `actual-http-api` to authenticate).
3. Create a new dummy budget file.
4. Click on **Budget name (top left)** -> **Settings** -> **Advanced**, and click **Enable Advanced Features**.
5. In that same Advanced settings menu, copy your **Sync ID**.

### Step 3: Configure the Test Environment

1. Open `.env.test` in the root of the project.
2. Paste your copied Budget ID (the API key is already hardcoded to `test_api_key` for testing):
   ```env
   ACTUAL_BUDGET_ID=your_copied_budget_id
   ```
3. Restart the Go application to pick up the new Budget ID:
   ```bash
   docker-compose -f docker-compose.test.yml restart who-owes-me
   ```

### Step 4: Test the Workflow

1. Go to the Who Owes Me App: **http://app.localhost:8080/login**
2. You will be automatically redirected to your local Authelia instance (`http://auth.app.localhost:9091`).
3. Log in with the pre-configured Admin test account:
   * **Username:** `admin`
   * **Password:** `password`
4. You will be redirected back to the Who Owes Me Admin dashboard. 
5. *(Optional)* Add dummy payees and transactions in your Actual Budget instance at `:5006` and watch them populate dynamically on the Admin dashboard!

*(To test the read-only user view, log into Authelia with the username `user` and password `password`.)*

---

## 🚀 Production Deployment

### 1. Configure Environment

Copy `.env.example` to `.env` in the project root and fill in your production Authelia and Actual Budget details:

```env
PORT=8080
DB_PATH=data.db
APP_ENV=development

# OIDC (Authelia) Configuration
OIDC_ISSUER_URL=https://auth.yourdomain.com
OIDC_CLIENT_ID=who-owes-me
OIDC_CLIENT_SECRET=your_client_secret_here
OIDC_REDIRECT_URL=http://localhost:8080/callback

# Actual Budget Configuration
ACTUAL_SERVER_URL=https://actual.yourdomain.com
ACTUAL_API_KEY=your_actual_api_key_here
ACTUAL_BUDGET_ID=your_budget_file_id_here

SPLIT_TAG=#gsu2026
```

### 2. Configure Authelia

Ensure you have created the `who-owes-me` OIDC client in your production Authelia `configuration.yml`.

### 3. Run via Docker Compose

**a) From GitHub Container Registry (recommended)**

The CI publishes images to `ghcr.io/<your-repo>/who-owes-me`. To use it, create a `docker-compose.yml`:

```yaml
services:
  who-owes-me:
    image: ghcr.io/your-org/who-owes-me:latest
    container_name: who-owes-me
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./.env:/app/.env
    restart: unless-stopped
```

> **Authenticating to GHCR**: If the repo is private, authenticate on the host:
> ```
> echo $GITHUB_TOKEN | docker login ghcr.io -u your-username --password-stdin
> ```

**Secrets via `_FILE` env vars (recommended over `.env` for secrets):**

Any env var can be provided via `<NAME>_FILE` pointing to a Docker secret or bind-mounted file. This is the preferred approach for sensitive values:

```yaml
services:
  who-owes-me:
    image: ghcr.io/your-org/who-owes-me:latest
    container_name: who-owes-me
    ports:
      - "8080:8080"
    volumes:
      - ./data:/app/data
    environment:
      - OIDC_ISSUER_URL=https://auth.yourdomain.com
      - OIDC_CLIENT_ID=who-owes-me
      - OIDC_CLIENT_SECRET_FILE=/run/secrets/oidc_client_secret
      - OIDC_REDIRECT_URL=http://localhost:8080/callback
      - ACTUAL_SERVER_URL=https://actual.yourdomain.com
      - ACTUAL_API_KEY_FILE=/run/secrets/actual_api_key
      - ACTUAL_BUDGET_ID=your_budget_file_id
      - SPLIT_TAG=#gsu2026
    secrets:
      - oidc_client_secret
      - actual_api_key
    restart: unless-stopped

secrets:
  oidc_client_secret:
    file: ./secrets/oidc_client_secret.txt
  actual_api_key:
    file: ./secrets/actual_api_key.txt
```

**b) Build locally**

```bash
docker compose up -d --build
```