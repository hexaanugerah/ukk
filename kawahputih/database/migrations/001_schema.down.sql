-- Reverse of 001_schema.up.sql — drop everything in dependency order.
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS ticket_verifications;
DROP TABLE IF EXISTS booking_details;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS auth_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS roles;
