+++
title = "Package Security Levels"
tags = ["reproducible_builds", "behavioral_detection", "attestations"]
+++

## Problem

We need a way to system to clearly explain the security level of a package, so that users/organisations can make informed decisions about which packages to use. This is also important for defining rules per organisation in #45.

## Recommendation

The way I see it we have these levels of verification (higher level = more secure, and will have all of the properties of the lower levels):
Level 0: NPM Mirror
Level 1: Behavioural Analysis -> NOTE: can be done at scale, we should be able to avoid any Level 0 packages, all higher levels will have this included aswell.
Level 2: Attestations (e.g. NPM provenance)
Level 3: Reproducible Builds (External) -> ie: OSS-rebuild, Nix PyPi Stuff
Level 4: Reproducible Builds (Internal) -> Manual (All packages in `/cmd/rep-build`)

Security levels will be defined per package version, not per package ie: `svelte:1.0.0` can be level 1, while `svelte:1.7.0` can be level 3.

Some packages may not require a higher level, they may not gain any security benefits from it. Such as Go modules, they are already verified by the checksum database, and the source code is distributed, not binaries. No reason to create a manual reproducible build process for them.

Organisations can then define rules such as "Only use packages with security level 3 or higher", and this will be enforced by the package manager. However, to allow for more fine control, we will have each organisations package have its own "security level" field via the `organisations_packages` table.

## Open Questions

1. What level should Go checksum database be?
2. Should this model be strictly linear? For instance, what if we could run reproducible builds but not behaviour analysis? Should it instance be like a list of properties instead of a security level? I think the security level is easier to understand, and adding this complexity isn't necessary for MVP, especially since it makes it easier for us to assign a security number like the client wanted.
