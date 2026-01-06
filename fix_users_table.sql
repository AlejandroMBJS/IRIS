-- Fix corrupted foreign key constraints in users table
-- SQLite doesn't support dropping individual constraints, so we need to recreate the table

BEGIN TRANSACTION;

-- Create new users table with correct schema
CREATE TABLE "users_new" (
    `id` text,
    `created_at` datetime,
    `updated_at` datetime,
    `deleted_at` datetime,
    `email` varchar(255) NOT NULL,
    `password_hash` varchar(255) NOT NULL,
    `role` text NOT NULL,
    `full_name` varchar(255) NOT NULL,
    `is_active` numeric DEFAULT true,
    `company_id` text NOT NULL,
    `last_login_at` datetime,
    `employee_id` text,
    `supervisor_id` text,
    `department` varchar(100),
    `area` varchar(100),
    PRIMARY KEY (`id`),
    CONSTRAINT `fk_users_employee` FOREIGN KEY (`employee_id`) REFERENCES `employees`(`id`),
    CONSTRAINT `fk_users_company` FOREIGN KEY (`company_id`) REFERENCES `companies`(`id`),
    CONSTRAINT `fk_users_supervisor` FOREIGN KEY (`supervisor_id`) REFERENCES `users_new`(`id`)
);

-- Copy all data from old table to new table
INSERT INTO users_new
SELECT id, created_at, updated_at, deleted_at, email, password_hash, role, full_name,
       is_active, company_id, last_login_at, employee_id, supervisor_id, department, area
FROM users;

-- Drop old table
DROP TABLE users;

-- Rename new table to users
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes
CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_company_id ON users(company_id);
CREATE INDEX idx_users_employee_id ON users(employee_id);

COMMIT;
