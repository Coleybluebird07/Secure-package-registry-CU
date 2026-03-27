--UP
CREATE TYPE SECURITY_LEVEL AS ENUM ('mirror', 'behavioural_analysis', 'attestations', 'repro_builds_external', 'repro_builds_internal');

ALTER TABLE package_versions ADD COLUMN security_level SECURITY_LEVEL DEFAULT 'mirror';

ALTER TABLE organization_packages ADD COLUMN security_level SECURITY_LEVEL DEFAULT 'mirror';


