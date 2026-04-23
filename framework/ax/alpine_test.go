package ax

import (
	"strings"
	"testing"

	"github.com/franchb/htmgo/framework/v2/h"
	"github.com/stretchr/testify/assert"
)

func renderAttr(attr h.Ren) string {
	return h.Render(h.Div(attr))
}

func TestAttributeConstants(t *testing.T) {
	t.Parallel()
	type c struct {
		got  Attribute
		want string
	}
	cases := []c{
		{DataAttr, "x-data"},
		{InitAttr, "x-init"},
		{ShowAttr, "x-show"},
		{BindAttr, "x-bind"},
		{OnAttr, "x-on"},
		{TextAttr, "x-text"},
		{HtmlAttr, "x-html"},
		{ModelAttr, "x-model"},
		{ModelableAttr, "x-modelable"},
		{CloakAttr, "x-cloak"},
		{RefAttr, "x-ref"},
		{IgnoreAttr, "x-ignore"},
		{TeleportAttr, "x-teleport"},
		{EffectAttr, "x-effect"},
		{IfAttr, "x-if"},
		{ForAttr, "x-for"},
		{IdAttr, "x-id"},
		{TransitionAttr, "x-transition"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, string(tc.got))
	}
}

func TestSimpleDirectives(t *testing.T) {
	t.Parallel()
	type c struct {
		name     string
		attr     h.Ren
		contains string
	}
	cases := []c{
		{"Data", Data("{ open: false }"), `x-data="{ open: false }"`},
		{"Init", Init("count = 0"), `x-init="count = 0"`},
		{"Show", Show("open"), `x-show="open"`},
		{"Text", Text("message"), `x-text="message"`},
		{"Html", Html("markup"), `x-html="markup"`},
		{"Model", Model("query"), `x-model="query"`},
		{"Effect", Effect("console.log(count)"), `x-effect="console.log(count)"`},
		{"Modelable", Modelable("value"), `x-modelable="value"`},
		{"If", If("visible"), `x-if="visible"`},
		{"For", For("item in items"), `x-for="item in items"`},
		{"Id", Id("['tab']"), `x-id="[&#39;tab&#39;]"`},
		{"Ref", Ref("input"), `x-ref="input"`},
		{"Teleport", Teleport("body"), `x-teleport="body"`},
		{"Transition", Transition(), `x-transition`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, strings.Contains(renderAttr(tc.attr), tc.contains),
				"rendered HTML %q did not contain %q", renderAttr(tc.attr), tc.contains)
		})
	}
}

func TestNoArgDirectives(t *testing.T) {
	t.Parallel()
	assert.Contains(t, renderAttr(Cloak()), `x-cloak`)
	assert.Contains(t, renderAttr(Ignore()), `x-ignore`)
}

func TestBindFamily(t *testing.T) {
	t.Parallel()
	type c struct {
		name     string
		attr     h.Ren
		contains string
	}
	cases := []c{
		{"Bind generic", Bind("data-foo", "bar"), `x-bind:data-foo="bar"`},
		{"BindClass", BindClass("{ active: isActive }"), `x-bind:class="{ active: isActive }"`},
		{"BindStyle", BindStyle("{ color: hex }"), `x-bind:style="{ color: hex }"`},
		{"BindHref", BindHref("url"), `x-bind:href="url"`},
		{"BindValue", BindValue("input"), `x-bind:value="input"`},
		{"BindDisabled", BindDisabled("locked"), `x-bind:disabled="locked"`},
		{"BindChecked", BindChecked("selected"), `x-bind:checked="selected"`},
		{"BindId", BindId("compId"), `x-bind:id="compId"`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, renderAttr(tc.attr), tc.contains)
		})
	}
}

func TestOnFamily(t *testing.T) {
	t.Parallel()
	type c struct {
		name     string
		attr     h.Ren
		contains string
	}
	cases := []c{
		{"On no modifiers", On("click", "count++"), `x-on:click="count++"`},
		{"On one modifier", On("click", "submit()", "prevent"), `x-on:click.prevent="submit()"`},
		{"On many modifiers", On("keydown", "close()", "escape", "prevent"), `x-on:keydown.escape.prevent="close()"`},
		{"OnClick plain", OnClick("count++"), `x-on:click="count++"`},
		{"OnClick with mod", OnClick("submit()", "prevent"), `x-on:click.prevent="submit()"`},
		{"OnSubmit", OnSubmit("save()"), `x-on:submit="save()"`},
		{"OnInput", OnInput("sync()"), `x-on:input="sync()"`},
		{"OnChange", OnChange("refresh()"), `x-on:change="refresh()"`},
		{"OnFocus", OnFocus("mark()"), `x-on:focus="mark()"`},
		{"OnBlur", OnBlur("unmark()"), `x-on:blur="unmark()"`},
		{"OnKeydown", OnKeydown("handle()"), `x-on:keydown="handle()"`},
		{"OnKeyup", OnKeyup("handle()"), `x-on:keyup="handle()"`},
		{"OnKeydown with modifiers", OnKeydown("handle()", "meta", "k", "prevent"), `x-on:keydown.meta.k.prevent="handle()"`},
		{"OnClickOutside", OnClickOutside("open = false"), `x-on:click.outside="open = false"`},
		{"OnKeydownEscape", OnKeydownEscape("open = false"), `x-on:keydown.escape="open = false"`},
		{"OnKeydownEnter", OnKeydownEnter("submit()"), `x-on:keydown.enter="submit()"`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, renderAttr(tc.attr), tc.contains)
		})
	}
}

func TestModelVariants(t *testing.T) {
	t.Parallel()
	type c struct {
		name     string
		attr     h.Ren
		contains string
	}
	cases := []c{
		{"ModelNumber", ModelNumber("age"), `x-model.number="age"`},
		{"ModelLazy", ModelLazy("title"), `x-model.lazy="title"`},
		{"ModelTrim", ModelTrim("name"), `x-model.trim="name"`},
		{"ModelFill", ModelFill("notes"), `x-model.fill="notes"`},
		{"ModelBoolean", ModelBoolean("checked"), `x-model.boolean="checked"`},
		{"ModelDebounce", ModelDebounce("query", "500ms"), `x-model.debounce.500ms="query"`},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, renderAttr(tc.attr), tc.contains)
		})
	}
}
