--UP
DO $$ BEGIN
    CREATE TYPE SECURITY_LEVEL AS ENUM ('mirror', 'behavioural_analysis', 'attestations', 'repro_builds_external', 'repro_builds_internal');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE package_versions ADD COLUMN IF NOT EXISTS security_level SECURITY_LEVEL DEFAULT 'mirror';

ALTER TABLE organization_packages ADD COLUMN IF NOT EXISTS security_level SECURITY_LEVEL DEFAULT 'mirror';
