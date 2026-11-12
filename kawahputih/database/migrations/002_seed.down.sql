-- Reverse of 002_seed.up.sql.
DELETE FROM tickets WHERE id BETWEEN 1 AND 20;
DELETE FROM roles WHERE id BETWEEN 1 AND 4;
