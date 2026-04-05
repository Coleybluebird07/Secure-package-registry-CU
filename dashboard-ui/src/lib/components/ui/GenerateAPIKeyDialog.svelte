<script lang="ts">
  let { onGenerate, onCancel } = $props<{
    onGenerate: (name: string) => void;
    onCancel: () => void;
  }>();

  let dialog: HTMLDialogElement;
  let keyName = $state("");

  export function open() {
    keyName = "";
    dialog.showModal();
  }

  export function close() {
    dialog.close();
  }

  function handleGenerate() {
    if (!keyName.trim()) return;
    onGenerate(keyName.trim());
    close();
  }
</script>

<dialog bind:this={dialog} class="dialog">
  <div class="dialog-content">
    <h2 class="dialog-title">Generate API key</h2>
    <p class="dialog-desc">
      Give your key a name so you can identify it later.
    </p>

    <label class="field-label" for="key-name">Key name</label>
    <input
      id="key-name"
      class="key-input"
      type="text"
      placeholder="e.g. CI pipeline, local dev"
      bind:value={keyName}
    />

    <div class="dialog-actions">
      <button class="btn-cancel" type="button" onclick={onCancel}>Cancel</button
      >
      <button
        class="btn-generate"
        type="button"
        disabled={!keyName.trim()}
        onclick={handleGenerate}
      >
        Generate
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
    max-width: 30rem;
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
    padding: 2rem;
  }

  .dialog-title {
    font-size: 1.25rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.375rem;
  }

  .dialog-desc {
    font-size: 0.875rem;
    color: var(--text-secondary);
    margin-bottom: 1.25rem;
  }

  .field-label {
    display: block;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    margin-bottom: 0.375rem;
  }

  .key-input {
    width: 100%;
    height: 2.5rem;
    padding: 0 0.75rem;
    font-size: 0.875rem;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    background: var(--bg-secondary);
    color: var(--text-primary);
    outline: none;
    margin-bottom: 1.5rem;
  }

  .key-input:focus {
    border-color: var(--accent);
    /* box-shadow: 0 0 0 2px rgba(29, 78, 216, 0.15); */
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
    /* background: transparent; */
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .btn-generate {
    padding: 0.5rem 1.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #fff;
    background: var(--accent);
    border: none;
    border-radius: 0.5rem;
    cursor: pointer;
    /* transition: opacity 0.15s; */
  }

  .btn-generate:disabled {
    /* opacity: 0.4; */
    cursor: not-allowed;
  }
</style>
