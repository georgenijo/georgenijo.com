#!/usr/bin/env python3
"""
collect-burn.py — build public-safe burn.json from ccusage + git log

Sources:
  - ccusage: npx ccusage@latest daily -j  +  monthly -j  (all agents)
  - git log: --since 45 days in ~/Documents/code/*

Output: burn.json (minimal, public-safe):
  {
    totalTokens: int,
    byModelTop: [{short, tokens}],
    last30: [{date, tokens, commits, topRepos, topMsgs}]
  }

Also generates usage-fleet.json (private, full) for fleet aggregates.

Run: python3 scripts/collect-burn.py
      python3 scripts/collect-burn.py --fleet   # also fleet exec --all (slow, ~30s)
"""
import json, subprocess, sys, os
from pathlib import Path
from collections import Counter, defaultdict
from datetime import datetime, timedelta
import argparse

CODE_ROOT = Path.home() / "Documents" / "code"
CUTOFF_DAYS = 45

def run(cmd, cwd=None, timeout=20):
    try:
        return subprocess.check_output(cmd, cwd=cwd, text=True, stderr=subprocess.DEVNULL, timeout=timeout, shell=isinstance(cmd, str))
    except Exception:
        return ""

def collect_local_ccusage():
    """Run ccusage daily -j on local host"""
    out = run(["npx", "--yes", "ccusage@latest", "daily", "-j"], timeout=30)
    try:
        return json.loads(out) if out else {}
    except:
        return {}

def collect_commits_local():
    commits_by_date = defaultdict(list)
    repos = [p for p in CODE_ROOT.iterdir() if p.is_dir() and (p / ".git").exists()] if CODE_ROOT.exists() else []
    for repo in repos:
        out = run(
            ["git", "log", f"--since={CUTOFF_DAYS} days ago", "--pretty=format:%cd|%h|%s", "--date=short", "--all"],
            cwd=repo, timeout=8
        )
        if not out:
            continue
        for line in out.splitlines():
            if not line.strip(): continue
            try:
                date, h, msg = line.split("|", 2)
                commits_by_date[date].append({"repo": repo.name, "hash": h, "msg": msg[:80]})
            except ValueError:
                continue
    return commits_by_date

def build_burn_from_local(code_host_commits=None):
    if code_host_commits is None:
        code_host_commits = collect_commits_local()
    cc = collect_local_ccusage()
    daily = cc.get("daily", [])
    totals = cc.get("totals", {})
    total_tokens = totals.get("totalTokens", sum(d.get("totalTokens",0) for d in daily))

    # Aggregate tokens per model from breakdowns (all time)
    per_model_tokens = Counter()
    for d in daily:
        for mb in d.get("modelBreakdowns", []):
            m = mb.get("modelName","")
            # ccusage modelBreakdown does not have totalTokens per model in some versions, sum components
            toks = mb.get("totalTokens")
            if toks is None:
                toks = mb.get("inputTokens",0)+mb.get("outputTokens",0)+mb.get("cacheCreationTokens",0)+mb.get("cacheReadTokens",0)
            per_model_tokens[m]+=toks

    # Fallback to monthly for totals if daily empty
    if not total_tokens:
        out_m = run(["npx","--yes","ccusage@latest","monthly","-j"], timeout=30)
        try:
            mj = json.loads(out_m)
            total_tokens = mj.get("totals",{}).get("totalTokens",0)
            for mon in mj.get("monthly",[]):
                for mb in mon.get("modelBreakdowns",[]):
                    m=mb.get("modelName","")
                    toks=mb.get("totalTokens") or (mb.get("inputTokens",0)+mb.get("outputTokens",0)+mb.get("cacheCreationTokens",0)+mb.get("cacheReadTokens",0))
                    per_model_tokens[m]+=toks
        except Exception:
            pass

    # top 5 by tokens
    top5 = per_model_tokens.most_common(5)
    def short_model(m):
        return m.replace("claude-","").replace("gpt-","").replace("-20251001","").replace("-20250929","")
    by_model_top = [{"short": short_model(k), "model": k, "tokens": int(v)} for k,v in top5]

    # last 30 days
    # daily may be sorted by date ascending already? sort to be safe
    daily_sorted = sorted(daily, key=lambda x: x.get("date","") or x.get("period",""))
    last30_raw = daily_sorted[-30:] if len(daily_sorted)>=30 else daily_sorted

    last30=[]
    for d in last30_raw:
        date = d.get("date") or d.get("period","")
        toks = d.get("totalTokens",0)
        commits = code_host_commits.get(date, [])
        c_counter = Counter(c["repo"] for c in commits)
        top_repos = [f"{repo} ({cnt})" for repo,cnt in c_counter.most_common(3)]
        top_msgs = [c["msg"][:70] for c in commits[:3]]
        # shorten common msgs
        top_msgs = [m.replace("Merge pull request","Merge PR").replace("Merge branch","Merge") for m in top_msgs]
        last30.append({
            "date": date,
            "tokens": int(toks),
            "commits": len(commits),
            "topRepos": top_repos,
            "topMsgs": top_msgs
        })

    # if we have fewer than 30 entries (new host), pad with zeros? Keep as is.
    return {
        "generatedAt": datetime.now().isoformat(),
        "totalTokens": int(total_tokens),
        "byModelTop": by_model_top,
        "last30": last30,
    }

def collect_fleet_and_merge():
    """Attempt fleet exec --all ccusage daily -j + monthly -j, aggregate"""
    # try fleet command
    fleet_cmd = 'bash -lc "export PATH=$HOME/.local/bin:$HOME/.nvm/versions/node/v20.20.2/bin:/opt/homebrew/bin:/usr/local/bin:$PATH; [ -f $HOME/.nvm/nvm.sh ] && . $HOME/.nvm/nvm.sh 2>/dev/null; npx --yes ccusage@latest daily -j 2>&1"'
    out = run(["fleet", "exec", "--all", "--json", fleet_cmd], timeout=120)
    if not out:
        print("fleet exec not available or timed out, skipping fleet aggregate", file=sys.stderr)
        return None
    try:
        nodes = json.loads(out)
    except Exception as e:
        print(f"fleet parse failed: {e}\n{out[:1000]}", file=sys.stderr)
        return None

    # aggregate daily across nodes, summing tokens per date
    by_day = {}
    per_model_tokens = Counter()
    total_tokens=0
    for node in nodes:
        if not node.get("ok"): continue
        try:
            j = json.loads(node.get("stdout",""))
        except:
            continue
        for d in j.get("daily",[]):
            date = d.get("date") or d.get("period","")
            if not date: continue
            if date not in by_day:
                by_day[date]=defaultdict(int)
                by_day[date]["_totalTokens"]=0
                by_day[date]["_models"]={}
            by_day[date]["_totalTokens"]+=int(d.get("totalTokens",0))
            total_tokens+=int(d.get("totalTokens",0))
            for mb in d.get("modelBreakdowns",[]):
                mn=mb.get("modelName","")
                toks=mb.get("totalTokens") or (mb.get("inputTokens",0)+mb.get("outputTokens",0)+mb.get("cacheCreationTokens",0)+mb.get("cacheReadTokens",0))
                per_model_tokens[mn]+=toks
                by_day[date]["_models"][mn]=by_day[date]["_models"].get(mn,0)+toks

    # build similar to local
    sorted_days = sorted(by_day.items(), key=lambda x: x[0])
    def short_model(m): return m.replace("claude-","").replace("gpt-","").replace("-20251001","")

    # also collect commits only from ubuntu (this host) for annotations — best effort
    commits = collect_commits_local()

    last30=[]
    for date, info in sorted_days[-30:]:
        toks=info["_totalTokens"]
        comms=commits.get(date,[])
        cnt=Counter(c["repo"] for c in comms)
        top_repos=[f"{r} ({c})" for r,c in cnt.most_common(3)]
        top_msgs=[c["msg"][:70] for c in comms[:3]]
        top_msgs=[m.replace("Merge pull request","Merge PR").replace("Merge branch","Merge") for m in top_msgs]
        last30.append({"date":date,"tokens":int(toks),"commits":len(comms),"topRepos":top_repos,"topMsgs":top_msgs})

    by_model_top=[{"short":short_model(k),"model":k,"tokens":int(v)} for k,v in per_model_tokens.most_common(5)]

    return {
        "generatedAt": datetime.now().isoformat(),
        "totalTokens": int(total_tokens),
        "byModelTop": by_model_top,
        "last30": last30,
        "_perNodeCheck": {n["node"]: ("ok" if n.get("ok") else n.get("error","fail")) for n in nodes}
    }

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument("--fleet", action="store_true", help="also aggregate fleet-wide (fleet exec --all)")
    parser.add_argument("--out", default="burn.json", help="output file (public safe)")
    parser.add_argument("--full", action="store_true", help="also write full usage-fleet.json")
    args=parser.parse_args()

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
    print(f"Wrote {out_path} — {len(burn.get('last30',[]))} days, {burn.get('totalTokens',0)/1e9:.2f}B tokens")

    # For TUI embed: also write a compact Go-embed friendly version
    # burn.tui.json with same data, but also a Go file if requested
    tui_json = out_dir / "ssh-tui" / "burn.json"
    try:
        with open(tui_json, "w") as f:
            json.dump(burn, f)
        print(f"Wrote {tui_json}")
    except Exception as e:
        print(f"Could not write TUI burn.json: {e}", file=sys.stderr)

if __name__=="__main__":
    main()
