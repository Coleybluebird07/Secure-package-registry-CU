import { redirect } from "@sveltejs/kit";
import { auth } from "$lib/auth";
import type { PageServerLoad } from "./$types";

export const load: PageServerLoad = async ({ request }) => {
	await auth.api.signOut({
		headers: request.headers,
	});
	redirect(302, "/");
};
