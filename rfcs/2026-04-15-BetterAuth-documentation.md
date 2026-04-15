## Purpose

Better Auth is integrated into this project for user authentication. With this there is a plugin `dash()` which enables a proxied connection between our localhost containerised instance and `dash.better-auth.com`. We are able to do this using Cloudflare tunnel. This creates a "tunnel" that enables the instance running locally to gain a connection to the dashboard.

> **Note:** The tunnel is necessary because `dash.better-auth.com` needs to reach your local instance over a public URL, it cannot connect to `localhost` directly. Each time you restart `cloudflared` with the `--url` flag, a **new URL is generated**, meaning you'll need to update `BETTER_AUTH_URL` and the Better Auth dashboard settings again. To avoid this, consider using a named tunnel instead.

## Prerequisites

In order to get this running and working you need a few dependencies installed.

- Podman/Docker installed
- Cloudflare tunnel CLI (cloudflared) installed and an authenticated cloudflared config. [See guide](https://developers.cloudflare.com/cloudflare-one/networks/connectors/cloudflare-tunnel/get-started/create-remote-tunnel-api/)
- The `.env` file needs to include some required values such as:
  - `BETTER_AUTH_SECRET` to generate this use `openssl rand -base64 32`
  - `BETTER_AUTH_URL` once running set this to the URL provided by the tunnel
  - `BETTER_AUTH_API_KEY` from the Better Auth dashboard in settings
  - `TRUSTED_ORIGINS` include tunnel URL + localhost ports

## Set up

- All migration files need to be ran before Better Auth can work, migrations `000002 core auth tables`, `000003 org packages` and `000004 API keys`.
- To ensure they have run, start up podman machine `podman machine start` _(macOS/Windows only, skip this step on Linux)_. Then `podman compose down; podman compose build; podman compose up`.
- Once podman is running and you can access the application on port 7001, start the cloudflare tunnel `cloudflared tunnel --url http://localhost:7001` and add the URL provided soon after this command is ran into Better Auth settings [https://dash.better-auth.com](https://dash.better-auth.com). Also update `BETTER_AUTH_URL` in `.env` with this URL at this point, and add the tunnel URL to `TRUSTED_ORIGINS` if not already set. Keep the cloudflare command running in terminal, and rebuild podman with the 3 commands above.
- Once you get your API key from Better Auth dashboard settings, paste into `BETTER_AUTH_API_KEY=` in `.env`.

To check it is working, register a user and login, enter the podman container and check `core_db` in postgres to check there is a valid record made for a user. To do this use these commands below:

- Find container name `podman ps`
- Enter bash terminal of the podman instance `podman exec -it <container_name> bash`
- Enter into the PSQL instance `psql -U <db_user> -d core_db`
- Then a simple query on the user table to check if the registered user has populated the table `SELECT * FROM "user";`

This means everything is working authentication/database wise on the application itself. Check on the Better Auth dashboard and you should see some user information displayed.
