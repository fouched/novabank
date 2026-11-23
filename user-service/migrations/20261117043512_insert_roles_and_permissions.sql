-- +goose Up
-- 1. Seed Roles
INSERT INTO roles (name) VALUES
    ('ADMIN'),
    ('CUSTOMER');

-- 2. Seed Permissions
INSERT INTO permissions (name) VALUES
    ('user.read.own'),
    ('user.read.all'),
    ('user.update'),
    ('user.delete'),
    ('user.search'),
    ('user.stats'),
    ('account.read.own'),
    ('account.read.all'),
    ('account.update'),
    ('account.delete'),
    ('account.search'),
    ('account.stats');

-- 3. Assign All Permissions to ADMIN
INSERT INTO role_permissions (role_id, permissions_id)
SELECT r.id, p.id
FROM roles r
    CROSS JOIN permissions p
WHERE r.name = 'ADMIN';

-- 4. Assign Customers Permission - possibly do below for ADMIN if you want to reduce their permissions
INSERT INTO role_permissions (role_id, permissions_id)
SELECT r.id, p.id
FROM (
    VALUES
        ('CUSTOMER', 'account.read.own'),
        ('CUSTOMER', 'user.read.own')
) AS rp(role_name, permission_name)
JOIN roles r ON r.name = rp.role_name
JOIN permissions p ON p.name = rp.permission_name;


-- +goose Down
TRUNCATE TABLE role_permissions, user_roles, roles, permissions RESTART IDENTITY;
