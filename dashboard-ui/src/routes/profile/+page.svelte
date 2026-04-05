<script lang="ts">
  import { page } from "$app/state";
  import { enhance } from "$app/forms";
  import ApiKeySlide from "$lib/components/ui/ApiKeySlide.svelte";
  import GenerateAPIKeyDialog from "$lib/components/ui/GenerateAPIKeyDialog.svelte";

  const user = $derived(page.data.user);

  let dialog: GenerateKeyDialog;

  let activeTab: "profile" | "APIKeys" = $state("profile");

  function handleGenerateKey(tokenName: string) {
    console.log("Generate new API key:", tokenName);
  }
</script>

<main>
  <div class="profile-card">
    <div class="tab-bar">
      <button
        class="tab-btn"
        class:active={activeTab === "profile"}
        onclick={() => (activeTab = "profile")}
      >
        Profile
      </button>
      <button
        class="tab-btn"
        class:active={activeTab === "APIKeys"}
        onclick={() => (activeTab = "APIKeys")}
      >
        API Keys
      </button>
    </div>

    {#if user}
      {#if activeTab === "profile"}
        <h1 class="profile-title">Profile</h1>
        <div class="info-section">
          <div class="info-row">
            <span class="info-label">Name</span>
            <span class="info-value">{user.name || "—"}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Email</span>
            <span class="info-value">{user.email}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Email Verified</span>
            <span class="info-value">{user.emailVerified ? "Yes" : "No"}</span>
          </div>
          <div class="info-row">
            <span class="info-label">Member Since</span>
            <span class="info-value">
              {new Date(user.createdAt).toLocaleDateString()}
            </span>
          </div>
        </div>

        <div class="actions">
          <form method="POST" action="/logout" use:enhance>
            <button type="submit" class="btn-logout">Log Out</button>
          </form>
        </div>
      {:else if activeTab === "APIKeys"}
        <h1 class="api-title">API Keys</h1>
        <p>Generate/Manage API Keys, for Pulling Packages</p>
        <div class="api-keys-list">
          <ApiKeySlide
            name="Default Key"
            preview="sk-****-1234"
            createdAt="2024-01-15"
            lastUsed="2024-06-10"
            onRevoke={() => alert("Revoke Default Key")}
          />
          <ApiKeySlide
            name="Secondary Key"
            preview="sk-****-5678"
            createdAt="2024-02-20"
            lastUsed="Never"
            onRevoke={() => alert("Revoke Secondary Key")}
          />
        </div>
        <button
          class="keygen-button"
          type="button"
          onclick={() => dialog.open()}
        >
          + Generate new key
        </button>

        <GenerateAPIKeyDialog
          bind:this={dialog}
          onGenerate={handleGenerateKey}
          onCancel={() => dialog.close()}
        />
      {/if}
    {/if}
  </div>
</main>

<style>
  main {
    display: flex;
    align-items: flex-start;
    justify-content: center;
    min-height: calc(100vh - 4rem);
    padding: 3rem 1rem;
  }

  .keygen-button {
    padding: 0.5rem 0.85rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-primary);
    border: 1px solid var(--border-stronger);
    border-radius: 0.5rem;
    cursor: pointer;
  }

  .api-keys-list {
    margin-top: 1.5rem;
    margin-bottom: 1.5rem;
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .tab-bar {
    display: flex;
    border-bottom: 1px solid var(--border);
    width: 100%;
  }

  .tab-btn {
    flex: 1;
    padding: 1rem;
    /* background: none; */
    border: none;
    font-size: 0.875rem;
    font-weight: 500;
    /* color: var(--text-secondary); */
    cursor: pointer;
    transition: all 0.15s;
  }

  .tab-btn:hover {
    /* color: var(--text-primary); */
    background: var(--bg-secondary);
  }

  .tab-btn.active {
    color: var(--accent);
    border-bottom: 2px solid var(--accent);
    margin-bottom: -1px;
  }

  .profile-card {
    width: 100%;
    max-width: 28rem;
    border-radius: 12px;
    border: 1px solid var(--card-border);
    background: var(--card-bg);
    padding: 0 2rem 2rem 2rem;
  }

  .api-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
  }

  .profile-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 1.5rem;
  }

  .info-section {
    display: flex;
    flex-direction: column;
    gap: 0;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem 0;
    border-bottom: 1px solid var(--border);
  }

  .info-row:last-child {
    border-bottom: none;
  }

  .info-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text-secondary);
  }

  .info-value {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--text-primary);
  }

  .actions {
    margin-top: 1.5rem;
    display: flex;
    justify-content: flex-end;
  }

  .btn-logout {
    padding: 0.5rem 1.25rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: #dc2626;
    background: rgba(220, 38, 38, 0.08);
    border: 1px solid rgba(220, 38, 38, 0.2);
    border-radius: 6px;
    text-decoration: none;
    transition:
      background 0.15s,
      border-color 0.15s;
  }

  .btn-logout:hover {
    background: rgba(220, 38, 38, 0.15);
    border-color: rgba(220, 38, 38, 0.4);
  }
</style>
