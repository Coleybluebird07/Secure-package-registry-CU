import { json } from "@sveltejs/kit";
import type { RequestHandler } from "./$types";

export const POST: RequestHandler = async ({
	params,
	request,
	locals,
	fetch,
}) => {
	const body = await request.json();
	const user = locals.user;

	if (!user) {
		return json({ error: "unauthorized" }, { status: 401 });
	}

	const reviewer = user.id;

	const response = await fetch(
		`http://caddy:8000/api/v1/admin/packages/${encodeURIComponent(params.ecosystem)}/${encodeURIComponent(params.identifier)}/review`,
		{
			body: JSON.stringify({
				maintainer_trust_level: body.maintainer_trust_level,
				notes: body.notes,
				reviewed_by: reviewer,
				status: body.status,
			}),
			headers: {
				"Content-Type": "application/json",
			},
			method: "POST",
		},
	);

	const data = await response.json().catch(() => ({
		error: "failed to parse backend response",
	}));

	return json(data, { status: response.status });
};
