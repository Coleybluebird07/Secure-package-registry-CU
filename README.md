# Secure Package Registry

Stopping supply chain attacks

## Developing

You can get all the nice `bun run dev` features (hot reloading etc) by running this other podman command:

```bash
podman compose -f compose.yml -f compose.dev.yml up -d
```

It will prioritise any settings in `compose.dev.yml` over `compose.yml`, so you can add any development-specific settings there. Right now only `dashboard-ui` is configured to work with this.
