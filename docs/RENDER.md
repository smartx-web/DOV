# Render deployment checklist

The source repository is public. Never commit credentials.

## Required steps
1. Create a Render PostgreSQL database or use another externally reachable managed PostgreSQL service.
2. Apply `migrations/001_init.sql` to that **hosted** database before starting the web service.
3. Create the Render web service from this repository using Dockerfile or render.yaml.
4. In Render environment settings provide `DATABASE_URL` (the hosted database connection URL), `ADMIN_EMAIL`, `ADMIN_PASSWORD_HASH`. Keep these values private.
5. Set `REGISTRATION_ENABLED=false` until the verified ward and polling-unit directory, privacy notice, and abuse protection are ready.
6. Set `COOKIE_SECURE=true` once cookie sessions are implemented. The current code uses Basic Auth and is not safe for production personal data.
7. Verify `/api/health`, `/api/lgas` and the registration-closed response before publishing links.

## Security status
The current repository uses HTTP Basic authentication for administrator API calls, no login throttling and no coordinator role model. It should be treated as a restricted demo, **not a live volunteer-data collection service**. Never use real volunteer details until those controls are implemented.
