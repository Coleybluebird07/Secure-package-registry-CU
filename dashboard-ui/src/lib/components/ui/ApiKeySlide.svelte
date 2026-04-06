<script lang="ts">
  import ConfirmDialog from "./ConfirmDialog.svelte";

  let { name, preview, createdAt, lastUsed, expiresAt, onRevoke, onRegen } =
    $props<{
      name: string;
      preview: string;
      createdAt: string;
      lastUsed: string;
      expiresAt: Date | null;
      onRevoke: (name: string) => void;
      onRegen: (name: string, expiryDays: number | null) => void;
    }>();

  const expiryLabel = $derived(
    expiresAt === null
      ? { text: "No expiry", class: "badge-never" }
      : new Date() > expiresAt
        ? { text: "Expired", class: "badge-expired" }
        : new Date(expiresAt.getTime() - 30 * 24 * 60 * 60 * 1000) < new Date()
          ? {
              text: `Expires ${expiresAt.toLocaleDateString("en-GB")}`,
              class: "badge-warn",
            }
          : {
              text: `Expires ${expiresAt.toLocaleDateString("en-GB")}`,
              class: "badge-ok",
            },
  );

  let revokeDialog: ConfirmDialog;
  let regenDialog: ConfirmDialog;
</script>

<div class="key-row">
  <div class="key-info">
    <span class="key-name">{name}</span>
    <span class="key-meta">
      Created {createdAt}
    </span>
    <span class="key-meta">
      {lastUsed === "Never" ? "never used" : `last used ${lastUsed}`}
    </span>
  </div>
  <span class="key-preview">{preview}</span>
  <div class="divider"></div>
  <span class="badge {expiryLabel.class}">{expiryLabel.text}</span>

  <div class="key-right">
    <button
      class="btn-revoke"
      type="button"
      onclick={() => revokeDialog.open()}
    >
      Revoke
    </button>

    <ConfirmDialog
      bind:this={revokeDialog}
      title={`Revoke ${name}?`}
      description="This key will stop working immediately. This cannot be undone."
      confirmLabel="Revoke"
      onConfirm={() => {
        revokeDialog.close();
        onRevoke();
      }}
      onCancel={() => revokeDialog.close()}
    />

    <button class="btn-revoke" type="button" onclick={() => regenDialog.open()}>
      Regenerate
    </button>

    <ConfirmDialog
      bind:this={regenDialog}
      title={`Regenerate ${name}?`}
      description="This will revoke the current key and generate a new one. This cannot be undone."
      confirmLabel="Regenerate"
      onConfirm={() => {
        regenDialog.close();
        onRegen();
      }}
      onCancel={() => regenDialog.close()}
    />
  </div>
</div>

<style>
  .badge {
    font-size: 0.75rem;
    padding: 0.2rem 0.6rem;
    border-radius: 9999px;
    font-weight: 500;
    white-space: nowrap;
  }

  .badge-never {
    background: var(--bg-secondary);
    color: var(--text-secondary);
  }

  .badge-ok {
    background: rgba(34, 197, 94, 0.1);
    color: #166534;
    border: 1px solid rgba(34, 197, 94, 0.3);
  }

  .badge-warn {
    background: rgba(251, 191, 36, 0.1);
    color: #92400e;
    border: 1px solid rgba(251, 191, 36, 0.3);
  }

  .badge-expired {
    background: rgba(220, 38, 38, 0.08);
    color: #dc2626;
    border: 1px solid rgba(220, 38, 38, 0.2);
  }

  .divider {
    width: 1px;
    height: 1.5rem;
    background: var(--border);
  }

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
    gap: 0.5rem;
  }

  .key-preview {
    font-family: monospace;
    font-size: 0.875rem;
    color: var(--text-secondary);
  }

  .btn-revoke {
    padding: 0.4rem 0.75rem;
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
