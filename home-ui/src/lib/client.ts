// Client-side Better Auth instance — runs in the browser.
// Communicates with the server via HTTP (e.g. POST /api/auth/sign-in/email).
// Methods like authClient.signIn.email() and authClient.signUp.email()
// are called from Svelte components.

import { createAuthClient } from "better-auth/svelte";

export const authClient = createAuthClient({
	// Must match BETTER_AUTH_URL in .env
	baseURL: "http://localhost:5173",
});
