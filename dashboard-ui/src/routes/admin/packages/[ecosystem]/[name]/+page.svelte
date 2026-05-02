<script lang="ts">
  import { page } from "$app/stores";
  import { packagesAPI, rebuildAPI } from "$lib/api";
  import type { PackageVersion, RebuildTaskDetail } from "$lib/types/api";
  import {
    ArrowLeft,
    AlertCircle,
    Loader2,
    Play,
    Activity,
    FileText,
    ScrollText,
    Archive,
  } from "lucide-svelte";

  let ecosystem = $derived($page.params.ecosystem ?? "");
  let name = $derived(decodeURIComponent($page.params.name ?? ""));

  let packageVersion = $state<PackageVersion | null>(null);
  let rebuildByVersion = $state<Record<string, RebuildTaskDetail | null>>({});
  let loading = $state(false);
  let error = $state<string | null>(null);
  let scanning = $state(false);
  let scanError = $state<string | null>(null);
  let scanSuccess = $state<string | null>(null);

  function rebuildStatusLabel(task: RebuildTaskDetail | null | undefined) {
    if (!task) return "not checked";
    if (task.status === "succeeded" && task.matched === true)
      return "reproducible";
    if (task.status === "succeeded" && task.matched === false)
      return "mismatch";
    return task.status;
  }

  function rebuildStatusClass(task: RebuildTaskDetail | null | undefined) {
    const label = rebuildStatusLabel(task);
    return `rebuild-${label.replace(" ", "-")}`;
  }

  async function loadRebuildForVersion(version: string) {
    try {
      const response = await rebuildAPI.getForPackage(ecosystem, name, version);
      rebuildByVersion = {
        ...rebuildByVersion,
        [version]: response,
      };
    } catch {
      rebuildByVersion = {
        ...rebuildByVersion,
        [version]: null,
      };
    }
  }

  async function loadVersions() {
    loading = true;
    error = null;
    rebuildByVersion = {};

    try {
      const response = await packagesAPI.versions(ecosystem, name);
      packageVersion = response;

      await Promise.all(
        response.versions.map((version) => loadRebuildForVersion(version)),
      );
    } catch (e) {
      error =
        e instanceof Error ? e.message : "Failed to load package versions";
    } finally {
      loading = false;
    }
  }

  async function handleScan() {
    scanning = true;
    scanError = null;
    scanSuccess = null;

    try {
      const response = await packagesAPI.scan(ecosystem, name, {});
      scanSuccess = `Scan initiated for version ${response.version} (Task ID: ${response.task_id})`;
      await loadVersions();
    } catch (e) {
      if (e instanceof Error) {
        scanError = e.message;
      } else {
        scanError = "Failed to trigger scan";
      }
    } finally {
      scanning = false;
    }
  }

  $effect(() => {
    if (ecosystem && name) {
      loadVersions();
    }
  });
</script>

<div class="page">
  <a href="/admin" class="back-link">
    <ArrowLeft class="icon-back" />
    Back to packages
  </a>

  <div class="page-header">
    <h1 class="page-title">{name}</h1>
    <div class="meta-row">
      <span class="eco-badge">
        {ecosystem}
      </span>
      {#if packageVersion?.latest_version}
        <span class="latest-text">Latest: {packageVersion.latest_version}</span>
      {/if}
    </div>
  </div>

  <div class="actions-bar">
    <button onclick={handleScan} disabled={scanning} class="btn-primary">
      {#if scanning}
        <Loader2 class="icon-spin" />
      {:else}
        <Play class="icon-sm" />
      {/if}
      Scan Latest
    </button>
  </div>

  {#if scanError}
    <div class="alert alert-error">
      <div class="alert-content">
        <AlertCircle class="icon-alert" />
        <span>{scanError}</span>
      </div>
    </div>
  {/if}

  {#if scanSuccess}
    <div class="alert alert-success">
      <div class="alert-content">
        <span>{scanSuccess}</span>
      </div>
    </div>
  {/if}

  {#if loading}
    <div class="loading-container">
      <Loader2 class="spinner" />
    </div>
  {:else if error}
    <div class="alert alert-error">
      <div class="alert-content">
        <AlertCircle class="icon-alert" />
        <span>{error}</span>
      </div>
    </div>
  {:else if packageVersion}
    <div class="table-wrapper">
      <table class="versions-table">
        <thead>
          <tr>
            <th>Version</th>
            <th>Rebuild Status</th>
            <th>Rebuild Details</th>
            <th class="col-actions">Actions</th>
          </tr>
        </thead>
        <tbody>
          {#each packageVersion.versions as version}
            {@const rebuild = rebuildByVersion[version]}
            <tr>
              <td class="cell-version">
                {version}
                {#if version === packageVersion.latest_version}
                  <span class="badge-latest">latest</span>
                {/if}
              </td>

              <td>
                <span class="rebuild-badge {rebuildStatusClass(rebuild)}">
                  {rebuildStatusLabel(rebuild)}
                </span>
              </td>

              <td class="rebuild-details">
                {#if rebuild}
                  {#if rebuild.failure_reason}
                    <div class="failure-reason">{rebuild.failure_reason}</div>
                  {:else if rebuild.status === "succeeded" && rebuild.matched === true}
                    <span class="detail-muted">
                      Distributed npm artifact matched rebuilt artifact.
                    </span>
                  {:else if rebuild.status === "succeeded" && rebuild.matched === false}
                    <span class="detail-muted">
                      Distributed npm artifact did not match rebuilt artifact.
                    </span>
                  {:else}
                    <span class="detail-muted">
                      Task created at {new Date(
                        rebuild.created_at,
                      ).toLocaleString()}
                    </span>
                  {/if}

                  <div class="artifact-links">
                    {#if rebuild.metadata_url}
                      <a href={rebuild.metadata_url} class="artifact-link">
                        <FileText class="icon-xs" />
                        Metadata
                      </a>
                    {/if}

                    {#if rebuild.logs_url}
                      <a href={rebuild.logs_url} class="artifact-link">
                        <ScrollText class="icon-xs" />
                        Logs
                      </a>
                    {/if}

                    {#if rebuild.diffoscope_url}
                      <a href={rebuild.diffoscope_url} class="artifact-link">
                        <Activity class="icon-xs" />
                        Diffoscope
                      </a>
                    {/if}

                    {#if rebuild.official_artifact_url}
                      <a
                        href={rebuild.official_artifact_url}
                        class="artifact-link"
                      >
                        <Archive class="icon-xs" />
                        Official
                      </a>
                    {/if}

                    {#if rebuild.rebuilt_artifact_url}
                      <a
                        href={rebuild.rebuilt_artifact_url}
                        class="artifact-link"
                      >
                        <Archive class="icon-xs" />
                        Rebuilt
                      </a>
                    {/if}
                  </div>
                {:else}
                  <span class="detail-muted">
                    No rebuild verification has been recorded for this version.
                  </span>
                {/if}
              </td>

              <td class="col-actions">
                <a
                  href="/admin/packages/{ecosystem}/{encodeURIComponent(
                    name,
                  )}/behavior?version={version}"
                  class="behavior-link"
                >
                  <Activity class="icon-xs" />
                  View Behavior
                </a>
              </td>
            </tr>
          {:else}
            <tr>
              <td colspan="4" class="cell-empty">No versions found</td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</div>

<style>
  .page {
    padding: 2rem;
  }

  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    margin-bottom: 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
    text-decoration: none;
    transition: color 0.15s;
  }

  .back-link:hover {
    color: var(--text-primary);
  }

  .back-link :global(.icon-back) {
    width: 1rem;
    height: 1rem;
  }

  .page-header {
    margin-bottom: 1.5rem;
  }

  .page-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.25rem;
  }

  .eco-badge {
    display: inline-block;
    padding: 0.25rem 0.5rem;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    border-radius: 999px;
    background: var(--bg-secondary);
    color: var(--text-secondary);
    border: 1px solid var(--border);
  }

  .latest-text {
    color: var(--text-secondary);
  }

  .actions-bar {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .btn-primary {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #fff;
    background: var(--accent);
    border: none;
    border-radius: 6px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    cursor: pointer;
    transition: background 0.15s;
  }

  .btn-primary:hover {
    background: var(--accent-hover);
  }

  .btn-primary:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary :global(.icon-spin) {
    width: 1rem;
    height: 1rem;
    animation: spin 1s linear infinite;
  }

  .btn-primary :global(.icon-sm) {
    width: 1rem;
    height: 1rem;
  }

  @keyframes spin {
    from {
      transform: rotate(0deg);
    }

    to {
      transform: rotate(360deg);
    }
  }

  .alert {
    margin-bottom: 1rem;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid;
  }

  .alert-error {
    border-color: rgba(220, 38, 38, 0.2);
    background: rgba(220, 38, 38, 0.08);
  }

  .alert-success {
    border-color: rgba(22, 163, 74, 0.2);
    background: rgba(22, 163, 74, 0.08);
  }

  .alert-content {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 0.875rem;
  }

  .alert-error .alert-content {
    color: #dc2626;
  }

  .alert-success .alert-content {
    color: #16a34a;
  }

  .alert :global(.icon-alert) {
    width: 1rem;
    height: 1rem;
    flex-shrink: 0;
  }

  .loading-container {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 3rem 0;
  }

  .loading-container :global(.spinner) {
    width: 2rem;
    height: 2rem;
    color: var(--text-secondary);
    animation: spin 1s linear infinite;
  }

  .table-wrapper {
    border-radius: 8px;
    border: 1px solid var(--card-border);
    background: var(--card-bg);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    overflow: auto;
  }

  .versions-table {
    width: 100%;
    border-collapse: collapse;
  }

  .versions-table thead tr {
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
  }

  .versions-table th {
    padding: 0.75rem 1rem;
    text-align: left;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
  }

  .versions-table tbody tr {
    border-bottom: 1px solid var(--border);
    transition: background 0.1s;
  }

  .versions-table tbody tr:last-child {
    border-bottom: none;
  }

  .versions-table tbody tr:hover {
    background: var(--bg-secondary);
  }

  .versions-table td {
    padding: 0.75rem 1rem;
    vertical-align: top;
  }

  .cell-version {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    color: var(--text-primary);
  }

  .badge-latest {
    display: inline-block;
    margin-left: 0.5rem;
    padding: 0.125rem 0.5rem;
    font-size: 0.7rem;
    font-weight: 600;
    font-family:
      system-ui,
      -apple-system,
      sans-serif;
    border-radius: 999px;
    background: rgba(22, 163, 74, 0.15);
    color: #15803d;
  }

  .col-actions {
    text-align: right;
  }

  .behavior-link,
  .artifact-link {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.375rem 0.75rem;
    font-size: 0.75rem;
    font-weight: 500;
    color: var(--text-primary);
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    border-radius: 6px;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    text-decoration: none;
    transition:
      background 0.15s,
      color 0.15s;
  }

  .behavior-link:hover,
  .artifact-link:hover {
    background: var(--bg-primary);
    color: var(--accent);
  }

  .behavior-link :global(.icon-xs),
  .artifact-link :global(.icon-xs) {
    width: 0.875rem;
    height: 0.875rem;
  }

  .rebuild-badge {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    font-size: 0.75rem;
    font-weight: 600;
    border-radius: 999px;
    white-space: nowrap;
  }

  .rebuild-not-checked {
    background: rgba(107, 114, 128, 0.15);
    color: #4b5563;
  }

  .rebuild-pending {
    background: rgba(234, 179, 8, 0.15);
    color: #a16207;
  }

  .rebuild-running {
    background: rgba(37, 99, 235, 0.15);
    color: #1d4ed8;
  }

  .rebuild-reproducible {
    background: rgba(22, 163, 74, 0.15);
    color: #15803d;
  }

  .rebuild-mismatch,
  .rebuild-failed {
    background: rgba(220, 38, 38, 0.15);
    color: #dc2626;
  }

  .rebuild-unavailable,
  .rebuild-cancelled {
    background: rgba(107, 114, 128, 0.15);
    color: #4b5563;
  }

  .rebuild-details {
    min-width: 20rem;
  }

  .detail-muted {
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .failure-reason {
    max-width: 32rem;
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .artifact-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0.375rem;
    margin-top: 0.5rem;
  }

  .cell-empty {
    padding: 2rem 1rem;
    text-align: center;
    color: var(--text-secondary);
  }
</style>
