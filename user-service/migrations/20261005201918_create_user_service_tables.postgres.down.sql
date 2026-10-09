-- Drop relational tables in reverse dependency sequence order
DROP TABLE IF EXISTS accounts;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;

-- Clear custom database type configurations
DROP TYPE IF EXISTS currency_type;
DROP TYPE IF EXISTS account_type;
DROP TYPE IF EXISTS account_status;
