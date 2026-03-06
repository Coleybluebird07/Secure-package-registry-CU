#!/bin/bash
set -e

echo "Starting core_db..."
podman compose up core_db -d

echo "Waiting for database to be ready..."
until podman compose exec core_db pg_isready -U core -d secure_registry -q; do
  sleep 1
done
echo "Database is ready."

echo "Running migrations..."

git show auth-schema:infra/migrations/000001_initial_schema.up.sql \
  | podman compose exec -T core_db psql -U core -d secure_registry -q

git show auth-schema:infra/migrations/000002_better_auth_organization.up.sql \
  | podman compose exec -T core_db psql -U core -d secure_registry -q 2>/dev/null || true

podman compose exec -T core_db psql -U core -d secure_registry -q << 'EOF'
CREATE TABLE IF NOT EXISTS "user" (
    "id" text NOT NULL PRIMARY KEY,
    "email" text NOT NULL UNIQUE,
    "name" text NOT NULL,
    "emailVerified" boolean NOT NULL DEFAULT false,
    "image" text,
    "createdAt" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updatedAt" timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL
);
CREATE TABLE IF NOT EXISTS "session" ("id" text not null primary key, "expiresAt" timestamptz not null, "token" text not null unique, "createdAt" timestamptz default CURRENT_TIMESTAMP not null, "updatedAt" timestamptz not null, "ipAddress" text, "userAgent" text, "userId" text not null references "user" ("id") on delete cascade, "activeOrganizationId" text);
CREATE TABLE IF NOT EXISTS "account" ("id" text not null primary key, "accountId" text not null, "providerId" text not null, "userId" text not null references "user" ("id") on delete cascade, "accessToken" text, "refreshToken" text, "idToken" text, "accessTokenExpiresAt" timestamptz, "refreshTokenExpiresAt" timestamptz, "scope" text, "password" text, "createdAt" timestamptz default CURRENT_TIMESTAMP not null, "updatedAt" timestamptz not null);
CREATE TABLE IF NOT EXISTS "member" ("id" text not null primary key, "organizationId" text not null references "organization" ("id") on delete cascade, "userId" text not null references "user" ("id") on delete cascade, "role" text not null, "createdAt" timestamptz not null);
CREATE TABLE IF NOT EXISTS "invitation" ("id" text not null primary key, "organizationId" text not null references "organization" ("id") on delete cascade, "email" text not null, "role" text, "status" text not null, "expiresAt" timestamptz not null, "createdAt" timestamptz default CURRENT_TIMESTAMP not null, "inviterId" text not null references "user" ("id") on delete cascade);
CREATE INDEX IF NOT EXISTS "session_userId_idx" on "session" ("userId");
CREATE INDEX IF NOT EXISTS "account_userId_idx" on "account" ("userId");
CREATE INDEX IF NOT EXISTS "member_organizationId_idx" on "member" ("organizationId");
CREATE INDEX IF NOT EXISTS "member_userId_idx" on "member" ("userId");
CREATE INDEX IF NOT EXISTS "invitation_organizationId_idx" on "invitation" ("organizationId");
CREATE INDEX IF NOT EXISTS "invitation_email_idx" on "invitation" ("email");
EOF

echo "Migrations complete."
echo ""
echo "Run 'cd home-ui && bun run dev' to start the app."
