<script lang="ts">
  import { PUBLIC_HOME_BASE_URL } from "$env/static/public";
  import { goto } from "$app/navigation";
  import { searchAPI } from "$lib/api";
  import type { Ecosystem, PackageSummary } from "$lib/types/api";

  interface Props {
    isDark?: boolean;
    onToggleTheme: () => void;
    isLoggedIn?: boolean;
  }

  let { isDark = true, onToggleTheme, isLoggedIn = false }: Props = $props();

  let searchQuery = $state("");
  let selectedEcosystem = $state<"" | Ecosystem>("");
  let suggestions = $state<PackageSummary[]>([]);
  let showSuggestions = $state(false);
  let loadingSuggestions = $state(false);

  let debounceTimer: ReturnType<typeof setTimeout> | null = null;

  const ecosystems = [
    { label: "All Ecosystems", value: "" as const },
    { label: "npm", value: "npm" as const },
    { label: "Go", value: "go" as const },
    { label: "Cargo", value: "cargo" as const },
    { label: "PyPI", value: "pypi" as const },
  ];

  function trustColor(score: number): string {
    if (score >= 90) return "trust-high";
    if (score >= 70) return "trust-medium";
    return "trust-low";
  }

  async function loadSuggestions() {
    const query = searchQuery.trim();

    if (!query) {
      suggestions = [];
      showSuggestions = false;
      return;
    }

    loadingSuggestions = true;

    try {
      const data = await searchAPI.search(
        query,
        selectedEcosystem || undefined,
      );
      suggestions = (data.items ?? []).slice(0, 6);
      showSuggestions = true;
    } catch {
      suggestions = [];
      showSuggestions = false;
    } finally {
      loadingSuggestions = false;
    }
  }

  function handleInput() {
    if (debounceTimer) clearTimeout(debounceTimer);

    debounceTimer = setTimeout(() => {
      loadSuggestions();
    }, 250);
  }

  async function submitSearch() {
    const params = new URLSearchParams();

    if (searchQuery.trim()) {
      params.set("q", searchQuery.trim());
    }

    if (selectedEcosystem) {
      params.set("ecosystem", selectedEcosystem);
    }

    showSuggestions = false;
    await goto(`/search?${params.toString()}`);
  }

  async function openSuggestion(pkg: PackageSummary) {
    showSuggestions = false;
    await goto(
      `/detail-page/${encodeURIComponent(pkg.identifier)}?ecosystem=${pkg.ecosystem}&version=${pkg.latest_version}`,
    );
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter") {
      event.preventDefault();
      submitSearch();
    }

    if (event.key === "Escape") {
      showSuggestions = false;
    }
  }

  function handleBlur() {
    setTimeout(() => {
      showSuggestions = false;
    }, 150);
  }
</script>

<header>
  <div class="header-left">
    <div class="logo">
      <a class="clear-a-stylings" href={`${PUBLIC_HOME_BASE_URL}/`}>SPR</a>
    </div>
    <a href={`${PUBLIC_HOME_BASE_URL}/docs`} class="nav-link">Docs</a>
    <a href={`${PUBLIC_HOME_BASE_URL}/pricing`} class="nav-link">Pricing</a>
  </div>

  <div class="header-middle">
    <div class="search-shell">
      <form
        class="search-bar"
        onsubmit={(event) => {
          event.preventDefault();
          submitSearch();
        }}
      >
        <input
          type="text"
          bind:value={searchQuery}
          placeholder="Search Packages..."
          class="search-input"
          oninput={handleInput}
          onkeydown={handleKeydown}
          onfocus={() => {
            if (suggestions.length > 0) showSuggestions = true;
          }}
          onblur={handleBlur}
        />

        <button type="submit" class="search-button">Search</button>
      </form>

      {#if showSuggestions}
        <div class="suggestions-dropdown">
          {#if loadingSuggestions}
            <div class="suggestion-empty">Loading...</div>
          {:else if suggestions.length > 0}
            {#each suggestions as pkg}
              <button
                type="button"
                class="suggestion-item"
                onclick={() => openSuggestion(pkg)}
              >
                <div class="suggestion-top">
                  <span class="suggestion-name">{pkg.identifier}</span>
                  <span class="eco-badge {pkg.ecosystem}">{pkg.ecosystem}</span>
                  <span class="trust-pill {trustColor(pkg.trustScore ?? 0)}">
                    {pkg.trustScore ?? 0}
                  </span>
                </div>

                {#if pkg.description}
                  <div class="suggestion-desc">{pkg.description}</div>
                {/if}
              </button>
            {/each}
          {:else}
            <div class="suggestion-empty">No matching packages</div>
          {/if}
        </div>
      {/if}
    </div>
  </div>

  {#if isLoggedIn}
    <div class="header-right">
      <button class="theme-toggle" onclick={onToggleTheme}>
        {isDark ? "Light" : "Dark"}
      </button>
      <a href="/login" class="navbar-button">Log Out</a>
    </div>
  {:else}
    <div class="header-right">
      <button class="theme-toggle" onclick={onToggleTheme}>
        {isDark ? "Light" : "Dark"}
      </button>
      <a href="/login" class="navbar-button">Log In</a>
      <a href="/signup" class="navbar-button">Sign Up</a>
    </div>
  {/if}
</header>

<style>
  header {
    padding: 0.75rem 2.5rem;
    display: flex;
    align-items: center;
    position: sticky;
    top: 0;
    background: var(--bg-primary);
    border-bottom: 1px solid var(--border);
    z-index: 100;
    backdrop-filter: blur(10px);
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .header-middle {
    flex: 1;
    display: flex;
    justify-content: center;
    min-width: 0;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    margin-left: auto;
  }

  .logo {
    font-size: 2rem;
    font-weight: 700;
    color: var(--accent);
  }

  .clear-a-stylings {
    text-decoration: none;
    color: inherit;
  }

  .nav-link {
    color: var(--text-secondary);
    text-decoration: none;
    font-size: 0.875rem;
    font-weight: 500;
    transition: color 0.3s;
  }

  .nav-link:hover {
    color: var(--accent);
  }

  .search-shell {
    position: relative;
    width: 100%;
    max-width: 760px;
  }

  .search-bar {
    width: 100%;
    display: flex;
    justify-content: center;
    margin-left: 2rem;
    margin-right: 2rem;
  }

  .search-input {
    padding: 0.5rem;
    font-size: 0.95rem;
    font-weight: 500;
    width: 100%;
    color: var(--text-secondary);
    background: var(--bg-primary);
    border: 1px solid var(--border);
    border-right: none;
    border-top-left-radius: 6px;
    border-bottom-left-radius: 6px;
    transition: background 0.2s;
  }

  .search-input::placeholder {
    color: var(--text-secondary);
  }

  .search-input:focus {
    outline: none;
    border-color: var(--accent);
  }

  .search-button {
    border: 1px solid var(--accent);
    border-top-right-radius: 6px;
    border-bottom-right-radius: 6px;
    background: var(--accent);
    color: white;
    padding: 0.5rem 1.75rem;
    font-size: 0.875rem;
    font-weight: 700;
    cursor: pointer;
    white-space: nowrap;
    transition: background 0.2s;
  }

  .search-button:hover {
    background: var(--accent-hover);
    border-color: var(--accent-hover);
  }

  .suggestions-dropdown {
    position: absolute;
    top: calc(100% + 0.4rem);
    left: 2rem;
    right: 2rem;
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.12);
    overflow: hidden;
    z-index: 200;
  }

  .suggestion-item {
    width: 100%;
    text-align: left;
    padding: 0.85rem 1rem;
    border: none;
    border-bottom: 1px solid var(--border);
    background: transparent;
    cursor: pointer;
    color: inherit;
  }

  .suggestion-item:last-child {
    border-bottom: none;
  }

  .suggestion-item:hover {
    background: var(--bg-secondary);
  }

  .suggestion-top {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.25rem;
    flex-wrap: wrap;
  }

  .suggestion-name {
    font-weight: 800;
    color: var(--text-primary);
  }

  .suggestion-desc {
    font-size: 0.85rem;
    color: var(--text-secondary);
    line-height: 1.4;
  }

  .suggestion-empty {
    padding: 0.9rem 1rem;
    color: var(--text-secondary);
    font-size: 0.9rem;
  }

  .navbar-button {
    background: var(--accent);
    color: var(--bg-primary);
    text-decoration: none;
    padding: 0.5rem 1.75rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    transition: background 0.2s;
  }

  .theme-toggle {
    border: none;
    color: var(--wb-bg);
    background: var(--wb-bg-invert);
    text-decoration: none;
    padding: 0.5rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    transition: color 0.3s;
    cursor: pointer;
  }

  .eco-badge {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    font-size: 0.7rem;
    font-weight: 800;
    text-transform: uppercase;
    letter-spacing: 0.025em;
    border-radius: 999px;
    border: 1px solid;
    flex-shrink: 0;
  }

  .eco-badge.npm {
    border-color: rgba(252, 165, 165, 0.4);
    color: #b91c1c;
    background: #fef2f2;
  }

  .eco-badge.go {
    border-color: rgba(103, 232, 249, 0.4);
    color: #0e7490;
    background: #ecfeff;
  }

  .eco-badge.cargo {
    border-color: rgba(253, 186, 116, 0.4);
    color: #c2410c;
    background: #fff7ed;
  }

  .eco-badge.pypi {
    border-color: rgba(147, 197, 253, 0.4);
    color: #1d4ed8;
    background: #eff6ff;
  }

  .trust-pill {
    flex-shrink: 0;
    min-width: 2.5rem;
    padding: 0.125rem 0.5rem;
    font-size: 0.8rem;
    font-weight: 800;
    border-radius: 999px;
    border: 1.5px solid;
    text-align: center;
    background: transparent;
  }

  .trust-high {
    color: #16a34a;
    border-color: #22c55e;
  }

  .trust-medium {
    color: #f59e0b;
    border-color: #fbbf24;
  }

  .trust-low {
    color: #dc2626;
    border-color: #ef4444;
  }

  @media (max-width: 1100px) {
    header {
      flex-wrap: wrap;
      padding: 0.75rem 1.25rem;
    }

    .header-middle {
      order: 3;
      width: 100%;
      flex-basis: 100%;
    }

    .search-shell {
      max-width: none;
    }

    .search-bar {
      margin-left: 0;
      margin-right: 0;
    }

    .suggestions-dropdown {
      left: 0;
      right: 0;
    }
  }

  @media (max-width: 640px) {
    .search-bar {
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .search-input,
    .search-button {
      width: 100%;
      border-radius: 10px;
      border-right: 1px solid var(--border);
    }

    .header-right {
      width: 100%;
      justify-content: flex-end;
      flex-wrap: wrap;
    }
  }
</style>
