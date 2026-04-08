<script lang="ts">
  import { page } from "$app/state";
  import { enhance } from "$app/forms";
  import ApiKeySlide from "$lib/components/ui/ApiKeySlide.svelte";
  import GenerateAPIKeyDialog from "$lib/components/ui/GenerateAPIKeyDialog.svelte";
  import { authClient } from "$lib/client";
  import type { ApiKey } from "@better-auth/api-key";

  const user = $derived(page.data.user);

  let dialog = $state<GenerateAPIKeyDialog | null>(null);

  let activeTab: "profile" | "APIKeys" = $state("profile");

  let apiKeys = $state<Omit<ApiKey, "key">[]>([]);

  async function updateKeys() {
    const { data, error } = await authClient.apiKey.list();
    if (!error) apiKeys = data.apiKeys;
  }

  updateKeys();

  async function handleGenerateKey(
    tokenName: string,
    expiryDays: number | null,
  ) {
    const { data, error } = await authClient.apiKey.create({
      name: tokenName,
      expiresIn: expiryDays ? expiryDays * 24 * 60 * 60 : undefined,
    });
    if (error) {
      console.error(error);
      return;
    }
    await updateKeys();
    dialog!.openWithKey(data.key);
  }

  async function handleRevokeKey(id: string) {
    await authClient.apiKey.delete({ keyId: id });
    await updateKeys();
  }

  async function handleRegenerateKey(
    id: string,
    name: string,
    expiryDays: number | null,
  ) {
    await authClient.apiKey.delete({ keyId: id });
    const { data, error } = await authClient.apiKey.create({
      name,
      expiresIn: expiryDays ? expiryDays * 24 * 60 * 60 : undefined,
    });
    if (error) {
      console.error(error);
      return;
    }
    await updateKeys();
    dialog!.openWithKey(data.key);
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

    <div class="tab-content">
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
              <span class="info-value">{user.emailVerified ? "Yes" : "No"}</span
              >
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
            <div class="api-keys-list">
              {#each apiKeys as key (key.id)}
                <ApiKeySlide
                  name={key.name ?? "Unnamed"}
                  preview={key.start ? `${key.start}••••` : "••••••••"}
                  createdAt={new Date(key.createdAt).toLocaleDateString(
                    "en-GB",
                  )}
                  lastUsed={key.lastRequest
                    ? new Date(key.lastRequest).toLocaleDateString("en-GB")
                    : "Never"}
                  expiresAt={key.expiresAt ? new Date(key.expiresAt) : null}
                  onRevoke={() => handleRevokeKey(key.id)}
                  onRegen={() =>
                    handleRegenerateKey(
                      key.id,
                      key.name ?? "Unnamed",
                      key.expiresAt
                        ? Math.round(
                            (new Date(key.expiresAt).getTime() - Date.now()) /
                              (24 * 60 * 60 * 1000),
                          )
                        : null,
                    )}
                />
              {/each}
            </div>
          </div>
          <button
            class="keygen-button"
            type="button"
            onclick={() => dialog!.open()}
          >
            + Generate new key
          </button>

          <GenerateAPIKeyDialog
            bind:this={dialog}
            onGenerate={handleGenerateKey}
            onCancel={() => dialog!.close()}
          />
        {/if}
      {/if}
    </div>
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

  .tab-content {
    padding: 0.5rem 2rem 2rem 2rem;
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

  .tab-btn:first-child {
    border-radius: 12px 0 0 0;
  }

  .tab-btn:last-child {
    border-radius: 0 12px 0 0;
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
    max-width: 45rem;
    border-radius: 12px;
    border: 1px solid var(--card-border);
    background: var(--card-bg);
    /* padding: 0 2rem 2rem 2rem; */
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
