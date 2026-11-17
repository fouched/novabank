-- +goose Up
-- 1. Create Custom PostgreSQL Enum Types
CREATE TYPE account_status AS ENUM ('ACTIVE', 'FROZEN', 'CLOSED');
CREATE TYPE account_type   AS ENUM ('SAVINGS', 'CREDIT', 'CURRENT');
CREATE TYPE currency_type   AS ENUM ('USD', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'SEK', 'NZD', 'ZAR');

-- 2. Create Reusable Trigger Function for updated_at
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- 3. Create Roles Table
CREATE TABLE roles (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- 4. Create Users Table
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT UNIQUE NOT NULL,
    password   TEXT NOT NULL,
    first_name TEXT,
    last_name  TEXT,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Create User Roles Join Table
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 6. Create Accounts Table
CREATE TABLE accounts (
    id             BIGSERIAL PRIMARY KEY,
    account_number TEXT UNIQUE NOT NULL,
    balance        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    currency       currency_type NOT NULL,
    account_type   account_type NOT NULL,
    account_status account_status NOT NULL DEFAULT 'ACTIVE',
    user_id        BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 7. Attach Automatic updated_at Triggers
CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

-- 8. Add Performance Index Optimisations
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_accounts_number ON accounts(account_number);

-- +goose Down
DROP TRIGGER IF EXISTS trg_accounts_updated_at ON accounts;
DROP TRIGGER IF EXISTS trg_users_updated_at ON users;

DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

DROP FUNCTION IF EXISTS set_updated_at();

DROP TYPE IF EXISTS currency_type;
DROP TYPE IF EXISTS account_type;
DROP TYPE IF EXISTS account_status;
