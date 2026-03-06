<script lang="ts">
	import { authClient } from '$lib/client';
	import { goto } from '$app/navigation';

	let firstName = $state('');
	let lastName = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}
		loading = true;
		error = '';

		const { data, error: authError } = await authClient.signUp.email({
			email,
			password,
			name: `${firstName} ${lastName}`.trim()
		});

		if (authError) {
			error = authError.message ?? 'Sign up failed';
		} else {
			goto('/');
		}
		loading = false;
	}
</script>

<div class="register-page">
	<div class="register-container">
		<div class="register-card">
			<h1>Create Account</h1>
			<p class="register-subtitle">Get started with secure package verification</p>

			<form class="register-form" onsubmit={handleSubmit}>
				<div class="form-row">
					<div class="form-group">
						<label for="firstName">First Name</label>
						<input type="text" id="firstName" name="firstName" placeholder="John" required bind:value={firstName} />
					</div>

					<div class="form-group">
						<label for="lastName">Last Name</label>
						<input type="text" id="lastName" name="lastName" placeholder="Doe" required bind:value={lastName} />
					</div>
				</div>

				<div class="form-group">
					<label for="email">Email</label>
					<input type="email" id="email" name="email" placeholder="you@company.com" required bind:value={email} />
				</div>

				<div class="form-group">
					<label for="company">Company Name</label>
					<input type="text" id="company" name="company" placeholder="Better NPM" />
				</div>

				<div class="form-group">
					<label for="password">Password</label>
					<input type="password" id="password" name="password" placeholder="•••••••" required bind:value={password} />
					<span class="helper-text">At least 8 characters</span>
				</div>

				<div class="form-group">
					<label for="confirmPassword">Confirm Password</label>
					<input
						type="password"
						id="confirmPassword"
						name="confirmPassword"
						placeholder="••••••••"
						required
						bind:value={confirmPassword}
					/>
				</div>

				{#if error}
					<p class="error-message">{error}</p>
				{/if}

				<label class="checkbox">
					<input type="checkbox" name="terms" required />
					<span
						>I agree to the <a href="/terms">Terms of Service</a> and
						<a href="/privacy">Privacy Policy</a></span
					>
				</label>

				<button type="submit" class="submit-button" disabled={loading}>
					{loading ? 'Creating account...' : 'Create Account'}
				</button>
			</form>

			<div class="form-footer">
				<p>Already have an account? <a href="/login">Sign in</a></p>
			</div>
		</div>
	</div>
</div>

<style>
	.register-page {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--bg-primary);
		padding: 2rem;
	}

	.register-container {
		width: 100%;
		max-width: 520px;
	}

	.register-card {
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

	.register-subtitle {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: 2.5rem;
		font-size: 0.95rem;
	}

	.register-form {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
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

	input[type='text'],
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

	input[type='text']:focus,
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

	.helper-text {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	.checkbox {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		cursor: pointer;
		color: var(--text-secondary);
		font-size: 0.875rem;
		line-height: 1.5;
	}

	.checkbox input[type='checkbox'] {
		width: 16px;
		height: 16px;
		margin-top: 0.15rem;
		cursor: pointer;
		flex-shrink: 0;
		accent-color: var(--accent);
	}

	.checkbox a {
		color: var(--accent);
		text-decoration: none;
	}

	.checkbox a:hover {
		text-decoration: underline;
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
		.register-card {
			padding: 2rem;
		}

		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
