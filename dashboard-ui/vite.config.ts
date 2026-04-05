import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

const proxyTarget = process.env.VITE_PROXY_TARGET ?? "http://localhost:8000";
const allowedHostsEnv = process.env.VITE_ALLOWED_HOSTS ?? "";
const allowedHosts = allowedHostsEnv.split(",");
console.log("VITE_ALLOWED_HOSTS env:", allowedHostsEnv);
console.log("allowedHosts config:", allowedHosts);

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		allowedHosts: allowedHosts,
		proxy: {
			"/api/v1": {
				changeOrigin: true,
				target: proxyTarget,
			},
		},
	},
});
