# Security Policy

Shoka stores and serves a private library, so reports about authorization,
path traversal, archive handling, and credential exposure are taken seriously.

## Supported versions

There are no maintained release branches. Only the current `main` is
supported; fixes land there and older commits or images receive no backports.

## Reporting a vulnerability

Use GitHub's private vulnerability reporting:

<https://github.com/Jqnx/Shoka/security/advisories/new>

Do not open a public issue, pull request, or discussion for a suspected
vulnerability.

Include reproduction steps, the affected commit or image tag, the expected and
actual behavior, and any proof-of-concept input. For archive-processing bugs, a
minimal malicious archive is more useful than a description of one.

## Scope

In scope:

- the Go API in `backend/`
- the SvelteKit frontend in `frontend/`
- the `Dockerfile` and entrypoints in `docker/`
- authentication, session, and permission handling
- archive extraction, metadata parsing, and image processing paths reachable
  from an uploaded or scanned file

Out of scope:

- vulnerabilities in third-party metadata sources
- misconfiguration of an operator's reverse proxy, mounts, or `PUID`/`PGID`
- denial of service from a deliberately oversized library on the reporter's own
  instance

## What to expect

This is a self-hosted project without a bounty program. Reports are
acknowledged on a best-effort basis, normally within seven days. Credit is
given in the advisory unless you ask otherwise.
