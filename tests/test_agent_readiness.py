"""Agent-readiness contract tests for georgenijo.com.

Static checks only: no network, no credentials. Verifies the raw-HTML
content, crawler allowlist, agent 404 recovery body, llms.txt/sitemap
consistency, and the acceptmarkdown.com negotiation artifacts.
"""

import re
import xml.etree.ElementTree as ET
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]

AGENT_UAS = [
    "GPTBot",
    "ChatGPT-User",
    "OAI-SearchBot",
    "ClaudeBot",
    "Claude-User",
    "Claude-SearchBot",
    "Google-Extended",
    "GoogleOther",
    "PerplexityBot",
    "Perplexity-User",
    "DeepSeekBot",
    "ora-agent",
    "Applebot",
    "Amazonbot",
    "CCBot",
    "meta-externalagent",
]


def visible_text(html):
    """Strip script/style blocks and tags, collapse whitespace."""
    body = re.sub(r"<(script|style)\b.*?</\1>", "", html, flags=re.S)
    text = re.sub(r"<[^>]+>", " ", body)
    return re.sub(r"\s+", " ", text).strip()


def test_homepage_has_static_nojs_content():
    html = (ROOT / "index.html").read_text()
    assert "<h1" in html, "homepage must have an H1 in raw HTML"
    # The static section must exist in raw markup...
    assert 'id="static-content"' in html
    static_start = html.index('id="static-content"')
    static_section = html[static_start : html.index("</main>", static_start)]
    # ...with 500+ chars of real text...
    assert len(visible_text(static_section)) >= 500
    # ...that JS removes on boot so the visual design is unchanged.
    assert 'getElementById("static-content")' in html
    assert "staticContent.remove()" in html
    # App must be visible without JS (no `hidden` in raw markup).
    assert '<div class="app" id="app">' in html


def test_robots_txt_allowlists_agent_user_agents():
    robots = (ROOT / "robots.txt").read_text()
    for ua in AGENT_UAS:
        block = re.search(
            rf"(?is)user-agent:\s*{re.escape(ua)}\s*\n((?:allow|disallow):.*\n?)+",
            robots,
        )
        assert block, f"robots.txt missing explicit block for {ua}"
        assert re.search(rf"(?i)^allow:\s*/\s*$", block.group(0), re.M), (
            f"{ua} must be allowed site-wide"
        )
    assert "Sitemap:" in robots


def test_llms_txt_format_and_links_resolve():
    llms = (ROOT / "llms.txt").read_text()
    lines = llms.splitlines()
    assert lines[0].startswith("# "), "llms.txt must start with an H1"
    assert any(line.startswith("> ") for line in lines), (
        "llms.txt must include a blockquote summary"
    )
    links = re.findall(r"\]\((https://georgenijo\.com/[^)]+)\)", llms)
    assert links, "llms.txt must link to site resources"
    for url in links:
        path = url.replace("https://georgenijo.com/", "")
        if path.endswith(".md"):
            assert (ROOT / path).is_file(), f"llms.txt links to missing {path}"


def test_markdown_variants_exist_and_are_substantive():
    for name in ["index.md", "about.md", "projects.md", "now.md", "burn.md", "contact.md"]:
        md = ROOT / name
        assert md.is_file(), f"missing markdown variant {name}"
        content = md.read_text()
        assert content.startswith("# "), f"{name} must start with an H1"
        assert len(content) >= 300, f"{name} too short to be useful"


def test_sitemap_xml_is_valid_and_urls_exist():
    root = ET.parse(ROOT / "sitemap.xml").getroot()
    ns = {"s": "http://www.sitemaps.org/schemas/sitemap/0.9"}
    locs = [el.text for el in root.findall(".//s:url/s:loc", ns)]
    assert locs, "sitemap.xml has no URLs"
    for loc in locs:
        assert loc.startswith("https://georgenijo.com/")
        path = loc.replace("https://georgenijo.com/", "") or "index.html"
        if not path.startswith("http"):
            assert (ROOT / path).is_file(), f"sitemap lists missing file {path}"


def test_404_has_agent_recovery_body():
    html = (ROOT / "404.html").read_text()
    for fragment in [
        "/llms.txt",
        "/sitemap.xml",
        "/index.md",
        "/projects.md",
        "/contact.md",
    ]:
        assert f'href="{fragment}"' in html, f"404.html missing recovery link {fragment}"
    assert len(visible_text(html)) >= 200


def test_pages_workflow_deploys_agent_files():
    workflow = (ROOT / ".github/workflows/pages.yml").read_text()
    for fragment in ["robots.txt", "llms.txt", "sitemap.xml", "*.md"]:
        assert fragment in workflow, f"pages.yml does not deploy {fragment}"
    # Preserve the existing release-contract fragments.
    assert "lab-health.js _site/" in workflow
    assert "cp -R data _site/" in workflow


def test_nginx_negotiation_config_follows_protocol():
    conf = (ROOT / "deploy/nginx-markdown-negotiation.conf").read_text()
    assert re.search(r"map\s+\$http_accept\s+\$preferred_ext", conf)
    assert '"~*text/markdown"' in conf
    assert "add_header Vary Accept always;" in conf
    assert "try_files $uri$preferred_ext $uri/index$preferred_ext $uri.html $uri/index.html =404;" in conf
    assert "default_type text/markdown;" in conf
    assert "charset utf-8;" in conf


def test_cloudflare_worker_follows_protocol():
    worker = (ROOT / "deploy/cloudflare-worker-markdown.js").read_text()
    assert '"text/markdown; charset=utf-8"' in worker
    assert '"Vary", "Accept, Accept-Encoding"' in worker or \
           '"Vary": "Accept, Accept-Encoding"' in worker
    assert "406" in worker


if __name__ == "__main__":
    import sys

    failures = 0
    for name, fn in sorted(globals().items()):
        if name.startswith("test_") and callable(fn):
            try:
                fn()
                print(f"PASS {name}")
            except AssertionError as exc:
                failures += 1
                print(f"FAIL {name}: {exc}", file=sys.stderr)
    if failures:
        sys.exit(f"{failures} agent-readiness test(s) failed")
    print("agent-readiness contract tests passed")
