# NovaBank

A microservices banking platform built with Go.

## Project Structure

```text
novabank/
├── http/                     # HTTP Client integration test suites
├── shared/                   # Shared packages (generic TOML config loader, common utilities)
├── notification-service/     # Handles email and event-driven notifications
├── transaction-service/      # Handles deposits, withdrawals, and transfers
├── user-service/             # Handles user registration, authentication, and accounts
├── .gitignore
└── README.md
```

Each service follows a standard internal layout (`cmd/`, `config/`, `internal/`, `migrations/`, `platform/`).

## Prerequisites

* **Go** (1.23+)
* **PostgreSQL**
* **Goose** (for database migrations)
* **Kafka / Redpanda** (for event streaming between services)
* **Consul** (optional, for service discovery)

## Getting Started (New Checkout)

### 1. Clone the Repository
```bash
git clone <repository-url>
cd novabank
```

### 2. Configure Local Secrets (Per Service)
Each service stores its committed non-sensitive defaults in `<service>/config/config.dev.toml` and merges a git-ignored `<service>/config/config.dev.private.toml` file on startup via `shared/config`.

Create a local `config.dev.private.toml` inside each service's `config/` directory (`user-service`, `transaction-service`, `notification-service`):

```bash
# Example for user-service
cp user-service/config/config.dev.toml user-service/config/config.dev.private.toml
```

Edit each `config.dev.private.toml` to supply your local credentials and secrets (removing any non-secret fields you don't need to override):

```toml
[postgres]
user = "postgres"
password = "your_local_password"

[authority]
jwt_secret = "your_local_jwt_secret"
```
*(Note: `*.private.toml` is ignored by Git so local credentials are never committed.)*

### 3. Database Setup
Ensure PostgreSQL is running locally and create the databases for each service:
* `novabank_user_db`
* `novabank_transaction_db`
* `novabank_notification_db`

Run migrations inside each service directory using Soda:
```bash
cd user-service && goose migrate
cd ../transaction-service && goose migrate
cd ../notification-service && goose migrate
```

### 4. Running Services Locally
By default, services run in `prod` mode (`NOVABANK_ENV=prod`). Set your IDE run configuration to dev, or export an environmental variable `NOVABANK_ENV=dev`

You can run each service from its own directory:

```bash
cd user-service
go run ./cmd/api
```

### 5. Running HTTP Client Tests
End-to-end API test suites are located in the top-level `http/` directory.
* Run requests sequentially in GoLand (e.g., `03_profile.http` before `08_accounts_admin.http`) so session variables like `{{jwt}}` and `{{account_number}}` populate in memory.
* Keep any local `http/http-client.private.env.json` file excluded from Git commits.

## Environment Overrides

In non-local environments, set `NOVABANK_ENV` (e.g., `prod`) and inject secrets via environment variables:
* `NOVABANK_ENV` — Target environment (`dev`, `prod`)
* `NOVABANK_DB_PASSWORD` — PostgreSQL password override
