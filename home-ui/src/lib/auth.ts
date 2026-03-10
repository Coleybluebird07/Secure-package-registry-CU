// Server-side Better Auth config. Only runs on the server — never import this from client code.

import { betterAuth } from "better-auth";
import { organization } from "better-auth/plugins";
import { sveltekitCookies } from "better-auth/svelte-kit";
import { Pool } from "pg";
import { getRequestEvent } from "$app/server";
// Reads private environment variables from .env (server-side only)
import { env } from "$env/dynamic/private";

export const auth = betterAuth({
	// URL of the app, used by Better Auth to build callback URLs
	baseURL: env.BETTER_AUTH_URL,

	// PostgreSQL connection pool — reuses connections rather than opening one per request
	database: new Pool({
		database: env.POSTGRES_DB,
		host: env.POSTGRES_HOST,
		password: env.POSTGRES_PASSWORD,
		port: env.POSTGRES_PORT ? Number.parseInt(env.POSTGRES_PORT) : 5432,
		user: env.POSTGRES_USER,
	}),

	// Enable email/password login and registration
	emailAndPassword: {
		enabled: true,
	},

	plugins: [
		// Adds organisation support with roles (owner, member, etc.)
		organization(),
		// Integrates Better Auth sessions with SvelteKit cookies — must be last
		sveltekitCookies(getRequestEvent),
	],
});
