CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_blob_digest ON artifact_blob (digest_blob)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_blob_digest_af ON artifact_blob (digest_af)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_digest ON artifact (digest)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_digest_project ON artifact (digest,project_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_project_id ON artifact (project_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_reference_child_id ON artifact_reference (child_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_repository_id ON artifact (repository_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_artifact_repository_name ON artifact (repository_name)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_blob ON project_blob (blob_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_id ON project_blob (id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_project_project ON project_blob (project_id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_quota_reference ON quota (reference, reference_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_report_digest ON scan_report (digest);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_report_id ON scan_report (id);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_scan_report_uuid ON scan_report (uuid);
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_status ON blob (status)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_tag_artifact_id ON tag (artifact_id)
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_version ON blob (version)
