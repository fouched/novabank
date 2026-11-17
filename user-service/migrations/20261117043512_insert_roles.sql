-- +goose Up
INSERT INTO roles (name) VALUES ('ADMIN'), ('CUSTOMER');

-- +goose Down
DELETE FROM roles WHERE name IN ('ADMIN', 'CUSTOMER');
