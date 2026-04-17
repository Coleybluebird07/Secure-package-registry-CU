<script lang="ts">
  import { page } from "$app/stores";
  import { packagesAPI } from "$lib/api";
  import type { AdminPackageDetail, ReviewStatus } from "$lib/types/api";
  import {
    ArrowLeft,
    AlertCircle,
    Loader2,
    Play,
    Activity,
    CheckCircle2,
    XCircle,
    FileText,
  } from "lucide-svelte";

  let ecosystem = $derived($page.params.ecosystem ?? "");
  let name = $derived(decodeURIComponent($page.params.name ?? ""));

  let packageVersion = $state<AdminPackageDetail | null>(null);
  let loading = $state(false);
  let error = $state<string | null>(null);

  let scanning = $state(false);
  let scanError = $state<string | null>(null);
  let scanSuccess = $state<string | null>(null);

  let reviewStatus = $state<ReviewStatus>("pending");
  let reviewNotes = $state("");
  let reviewSubmitting = $state(false);
  let reviewError = $state<string | null>(null);
  let reviewSuccess = $state<string | null>(null);
  let maintainerTrustLevel = $state(0);

  async function loadVersions() {
    loading = true;
    error = null;

    try {
      const response = await packagesAPI.versionDetails(ecosystem, name);
      packageVersion = response;
      reviewStatus = response.review?.status ?? "pending";
      reviewNotes = response.review?.notes ?? "";
      maintainerTrustLevel = response.maintainer_trust_level ?? 0;
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

  async function submitReview(status: ReviewStatus) {
    reviewSubmitting = true;
    reviewError = null;
    reviewSuccess = null;

    try {
      const res = await fetch(
        `/admin/packages/${encodeURIComponent(ecosystem)}/${encodeURIComponent(name)}/review`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            status,
            notes: reviewNotes,
            maintainer_trust_level: maintainerTrustLevel,
          }),
        },
      );

      let response;

      try {
        response = await res.json();
      } catch {
        throw new Error("Invalid server response");
      }

      if (!res.ok) {
        throw new Error(response?.error ?? "Failed to submit review decision");
      }

      reviewStatus = response.review.status;
      reviewNotes = response.review.notes ?? "";

      if (packageVersion) {
        packageVersion = {
          ...packageVersion,
          maintainer_trust_level: maintainerTrustLevel,
          review: response.review,
        };
      }

      reviewSuccess =
        status === "approved"
          ? "Package approved successfully"
          : status === "rejected"
            ? "Package rejected successfully"
            : "Package review updated.";
    } catch (e) {
      reviewError =
        e instanceof Error ? e.message : "Failed to submit review decision";
    } finally {
      reviewSubmitting = false;
    }
  }

  function trustColorClass(score: number) {
    if (score >= 90) return "trust-high";
    if (score >= 70) return "trust-medium";
    return "trust-low";
  }

  function reviewBadgeClass(status: ReviewStatus) {
    if (status === "approved") return "badge-approved";
    if (status === "rejected") return "badge-rejected";
    return "badge-pending";
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

  {#if reviewError}
    <div class="alert alert-error">
      <div class="alert-content">
        <AlertCircle class="icon-alert" />
        <span>{reviewError}</span>
      </div>
    </div>
  {/if}

  {#if reviewSuccess}
    <div class="alert alert-success">
      <div class="alert-content">
        <AlertCircle class="icon-alert" />
        <span>{reviewSuccess}</span>
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
    <div class="content-grid">
      <div class="main-column">
        <section class="card-section">
          <div class="section-header">
            <h2>Package details</h2>
          </div>

          <div class="details-grid">
            <div class="detail-item">
              <span class="detail-label">Package name</span>
              <span class="detail-value mono">{name}</span>
            </div>

            <div class="detail-item">
              <span class="detail-label">Ecosystem</span>
              <span class="detail-value mono">{ecosystem}</span>
            </div>

            <div class="detail-item">
              <span class="detail-label">Latest version</span>
              <span class="detail-value mono">
                {packageVersion.latest_version ?? "N/A"}
              </span>
            </div>

            <div class="detail-item">
              <span class="detail-label">Maintainer trust level</span>
              <span class="detail-value"
                >{packageVersion.maintainer_trust_level}</span
              >
            </div>

            <div class="detail-item">
              <span class="detail-label">Available versions</span>
              <span class="detail-value">
                {packageVersion.versions?.length ?? 0}
              </span>
            </div>
          </div>
        </section>

        <section class="card-section">
          <div class="section-header">
            <h2>Versions</h2>
          </div>

          <div class="table-wrapper">
            <table class="versions-table">
              <thead>
                <tr>
                  <th>Version</th>
                  <th class="col-actions">Actions</th>
                </tr>
              </thead>
              <tbody>
                {#each packageVersion.versions as version}
                  <tr>
                    <td class="cell-version">
                      {version}
                      {#if version === packageVersion.latest_version}
                        <span class="badge-latest">latest</span>
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
                    <td colspan="2" class="cell-empty">No versions found</td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </section>
      </div>

      <aside class="side-column">
        <section class="card-section">
          <div class="section-header">
            <h2>Current review status</h2>
          </div>

          <div class="review-status-panel">
            <div class="review-status-top">
              <span class="review-badge {reviewBadgeClass(reviewStatus)}">
                {reviewStatus}
              </span>

              <span class="trust-pill {trustColorClass(maintainerTrustLevel)}">
                Trust: {maintainerTrustLevel}
              </span>
            </div>

            <div class="review-status-meta">
              <div class="status-meta-item">
                <span class="status-meta-label">Updated by</span>
                <span class="status-meta-value">
                  {packageVersion.review?.updated_by ?? "Not recorded"}
                </span>
              </div>

              <div class="status-meta-item">
                <span class="status-meta-label">Last updated</span>
                <span class="status-meta-value">
                  {packageVersion.review?.updated_at
                    ? new Date(
                        packageVersion.review.updated_at,
                      ).toLocaleString()
                    : "Not recorded"}
                </span>
              </div>
            </div>

            {#if packageVersion.review?.notes}
              <div class="review-status-notes">
                <span class="status-meta-label">Latest notes</span>
                <p>{packageVersion.review.notes}</p>
              </div>
            {/if}
          </div>
        </section>

        <section class="card-section">
          <div class="section-header">
            <h2>Review package</h2>
          </div>

          <div class="review-form">
            <div class="trust-score-field">
              <label class="notes-label" for="trust-score">
                Maintainer trust score
              </label>

              <input
                id="trust-score"
                type="number"
                min="0"
                max="100"
                bind:value={maintainerTrustLevel}
                class="trust-score-input"
              />
            </div>

            <div class="notes-group">
              <label class="notes-label" for="review-notes">
                <FileText class="icon-xs" />
                Review notes
              </label>

              <textarea
                id="review-notes"
                bind:value={reviewNotes}
                class="notes-input"
                rows="6"
                placeholder="Risk observations, approval reasoning, or rejection notes..."
              ></textarea>
            </div>

            <div class="review-actions">
              <button
                class="btn-approve"
                type="button"
                onclick={() => submitReview("approved")}
                disabled={reviewSubmitting}
              >
                <CheckCircle2 class="icon-sm" />
                Approve
              </button>

              <button
                class="btn-reject"
                type="button"
                onclick={() => submitReview("rejected")}
                disabled={reviewSubmitting}
              >
                <XCircle class="icon-sm" />
                Reject
              </button>
            </div>
          </div>
        </section>
      </aside>
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
    margin: 0;
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.25rem;
    flex-wrap: wrap;
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

  .content-grid {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 1rem;
    align-items: start;
  }

  .main-column,
  .side-column {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .card-section {
    border-radius: 8px;
    border: 1px solid var(--card-border);
    background: var(--card-bg);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    padding: 1rem;
  }

  .section-header {
    margin-bottom: 1rem;
  }

  .section-header h2 {
    margin: 0;
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .details-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--bg-secondary);
  }

  .detail-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--text-secondary);
  }

  .detail-value {
    font-size: 0.95rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .mono {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
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

  .btn-primary:disabled,
  .btn-approve:disabled,
  .btn-reject:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary :global(.icon-spin) {
    width: 1rem;
    height: 1rem;
    animation: spin 1s linear infinite;
  }

  .btn-primary :global(.icon-sm),
  .btn-approve :global(.icon-sm),
  .btn-reject :global(.icon-sm) {
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

  .behavior-link {
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

  .behavior-link:hover {
    background: var(--bg-primary);
    color: var(--accent);
  }

  .behavior-link :global(.icon-xs),
  .notes-label :global(.icon-xs) {
    width: 0.875rem;
    height: 0.875rem;
  }

  .cell-empty {
    padding: 2rem 1rem;
    text-align: center;
    color: var(--text-secondary);
  }

  .review-status-panel {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .review-status-top {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .review-status-meta {
    display: grid;
    gap: 0.75rem;
  }

  .status-meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--bg-secondary);
  }

  .status-meta-label {
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: var(--text-secondary);
  }

  .status-meta-value {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .review-status-notes {
    padding: 0.85rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--bg-secondary);
  }

  .review-status-notes p {
    margin: 0.45rem 0 0;
    color: var(--text-primary);
    line-height: 1.5;
    font-size: 0.9rem;
  }

  .trust-pill {
    display: inline-flex;
    align-items: center;
    padding: 0.4rem 0.75rem;
    border-radius: 999px;
    border: 1px solid;
    font-size: 0.8rem;
    font-weight: 700;
  }

  .trust-high {
    color: #15803d;
    background: rgba(22, 163, 74, 0.12);
    border-color: rgba(22, 163, 74, 0.3);
  }

  .trust-medium {
    color: #a16207;
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.3);
  }

  .trust-low {
    color: #b91c1c;
    background: rgba(220, 38, 38, 0.12);
    border-color: rgba(220, 38, 38, 0.3);
  }

  .review-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .trust-score-field {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .trust-score-input {
    width: 100%;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--bg-secondary);
    color: var(--text-primary);
    font: inherit;
    box-sizing: border-box;
  }

  .trust-score-input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .notes-group {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .notes-input {
    width: 100%;
    resize: vertical;
    min-height: 140px;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--bg-secondary);
    color: var(--text-primary);
    font: inherit;
    box-sizing: border-box;
  }

  .notes-input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .review-actions {
    display: flex;
    gap: 0.75rem;
    flex-wrap: wrap;
  }

  .btn-approve,
  .btn-reject {
    display: inline-flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem 1rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 600;
    cursor: pointer;
    border: 1px solid;
  }

  .btn-approve {
    color: #fff;
    background: #16a34a;
    border-color: #16a34a;
  }

  .btn-approve:hover {
    background: #15803d;
    border-color: #15803d;
  }

  .btn-reject {
    color: #fff;
    background: #dc2626;
    border-color: #dc2626;
  }

  .btn-reject:hover {
    background: #b91c1c;
    border-color: #b91c1c;
  }

  .badge-pending {
    color: #a16207;
    background: rgba(245, 158, 11, 0.12);
    border-color: rgba(245, 158, 11, 0.3);
  }

  .badge-approved {
    color: #15803d;
    background: rgba(22, 163, 74, 0.12);
    border-color: rgba(22, 163, 74, 0.3);
  }

  .badge-rejected {
    color: #b91c1c;
    background: rgba(220, 38, 38, 0.12);
    border-color: rgba(220, 38, 38, 0.3);
  }

  .trust-pill {
    display: inline-flex;
    align-items: center;
    padding: 0.4rem 0.75rem;
    border-radius: 999px;
    background: var(--bg-secondary);
    border: 1px solid var(--border);
    color: var(--text-primary);
    font-size: 0.8rem;
    font-weight: 600;
  }

  .review-status-meta {
    display: grid;
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }

  .status-meta-item {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.75rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--bg-secondary);
  }

  .status-meta-label {
    font-size: 0.72rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: var(--text-secondary);
  }

  .status-meta-value {
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .review-status-notes {
    padding: 0.85rem;
    border: 1px solid var(--border);
    border-radius: 8px;
    background: var(--bg-secondary);
  }

  .review-status-notes p {
    margin: 0.45rem 0 0;
    color: var(--text-primary);
    line-height: 1.5;
    font-size: 0.9rem;
  }

  .trust-score-field {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }

  .trust-score-input {
    width: 100%;
    padding: 0.75rem;
    border-radius: 8px;
    border: 1px solid var(--border);
    background: var(--bg-secondary);
    color: var(--text-primary);
    font: inherit;
    box-sizing: border-box;
  }

  .trust-score-input:focus {
    outline: none;
    border-color: var(--accent);
  }
</style>
