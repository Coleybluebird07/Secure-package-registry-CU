import { apiKey } from "@better-auth/api-key";
import { betterAuth } from "better-auth";
import { admin, organization } from "better-auth/plugins";
import { sveltekitCookies } from "better-auth/svelte-kit";
import nodemailer from "nodemailer";
import { Pool } from "pg";
import { getRequestEvent } from "$app/server";

const transporter = nodemailer.createTransport({
	host: process.env.SMTP_HOST ?? "mailpit",
	port: Number(process.env.SMTP_PORT ?? 1025),
	secure: false,
});

export const auth = betterAuth({
	database: new Pool({
		database: process.env.POSTGRES_DB,
		host: process.env.POSTGRES_HOST,
		password: process.env.POSTGRES_PASSWORD,
		port: process.env.POSTGRES_PORT
			? Number.parseInt(process.env.POSTGRES_PORT, 10)
			: 5432,
		user: process.env.POSTGRES_USER,
	}),

	emailAndPassword: {
		autoSignIn: false,
		enabled: true,
		requireEmailVerification: true,
	},

	emailVerification: {
		callbackURL: "/verified",
		sendVerificationEmail: async ({ user, url }) => {
			const verificationUrl = new URL(url);
			verificationUrl.searchParams.set("callbackURL", "/verified");
			await transporter.sendMail({
				from: "noreply@spr.local",
				html: `<p>Click the link below to verify your email address:</p><p><a href="${verificationUrl.toString()}">${verificationUrl.toString()}</a></p>`,
				subject: "Verify your email address",
				to: user.email,
			});
		},
	},

	plugins: [
		admin(),
		organization(),
		apiKey(),
		sveltekitCookies(getRequestEvent), // make sure this is the last plugin in the array
	],

	trustedOrigins: process.env.TRUSTED_ORIGINS
		? process.env.TRUSTED_ORIGINS.split(",")
		: [
				"http://localhost:8000",
				"http://localhost:5174",
				"http://localhost:5173",
				"http://localhost:3001",
			],
});
