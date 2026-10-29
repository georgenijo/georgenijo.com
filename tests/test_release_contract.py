"""Static release checks for workflows; requires no credentials or network."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
REFRESH = (ROOT / ".github/workflows/refresh-lab.yml").read_text()
PAGES = (ROOT / ".github/workflows/pages.yml").read_text()


def require(text, fragment, context):
    if fragment not in text:
        raise AssertionError(f"{context}: missing {fragment!r}")


require(REFRESH, "on:\n  workflow_dispatch:", "manual event")
if "repository_dispatch" in REFRESH or "schedule:" in REFRESH:
    raise AssertionError("refresh workflow must stay manual-only")
require(REFRESH, "EVENT_PROJECT: ${{ inputs.project }}", "project mapping")
require(REFRESH, "GH_TOKEN: ${{ secrets.LAB_GITHUB_TOKEN }}", "central secret")
require(REFRESH, "permissions:\n  actions: write\n  contents: write", "refresh permissions")
require(REFRESH, "git diff --quiet -- data/lab-projects.json", "commit scope")
require(REFRESH, "gh workflow run pages.yml --ref main", "manual-refresh Pages deploy")
require(PAGES, "permissions:\n  contents: read\n  pages: write\n  id-token: write", "Pages permissions")
require(PAGES, "uses: actions/configure-pages@v5", "Pages configuration")
require(PAGES, "uses: actions/upload-pages-artifact@v3", "Pages artifact")
require(PAGES, "uses: actions/deploy-pages@v4", "Pages deployment")
require(PAGES, "lab-health.js _site/", "Lab browser contract artifact")
require(PAGES, "cp -R data docs about contact privacy _site/", "public directories artifact")
require(PAGES, "cp index.html dossier.html terminal.html", "homepage and archived dossier")

print("release workflow contract tests passed")
