-- Ensure the email column exists on the user table.
-- Migration 000002 added it via ALTER TABLE, but if that ran against a pre-existing
-- table or failed partway, the column may be absent.
ALTER TABLE "user" ADD COLUMN IF NOT EXISTS "email" text NOT NULL DEFAULT '';
ALTER TABLE "user" ALTER COLUMN "email" DROP DEFAULT;
CREATE UNIQUE INDEX IF NOT EXISTS "user_email_uidx" ON "user" ("email");
