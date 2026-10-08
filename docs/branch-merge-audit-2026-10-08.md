# Branch integration audit — 8 October 2026

Reviewed main: `76003e15f761e0a7a75def91f8a586cb4b26b5ed`. All remote branches were inventoried. No open pull requests existed at the start. No branches were deleted or force-pushed.

## Decision and order

1. Integrate `dev` into current `main` on a separate branch, resolving both implementations together.
2. Preserve the two independent historical RFC branches (`rfc-auth-data`, `rfc-malware-case-studies`).
3. Validate the combined result before merging the integration pull request into `main` with a merge commit, retaining ancestry.

Do not merge every remaining branch. Branch names and ahead counts do not establish whether their changes are current or useful. Rebased equivalents were checked using patch comparisons as well as ancestry.

## Integration repairs

- Preserve rebuild worker, completion consumer, API routes, artifacts, manual retries, side-by-side comparison, and PyPI/Cargo watcher support.
- Retain new project processing, review tags, registry policy and admin authorization from dev.
- Keep local Gitea configuration and this fork's mirror destinations.
- Keep rebuild migration 000005; move security levels to 000006 and assign unique later versions to dev features. Add the text enum value before inserting text-valued tag types. Regenerate database bindings including the previously missing security-level type.
- Preserve package-creation tests and adapt them to the atomic upsert result instead of deleting them.
- Authenticate rebuild requests and artifact downloads under the new admin middleware.
- Require a runtime BetterAuth secret; use a build-only placeholder during static analysis/build. Update setup instructions and Go version.
- Repair existing rollback ordering/table-name errors found by actual rollback testing.

## Validation

- Both web interfaces: type checks, configured formatting/lint checks and production builds pass.
- Go lint: zero issues.
- All Go packages compile; rebuild-worker, package-watcher, private API, config, Gitea, npm, lockfile and OSS Rebuild test suites pass.
- Full `go test ./...` still fails seven behavior tests because the repository does not include `context-references/sample-behaviors/{safe,safe-2,malicious}.jsonl`. The untouched main baseline has the same missing-fixture failures and additionally fails compilation with undefined `NullSecurityLevel`; the compilation error is repaired here. Tests were not skipped or weakened to conceal missing fixtures.
- All 15 migration files apply successfully to a fresh PostgreSQL database. An upgrade sequence preserves an existing package row. Rollback was exercised in an isolated temporary schema.
- Compose configuration validates with a supplied test secret.
- Full Docker application startup and live external rebuild execution have not been validated. Passing builds is not a production-readiness claim.

## Database compatibility

This integration targets main's migration history, retaining rebuild migration version 5. A database previously initialized from dev uses different meanings for migration versions 5–12: do not point that database at this migration sequence without a schema-specific conversion. Back up existing data before deployment. No user's running database was modified during this audit.

If an installation manually applied the old security-level migration as version 5 instead of the rebuild migration, inspect and reconcile its rebuild_tasks schema first. The original main directory had two version-5 migrations and could not be consumed as a valid migration sequence.

## Every feature branch

| Branch | Decision | Reason |
| --- | --- | --- |
| `19-as-an-authenticated-user-i-want-to-search-for-packages-by-name-so-that-i-can-quickly-locate-a` | Superseded; retain branch | dev has URL-initialized, paginated search and a newer package details route. The older header/search replacement should not overwrite these. |
| `26-as-an-authenticated-user-i-want-to-see-an-overview-dashboard-of-my-organisation-s-spr-status-so` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `31-as-a-developer-i-want-users-to-authenticate-on-our-application-so-they-can-have-specific-user` | Superseded; retain branch | Early home-ui authentication and Compose prototype, including accidental files. Current authentication lives in dashboard-ui. |
| `31_new_auth` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `42-better-auth-dashboard` | Superseded; retain branch | Older auth integration includes duplicate hook files and changes previously applied migrations. |
| `57-package-review-interface` | Alternative implementation; retain branch | Separate package-review tables conflict with current schemas and admin UI. dev uses per-version review tags and a review queue; do not combine both policies blindly. |
| `58-rfc-for-betterauth-dashboard` | Already present as equivalent content | Its only changed RFC is byte-identical to the version on main; merging adds no content. |
| `api-token-frontend` | Superseded key model; retain branch | Creates user-level BetterAuth keys, whereas dev registry authorization now requires project-scoped keys. Merging this UI would expose keys that do not authenticate to the registry. |
| `auth-ui-init` | Superseded; retain branch | Earlier auth consolidation; merging wholesale would remove current admin-route protection and overwrite newer auth changes. |
| `auth_email_verification` | Needs separate port; retain branch | Email verification is unique unfinished integration work. Port SMTP and verification into current dashboard auth, with forward migrations and tests; do not merge old home-ui auth or replace existing migrations. |
| `dev` | Integrated with repairs | Project dashboard, dependency resolution, verification, review queue, admin authorization and project-scoped registry keys. Preserve main rebuild workflow; resolve conflicts and migration collisions. |
| `fix-branch` | Superseded; retain branch | Pipeline ref fix is already incorporated; old mirror configuration targets a different owner. Current fork mirror destinations are preserved. |
| `formatting` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `formatting-rfc` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `future-wireframe` | Conflicting historical proposal; retain branch | Conflicts with the existing wireframe document and describes an earlier interface. Preserve the branch for deliberate documentation reconciliation. |
| `gitea-secrets-tokens-2` | Obsolete and unsuitable; retain branch | Old scaffolding and a committed token file. Current local Gitea setup generates tokens at runtime; do not reintroduce the token file. |
| `integration-clients` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `package-security-levels` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `readme` | Superseded; retain branch | Current README preserves portfolio attribution and a fuller setup guide; old draft would replace this. |
| `redis-database` | Superseded infrastructure; retain branch | Old scaffolding/Valkey implementation. Current Compose already includes Valkey and newer runtime integration. |
| `redis-database-2` | Obsolete and unsuitable; retain branch | Older Valkey setup plus committed token file; current setup is retained. |
| `refactor-details` | Do not merge | The final net diff from its main ancestor only deletes the detail-page route. It does not deliver the feature suggested by its name. |
| `rep-build-prettier` | Included through dev equivalents | All non-merge patches have equivalents in dev; avoid replaying old build integration. |
| `repro-backend` | Included through dev equivalents | All non-merge patches have equivalents in dev. |
| `revert-f9a27e1c` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `review-registries` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `rfc-auth-data` | Merged as historical design | Preserve the original authentication proposal; it is not the current database schema or executable migration. |
| `rfc-malware-case-studies` | Merged as research documentation | Preserve the supply-chain case studies; no referenced samples were downloaded or executed. |
| `rfc-schema` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `test/rebuild-verification` | Already merged into main | Branch tip is an ancestor of the reviewed main commit; no additional merge needed. |
| `ui-att-rep` | Included/evolved through dev | Verification UI and behavior tags are present in dev, including subsequent fixes; old UI should not overwrite them. |
| `upstream-attest` | Included through dev equivalents | Attestation clients and tests are included through equivalent dev commits. |
| `user-dashboard` | Included/evolved through dev | Project dashboard and dependency tracking were rebased and developed further in dev. Old user-key/CLI authentication was explicitly removed later. |

## Remaining work

Restore the original behavior fixtures, separately port email verification, and run a complete local-stack/external-service acceptance test. Historical proposals remain proposals; the merged RFCs do not change runtime authorization policy.
