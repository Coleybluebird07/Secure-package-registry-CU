<script lang="ts">
	import { authClient } from '$lib/client';
	import { goto } from '$app/navigation';

	let email = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		loading = true;
		error = '';

		const { data, error: authError } = await authClient.signIn.email({ email, password });

		if (authError) {
			error = authError.message ?? 'Sign in failed';
		} else {
			goto('/');
		}
		loading = false;
	}
</script>

<div class="login-page">
	<div class="login-container">
		<div class="login-card">
			<h1>Sign In</h1>
			<p class="login-subtitle">Access your secure package registry</p>

			<form class="login-form" onsubmit={handleSubmit}>
				<div class="form-group">
					<label for="email">Email</label>
					<input type="email" id="email" name="email" placeholder="you@example.com" required bind:value={email} />
				</div>

				<div class="form-group">
					<label for="password">Password</label>
					<input type="password" id="password" name="password" placeholder="••••••••" required bind:value={password} />
				</div>

				{#if error}
					<p class="error-message">{error}</p>
				{/if}

				<div class="form-options">
					<label class="checkbox">
						<input type="checkbox" name="remember" />
						<span>Remember me</span>
					</label>
					<a href="/forgot-password" class="forgot-link">Forgot password?</a>
				</div>

				<button type="submit" class="submit-button" disabled={loading}>
					{loading ? 'Signing in...' : 'Sign In'}
				</button>
			</form>

			<div class="form-footer">
				<p>Don't have an account? <a href="/register">Sign up</a></p>
			</div>
		</div>
	</div>
</div>

<style>
	.login-page {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-primary);
		padding: 2rem;
	}

	.login-container {
		width: 100%;
		max-width: 440px;
	}

	.login-card {
		background: var(--card-bg);
		border: 1px solid var(--card-border);
		border-radius: 16px;
		padding: 3rem;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
	}

	h1 {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
		text-align: center;
	}

	.login-subtitle {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: 2.5rem;
		font-size: 0.95rem;
	}

	.login-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	label {
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	input[type='email'],
	input[type='password'] {
		padding: 0.875rem 1rem;
		border: 1px solid var(--border);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--bg-primary);
		color: var(--text-primary);
		transition: all 0.2s;
	}

	input[type='email']:focus,
	input[type='password']:focus {
		outline: none;
		border-color: var(--accent);
		box-shadow: 0 0 0 3px rgba(79, 195, 247, 0.1);
	}

	input::placeholder {
		color: var(--text-secondary);
		opacity: 0.5;
	}

	.form-options {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.875rem;
	}

	.checkbox {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		cursor: pointer;
		color: var(--text-secondary);
	}

	.checkbox input[type='checkbox'] {
		width: 16px;
		height: 16px;
		cursor: pointer;
		accent-color: var(--accent);
	}

	.forgot-link {
		color: var(--accent);
		text-decoration: none;
		transition: color 0.2s;
	}

	.forgot-link:hover {
		color: var(--accent-hover);
	}

	.submit-button {
		background: var(--accent);
		color: var(--bg-primary);
		padding: 0.875rem 2rem;
		border: none;
		border-radius: 8px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s;
		margin-top: 0.5rem;
	}

	.error-message {
		color: #ef4444;
		font-size: 0.875rem;
		text-align: center;
	}

	.submit-button:hover {
		background: var(--accent-hover);
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(79, 195, 247, 0.3);
	}

	.form-footer {
		margin-top: 2rem;
		padding-top: 2rem;
		border-top: 1px solid var(--border);
		text-align: center;
	}

	.form-footer p {
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.form-footer a {
		color: var(--accent);
		text-decoration: none;
		font-weight: 600;
		transition: color 0.2s;
	}

	.form-footer a:hover {
		color: var(--accent-hover);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.login-card {
			padding: 2rem;
		}

		.form-options {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}
	}
</style>
