<script lang="ts">
  import { page } from "$app/stores";
  import { rebuildAPI } from "$lib/api";
  import type { RebuildMetadata } from "$lib/types/api";
  import {
    Archive,
    ArrowLeft,
    FileText,
    GitCompare,
    Loader2,
    ShieldAlert,
  } from "lucide-svelte";

  const taskId = $derived(Number($page.params.id));

  let loading = $state(false);
  let error = $state<string | null>(null);

  let metadata = $state<RebuildMetadata | null>(null);
  let officialBlob = $state<Blob | null>(null);
  let rebuiltBlob = $state<Blob | null>(null);

  let officialPreview = $state<string>("");
  let rebuiltPreview = $state<string>("");

  let officialHash = $state<string>("");
  let rebuiltHash = $state<string>("");

  async function sha256(blob: Blob): Promise<string> {
    const buffer = await blob.arrayBuffer();
    const digest = await crypto.subtle.digest("SHA-256", buffer);
    return Array.from(new Uint8Array(digest))
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
  }

  async function hexPreview(blob: Blob, maxBytes = 256): Promise<string> {
    const buffer = await blob.slice(0, maxBytes).arrayBuffer();
    const bytes = new Uint8Array(buffer);

    const lines: string[] = [];
    for (let i = 0; i < bytes.length; i += 16) {
      const chunk = bytes.slice(i, i + 16);
      const offset = i.toString(16).padStart(8, "0");
      const hex = Array.from(chunk)
        .map((b) => b.toString(16).padStart(2, "0"))
        .join(" ");
      const ascii = Array.from(chunk)
        .map((b) => (b >= 32 && b <= 126 ? String.fromCharCode(b) : "."))
        .join("");

      lines.push(`${offset}  ${hex.padEnd(47, " ")}  ${ascii}`);
    }

    return lines.join("\n");
  }

  async function loadComparison() {
    if (!taskId || Number.isNaN(taskId)) {
      error = "Invalid rebuild task ID";
      return;
    }

    loading = true;
    error = null;

    try {
      const metadataText = await rebuildAPI.getArtifactText(taskId, "metadata");
      metadata = JSON.parse(metadataText) as RebuildMetadata;

      officialBlob = await rebuildAPI.getArtifactBlob(taskId, "official");
      rebuiltBlob = await rebuildAPI.getArtifactBlob(taskId, "rebuilt");

      const [officialPreviewText, rebuiltPreviewText, officialSha, rebuiltSha] =
        await Promise.all([
          hexPreview(officialBlob),
          hexPreview(rebuiltBlob),
          sha256(officialBlob),
          sha256(rebuiltBlob),
        ]);

      officialPreview = officialPreviewText;
      rebuiltPreview = rebuiltPreviewText;
      officialHash = officialSha;
      rebuiltHash = rebuiltSha;
    } catch (e) {
      error =
        e instanceof Error ? e.message : "Failed to load rebuild comparison";
    } finally {
      loading = false;
    }
  }

  $effect(() => {
    loadComparison();
  });
</script>

<div class="page">
  <a href="/admin" class="back-link">
    <ArrowLeft class="icon-back" />
    Back to dashboard
  </a>

  <div class="page-header">
    <div>
      <h1 class="page-title">
        <GitCompare class="title-icon" />
        Rebuild Artifact Comparison
      </h1>

      {#if metadata}
        <div class="meta-row">
          <span class="eco-badge">{metadata.ecosystem}</span>
          <span class="meta-name">{metadata.package}</span>
          <span class="meta-version">v{metadata.version}</span>
          <span class="source-badge">{metadata.source}</span>
        </div>
      {/if}
    </div>
  </div>

  {#if loading}
    <div class="loading-container">
      <Loader2 class="spinner" />
    </div>
  {:else if error}
    <div class="alert alert-error">
      <ShieldAlert class="alert-icon" />
      <span>{error}</span>
    </div>
  {:else if metadata && officialBlob && rebuiltBlob}
    <div class="summary-card">
      <div class="summary-main">
        <ShieldAlert class="summary-icon" />
        <div>
          <h2 class="summary-title">Artifact mismatch detected</h2>
          <p class="summary-text">
            The package distributed on npm does not match the artifact produced
            by OSS Rebuild. This view compares the two artifacts side-by-side.
          </p>
        </div>
      </div>

      <div class="summary-grid">
        <div>
          <span class="summary-label">Official npm SHA256</span>
          <code>{officialHash}</code>
        </div>
        <div>
          <span class="summary-label">Rebuilt SHA256</span>
          <code>{rebuiltHash}</code>
        </div>
        <div>
          <span class="summary-label">Official size</span>
          <code>{officialBlob.size} bytes</code>
        </div>
        <div>
          <span class="summary-label">Rebuilt size</span>
          <code>{rebuiltBlob.size} bytes</code>
        </div>
      </div>
    </div>

    <div class="comparison-toolbar">
      <a
        href={rebuildAPI.artifactURL(taskId, "diffoscope")}
        class="toolbar-link"
      >
        <FileText class="toolbar-icon" />
        Open Diffoscope Report
      </a>
      <a href={rebuildAPI.artifactURL(taskId, "official")} class="toolbar-link">
        <Archive class="toolbar-icon" />
        Download Official
      </a>
      <a href={rebuildAPI.artifactURL(taskId, "rebuilt")} class="toolbar-link">
        <Archive class="toolbar-icon" />
        Download Rebuilt
      </a>
    </div>

    <div class="compare-grid">
      <section class="compare-panel panel-official">
        <div class="panel-header">
          <div>
            <h2>Official npm package</h2>
            <p>The artifact downloaded from the npm registry.</p>
          </div>
          <span class="side-badge official-badge">Expected</span>
        </div>

        <div class="panel-meta">
          <div>
            <span>File</span>
            <code>official.tgz</code>
          </div>
          <div>
            <span>SHA256</span>
            <code>{officialHash}</code>
          </div>
          <div>
            <span>Size</span>
            <code>{officialBlob.size} bytes</code>
          </div>
        </div>

        <pre class="hex-view">{officialPreview}</pre>
      </section>

      <section class="compare-panel panel-rebuilt">
        <div class="panel-header">
          <div>
            <h2>Rebuilt artifact</h2>
            <p>The artifact produced by OSS Rebuild.</p>
          </div>
          <span class="side-badge rebuilt-badge">Actual</span>
        </div>

        <div class="panel-meta">
          <div>
            <span>File</span>
            <code>rebuilt.tgz</code>
          </div>
          <div>
            <span>SHA256</span>
            <code>{rebuiltHash}</code>
          </div>
          <div>
            <span>Size</span>
            <code>{rebuiltBlob.size} bytes</code>
          </div>
        </div>

        <pre class="hex-view">{rebuiltPreview}</pre>
      </section>
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
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 0.875rem;
  }

  .back-link:hover {
    color: var(--text-primary);
  }

  .back-link :global(.icon-back) {
    width: 1rem;
    height: 1rem;
  }

  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
  }

  .page-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .page-title :global(.title-icon) {
    width: 1.5rem;
    height: 1.5rem;
    color: var(--accent);
  }

  .meta-row {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-top: 0.5rem;
  }

  .eco-badge,
  .source-badge {
    display: inline-flex;
    padding: 0.25rem 0.5rem;
    font-size: 0.7rem;
    font-weight: 600;
    text-transform: uppercase;
    border-radius: 999px;
    background: var(--bg-secondary);
    color: var(--text-secondary);
    border: 1px solid var(--border);
  }

  .meta-name {
    font-weight: 600;
    color: var(--text-primary);
  }

  .meta-version {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    color: var(--text-secondary);
  }

  .loading-container {
    display: flex;
    justify-content: center;
    padding: 3rem;
  }

  .loading-container :global(.spinner) {
    width: 2rem;
    height: 2rem;
    animation: spin 1s linear infinite;
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
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 1rem;
    border-radius: 8px;
    border: 1px solid rgba(220, 38, 38, 0.2);
    background: rgba(220, 38, 38, 0.08);
    color: #dc2626;
  }

  .alert :global(.alert-icon) {
    width: 1.25rem;
    height: 1.25rem;
  }

  .summary-card {
    margin-bottom: 1rem;
    padding: 1.25rem;
    border-radius: 8px;
    border: 1px solid rgba(220, 38, 38, 0.2);
    background: rgba(220, 38, 38, 0.08);
  }

  .summary-main {
    display: flex;
    gap: 0.75rem;
    margin-bottom: 1rem;
  }

  .summary-main :global(.summary-icon) {
    width: 1.5rem;
    height: 1.5rem;
    color: #dc2626;
  }

  .summary-title {
    font-size: 1rem;
    font-weight: 700;
    color: #dc2626;
  }

  .summary-text {
    margin-top: 0.25rem;
    color: var(--text-secondary);
    font-size: 0.875rem;
  }

  .summary-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.75rem;
  }

  .summary-grid div,
  .panel-meta div {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .summary-label,
  .panel-meta span {
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.03em;
    color: var(--text-secondary);
  }

  code {
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    font-size: 0.75rem;
    word-break: break-all;
    color: var(--text-primary);
  }

  .comparison-toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    margin-bottom: 1rem;
  }

  .toolbar-link {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: var(--bg-secondary);
    color: var(--text-primary);
    text-decoration: none;
    font-size: 0.8125rem;
    font-weight: 500;
  }

  .toolbar-link:hover {
    color: var(--accent);
    background: var(--bg-primary);
  }

  .toolbar-link :global(.toolbar-icon) {
    width: 0.875rem;
    height: 0.875rem;
  }

  .compare-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 1rem;
  }

  .compare-panel {
    border: 1px solid var(--card-border);
    border-radius: 8px;
    background: var(--card-bg);
    overflow: hidden;
  }

  .panel-header {
    display: flex;
    justify-content: space-between;
    gap: 1rem;
    padding: 1rem;
    border-bottom: 1px solid var(--border);
    background: var(--bg-secondary);
  }

  .panel-header h2 {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .panel-header p {
    margin-top: 0.25rem;
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .side-badge {
    align-self: flex-start;
    padding: 0.125rem 0.5rem;
    border-radius: 999px;
    font-size: 0.7rem;
    font-weight: 700;
    text-transform: uppercase;
  }

  .official-badge {
    background: rgba(22, 163, 74, 0.15);
    color: #15803d;
  }

  .rebuilt-badge {
    background: rgba(220, 38, 38, 0.15);
    color: #dc2626;
  }

  .panel-meta {
    display: grid;
    gap: 0.75rem;
    padding: 1rem;
    border-bottom: 1px solid var(--border);
  }

  .hex-view {
    margin: 0;
    padding: 1rem;
    max-height: 32rem;
    overflow: auto;
    background: #111827;
    color: #f9fafb;
    font-size: 0.75rem;
    line-height: 1.5;
    font-family:
      ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }

  @media (max-width: 900px) {
    .compare-grid {
      grid-template-columns: 1fr;
    }

    .summary-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
