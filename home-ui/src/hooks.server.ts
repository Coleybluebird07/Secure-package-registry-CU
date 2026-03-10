// SvelteKit server hook — runs on every request before it reaches a route.
// Used here to check the session once globally and attach it to event.locals,
// making the current user available to any server-side page or endpoint.

import type { Handle } from "@sveltejs/kit";
import { svelteKitHandler } from "better-auth/svelte-kit";
import { building } from "$app/environment";
import { auth } from "$lib/auth";

export const handle: Handle = async ({ event, resolve }) => {
	// Look up the session from the request cookies
	const session = await auth.api.getSession({
		headers: event.request.headers,
	});

	// Attach session and user to locals so routes can access them server-side
	if (session) {
		event.locals.session = session.session;
		event.locals.user = session.user;
	}

	// Hands off to Better Auth — this also handles the /api/auth/* routes automatically
	return svelteKitHandler({ auth, building, event, resolve });
};
