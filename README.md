# David Olofu Volunteer Portal

Go/Gin backend, PostgreSQL schema, responsive volunteer form and prototype administrator dashboard.

## Install PostgreSQL (Ubuntu)
```bash
sudo apt update
sudo apt install postgresql postgresql-client
sudo systemctl enable --now postgresql
sudo -u postgres createuser --pwprompt dov_user
sudo -u postgres createdb -O dov_user dov
```

## Run
```bash
git clone https://github.com/smartx-web/DOV.git
cd DOV
cp .env.example .env
# Edit .env and set your DATABASE_URL with your chosen password
psql "postgres://dov_user:YOUR_PASSWORD@localhost:5432/dov?sslmode=disable" -v ON_ERROR_STOP=1 -f migrations/001_init.sql
go mod tidy
set -a; source .env; set +a
go run ./cmd/server
```
Open http://localhost:8080. Verify http://localhost:8080/api/health.

## Official INEC electoral data
Only the nine LGAs are seeded. Import verified wards and polling units yourself before enabling registration. Example (replace placeholders with verified names):
```sql
INSERT INTO wards(lga_id,name) SELECT id,'VERIFIED WARD' FROM lgas WHERE name='Agatu';
INSERT INTO polling_units(ward_id,code,name) SELECT id,'VERIFIED CODE','VERIFIED POLLING UNIT' FROM wards WHERE name='VERIFIED WARD';
```

## Admin configuration
For local testing only, generate a bcrypt hash for your own password using a trusted bcrypt tool. Set ADMIN_EMAIL and ADMIN_PASSWORD_HASH in .env. The admin screen uses HTTP Basic authentication and is **not production secure**.

**Do not publicly deploy or collect real volunteer data yet.** Before launch: implement secure sessions, HTTPS, CSRF protection, rate limiting, audit logs, backups, privacy notice, data retention and role-based access. No Google Sheets integration yet. No real INEC ward/polling-unit dataset has been imported.
