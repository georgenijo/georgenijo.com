// Cloudflare Worker: markdown content negotiation at the edge.
//
// georgenijo.com is fronted by Cloudflare in front of a static GitHub Pages
// origin, which cannot negotiate on Accept. Deploying this worker on the
// zone's route (georgenijo.com/*) implements the acceptmarkdown.com
// protocol without touching the origin:
//
//   - Accept: text/markdown  -> serves the .md sibling from the same path,
//     Content-Type: text/markdown; charset=utf-8
//   - Vary: Accept, Accept-Encoding on every response
//   - unsupported types (per q-values) -> 406
//
// Deployment requires Cloudflare dashboard/API access (not in this repo).

const MARKDOWN_PATHS = [
  "/",
  "/index",
  "/about",
  "/projects",
  "/now",
  "/burn",
  "/contact",
];

function acceptsMarkdown(request) {
  const header = request.headers.get("Accept") || "";
  const entries = header.split(",").map((part) => {
    const [type, ...params] = part.trim().split(";");
    let q = 1;
    for (const param of params) {
      const [key, value] = param.trim().split("=");
      if (key === "q") {
        const parsed = parseFloat(value);
        if (!Number.isNaN(parsed)) q = parsed;
      }
    }
    return { type: type.trim().toLowerCase(), q };
  });

  const markdown = entries.find((e) => e.type === "text/markdown");
  const html = entries.find((e) => e.type === "text/html");
  const wildcard = entries.find((e) => e.type === "*/*");

  const markdownQ = markdown ? markdown.q : null;
  const fallbackQ = Math.max(
    html ? html.q : 0,
    wildcard ? wildcard.q : 0
  );

  // No explicit markdown mention: only serve it if nothing else matches.
  if (markdownQ === null) return fallbackQ === 0 ? "406" : "html";
  // Explicitly rejected markdown.
  if (markdownQ === 0) return fallbackQ > 0 ? "html" : "406";
  // Prefer whichever type has the higher q-value.
  return markdownQ >= fallbackQ ? "markdown" : "html";
}

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const vary = { Vary: "Accept, Accept-Encoding" };
    const decision = acceptsMarkdown(request);

    if (decision === "406") {
      return new Response("Not Acceptable: no supported representation\n", {
        status: 406,
        headers: { "Content-Type": "text/plain; charset=utf-8", ...vary },
      });
    }

    if (decision === "markdown" && MARKDOWN_PATHS.includes(url.pathname)) {
      const mdUrl = new URL(request.url);
      mdUrl.pathname = url.pathname === "/" ? "/index.md" : url.pathname + ".md";
      const mdResponse = await fetch(mdUrl.toString(), { cf: { cacheTtl: 3600 } });
      if (mdResponse.ok) {
        return new Response(mdResponse.body, {
          status: 200,
          headers: {
            "Content-Type": "text/markdown; charset=utf-8",
            ...vary,
          },
        });
      }
      // No markdown variant: fall through to HTML rather than failing.
    }

    const response = await fetch(request);
    const headers = new Headers(response.headers);
    headers.set("Vary", "Accept, Accept-Encoding");
    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers,
    });
  },
};
