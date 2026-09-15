# Secure Package Registry

A university prototype for investigating software supply-chain risks by rebuilding packages and comparing the results with published artifacts.

Developed by a five-person Cardiff University team. This repository is my portfolio fork of the team project, including my additions to the rebuild and verification workflows.

## My contribution

I am David Cole. My work included:

- Implementing database and background-worker workflows for package rebuilding and verification.
- Adding Diffoscope comparison reports and APIs for retrieving build artifacts.
- Adding manual controls, side-by-side comparisons, failure handling and resource limits.
- Extending package support to PyPI and Cargo.
- Adding verification tests and GitLab CI automation.

The original platform and wider architecture were developed collaboratively. This fork includes that shared code as well as my contributions.

## What the system does

The rebuild workflow lets a developer investigate whether a published package matches the output of a rebuild:

1. A package version is submitted for verification.
2. A background worker retrieves the published artifact and invokes the configured rebuild command.
3. The worker compares the original and rebuilt artifacts.
4. When they differ, Diffoscope produces a report to help inspect the differences.
5. Results and available artifacts are stored and exposed through the API and dashboard.

A mismatch is a finding to investigate, not proof of malicious code. Likewise, a matching rebuild alone does not prove that a package is safe.

The wider project includes package watching, registry services and behavioural analysis components.

## Technology

| Area | Tools |
| --- | --- |
| Backend and workers | Go |
| Web interfaces | Svelte, TypeScript |
| Database | PostgreSQL |
| Messaging | RabbitMQ |
| Artifact storage | MinIO |
| Supporting services | Gitea, Valkey, Caddy |
| Rebuild investigation | OSS Rebuild, Diffoscope |
| Development and CI | Docker Compose, GitLab CI/CD, Go tests |

## Project structure

```text
cmd/                         Service entry points and utilities
pkg/services/rebuild-worker/ Rebuild execution and comparison
pkg/services/package-watcher/ Package version watching
pkg/services/core-svc/       Core service and API handlers
pkg/config/                  Runtime configuration
dashboard-ui/                Management dashboard
home-ui/                     Public-facing interface
internal/                    Internal code and generated database access
scripts/                     Rebuild integration scripts
docs/svc/                    OpenAPI specification
rfcs/                        Team design proposals and decisions
infra/                       Supporting infrastructure configuration
compose.yml                  Local service stack
```

## Run locally

### Requirements

- Git and Docker with Docker Compose.
- For Go development, a toolchain compatible with `go.mod`, which specifies Go 1.25.6 in this snapshot.
- For frontend development and repository checks, Bun. The optional `just` task runner exposes the commands in `justfile`.

### Start the development stack

```bash
git clone https://github.com/Coleybluebird07/Secure-package-registry-CU.git
cd Secure-package-registry-CU
docker compose up --build
```

The Compose configuration defines these local entry points:

| Service | Address |
| --- | --- |
| Dashboard and proxied API | `http://localhost:7001` |
| Home interface | `http://localhost:7003` |
| RabbitMQ management | `http://localhost:10001` |
| MinIO console | `http://localhost:10004` |
| Gitea | `http://localhost:10005` |

The default configuration sets `SPR_MOCK=true`, which seeds development data. It does not mean that external rebuild dependencies are simulated or that every package can be rebuilt.

This is a local development configuration with demonstration credentials. The main service mounts the Docker socket to support rebuilds, giving it access to the host Docker daemon. Run it in a development environment you control.

### Rebuild dependencies

The default integration script is `scripts/oss-rebuild-artifact.sh`. Actual rebuilds depend on the OSS Rebuild CLI, Docker access and an available rebuild record for the selected package version. Diffoscope must also be available to generate detailed comparisons.

The script handles npm, PyPI and Cargo packages. Its Go-module rebuild path is explicitly unsupported in this snapshot, even though Go appears elsewhere in the project's package model.

Relevant settings are defined in `compose.yml` and `pkg/config/config.go`:

| Setting | Purpose |
| --- | --- |
| `OSS_REBUILD_CMD` | Command used to produce the rebuilt artifact |
| `DIFFOSCOPE_CMD` | Diffoscope executable |
| `OSS_REBUILD_TIMEOUT` | Time limit for rebuilding |
| `DIFFOSCOPE_TIMEOUT` | Time limit for comparison reporting |
| `MAX_REBUILD_ARTIFACT_SIZE` | Maximum accepted artifact size |
| `MAX_DIFFOSCOPE_REPORT_SIZE` | Maximum accepted comparison report size |
| `GITHUB_TOKEN`, `GITHUB_OWNER`, `GITHUB_REPO` | Configuration for GitHub-backed workflows when used |

Keep access tokens out of version control. Some original university infrastructure or workflow configuration may need to be replaced for an independent setup.

## Tests and checks

From the repository root:

```bash
go test ./...
```

Targeted tests for the rebuild and watcher components:

```bash
go test ./pkg/services/rebuild-worker ./pkg/services/package-watcher
```

The rebuild-worker tests include report generation, comparison failure and timeout cases. Some tests use fake executables to check worker behaviour without performing a real external rebuild.

With the required development tools installed:

```bash
just check
just integration
```

`just check` runs the repository's formatting, lint and frontend checks. `just integration` runs integration-tagged Go tests and requires the relevant services and configuration. These commands document the available checks, not a claim that they have all passed in a fresh environment.

## Design documentation

- [Team design proposals](rfcs/)
- [API specification](docs/svc/openapi.yaml)
- [Runtime configuration](pkg/config/config.go)
- [Local service configuration](compose.yml)

## Project scope

This is an educational prototype, not a production security guarantee. Package and version coverage depends on the configured rebuild tooling. Development credentials, Docker permissions, deployment settings and external integrations need review before wider use.

## Credits

Built collaboratively as a Cardiff University software engineering project. This fork presents my additions alongside the original team's work.

[David Cole on GitHub](https://github.com/Coleybluebird07)
