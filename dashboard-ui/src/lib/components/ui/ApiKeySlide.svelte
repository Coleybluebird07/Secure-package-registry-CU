<script lang="ts">
  let { name, preview, createdAt, lastUsed, onRevoke } = $props<{
    name: string;
    preview: string;
    createdAt: string;
    lastUsed: string;
    onRevoke: () => void;
  }>();
</script>

<div class="key-row">
  <div class="key-info">
    <span class="key-name">{name}</span>
    <span class="key-meta">
      Created {createdAt} · {lastUsed === "Never"
        ? "never used"
        : `last used ${lastUsed}`}
    </span>
  </div>
  <div class="key-right">
    <span class="key-preview">{preview}</span>
    <button class="btn-revoke" type="button" onclick={onRevoke}>Revoke</button>
  </div>
</div>

<style>
  .key-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.875rem 1rem;
    border: 1px solid var(--border);
    border-radius: 10px;
    background: var(--bg-secondary);
  }

  .key-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .key-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .key-meta {
    font-size: 0.8125rem;
    color: var(--text-secondary);
  }

  .key-right {
    display: flex;
    align-items: center;
    gap: 1rem;
  }

  .key-preview {
    font-family: monospace;
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .btn-revoke {
    padding: 0.4rem 1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    cursor: pointer;
    transition:
      background 0.15s,
      border-color 0.15s;
  }

  .btn-revoke:hover {
    background: var(--card-bg);
    border-color: var(--text-secondary);
  }
</style>
