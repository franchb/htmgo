"""Compile the published Go examples against the framework in this checkout."""

import os
import re
import shutil
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FRAMEWORK = "github.com/franchb/htmgo/framework/v2"
SITE = ROOT / "htmgo-site/pages/docs"
SKILL = ROOT / ".claude/skills/htmgo-guidance/SKILL.md"


def markdown_example(path, signature):
    blocks = re.findall(r"^```go\n(.*?)^```", path.read_text(), re.M | re.S)
    matches = [block for block in blocks if signature in block]
    if len(matches) != 1:
        raise ValueError(f"{path}: expected one Go block containing {signature!r}")
    return matches[0]


def site_example(path, name):
    match = re.search(rf"^(?:var|const) {name} = `([^`]*)`", path.read_text(), re.M)
    if match is None:
        raise ValueError(f"{path}: missing Go example {name!r}")
    return match[1].strip()


def declarations(source, *imports):
    paths = "\n".join(f'"{path}"' for path in imports)
    return f"package example\nimport (\n{paths}\n)\n{source}\n"


def write_examples(app):
    def write(relative, source):
        dest = app / relative
        dest.parent.mkdir(parents=True, exist_ok=True)
        dest.write_text(source)

    h = FRAMEWORK + "/h"
    js = FRAMEWORK + "/js"
    hx = FRAMEWORK + "/hx"
    for name, path in [
        ("readme", ROOT / "README.md"),
        ("site_intro", ROOT / "htmgo-site/md/index.md"),
    ]:
        source = markdown_example(path, "func IndexPage")
        write(f"{name}/page.go", declarations(source, "time", h))

    for handler, path in [
        ("CounterPage", "pages/counter.go"),
        ("PostIncrement", "partials/counter/increment.go"),
        ("SearchBox", "pages/search.go"),
        ("GetSearch", "partials/search/search.go"),
    ]:
        write(path, markdown_example(SKILL, f"func {handler}("))

    # Each standalone example gets its own package, so reused handler names are safe.
    for filename, name, imports in [
        ("introduction.go", "IntroSnippet", ("time", h)),
        ("core-concepts/components.go", "ComponentExampleSnippet", (h,)),
        ("control/loops.go", "ListExample", (h,)),
        ("control/loops.go", "MapExample", (h,)),
        ("htmx-extensions/alpine-compat.go", "PopoverExample", (FRAMEWORK + "/ax", h, hx)),
        ("htmx-extensions/mutation-error.go", "MutationErrorExample", (h, js)),
        ("interactivity/events.go", "EventsExample1", (h, js)),
        ("interactivity/events.go", "EvalArbitraryJavascriptSnippet", (h, js)),
        ("interactivity/eval-commands.go", "EvalCommandsSnippet", (h, js)),
        ("interactivity/eval-commands.go", "EvalCommandsSnippetWithSelector", (h, js)),
        ("performance/caching-globally.go", "CachedGloballyExample", ("time", h)),
        ("performance/caching-per-key.go", "CachedPerKeyExample", ("time", h)),
        ("pushing-data/sse.go", "SseConnectAttribute", (h,)),
        ("pushing-data/sse.go", "SseErrorHandlingExample", (h, js)),
        ("pushing-data/sse.go", "SseClearInputExample", (h, js)),
    ]:
        source = site_example(SITE / filename, name)
        write(f"site/{name}/example.go", declarations(source, *imports))

    pages = SITE / "core-concepts/pages.go"
    swaps = SITE / "interactivity/swapping.go"
    counter = [
        site_example(pages, "RootPageSnippet"),
        site_example(swaps, "SwapExample"),
        site_example(swaps, "SwapGetPartialPathWithQsExample"),
        site_example(swaps, "SwapGetPartialPathExampleOnBlur"),
    ]
    write("site/swap/example.go", declarations("\n".join(counter), "strconv", h, hx))

    cache_readme = ROOT / "framework/h/cache/README.md"
    cache_page = SITE / "performance/pluggable-caches.go"
    cache_examples = [
        ("cache_readme", markdown_example, cache_readme, "func NewProfiles", "type Store["),
        ("cache_page", site_example, cache_page, "PerComponentExample", "CacheStoreInterface"),
    ]
    # Both assignments must compile so omissions and additions to the interface are caught.
    assertions = """
var _ cache.Store[int, string] = Store[int, string](nil)
var _ Store[int, string] = cache.Store[int, string](nil)
"""
    for name, extract, path, profile, contract in cache_examples:
        write(f"{name}/profile.go", extract(path, profile))
        source = extract(path, contract) + assertions
        write(f"{name}/contract/store.go", declarations(source, "time", FRAMEWORK + "/h/cache"))


if __name__ == "__main__":
    with tempfile.TemporaryDirectory() as temp:
        app = Path(temp)
        framework = ROOT / "framework"
        version = re.search(r"^go ([\d.]+)$", (framework / "go.mod").read_text(), re.M)[1]
        (app / "go.mod").write_text(
            f"module example.com/htmgo-docs\n\ngo {version}\n"
            f"require {FRAMEWORK} v2.0.0\nreplace {FRAMEWORK} => \"{framework.as_posix()}\"\n"
        )
        shutil.copyfile(framework / "go.sum", app / "go.sum")
        write_examples(app)
        subprocess.run(
            ["go", "test", "-mod=mod", "./..."],
            cwd=app, env={**os.environ, "GOWORK": "off"}, check=True,
        )
