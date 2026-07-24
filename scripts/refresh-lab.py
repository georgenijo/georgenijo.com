#!/usr/bin/env python3
"""Refresh the checked-in George Lab snapshot from the GitHub REST API."""

import argparse
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
from datetime import datetime, timezone
from pathlib import Path

OWNER = "georgenijo"
OUTPUT = Path(__file__).resolve().parents[1] / "data" / "lab-projects.json"
PROJECTS = [
    {
        "repo": "murmur-app",
        "status": "active",
        "category": "macOS utility",
        "note": "Local, offline voice-to-text for macOS, built around a hold-to-talk workflow.",
    },
    {
        "repo": "agentos",
        "status": "active",
        "category": "agent platform",
        "note": "A personal and home agent operating system bringing long-running agents under one control plane.",
    },
    {
        "repo": "agent-mesh",
        "status": "active",
        "category": "agent coordination",
        "note": "Coordination infrastructure for multiple coding agents sharing tasks, files, and decisions.",
    },
    {
        "repo": "fleet",
        "status": "active",
        "category": "developer infrastructure",
        "note": "Tools for running commands and agent work across George's machines as a small fleet.",
    },
    {
        "repo": "ghosthands",
        "status": "active",
        "category": "macOS automation",
        "note": "Local computer-use for native macOS apps through the accessibility tree.",
    },
]


def request_json(path, token):
    request = urllib.request.Request(
        "https://api.github.com" + path,
        headers={
            "Accept": "application/vnd.github+json",
            "User-Agent": "georgenijo.com-lab-refresh",
            "X-GitHub-Api-Version": "2022-11-28",
            **({"Authorization": f"Bearer {token}"} if token else {}),
        },
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.load(response)


def latest_release(repo, token):
    try:
        release = request_json(f"/repos/{OWNER}/{repo}/releases/latest", token)
    except urllib.error.HTTPError as error:
        if error.code == 404:
            return None
        raise
    return {
        "tag": release["tag_name"],
        "name": release.get("name") or release["tag_name"],
        "url": release["html_url"],
        "publishedAt": release.get("published_at"),
    }


def search_count(query, token):
    encoded = urllib.parse.quote(query)
    return int(request_json(f"/search/issues?q={encoded}&per_page=1", token)["total_count"])


def unavailable_project(config):
    repo = config["repo"]
    return {
        "id": repo,
        "name": repo,
        "repository": f"{OWNER}/{repo}",
        "repositoryUrl": f"https://github.com/{OWNER}/{repo}",
        "status": config["status"],
        "category": config["category"],
        "note": config["note"],
        "latestRelease": None,
        "openIssues": None,
        "openPullRequests": None,
        "lastActivityAt": None,
        "metadataAvailable": False,
    }


def build_project(config, token):
    repo = config["repo"]
    try:
        metadata = request_json(f"/repos/{OWNER}/{repo}", token)
    except urllib.error.HTTPError as error:
        if error.code == 404 and not token:
            return unavailable_project(config)
        raise
    base_query = f"repo:{OWNER}/{repo} is:open"
    return {
        "id": repo,
        "name": repo,
        "repository": f"{OWNER}/{repo}",
        "repositoryUrl": metadata["html_url"],
        "status": config["status"],
        "category": config["category"],
        "note": config["note"],
        "latestRelease": latest_release(repo, token),
        "openIssues": search_count(f"{base_query} is:issue", token),
        "openPullRequests": search_count(f"{base_query} is:pr", token),
        "lastActivityAt": metadata["pushed_at"],
        "metadataAvailable": True,
    }


def validate_project(project):
    known = {config["repo"] for config in PROJECTS}
    if project and project not in known:
        raise ValueError(
            f"unknown project {project!r}; expected one of: {', '.join(sorted(known))}"
        )


def build_snapshot(token, generated_at=None):
    return {
        "schemaVersion": 1,
        "generatedAt": (
            generated_at
            or datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")
        ),
        "source": "github-api",
        "projects": [build_project(project, token) for project in PROJECTS],
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--project",
        help="Validate a repository_dispatch project id; the snapshot still refreshes all projects.",
    )
    parser.add_argument("--output", type=Path, default=OUTPUT)
    args = parser.parse_args()
    try:
        validate_project(args.project)
    except ValueError as error:
        parser.error(str(error))

    token = os.environ.get("GH_TOKEN") or os.environ.get("GITHUB_TOKEN")
    try:
        payload = build_snapshot(token)
    except urllib.error.HTTPError as error:
        print(f"GitHub API request failed: {error.code} {error.reason}", file=sys.stderr)
        return 1

    args.output.parent.mkdir(parents=True, exist_ok=True)
    args.output.write_text(json.dumps(payload, indent=2) + "\n")
    print(f"Wrote {args.output} with {len(payload['projects'])} projects")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
