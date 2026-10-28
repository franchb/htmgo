package h

import (
	"testing"

	"github.com/franchb/htmgo/framework/v2/hx"
	"github.com/stretchr/testify/assert"
)

func validateTestPartial(ctx *RequestContext) *Partial { return NewPartial(Fragment()) }

func TestValidation_Empty(t *testing.T) {
	assert.Equal(t, `<input type="text"/>`, Render(Input("text", Validation())))
}

func TestValidation_RequiredAndPattern(t *testing.T) {
	out := Render(Input("text", Validation(
		ValidateRequired(),
		ValidatePattern(`[a-z]+`, "Lowercase letters only"),
	)))
	assert.Equal(t, `<input type="text" required pattern="[a-z]+" title="Lowercase letters only"/>`, out)
}

func TestValidation_PatternWithoutMessageHasNoTitle(t *testing.T) {
	out := Render(Input("text", Validation(ValidatePattern(`\d+`, ""))))
	assert.Equal(t, `<input type="text" pattern="\d+"/>`, out)
}

func TestValidation_ServerPath(t *testing.T) {
	out := Render(Input("text", Name("email"), Validation(ValidatePath("/validate/email"))))
	assert.Equal(t,
		`<input type="text" name="email" hx-post="/validate/email" hx-trigger="change" hx-target="next div" hx-swap="innerHTML" hx-sync="this:replace"/>`,
		out)
}

func TestValidation_ServerPartialAndOverrides(t *testing.T) {
	out := Render(Input("text", Validation(
		ValidatePartial(validateTestPartial),
		ValidateOn(hx.BlurEvent),
		ValidateTarget("#email-error"),
		ValidateSwap("outerHTML"),
	)))
	assert.Contains(t, out, `hx-post="`+GetPartialPath(validateTestPartial)+`"`)
	assert.Contains(t, out, `hx-trigger="blur"`)
	assert.Contains(t, out, `hx-target="#email-error"`)
	assert.Contains(t, out, `hx-swap="outerHTML"`)
}

func TestValidation_PatternIsEscaped(t *testing.T) {
	out := Render(Input("text", Validation(ValidatePattern(`"><script>`, ""))))
	assert.NotContains(t, out, "<script>")
	assert.Contains(t, out, `pattern="&#34;&gt;&lt;script&gt;"`)
}
