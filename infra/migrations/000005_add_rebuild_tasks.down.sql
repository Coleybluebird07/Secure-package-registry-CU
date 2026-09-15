DROP INDEX IF EXISTS idx_rebuild_tasks_package_version;
DROP INDEX IF EXISTS idx_rebuild_tasks_heartbeat;
DROP INDEX IF EXISTS idx_rebuild_tasks_status;

DROP TABLE IF EXISTS rebuild_tasks;

DROP TYPE IF EXISTS REBUILD_TASK_STATUS;