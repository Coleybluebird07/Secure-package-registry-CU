<script lang="ts">
</script>

<div class="verified-page">
  <div class="verified-card">
    <div class="icon">✓</div>
    <h1>Email Verified</h1>
    <p>Your account has been verified. You can now sign in.</p>
    <a href="/login" class="sign-in-button">Sign In</a>
  </div>
</div>

<style>
  .verified-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-primary);
  }

  .verified-card {
    background: var(--card-bg);
    border: 1px solid var(--card-border);
    border-radius: 16px;
    padding: 3rem 2.5rem;
    text-align: center;
    max-width: 400px;
    width: 100%;
  }

  .icon {
    width: 64px;
    height: 64px;
    background: #22c55e;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 2rem;
    color: white;
    margin: 0 auto 1.5rem;
  }

  h1 {
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--text-primary);
    margin-bottom: 0.75rem;
  }

  p {
    color: var(--text-secondary);
    margin-bottom: 2rem;
  }

  .sign-in-button {
    display: inline-block;
    background: var(--accent);
    color: var(--bg-primary);
    padding: 0.875rem 2rem;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    text-decoration: none;
    transition: all 0.3s;
  }

  .sign-in-button:hover {
    background: var(--accent-hover);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(79, 195, 247, 0.3);
  }
</style>
