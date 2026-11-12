# Backend

## Local development

1. Clone the repository and open this `backend` directory.
2. Start a local MySQL server and create the database named by `DB_NAME` (the example uses `kawah_putih_db`).
3. Copy `.env.example` to `.env`, then configure `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`. Keep `.env` local and do not commit it.
4. If you want login sessions to survive restarts, set unique local values for `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET`. If left blank, the API generates different temporary values each run.
5. Optionally set unique strong `SEED_ADMIN_PASSWORD`, `SEED_MANAGER_PASSWORD`, and `SEED_PETUGAS_PASSWORD` values to create initial staff accounts. No staff account is created when its password is blank; the application has no built-in staff passwords.
6. Start the API with `go run ./cmd/server`. It uses the configured MySQL database and runs schema migrations and seed data at startup.
7. Run the automated checks with `go test ./...`. Unit tests use isolated in-memory SQLite databases and do not connect to MySQL.

## Staging and production

Set `APP_ENV` to `staging` or `production`, then configure the JWT secrets, `MIDTRANS_SERVER_KEY`, and MySQL connection values in the deployment environment. Production requires `MIDTRANS_IS_PRODUCTION=true`; the API fails during startup if required secrets or database settings are missing. The schema migration and seed routine run during startup.

The standalone scheduler uses the configured MySQL connection regardless of `APP_ENV`.

If an existing database was seeded by an earlier version, reset the passwords of any staff accounts created with the old built-in seed credentials. The new seed does not change existing passwords.
