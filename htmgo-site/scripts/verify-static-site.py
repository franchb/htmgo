"""Check the published HTML for controls that need the Go server."""

import sys
from html.parser import HTMLParser
from pathlib import Path
from urllib.parse import urlsplit


class Page(HTMLParser):
    def __init__(self, path):
        super().__init__()
        self.tags = []
        self.feed(path.read_text(encoding="utf-8"))
        self.close()

    def handle_starttag(self, tag, attrs):
        self.tags.append((tag, dict(attrs)))

    def has(self, tag, **attrs):
        return any(
            name == tag and actual.items() >= attrs.items()
            for name, actual in self.tags
        )


def require(condition, message):
    if not condition:
        raise SystemExit(message)


def verify(root):
    for asset in ("public/main.css", "public/htmgo.js"):
        require((root / asset).is_file(), f"The export is missing /{asset}")

    pages = {
        path.parent.relative_to(root).as_posix(): Page(path)
        for path in root.rglob("index.html")
    }
    home = pages.get(".")
    require(home and home.has("summary", **{"aria-label": "Menu"}),
            "Missing the native mobile menu on the home page")

    clock = pages.get("docs/core-concepts/partials")
    require(clock and clock.has("time"), "Missing the browser clock preview")

    for route in ("html-to-go", "examples/form",
                  "examples/form-with-blur-validation", "examples/click-to-edit"):
        page = pages.get(route)
        require(page and page.has("a", href="https://htmgo.dev/" + route),
                f"/{route} is missing its live Go service link")

    request_attrs = ("action", "hx-get", "hx-post", "hx-put", "hx-patch", "hx-delete")
    for route, page in pages.items():
        for _, attrs in page.tags:
            for name in request_attrs:
                target = attrs.get(name)
                if target and urlsplit(target).path.startswith("/htmgo-site/"):
                    raise SystemExit(f"/{route} contains a Go partial request to {target}")


if __name__ == "__main__":
    verify(Path(sys.argv[1]))
