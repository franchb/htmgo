package h

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"html"
	"log/slog"
	"strings"
)

// CSPNoncePlaceholder is replaced with the per-request nonce in
// AppOpts.ContentSecurityPolicy, e.g. "script-src 'nonce-{nonce}' 'strict-dynamic'".
const CSPNoncePlaceholder = "{nonce}"

// newCSPNonce returns 128 bits of randomness, base64url-encoded without padding.
func newCSPNonce() string {
	b := make([]byte, 16)
	// crypto/rand.Read never returns an error since Go 1.24; it crashes the
	// program if the OS cannot supply randomness.
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// CSPNonce returns the Content-Security-Policy nonce for this request,
// generating it on first use. Every <script> tag in a full page rendered by
// htmgo for this request carries the same nonce. Use it in your own
// middleware if you set the CSP header yourself.
func (c *RequestContext) CSPNonce() string {
	if c.cspNonce == "" {
		c.cspNonce = newCSPNonce()
	}
	return c.cspNonce
}

// WithScriptNonce adds nonce="..." to every <script> tag the render emits:
// h.Script, h.ScriptWithVersion, h.UnsafeRawScript and the helper scripts
// generated for complex JS commands. An empty nonce is a no-op.
func WithScriptNonce(nonce string) RenderOpt {
	return func(context *RenderContext, opt *RenderOptions) {
		context.nonce = nonce
	}
}

func writeNonceAttr(context *RenderContext) {
	if context.nonce == "" {
		return
	}
	context.builder.WriteString(` nonce="`)
	context.builder.WriteString(html.EscapeString(context.nonce))
	context.builder.WriteString(`"`)
}

// htmx 4 reads its config from a single <meta name="htmx-config"> tag, so we
// only add ours when the page has none.
const htmxConfigMetaName = `name="htmx-config"`

// injectHtmxNonceConfig tells htmx to put the page nonce on scripts it
// inserts during swaps (inlineScriptNonce) and on its own indicator <style>
// (inlineStyleNonce). Without this, scripts inside swapped partials are
// blocked by the page's CSP.
func injectHtmxNonceConfig(rendered string, nonce string) string {
	if strings.Contains(rendered, htmxConfigMetaName) {
		slog.Warn("page already has a <meta name=\"htmx-config\">; add inlineScriptNonce/inlineStyleNonce to it yourself for CSP nonces to work in swapped content")
		return rendered
	}
	cfg, _ := json.Marshal(map[string]string{
		"inlineScriptNonce": nonce,
		"inlineStyleNonce":  nonce,
	})
	meta := `<meta name="htmx-config" content="` + html.EscapeString(string(cfg)) + `">`
	loc := headOpenRe.FindStringIndex(rendered)
	if loc == nil {
		return rendered
	}
	return rendered[:loc[1]] + meta + rendered[loc[1]:]
}

// Cached components are rendered once and reused across requests, so they
// can't contain a real per-request nonce. They are rendered with this token
// instead, and writeCached swaps in the current request's nonce (or removes
// the attribute when CSP is off).
const cachedNonceToken = "__htmgo_csp_nonce__"

var cachedNonceAttr = ` nonce="` + cachedNonceToken + `"`

func withCachedNonceToken() RenderOpt {
	return WithScriptNonce(cachedNonceToken)
}

func writeCached(context *RenderContext, cached string) {
	if !strings.Contains(cached, cachedNonceAttr) {
		context.builder.WriteString(cached)
		return
	}
	replacement := ""
	if context.nonce != "" {
		replacement = ` nonce="` + html.EscapeString(context.nonce) + `"`
	}
	context.builder.WriteString(strings.ReplaceAll(cached, cachedNonceAttr, replacement))
}
