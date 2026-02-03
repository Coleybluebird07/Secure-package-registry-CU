+++
title = "Initial UI Wireframe"
author = "Bolajims@cardiff.ac.uk"
reviewer = [""]
date = "2026-02-02"
status = "draft"
+++
## Summary
This RFC defines the initial user interface structure and **data requirements**
for the Secure Package Registry. The accompanying wireframe is **low-fidelity**
and intended to illustrate layout and information hierarchy rather than final
visual design.

## Problem
Users need a clear way to:
- Search for packages in the secure registry
- View package verification status
- Request verification for unverified packages

The UI must reflect the system’s role as an **alternative package registry**,
rather than a vulnerability scanning or threat-intelligence platform.

## Data Requirements (Initial Version)

The primary interface must display the following information:

- Package name
- Package version
- Verification status:
    - Reproducible
    - Behavioral consistency
    - Verification pending
    - Verification failed
- Date of last verification
- Action to request verification (if unverified)

## User Interaction Flow (v1)

1. User searches for a package
2. Registry returns matching packages and versions
3. Verification status is displayed for each result
4. If a package is unverified, the user may request verification
5. Status updates asynchronously once analysis completes

## Recommendation
A low-fidelity Excalidraw mockup accompanies this RFC and focuses on
data requirements and interaction flow rather than visual design.

The wireframe is stored at:
`design/wireframes/initial-ui-wireframe.excalidraw`

## Open Questions
- Should verification status be shown as tags, icons, or text labels?
- Is a package detail view required in v1, or are search results sufficient?
- What metadata (if any) should be exposed beyond verification status?