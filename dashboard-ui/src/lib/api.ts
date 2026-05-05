import type {
  AddPackageRequest,
  AddPackageResponse,
  Ecosystem,
  ListPackagesResponse,
  ListRebuildTasksResponse,
  ListTasksResponse,
  PackageVersion,
  PackageVersionDetail,
  ProcessTree,
  RebuildArtifactKind,
  RebuildTaskDetail,
  SearchResult,
  TriggerRebuildRequest,
  TriggerRebuildResponse,
  TriggerScanRequest,
  TriggerScanResponse,
} from "$lib/types/api.js";

const API_BASE = "/api/v1";

class APIError extends Error {
  constructor(
    public status: number,
    public override message: string,
  ) {
    super(message);
  }
}

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...options?.headers,
    },
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Unknown error" }));
    throw new APIError(
      response.status,
      error.error || `HTTP ${response.status}`,
    );
  }

  return response.json();
}

// Admin API — for the /admin dashboard
export const packagesAPI = {
  add: (data: AddPackageRequest): Promise<AddPackageResponse> => {
    return fetchJSON(`${API_BASE}/admin/packages`, {
      body: JSON.stringify(data),
      method: "POST",
    });
  },

  behavior: (
    ecosystem: string,
    identifier: string,
    version: string,
  ): Promise<ProcessTree> => {
    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/behavior?version=${encodeURIComponent(version)}`,
    );
  },

  behaviorRaw: (
    ecosystem: string,
    identifier: string,
    version: string,
  ): Promise<ProcessTree> => {
    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/behavior/raw?version=${encodeURIComponent(version)}`,
    );
  },

  list: (ecosystem: Ecosystem): Promise<ListPackagesResponse> => {
    return fetchJSON(`${API_BASE}/admin/packages?ecosystem=${ecosystem}`);
  },

  scan: (
    ecosystem: string,
    identifier: string,
    data: TriggerScanRequest,
  ): Promise<TriggerScanResponse> => {
    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/scan`,
      {
        body: JSON.stringify(data),
        method: "POST",
      },
    );
  },

  versions: (
    ecosystem: string,
    identifier: string,
  ): Promise<PackageVersion> => {
    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/versions`,
    );
  },
};

export const tasksAPI = {
  downloadArtifact: (taskId: number): Promise<Response> => {
    return fetch(`${API_BASE}/admin/tasks/${taskId}/artifact`);
  },

  list: (params?: {
    ecosystem?: Ecosystem;
    page?: number;
    page_size?: number;
  }): Promise<ListTasksResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.ecosystem) searchParams.set("ecosystem", params.ecosystem);
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = queryString
      ? `${API_BASE}/admin/tasks?${queryString}`
      : `${API_BASE}/admin/tasks`;

    return fetchJSON(url);
  },
};

export const rebuildAPI = {
  artifactURL: (taskId: number, kind: RebuildArtifactKind): string => {
    return `${API_BASE}/admin/rebuild-tasks/${taskId}/artifact/${kind}`;
  },

  downloadArtifact: (
    taskId: number,
    kind: RebuildArtifactKind,
  ): Promise<Response> => {
    return fetch(`${API_BASE}/admin/rebuild-tasks/${taskId}/artifact/${kind}`);
  },

  getArtifactBlob: async (
    taskId: number,
    kind: RebuildArtifactKind,
  ): Promise<Blob> => {
    const response = await fetch(
      `${API_BASE}/admin/rebuild-tasks/${taskId}/artifact/${kind}`,
    );

    if (!response.ok) {
      throw new APIError(response.status, `Failed to fetch ${kind} artifact`);
    }

    return response.blob();
  },

  getArtifactText: async (
    taskId: number,
    kind: RebuildArtifactKind,
  ): Promise<string> => {
    const response = await fetch(
      `${API_BASE}/admin/rebuild-tasks/${taskId}/artifact/${kind}`,
    );

    if (!response.ok) {
      throw new APIError(response.status, `Failed to fetch ${kind} artifact`);
    }

    return response.text();
  },

  getForPackage: (
    ecosystem: string,
    identifier: string,
    version: string,
    source = "oss-rebuild",
  ): Promise<RebuildTaskDetail> => {
    const params = new URLSearchParams({
      source,
      version,
    });

    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/rebuild?${params.toString()}`,
    );
  },

  list: (params?: {
    ecosystem?: Ecosystem;
    page?: number;
    page_size?: number;
  }): Promise<ListRebuildTasksResponse> => {
    const searchParams = new URLSearchParams();
    if (params?.ecosystem) searchParams.set("ecosystem", params.ecosystem);
    if (params?.page) searchParams.set("page", params.page.toString());
    if (params?.page_size)
      searchParams.set("page_size", params.page_size.toString());

    const queryString = searchParams.toString();
    const url = queryString
      ? `${API_BASE}/admin/rebuild-tasks?${queryString}`
      : `${API_BASE}/admin/rebuild-tasks`;

    return fetchJSON(url);
  },

  trigger: (
    ecosystem: string,
    identifier: string,
    data: TriggerRebuildRequest,
  ): Promise<TriggerRebuildResponse> => {
    return fetchJSON(
      `${API_BASE}/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(identifier)}/rebuild`,
      {
        body: JSON.stringify(data),
        method: "POST",
      },
    );
  },
};

// Public search API — for /api/v1/svc/ endpoints
export const searchAPI = {
  getVersion: (
    ecosystem: string,
    identifier: string,
    version: string,
  ): Promise<PackageVersionDetail> => {
    const safeIdentifier = encodeURIComponent(identifier);
    return fetchJSON(
      `${API_BASE}/svc/packages/${ecosystem}/${safeIdentifier}/${version}`,
    );
  },

  search: (query?: string, ecosystem?: string): Promise<SearchResult> => {
    const params = new URLSearchParams();
    if (query?.trim()) params.append("q", query);
    if (ecosystem) params.append("ecosystem", ecosystem);

    const queryString = params.toString();
    const url = queryString
      ? `${API_BASE}/svc/packages?${queryString}`
      : `${API_BASE}/svc/packages`;

    return fetchJSON(url);
  },
};

export type {
  CollectionTaskStatus,
  Ecosystem,
  RebuildArtifactKind,
  RebuildTaskStatus,
} from "$lib/types/api.js";
