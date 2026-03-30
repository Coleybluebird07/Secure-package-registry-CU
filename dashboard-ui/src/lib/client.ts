import { apiKeyClient } from "@better-auth/api-key/client";
import { dashClient } from "@better-auth/infra/client";
import { createAuthClient } from "better-auth/svelte";

export const authClient = createAuthClient({
	plugins: [apiKeyClient(), dashClient()],
});

export type Session = typeof authClient.$Infer.Session;
