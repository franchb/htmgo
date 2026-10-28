package h

import "github.com/franchb/htmgo/framework/v2/hx"

// ValidationOpt configures Validation.
type ValidationOpt func(*validationConfig)

type validationConfig struct {
	required bool
	pattern  string
	message  string
	path     string
	trigger  string
	target   string
	swap     string
}

// Validation returns the attributes that validate a form field. Put it inside
// an input, select or textarea:
//
//	h.Input("email",
//		h.Name("email"),
//		h.Validation(
//			h.ValidateRequired(),
//			h.ValidatePattern(`[^@\s]+@[^@\s]+`, "Enter an email address"),
//			h.ValidatePartial(partials.ValidateEmail),
//		),
//	),
//	h.Div(), // the server's message is swapped in here
//
// Required and pattern checks are native HTML constraint validation: they run
// in the browser with no JavaScript and block form submission. Server
// validation posts the field to a partial when it changes (hx-post on the
// field, so the request carries the field's name and value) and swaps the
// partial's response into the next <div> by default. An empty response clears
// the previous message. Server validation is advisory: always validate again
// in the handler that processes the submitted form.
func Validation(opts ...ValidationOpt) *ChildList {
	cfg := validationConfig{
		trigger: string(hx.ChangeEvent),
		target:  "next div",
		swap:    "innerHTML",
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	attrs := make([]Ren, 0, 8)
	if cfg.required {
		attrs = append(attrs, Required())
	}
	if cfg.pattern != "" {
		attrs = append(attrs, Pattern(cfg.pattern))
		if cfg.message != "" {
			// Browsers show the title as part of the pattern-mismatch message.
			attrs = append(attrs, Attribute("title", cfg.message))
		}
	}
	if cfg.path != "" {
		attrs = append(attrs,
			Post(cfg.path, cfg.trigger),
			Attribute(hx.TargetAttr, cfg.target),
			Attribute(hx.SwapAttr, cfg.swap),
			// A newer edit replaces an in-flight validation request, so a
			// slow response for an old value never overwrites a newer one.
			Attribute(hx.SyncAttr, "this:replace"),
		)
	}
	return Children(attrs...)
}

// ValidateRequired marks the field as required (the HTML required attribute).
func ValidateRequired() ValidationOpt {
	return func(c *validationConfig) { c.required = true }
}

// ValidatePattern sets the HTML pattern attribute: the whole value must match
// the regular expression (JavaScript syntax, the browser anchors it with ^ and $).
// message, if not empty, is set as the title, which browsers include in their
// mismatch message.
func ValidatePattern(pattern string, message string) ValidationOpt {
	return func(c *validationConfig) {
		c.pattern = pattern
		c.message = message
	}
}

// ValidatePartial validates the field on the server with the given partial.
func ValidatePartial(partial PartialFunc) ValidationOpt {
	return ValidatePath(GetPartialPath(partial))
}

// ValidatePath validates the field on the server by posting it to path.
func ValidatePath(path string) ValidationOpt {
	return func(c *validationConfig) { c.path = path }
}

// ValidateOn changes the event that triggers server validation
// (default hx.ChangeEvent). For example hx.BlurEvent, or hx.InputEvent with
// your own delay modifier.
func ValidateOn(event hx.Event) ValidationOpt {
	return func(c *validationConfig) { c.trigger = string(event) }
}

// ValidateTarget changes where the server's response is swapped
// (default "next div"). Any htmx target selector works, e.g. "#email-error".
func ValidateTarget(selector string) ValidationOpt {
	return func(c *validationConfig) { c.target = selector }
}

// ValidateSwap changes the hx-swap value for the server response
// (default "innerHTML").
func ValidateSwap(swap string) ValidationOpt {
	return func(c *validationConfig) { c.swap = swap }
}
