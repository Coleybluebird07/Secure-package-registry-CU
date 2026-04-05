<script lang="ts">
  let {
    onConfirm,
    onCancel,
    title,
    description,
    confirmLabel = "Confirm",
    confirmDanger = false,
  } = $props<{
    onConfirm: () => void;
    onCancel: () => void;
    title: string;
    description: string;
    confirmLabel?: string;
    confirmDanger?: boolean;
  }>();

  let dialog: HTMLDialogElement;

  export function open() {
    dialog.showModal();
  }

  export function close() {
    dialog.close();
  }
</script>

<dialog bind:this={dialog}>
  <div class="dialog-content">
    <h2 class="dialog-title">{title}</h2>
    <p class="dialog-desc">{description}</p>
    <div class="dialog-actions">
      <button class="btn-cancel" type="button" onclick={onCancel}>Cancel</button
      >
      <button
        class="btn-confirm"
        class:danger={confirmDanger}
        type="button"
        onclick={onConfirm}
      >
        {confirmLabel}
      </button>
    </div>
  </div>
</dialog>

<style>
  dialog {
    border: 1px solid var(--border);
    border-radius: 1rem;
    background: var(--card-bg-no-alpha);
    padding: 0;
    max-width: 22rem;
    width: 100%;
    position: fixed;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
  }

  dialog::backdrop {
    background: rgba(0, 0, 0, 0.4);
  }

  .dialog-content {
    padding: 1.5rem;
  }

  .dialog-title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.5rem;
  }

  .dialog-desc {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin-bottom: 1.25rem;
  }

  .dialog-actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.75rem;
  }

  .btn-cancel {
    padding: 0.5rem 1.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .btn-cancel:hover {
    background: var(--bg-secondary);
  }

  .btn-confirm {
    padding: 0.5rem 1.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #fff;
    background: var(--accent);
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .btn-confirm.danger {
    background: #dc2626;
  }

  .btn-confirm.danger:hover {
    background: #b91c1c;
  }
</style>
