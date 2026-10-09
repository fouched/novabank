-- 1. Create Custom PostgreSQL Enum Types
CREATE TYPE account_status AS ENUM ('ACTIVE', 'FROZEN', 'CLOSED');
CREATE TYPE account_type   AS ENUM ('SAVINGS', 'CREDIT', 'CURRENT');
CREATE TYPE currency_type   AS ENUM ('USD', 'EUR', 'GBP', 'JPY', 'AUD', 'CAD', 'CHF', 'CNY', 'SEK', 'NZD', 'ZAR');

-- 2. Create Roles Table
CREATE TABLE roles (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- 3. Create Users Table
CREATE TABLE users (
    id         BIGSERIAL PRIMARY KEY,
    email      TEXT UNIQUE NOT NULL,
    password   TEXT NOT NULL, -- Will store secure hashed bcrypt string strings
    first_name TEXT,
    last_name  TEXT,
    enabled    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Create User Roles Join Table (Handles Many-to-Many Relationship)
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 5. Create Accounts Table (Handles One-to-One Relationship with User)
CREATE TABLE accounts (
    id             BIGSERIAL PRIMARY KEY,
    account_number TEXT UNIQUE NOT NULL, -- Generous limit accommodates international formats
    balance        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000, -- High precision numeric avoids floating point tracking inaccuracies
    currency       currency_type NOT NULL,
    account_type   account_type NOT NULL,
    account_status account_status NOT NULL DEFAULT 'ACTIVE',
    user_id        BIGINT UNIQUE NOT NULL REFERENCES users(id) ON DELETE RESTRICT, -- Unique constraint enforces 1:1 rule layout
    created_at     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 6. Add Performance Visual Index Optimisations
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_accounts_number ON accounts(account_number);
