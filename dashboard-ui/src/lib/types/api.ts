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
	identifier: string;
	ecosystem: "npm" | "go" | "cargo" | "pypi";
	latest_version: string;
	description: string;
	author: string;
	updatedAgo: string;
	trustScore: number;
	tier: string;
	tags: string[];
}

export interface SearchResult {
	items: PackageSummary[];
}

export interface PackageVersionDetail {
	identifier: string;
	ecosystem: string;
	version: string;
	latest: boolean;
	source: {
		url: string;
		tag: string;
		commit: string;
	};
	trust_level: number;
	maintainer_notes: string;
	tags: Array<{
		label: string;
		value_type: "boolean" | "integer" | "float";
		data: string; // base64 encoded JSON
	}>;
}
