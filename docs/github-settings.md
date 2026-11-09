# GitHub repository settings

Applied to `Jqnx/Shoka` (public) on 2026-11-09 via the GitHub REST API with `gh`.
This file records the baseline, the settings that were changed, and the ones
deliberately left alone, so the configuration can be audited and reproduced
without reading repository settings in the web UI.

## Baseline before the change

| Setting | Value |
| --- | --- |
| Branch protection / rulesets on `main` | none |
| Secret scanning | disabled |
| Secret scanning push protection | disabled |
| Secret scanning validity checks | disabled |
| Secret scanning non-provider patterns | disabled |
| Dependabot alerts | enabled |
| Dependabot security updates | disabled |
| Private vulnerability reporting | disabled |
| Actions allowed | all |
| Actions require SHA-pinned actions | false |
| Actions default `GITHUB_TOKEN` permissions | read (already correct) |
| Actions can approve pull request reviews | false (already correct) |
| Fork pull request approval | first-time contributors |
| Merge methods | merge commit, squash, rebase |
| Delete head branches on merge | false |
| `SECURITY.md` / `.github/dependabot.yml` | missing |
| `v2` branch | 0 commits ahead of `main`, 3 behind |

## Applied

### Secret scanning and Dependabot

```bash
gh api -X PATCH repos/Jqnx/Shoka --input - <<'JSON'
{
  "security_and_analysis": {
    "secret_scanning": {"status": "enabled"},
    "secret_scanning_push_protection": {"status": "enabled"},
    "secret_scanning_validity_checks": {"status": "enabled"},
    "secret_scanning_non_provider_patterns": {"status": "enabled"},
    "dependabot_security_updates": {"status": "enabled"}
  }
}
JSON

gh api -X PUT repos/Jqnx/Shoka/private-vulnerability-reporting -F enabled=true
```

Verified state:

```text
secret_scanning                         enabled
secret_scanning_push_protection         enabled
dependabot_security_updates             enabled
secret_scanning_validity_checks         disabled (see Deferred)
secret_scanning_non_provider_patterns   disabled (see Deferred)
private vulnerability reporting         enabled
```

### Merge hygiene

```bash
gh api -X PATCH repos/Jqnx/Shoka \
  -F allow_merge_commit=false \
  -F allow_rebase_merge=false \
  -F delete_branch_on_merge=true
```

Verified: `allow_squash_merge=true`, `allow_merge_commit=false`,
`allow_rebase_merge=false`, `delete_branch_on_merge=true`.

Squash-only keeps `main` linear; the ruleset also enforces linear history, so
merge commits are rejected at the branch level too.

### Ruleset `main` (id 26892894)

Encoded via:

```bash
gh api -X POST repos/Jqnx/Shoka/rulesets --input - <<'JSON'
{
  "name": "main",
  "target": "branch",
  "enforcement": "active",
  "conditions": {"ref_name": {"include": ["~DEFAULT_BRANCH"], "exclude": []}},
  "rules": [
    {"type": "deletion"},
    {"type": "non_fast_forward"},
    {"type": "required_linear_history"},
    {"type": "required_status_checks", "parameters": {
      "strict_required_status_checks_policy": true,
      "required_status_checks": [
        {"context": "Backend (lint + tests)"},
        {"context": "Frontend (check + lint + build)"},
        {"context": "Docker image"}
      ]
    }},
    {"type": "pull_request", "parameters": {
      "required_approving_review_count": 0,
      "dismiss_stale_reviews_on_push": true,
      "require_code_owner_review": false,
      "require_last_push_approval": false,
      "required_review_thread_resolution": true
    }}
  ]
}
JSON
```

Effective rules:

- no branch deletion, no force pushes
- linear history only (squash merges)
- pull requests required for every change to `main`
- the three checks from `.github/workflows/ci.yml` must pass, and the branch
  must be up to date with `main` before merging
- review conversations must be resolved
- zero required approvals: the repository has a single maintainer, and GitHub
  does not let an author approve their own pull request. Raise
  `required_approving_review_count` to 1 when a second reviewer exists.

### Actions policy

```bash
gh api -X PUT repos/Jqnx/Shoka/actions/permissions \
  -F enabled=true -f allowed_actions=selected

gh api -X PUT repos/Jqnx/Shoka/actions/permissions/selected-actions \
  -F github_owned_allowed=true \
  -F verified_allowed=true \
  -f 'patterns_allowed[]=golangci/golangci-lint-action@*'

gh api -X PUT repos/Jqnx/Shoka/actions/permissions/fork-pr-contributor-approval \
  -f approval_policy=all_external_contributors
```

Verified:

```text
allowed_actions               selected
github_owned_allowed          true
verified_allowed              true
patterns_allowed              ["golangci/golangci-lint-action@*"]
sha_pinning_required          false
fork PR approval              all_external_contributors
default workflow permissions  read (unchanged)
```

This blocks arbitrary marketplace actions while retaining everything the CI
workflow uses: `actions/checkout`, `actions/setup-go`, `actions/setup-node`
(GitHub-owned) and `golangci/golangci-lint-action` (explicit pattern).
External contributors' pull requests now require maintainer approval before
workflows run.

### Repository files

- `SECURITY.md` — private advisory reporting, scope, response expectations.
- `.github/dependabot.yml` — weekly updates for `gomod`, `github-actions`, and
  `docker`. The frontend is intentionally excluded: `package-lock.json` and
  `bun.lock` must move together, and Dependabot's npm ecosystem updates only
  the former, which would break the Docker build's
  `bun install --frozen-lockfile` step.

## Deferred

| Item | Why not applied |
| --- | --- |
| Secret scanning validity checks | Requested; the API returns the setting as `disabled` for this repository. Enable in **Settings → Code security** if the plan offers it. |
| Secret scanning non-provider patterns | Same as above. |
| SHA-pinned actions (`sha_pinning_required`) | The workflow uses major tags (`@v7`, `@v9`). Enabling this before rewriting every `uses:` to a full SHA would fail all jobs. Pair the rewrite with Dependabot's `github-actions` updater. |
| CodeQL / code scanning | Go autobuild fails on the `govips` CGO dependency (needs `libvips-dev` and `-tags sqlite_fts5`). A custom workflow or a JS/TS-only default setup is a separate decision. |
| `v2` branch deletion | Fully merged into `main` (0 ahead, 3 behind), so it is safe to tag as `v2-final` and delete. Left intact rather than destroying a ref without an explicit request. |
| `CONTRIBUTING.md`, issue/PR templates, `CODEOWNERS` | Useful for contributors; not security settings. |
| Signed commits / DCO | High friction for a single-maintainer repository; no contributor identity problem to solve yet. |

## Verification

```bash
gh api repos/Jqnx/Shoka --jq '.security_and_analysis'
gh api repos/Jqnx/Shoka/private-vulnerability-reporting
gh api repos/Jqnx/Shoka/actions/permissions
gh api repos/Jqnx/Shoka/actions/permissions/selected-actions
gh api repos/Jqnx/Shoka/actions/permissions/fork-pr-contributor-approval
gh api repos/Jqnx/Shoka/rulesets
gh api repos/Jqnx/Shoka/rulesets/26892894
```

The ruleset is only satisfiable because `.github/workflows/ci.yml` exports the
three required check names; renaming a job in that workflow requires updating
the ruleset's `required_status_checks` contexts in the same change, or `main`
becomes unmergeable.
