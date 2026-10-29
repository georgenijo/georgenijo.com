# George Lab update workflow

George Lab is a static project ledger that changes only when George updates it.
Nothing publishes automatically: there is no schedule, no `repository_dispatch`
trigger, and no machine pushing snapshots. Every change reaches the site through
a commit to `main` (which runs **Site checks** and **Deploy Pages**) or a manual
run of **Refresh George Lab**.

Earlier versions accepted `george-lab-project-updated` dispatch events from
satellite repositories and received a Fleet health snapshot from the Mac mini
every ten minutes. Both were retired on 2026-10-29. Any leftover
`LAB_DISPATCH_TOKEN` secrets in satellite repositories are unused and can be
deleted.

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

## Manual refresh

In `georgenijo/georgenijo.com`, open Actions → Refresh George Lab → Run workflow.
The project input is optional. Validate the input contract locally:

```sh
python3 -m unittest tests/test_refresh_lab.py
python3 scripts/refresh-lab.py --project not-a-project --output /tmp/lab.json
```

The test is network-free and uses no token. The second command must fail with an
unknown-project error. A live local refresh is optional and must use a token supplied through
the process environment, never a checked-in file.

## Fleet health snapshot

`data/fleet-health.json` is a point-in-time snapshot using Fleet's
privacy-safe schema version 1. The browser copies only the documented public
fields, shows the snapshot with its collection date, and shows nothing for
missing, malformed, or unsupported data. It never reports a snapshot as stale.

To update it by hand, run Fleet's collector and commit the file:

```sh
fleet health --output data/fleet-health.json
git add data/fleet-health.json
git commit -m "chore(lab): update Fleet health snapshot"
```

Do not publish the Fleet manifest, logs, SSH configuration, Tailscale state,
raw probe output, or machine addresses. See the Fleet repository's
`docs/george-lab-health.md` for the collector and schema.

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
accepted/rejected project ids, manual refresh without a project id, the
manual-only workflow trigger, secret references, least-privilege workflow
permissions, Pages artifact contents, and JSON syntax.
