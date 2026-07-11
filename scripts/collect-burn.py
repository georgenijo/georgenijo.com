#!/usr/bin/env python3
"""
collect-burn.py — build public-safe burn.json from ccusage + git log

Sources:
  - ccusage: npx ccusage@latest daily -j  (+ monthly fallback)
  - git log: --since 45 days in ~/Documents/code/* (public repos only)

Output: burn.json (minimal, public-safe):
  {
    totalTokens: int,          # all-time (ccusage totals), same for local + fleet
    byModelTop: [{short, tokens}],
    last30: [{date, tokens, commits, topRepos, topMsgs}]  # calendar 30d
  }

Run: python3 scripts/collect-burn.py
      python3 scripts/collect-burn.py --fleet   # also fleet exec --all (slow, ~30s)
"""
import json, subprocess, sys, os
from pathlib import Path
from collections import Counter, defaultdict
from datetime import datetime, timedelta, date
import argparse

CODE_ROOT = Path.home() / "Documents" / "code"
CUTOFF_DAYS = 45
CALENDAR_DAYS = 30

# Only annotate commits from repos already listed on the public site (+ site itself).
PUBLIC_REPOS = frozenset({
    "agent-mesh", "usher", "hangar", "ghosthands", "murmur-app",
    "fleetmap", "gauge", "whoop-dashboard", "aperture",
    "georgenijo.com", "agentos",
})


def run(cmd, cwd=None, timeout=20):
    try:
        return subprocess.check_output(
            cmd, cwd=cwd, text=True, stderr=subprocess.DEVNULL,
            timeout=timeout, shell=isinstance(cmd, str),
        )
    except Exception:
        return ""


def parse_json_loose(s):
    """Parse JSON, tolerating leading/trailing non-JSON noise."""
    if not s:
        raise ValueError("empty")
    s = s.strip()
    try:
        return json.loads(s)
    except json.JSONDecodeError:
        i, j = s.find("{"), s.rfind("}")
        if i >= 0 and j > i:
            return json.loads(s[i : j + 1])
        raise


def short_model(m):
    return (
        m.replace("claude-", "")
        .replace("gpt-", "")
        .replace("-20251001", "")
        .replace("-20250929", "")
    )


def model_tokens(mb):
    toks = mb.get("totalTokens")
    if toks is None:
        toks = (
            mb.get("inputTokens", 0)
            + mb.get("outputTokens", 0)
            + mb.get("cacheCreationTokens", 0)
            + mb.get("cacheReadTokens", 0)
        )
    return int(toks or 0)


def collect_local_ccusage():
    """Run ccusage daily -j on local host"""
    out = run(["npx", "--yes", "ccusage@latest", "daily", "-j"], timeout=30)
    try:
        return parse_json_loose(out) if out else {}
    except Exception:
        return {}


def collect_commits_local():
    """Commits keyed by date, newest-first within each day (by unix time)."""
    commits_by_date = defaultdict(list)
    if not CODE_ROOT.exists():
        return commits_by_date
    repos = [
        p for p in CODE_ROOT.iterdir()
        if p.is_dir() and (p / ".git").exists() and p.name in PUBLIC_REPOS
    ]
    for repo in repos:
        out = run(
            [
                "git", "log", f"--since={CUTOFF_DAYS} days ago",
                "--pretty=format:%cd|%ct|%h|%s", "--date=short", "--all",
            ],
            cwd=repo, timeout=8,
        )
        if not out:
            continue
        for line in out.splitlines():
            if not line.strip():
                continue
            try:
                d, ts, h, msg = line.split("|", 3)
                commits_by_date[d].append({
                    "repo": repo.name,
                    "hash": h,
                    "msg": msg[:80],
                    "ts": int(ts),
                })
            except ValueError:
                continue
    for d in commits_by_date:
        commits_by_date[d].sort(key=lambda c: c["ts"], reverse=True)
    return commits_by_date


def day_annotation(commits):
    c_counter = Counter(c["repo"] for c in commits)
    top_repos = [f"{repo} ({cnt})" for repo, cnt in c_counter.most_common(3)]
    top_msgs = [c["msg"][:70] for c in commits[:3]]
    top_msgs = [
        m.replace("Merge pull request", "Merge PR").replace("Merge branch", "Merge")
        for m in top_msgs
    ]
    return len(commits), top_repos, top_msgs


def calendar_last30(tokens_by_date, commits_by_date, end=None):
    """Build exactly CALENDAR_DAYS entries ending at end (default: today)."""
    if end is None:
        end = date.today()
    last30 = []
    for i in range(CALENDAR_DAYS - 1, -1, -1):
        d = end - timedelta(days=i)
        date_str = d.isoformat()
        commits = commits_by_date.get(date_str, [])
        n, top_repos, top_msgs = day_annotation(commits)
        last30.append({
            "date": date_str,
            "tokens": int(tokens_by_date.get(date_str, 0)),
            "commits": n,
            "topRepos": top_repos,
            "topMsgs": top_msgs,
        })
    return last30


def accumulate_models(daily, per_model_tokens):
    for d in daily:
        for mb in d.get("modelBreakdowns", []):
            per_model_tokens[mb.get("modelName", "")] += model_tokens(mb)


def build_burn_from_local(code_host_commits=None):
    if code_host_commits is None:
        code_host_commits = collect_commits_local()
    cc = collect_local_ccusage()
    daily = cc.get("daily", [])
    totals = cc.get("totals", {})
    total_tokens = int(totals.get("totalTokens") or 0)
    if not total_tokens:
        total_tokens = sum(int(d.get("totalTokens", 0) or 0) for d in daily)

    per_model_tokens = Counter()
    accumulate_models(daily, per_model_tokens)

    if not total_tokens:
        out_m = run(["npx", "--yes", "ccusage@latest", "monthly", "-j"], timeout=30)
        try:
            mj = parse_json_loose(out_m)
            total_tokens = int(mj.get("totals", {}).get("totalTokens", 0) or 0)
            for mon in mj.get("monthly", []):
                for mb in mon.get("modelBreakdowns", []):
                    per_model_tokens[mb.get("modelName", "")] += model_tokens(mb)
        except Exception:
            pass

    tokens_by_date = {}
    for d in daily:
        date_str = d.get("date") or d.get("period", "")
        if date_str:
            tokens_by_date[date_str] = int(d.get("totalTokens", 0) or 0)

    by_model_top = [
        {"short": short_model(k), "model": k, "tokens": int(v)}
        for k, v in per_model_tokens.most_common(5)
    ]

    return {
        "generatedAt": datetime.now().isoformat(),
        "totalTokens": int(total_tokens),
        "byModelTop": by_model_top,
        "last30": calendar_last30(tokens_by_date, code_host_commits),
    }


def collect_fleet_and_merge():
    """fleet exec --all ccusage daily -j, aggregate all-time totals + calendar last30."""
    # No 2>&1 — npm noise must not pollute JSON stdout.
    # No hardcoded nvm version — source nvm.sh if present, else rely on PATH.
    fleet_cmd = (
        'bash -lc \''
        'export PATH="$HOME/.local/bin:/opt/homebrew/bin:/usr/local/bin:$PATH"; '
        '[ -s "$HOME/.nvm/nvm.sh" ] && . "$HOME/.nvm/nvm.sh" >/dev/null 2>&1; '
        'npx --yes ccusage@latest daily -j'
        '\''
    )
    out = run(["fleet", "exec", "--all", "--json", fleet_cmd], timeout=120)
    if not out:
        print("fleet exec not available or timed out, skipping fleet aggregate", file=sys.stderr)
        return None
    try:
        nodes = parse_json_loose(out)
    except Exception as e:
        print(f"fleet parse failed: {e}\n{out[:1000]}", file=sys.stderr)
        return None

    tokens_by_date = defaultdict(int)
    per_model_tokens = Counter()
    total_tokens = 0
    nodes_ok = 0

    for node in nodes:
        if not node.get("ok"):
            continue
        try:
            j = parse_json_loose(node.get("stdout", ""))
        except Exception:
            print(
                f"  fleet node: {node.get('node', '?')} → bad JSON (dropped)",
                file=sys.stderr,
            )
            continue
        nodes_ok += 1
        # Prefer ccusage all-time totals so fleet matches local semantics.
        node_total = int((j.get("totals") or {}).get("totalTokens") or 0)
        daily = j.get("daily", [])
        if not node_total:
            node_total = sum(int(d.get("totalTokens", 0) or 0) for d in daily)
        total_tokens += node_total
        accumulate_models(daily, per_model_tokens)
        for d in daily:
            date_str = d.get("date") or d.get("period", "")
            if date_str:
                tokens_by_date[date_str] += int(d.get("totalTokens", 0) or 0)

    if nodes_ok == 0:
        print("fleet: no nodes returned parseable ccusage JSON", file=sys.stderr)
        return None

    commits = collect_commits_local()
    by_model_top = [
        {"short": short_model(k), "model": k, "tokens": int(v)}
        for k, v in per_model_tokens.most_common(5)
    ]

    for n in nodes:
        status = "ok" if n.get("ok") else n.get("error", "fail")
        print(f"  fleet node: {n.get('node', '?')} → {status}", file=sys.stderr)

    return {
        "generatedAt": datetime.now().isoformat(),
        "totalTokens": int(total_tokens),
        "byModelTop": by_model_top,
        "last30": calendar_last30(tokens_by_date, commits),
    }


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--fleet", action="store_true", help="also aggregate fleet-wide (fleet exec --all)")
    parser.add_argument("--out", default="burn.json", help="output file (public safe)")
    args = parser.parse_args()

    out_dir = Path(__file__).parent.parent
    os.chdir(out_dir)

    commits = collect_commits_local()

    if args.fleet:
        print("Collecting fleet-wide usage (this may take ~30s)...", file=sys.stderr)
        merged = collect_fleet_and_merge()
        if merged:
            burn = merged
        else:
            print("Fleet merge failed, falling back to local", file=sys.stderr)
            burn = build_burn_from_local(commits)
    else:
        burn = build_burn_from_local(commits)

    out_path = out_dir / args.out
    with open(out_path, "w") as f:
        json.dump(burn, f, indent=2)
        f.write("\n")
    print(f"Wrote {out_path} — {len(burn.get('last30', []))} days, {burn.get('totalTokens', 0)/1e9:.2f}B tokens")

    tui_json = out_dir / "ssh-tui" / "burn.json"
    try:
        with open(tui_json, "w") as f:
            json.dump(burn, f)
            f.write("\n")
        print(f"Wrote {tui_json}")
    except Exception as e:
        print(f"Could not write TUI burn.json: {e}", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
