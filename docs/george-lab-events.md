# George Lab event workflow

George Lab is a static, event-updated project ledger. Nothing listens continuously:
satellite repositories emit a GitHub `repository_dispatch` event after meaningful
activity, and the central site repository refreshes all five projects from GitHub.

## Central repository setup

Create a fine-grained personal access token with this exact configuration:

- **Resource owner:** `georgenijo`.
- **Repository access:** only `georgenijo.com`, `murmur-app`, `agentos`,
  `agent-mesh`, `fleet`, and `ghosthands`.
- **Repository permissions:** **Contents: Read-only**, **Issues: Read-only**,
  and **Pull requests: Read-only**. GitHub includes **Metadata: Read-only**
  automatically. No account permissions are needed.
- **Expiration:** choose the shortest practical lifetime and record its owner
  and renewal date outside the repository.

Add the token to the central repository at Settings → Secrets and variables →
Actions → New repository secret, with the exact secret name
`LAB_GITHUB_TOKEN`. Paste the value only into GitHub's secret-value field.
Do not put it in a file, command example, issue, pull request, or Actions
variable.

The refresh workflow uses this token only to read project metadata, releases,
issues, and pull requests. Its built-in `GITHUB_TOKEN`—scoped to
`contents: write` and `actions: write`—performs the snapshot commit and starts
the existing Pages workflow when the snapshot changes. The personal token does
not need write access.

## Satellite event contract

Send this request to the central repository:

```http
POST /repos/georgenijo/georgenijo.com/dispatches
Authorization: Bearer <token>
Accept: application/vnd.github+json
X-GitHub-Api-Version: 2022-11-28
Content-Type: application/json

{
  "event_type": "george-lab-project-updated",
  "client_payload": {
    "project": "murmur-app"
  }
}
```

`client_payload.project` is required and must be exactly one of `murmur-app`,
`agentos`, `agent-mesh`, `fleet`, or `ghosthands`. The event identifies the
source and validates the contract; every run refreshes all projects so the
checked-in snapshot remains internally consistent.

Create a separate fine-grained personal access token for dispatch with:

- **Resource owner:** `georgenijo`.
- **Repository access:** only `georgenijo/georgenijo.com`.
- **Repository permissions:** **Contents: Read and write** (required by
  GitHub's repository-dispatch endpoint). GitHub includes **Metadata:
  Read-only** automatically. No other repository or account permission is
  needed.
- **Expiration:** choose the shortest practical lifetime and track renewal
  outside the repository.

In each satellite that will emit an event, open Settings → Secrets and
variables → Actions → New repository secret and store the token with the exact
name `LAB_DISPATCH_TOKEN`. The same narrowly scoped token may be stored in
multiple satellites; use separate per-satellite tokens instead if independent
revocation or audit identity is worth the extra rotation work.

Example satellite workflow step:

```yaml
- name: Notify George Lab
  env:
    GH_TOKEN: ${{ secrets.LAB_DISPATCH_TOKEN }}
  run: |
    gh api --method POST repos/georgenijo/georgenijo.com/dispatches \
      -f event_type=george-lab-project-updated \
      -F 'client_payload[project]=murmur-app'
```

Trigger this after a release, a merge to the default branch, or another event
that should be reflected on the site. Avoid issue and pull-request event triggers
unless near-real-time counts justify the extra workflow traffic.

## Manual refresh

In `georgenijo/georgenijo.com`, open Actions → Refresh George Lab → Run workflow.
The project input is optional for manual runs. Before credentials are added or
anything is deployed, validate the same input contract locally:

```sh
python3 -m unittest tests/test_refresh_lab.py
python3 scripts/refresh-lab.py --project not-a-project --output /tmp/lab.json
```

The test is network-free and uses no token. The second command must fail with an
unknown-project error, demonstrating the rejection path used by dispatch
events. A live local refresh is optional and must use a token supplied through
the process environment, never a checked-in file.

## Fleet health end-to-end setup

Fleet health uses the separate, privacy-safe contract in
`data/fleet-health.json`. The browser accepts only Fleet schema version 1,
copies only the documented public fields, and treats data as unavailable when
it is missing, malformed, more than 15 minutes old, or dated more than one
minute in the future. It never falls back to old node details.

The event flow is:

1. A five-minute scheduled, short-lived job runs `fleet health --output
   /absolute/path/to/georgenijo.com/data/fleet-health.json`.
2. Fleet probes, writes the allowlisted snapshot atomically, and exits. There is
   no listener or resident George Lab agent.
3. A separate publisher notices or is invoked after the successful file write,
   commits only `data/fleet-health.json`, and pushes it to the central site
   repository.
4. The existing Pages workflow packages `data/`, `index.html`, and
   `lab-health.js`. The browser fetches both project and Fleet snapshots.

Step 3 is intentionally not installed here. Before enabling it, choose a
publishing mechanism and give it the narrowest credentials possible. Do not
publish the Fleet manifest, logs, SSH configuration, Tailscale state, raw probe
output, or machine addresses. The checked-in epoch-dated empty snapshot is a
safe placeholder and intentionally renders as stale until a real publisher is
approved.

For the collector command, JSON schema, and optional launchd example, see the
Fleet repository's `docs/george-lab-health.md`. Installing the schedule,
creating tokens, committing snapshots, pushing, and changing Pages settings all
require explicit operator action.

## Validation

Run the full release validation with:

```sh
node tests/lab-health.test.js
python3 -m unittest tests/test_refresh_lab.py
python3 tests/test_release_contract.py
python3 -m json.tool data/lab-projects.json >/dev/null
python3 -m json.tool data/fleet-health.json >/dev/null
```

These commands cover Fleet snapshot safety, deterministic metadata refresh,
accepted/rejected event project ids, manual refresh without a project id, the
workflow event mapping, secret references, least-privilege workflow
permissions, Pages artifact contents, and JSON syntax.

## PR, review, and deploy checklist

No item below authorizes a credential, repository-setting, commit, push, or
deployment change. The operator performs each external action explicitly.

### Before opening the PR

- [ ] Review the complete diff in both the website and Fleet worktrees.
- [ ] Run every command in **Validation** and the Fleet repository's test suite.
- [ ] Confirm secret scanning finds names/placeholders only, never a token value.
- [ ] Confirm `.github/workflows/pages.yml` still uses `actions/configure-pages`,
  `actions/upload-pages-artifact`, and `actions/deploy-pages`.
- [ ] Confirm the Pages workflow adds only `lab-health.js` and `data/` to the
  existing artifact and retains `contents: read`, `pages: write`, and
  `id-token: write`.

### PR review

- [ ] Verify `data/fleet-health.json` is the epoch-dated empty placeholder.
- [ ] Verify the Lab UI fails closed for stale, future, unsupported, or malformed
  Fleet data and never displays unexpected node fields.
- [ ] Verify `repository_dispatch` requires an allowlisted
  `client_payload.project`, while `workflow_dispatch` permits an empty project.
- [ ] Verify refresh commits only `data/lab-projects.json`.
- [ ] Verify branch protection and required reviews are satisfied before merge.

### Credential and release handoff

- [ ] Create and store `LAB_GITHUB_TOKEN` using the central-token instructions.
- [ ] Create and store `LAB_DISPATCH_TOKEN` only in approved satellite repos.
- [ ] Manually run **Refresh George Lab** and review its generated snapshot.
- [ ] Emit one test dispatch from an approved satellite and verify one refresh.
- [ ] Merge through the normal reviewed PR path; do not change Pages settings.
- [ ] Observe **Deploy Pages** succeed and inspect the published Lab page.
- [ ] Approve and configure Fleet snapshot publishing separately; until then,
  leave the safe placeholder in place.
