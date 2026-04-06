<script lang="ts">
  let { onGenerate, onCancel } = $props<{
    onGenerate: (name: string, expiryDays: number | null) => void;
    onCancel: () => void;
  }>();

  let dialog: HTMLDialogElement;
  let keyName = $state("");
  let revealedKey = $state<string | null>(null);
  let copied = $state(false);
  let expiryDays = $state<number | null>(30);

  export function open() {
    keyName = "";
    revealedKey = null;
    expiryDays = 30;
    dialog.showModal();
  }

  export function openWithKey(key: string) {
    keyName = "";
    revealedKey = key;
    expiryDays = 30;
    dialog.showModal();
  }

  export function close() {
    revealedKey = null;
    copied = false;
    dialog.close();
  }

  export function reveal(key: string) {
    revealedKey = key;
  }

  function handleGenerate(expiryDays: number | null) {
    if (!keyName.trim()) return;
    onGenerate(keyName.trim(), expiryDays);
  }

  async function handleCopy() {
    await navigator.clipboard.writeText(revealedKey!);
    copied = true;
  }
</script>

<dialog bind:this={dialog} class="dialog">
  <div class="dialog-content">
    {#if revealedKey}
      <h2 class="dialog-title">Your new API key</h2>
      <div class="warn-box">Copy this key now — it won't be shown again.</div>
      <div class="key-box">
        <span class="key-box-label">API key</span>
        <code class="key-text">{revealedKey}</code>
        <div class="flex">
          <button class="btn-copy" type="button" onclick={handleCopy}>
            Copy
          </button>
          {#if copied}
            <span class="clipboard-copy-alert">Copied to clipboard!</span>
          {/if}
        </div>
      </div>
      <div class="dialog-actions">
        <button class="btn-cancel" type="button" onclick={close}>Done</button>
      </div>
    {:else}
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

      <label class="field-label" for="expiry">Expiration</label>
      <select id="expiry" class="key-input" bind:value={expiryDays}>
        <option value={30}>30 days</option>
        <option value={60}>60 days</option>
        <option value={90}>90 days</option>
        <option value={365}>1 year</option>
        <option value={null}>Never</option>
      </select>

      <div class="dialog-actions">
        <button class="btn-cancel" type="button" onclick={onCancel}
          >Cancel</button
        >
        <button
          class="btn-generate"
          type="button"
          disabled={!keyName.trim()}
          onclick={() => handleGenerate(expiryDays)}
        >
          Generate
        </button>
      </div>
    {/if}
  </div>
</dialog>

<style>
  .flex {
    display: flex;
  }

  .clipboard-copy-alert {
    padding: 0.375rem;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--text-primary);
    background: transparent;
  }

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

  .warn-box {
    font-size: 0.875rem;
    color: #92400e;
    background: rgba(251, 191, 36, 0.1);
    border: 1px solid rgba(251, 191, 36, 0.4);
    border-radius: 0.5rem;
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
  }

  .key-box {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: rgba(34, 197, 94, 0.08);
    border: 1px solid rgba(34, 197, 94, 0.3);
    border-radius: 0.5rem;
    padding: 0.875rem 1rem;
    margin-bottom: 1.5rem;
  }

  .key-box-label {
    font-size: 0.75rem;
    font-weight: 600;
    color: #166534;
  }

  .key-text {
    font-family: monospace;
    font-size: 0.8125rem;
    color: var(--text-primary);
    word-break: break-all;
  }

  .btn-copy {
    align-self: flex-start;
    padding: 0.375rem 0.875rem;
    font-size: 0.8125rem;
    font-weight: 500;
    color: var(--text-primary);
    background: transparent;
    border: 1px solid var(--border);
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .btn-copy:hover {
    background: var(--bg-secondary);
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
  }

  .btn-generate:disabled {
    cursor: not-allowed;
  }
</style>
