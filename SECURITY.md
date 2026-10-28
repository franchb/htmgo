# Security Policy

## Supported versions

Security fixes are released for the latest v2.x release of each module
only. Upgrade to the newest tag before reporting.

| Module                                             | Supported     |
| -------------------------------------------------- | ------------- |
| `github.com/franchb/htmgo/framework/v2`            | latest v2.x   |
| `github.com/franchb/htmgo/framework-ui/v2`         | latest v2.x   |
| `github.com/franchb/htmgo/extensions/websocket/v2` | latest v2.x   |
| `github.com/franchb/htmgo/tools/html-to-htmgo/v2`  | latest v2.x   |
| `github.com/franchb/htmgo/cli/htmgo/v2` (CLI)      | latest v2.x   |
| v1.x and the upstream `maddalax/htmgo` modules     | not supported |

## Reporting a vulnerability

Please **do not** open a public issue, pull request or discussion for a
security problem.

Report it privately through GitHub:
[Security → Report a vulnerability](https://github.com/franchb/htmgo/security/advisories/new).
Include:

- the affected module(s) and version(s),
- a description of the issue and its impact,
- steps or a minimal program to reproduce it,
- any known workaround or suggested fix.

You can expect an acknowledgement within 7 days. Once the issue is
confirmed, a fix is prepared in a private fork, released as a new patch
version, and published as a GitHub Security Advisory crediting you
unless you ask otherwise.

## Dependencies

Vulnerabilities in dependencies (for example `golang.org/x/net`) are
tracked with `govulncheck`, which runs in CI for every library module.
If govulncheck reports a vulnerability as reachable from htmgo code,
please report it as above.
