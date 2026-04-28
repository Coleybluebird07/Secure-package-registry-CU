CREATE TYPE REBUILD_TASK_STATUS AS ENUM (
    'pending',
    'running',
    'succeeded',
    'failed',
    'cancelled',
    'unavailable'
);

CREATE TABLE rebuild_tasks (
    id SERIAL PRIMARY KEY,
    package_version_id INTEGER NOT NULL REFERENCES package_versions(id) ON DELETE CASCADE,
    source TEXT NOT NULL DEFAULT 'oss-rebuild',
    status REBUILD_TASK_STATUS NOT NULL DEFAULT 'pending',

    matched BOOLEAN,

    official_artifact_bucket TEXT,
    official_artifact_key TEXT,

    rebuilt_artifact_bucket TEXT,
    rebuilt_artifact_key TEXT,

    diffoscope_bucket TEXT,
    diffoscope_key TEXT,

    logs_bucket TEXT,
    logs_key TEXT,

    metadata_bucket TEXT,
    metadata_key TEXT,

    started_at TIMESTAMP WITH TIME ZONE,
    heartbeat_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    failure_reason TEXT,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(package_version_id, source)
);

CREATE INDEX idx_rebuild_tasks_status ON rebuild_tasks (status);
CREATE INDEX idx_rebuild_tasks_heartbeat ON rebuild_tasks (status, heartbeat_at) WHERE status = 'running';
CREATE INDEX idx_rebuild_tasks_package_version ON rebuild_tasks (package_version_id);