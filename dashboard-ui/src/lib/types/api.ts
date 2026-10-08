export type Ecosystem = "npm" | "go" | "cargo" | "pypi";

export type CollectionTaskStatus =
	| "pending"
	| "running"
	| "succeeded"
	| "failed"
	| "cancelled";

export type RebuildTaskStatus =
	| "pending"
	| "running"
	| "succeeded"
	| "failed"
	| "cancelled"
	| "unavailable";

export type RebuildArtifactKind =
	| "official"
	| "rebuilt"
	| "diffoscope"
	| "logs"
	| "metadata";

export interface Package {
	id: number;
	identifier: string;
	ecosystem: string;
	latest_version?: string;
}

export interface PackageVersion {
	identifier: string;
	ecosystem: string;
	latest_version?: string;
	versions: string[];
}

export interface CollectionTask {
	id: number;
	identifier: string;
	ecosystem: string;
	version: string;
	source: string;
	status: CollectionTaskStatus;
	failure_reason?: string;
	has_artifact: boolean;
	started_at?: string;
	completed_at?: string;
	created_at: string;
}

export interface RebuildTask {
	id: number;
	identifier: string;
	ecosystem: string;
	version: string;
	source: string;
	status: RebuildTaskStatus;
	matched?: boolean;
	failure_reason?: string;
	has_diffoscope: boolean;
	has_logs: boolean;
	has_metadata: boolean;
	started_at?: string;
	completed_at?: string;
	created_at: string;
}

export interface RebuildTaskDetail {
	id: number;
	package_version_id: number;
	source: string;
	status: RebuildTaskStatus;
	matched?: boolean;
	failure_reason?: string;
	has_official_artifact: boolean;
	has_rebuilt_artifact: boolean;
	has_diffoscope: boolean;
	has_logs: boolean;
	has_metadata: boolean;
	official_artifact_url?: string;
	rebuilt_artifact_url?: string;
	diffoscope_url?: string;
	logs_url?: string;
	metadata_url?: string;
	started_at?: string;
	heartbeat_at?: string;
	completed_at?: string;
	created_at: string;
	updated_at: string;
}

export interface RebuildMetadata {
	task_id: number;
	ecosystem: string;
	package: string;
	version: string;
	source: string;
	matched: boolean;
	official_sha256: string;
	rebuilt_sha256?: string;
	official_tarball: string;
	generated_at: string;
}

export interface TriggerRebuildRequest {
	version?: string;
	source?: string;
}

export interface TriggerRebuildResponse {
	task_id: number;
	identifier: string;
	ecosystem: string;
	version: string;
	source: string;
	status: string;
	retried: boolean;
	already_active: boolean;
}

export interface ListPackagesResponse {
	items: Package[];
}

export interface AddPackageRequest {
	identifier: string;
	ecosystem: Ecosystem;
}

export interface AddPackageResponse {
	id: number;
	identifier: string;
	ecosystem: string;
	already_exists: boolean;
}

export interface TriggerScanRequest {
	version?: string;
}

export interface TriggerScanResponse {
	task_id: number;
	identifier: string;
	ecosystem: string;
	version: string;
	status: string;
}

export interface ListTasksResponse {
	items: CollectionTask[];
}

export interface ListRebuildTasksResponse {
	items: RebuildTask[];
}

// Behavioral analysis types — mirrors Go pkg/behavior types

export interface ProcessTree {
	root: ProcessNode;
}

export interface ProcessNode {
	name: string;
	identity: string;
	behaviors: ProcessBehaviors;
	children?: ProcessNode[];
}

export interface ProcessBehaviors {
	execs?: ExecBehavior[];
	files_opened?: string[];
	connections?: ConnectionBehavior[];
	dns_queries?: string[];
}

export interface ExecBehavior {
	pathname: string;
	argv: string[];
}

export interface ConnectionBehavior {
	addr: string;
	port: number;
}

// Public search types (used by /api/v1/svc/ endpoints)

export interface PackageSummary {
	ecosystem: Ecosystem;
	identifier: string;
	latest_version: string;
}

export interface SearchResult {
	items: PackageSummary[];
	page: number;
	page_size: number;
	total_count: number;
}

export interface PackageVersionDetail {
	ecosystem: string;
	identifier: string;
	latest: boolean;
	maintainer_notes: string;
	source: {
		commit: string;
		tag: string;
		url: string;
	};
	tags: Array<{
		data: string; // base64 encoded JSON
		label: string;
		value_type: "boolean" | "integer" | "float";
	}>;
	trust_level: number;
	version: string;
}

export interface VerifyResponse {
	ecosystem: string;
	identifier: string;
	oss_rebuild: boolean;
	upstream_attestation: boolean;
	version: string;
}

// Version list types (public endpoint)

export interface VersionSummary {
	behavior_passed: boolean | null;
	has_attestation: boolean;
	has_oss_rebuild: boolean;
	has_reproducible: boolean;
	latest: boolean;
	manually_approved: boolean | null;
	review_comment: string | null;
	source: {
		commit: string;
		tag: string;
		url: string;
	};
	version: string;
}

export interface VersionListResult {
	ecosystem: Ecosystem;
	identifier: string;
	versions: VersionSummary[];
}

// Project dependency tracking types

export type DependencyType = "direct" | "transitive";

export interface Project {
	id: number;
	name: string;
	source_type: string;
	created_at?: string;
	updated_at?: string;
}

export interface ProjectDependency {
	behavior_passed: boolean | null;
	dependency_type: DependencyType;
	ecosystem: string;
	has_attestation: boolean | null;
	has_oss_rebuild: boolean | null;
	has_reproducible: boolean | null;
	id: number;
	identifier: string;
	version: string;
	version_constraint?: string;
}

export interface ProjectSummaryRow {
	behavior_passed: number;
	dependency_type: DependencyType;
	has_attestation: number;
	has_oss_rebuild: number;
	has_reproducible: number;
	total: number;
}

export interface ListProjectsResponse {
	items: Project[];
}

export interface UploadProjectResponse {
	id: number;
	name: string;
	source_type: string;
	total_deps: number;
	direct_deps: number;
}

export interface ListProjectDependenciesResponse {
	items: ProjectDependency[];
}

export interface ProjectSummaryResponse {
	project_id: number;
	summary: ProjectSummaryRow[];
}

// Admin review queue types

export interface ReviewQueueItem {
	ecosystem: string;
	identifier: string;
	is_latest: boolean;
	manually_approved: boolean | null;
	review_comment: string | null;
	version: string;
}

export interface ReviewQueueResponse {
	items: ReviewQueueItem[];
}

export interface ReviewStatusResponse {
	manually_approved: boolean | null;
	review_comment: string | null;
}

// Project policy types

export interface ProjectPolicy {
	project_id: number;
	require_provenance: boolean;
	require_behavior: boolean;
	allow_manual_review: boolean;
}

export interface UpdatePolicyRequest {
	require_provenance?: boolean;
	require_behavior?: boolean;
	allow_manual_review?: boolean;
}

// Project API key types

export interface ProjectAPIKey {
	id: string;
	name: string;
	prefix: string;
	expires_at?: string;
	created_at: string;
}

export interface CreateAPIKeyResponse extends ProjectAPIKey {
	key: string; // raw key, shown only once
}

export interface ProjectAPIKeyListResponse {
	items: ProjectAPIKey[];
}
