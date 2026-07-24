"""Static release checks for workflows; requires no credentials or network."""

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
REFRESH = (ROOT / ".github/workflows/refresh-lab.yml").read_text()
PAGES = (ROOT / ".github/workflows/pages.yml").read_text()


def require(text, fragment, context):
    if fragment not in text:
        raise AssertionError(f"{context}: missing {fragment!r}")


require(REFRESH, "types: [george-lab-project-updated]", "dispatch event")
require(REFRESH, "workflow_dispatch:", "manual event")
require(REFRESH, 'EVENT_NAME: ${{ github.event_name }}', "event mapping")
require(
    REFRESH,
    'EVENT_PROJECT: ${{ github.event.client_payload.project || inputs.project }}',
    "project mapping",
)
require(REFRESH, 'if [[ "$EVENT_NAME" == "repository_dispatch" && -z "$EVENT_PROJECT" ]]', "dispatch validation")
require(REFRESH, "GH_TOKEN: ${{ secrets.LAB_GITHUB_TOKEN }}", "central secret")
require(REFRESH, "permissions:\n  contents: write", "refresh permissions")
require(REFRESH, "git diff --quiet -- data/lab-projects.json", "commit scope")
require(PAGES, "permissions:\n  contents: read\n  pages: write\n  id-token: write", "Pages permissions")
require(PAGES, "uses: actions/configure-pages@v5", "Pages configuration")
require(PAGES, "uses: actions/upload-pages-artifact@v3", "Pages artifact")
require(PAGES, "uses: actions/deploy-pages@v4", "Pages deployment")
require(PAGES, "lab-health.js _site/", "Lab browser contract artifact")
require(PAGES, "cp -R data _site/", "Lab snapshots artifact")

print("release workflow contract tests passed")
