--DOWN
ALTER TABLE organization_packages DROP COLUMN security_level;
ALTER TABLE package_versions DROP COLUMN security_level;
DROP TYPE SECURITY_LEVEL;
