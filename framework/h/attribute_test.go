package h

import (
	"github.com/franchb/htmgo/framework/hx"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestAttributes(t *testing.T) {
	tests := []struct {
		name          string
		attribute     *AttributeR
		expectedKey   string
		expectedValue string
	}{
		{"NoSwap", NoSwap(), "hx-swap", "none"},
		{"Checked", Checked().(*AttributeR), "checked", ""},
		{"Id", Id("myID").(*AttributeR), "id", "myID"},
		{"Disabled", Disabled(), "disabled", ""},
		{"HxTarget", HxTarget("#myTarget").(*AttributeR), "hx-target", "#myTarget"},
		{"Name", Name("myName").(*AttributeR), "name", "myName"},
		{"HxConfirm", HxConfirm("Are you sure?").(*AttributeR), "hx-confirm", "Are you sure?"},
		{"Class", Class("class1", "class2"), "class", "class1 class2 "},
		{"ReadOnly", ReadOnly(), "readonly", ""},
		{"Required", Required(), "required", ""},
		{"Multiple", Multiple(), "multiple", ""},
		{"Selected", Selected(), "selected", ""},
		{"MaxLength", MaxLength(10), "maxlength", "10"},
		{"MinLength", MinLength(5), "minlength", "5"},
		{"Size", Size(3), "size", "3"},
		{"Width", Width(100), "width", "100"},
		{"Height", Height(200), "height", "200"},
		{"Download", Download(true), "download", "true"},
		{"Rel", Rel("noopener"), "rel", "noopener"},
		{"Pattern", Pattern("[A-Za-z]+"), "pattern", "[A-Za-z]+"},
		{"Action", Action("/submit"), "action", "/submit"},
		{"Method", Method("POST"), "method", "POST"},
		{"Enctype", Enctype("multipart/form-data"), "enctype", "multipart/form-data"},
		{"AutoComplete", AutoComplete("on"), "autocomplete", "on"},
		{"AutoFocus", AutoFocus(), "autofocus", ""},
		{"NoValidate", NoValidate(), "novalidate", ""},
		{"Step", Step("0.1"), "step", "0.1"},
		{"Max", Max("100"), "max", "100"},
		{"Min", Min("0"), "min", "0"},
		{"Cols", Cols(30), "cols", "30"},
		{"Rows", Rows(10), "rows", "10"},
		{"Wrap", Wrap("soft"), "wrap", "soft"},
		{"Role", Role("button"), "role", "button"},
		{"AriaLabel", AriaLabel("Close Dialog"), "aria-label", "Close Dialog"},
		{"AriaHidden", AriaHidden(true), "aria-hidden", "true"},
		{"TabIndex", TabIndex(1), "tabindex", "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedKey, tt.attribute.Name)
			assert.Equal(t, tt.expectedValue, tt.attribute.Value)
		})
	}
}

func TestClassF(t *testing.T) {
	attribute := ClassF("class-%d", 123)
	assert.Equal(t, "class", attribute.Name)
	assert.Equal(t, "class-123", attribute.Value)
}

func TestClassX(t *testing.T) {
	classMap := ClassMap{"visible": true, "hidden": false}
	attribute := ClassX("base", classMap).(*AttributeR)
	assert.Equal(t, "class", attribute.Name)
	assert.Equal(t, "base visible ", attribute.Value)
}

func TestJoinAttributes(t *testing.T) {
	attr1 := Attribute("data-attr", "one")
	attr2 := Attribute("data-attr", "two")
	joined := JoinAttributes(", ", attr1, attr2)
	assert.Equal(t, "data-attr", joined.Name)
	assert.Equal(t, "one, two", joined.Value)
}

func TestTarget(t *testing.T) {
	attr := Target("_blank")
	assert.Equal(t, "target", attr.(*AttributeR).Name)
	assert.Equal(t, "_blank", attr.(*AttributeR).Value)
}

func TestD(t *testing.T) {
	attr := D("M10 10 H 90 V 90 H 10 Z")
	assert.Equal(t, "d", attr.(*AttributeR).Name)
	assert.Equal(t, "M10 10 H 90 V 90 H 10 Z", attr.(*AttributeR).Value)
}

func TestHxTrigger(t *testing.T) {
	trigger := hx.NewTrigger(hx.OnClick()) // This assumes hx.NewTrigger is a correct call
	attr := HxTrigger(hx.OnClick())
	assert.Equal(t, "hx-trigger", attr.Name)
	assert.Equal(t, trigger.ToString(), attr.Value)
}

func TestHxTriggerClick(t *testing.T) {
	attr := HxTriggerClick() // Assuming no options for simplicity
	assert.Equal(t, "hx-trigger", attr.Name)
	assert.Equal(t, "click", attr.Value)
}


func TestHxInclude(t *testing.T) {
	attr := HxInclude(".include-selector")
	assert.Equal(t, "hx-include", attr.(*AttributeR).Name)
	assert.Equal(t, ".include-selector", attr.(*AttributeR).Value)
}

func TestHxIndicator(t *testing.T) {
	attr := HxIndicator("#my-indicator")
	assert.Equal(t, "hx-indicator", attr.Name)
	assert.Equal(t, "#my-indicator", attr.Value)
}

func TestHidden(t *testing.T) {
	attr := Hidden()
	assert.Equal(t, "style", attr.(*AttributeR).Name)
	assert.Equal(t, "display:none", attr.(*AttributeR).Value)
}

func TestControls(t *testing.T) {
	attr := Controls()
	assert.Equal(t, "controls", attr.(*AttributeR).Name)
	assert.Equal(t, "", attr.(*AttributeR).Value)
}

func TestPlaceholder(t *testing.T) {
	attr := Placeholder("Enter text")
	assert.Equal(t, "placeholder", attr.(*AttributeR).Name)
	assert.Equal(t, "Enter text", attr.(*AttributeR).Value)
}

func TestBoost(t *testing.T) {
	attr := Boost()
	assert.Equal(t, "hx-boost", attr.(*AttributeR).Name)
	assert.Equal(t, "true", attr.(*AttributeR).Value)
}

func TestHxInheritedAttributes(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		attr  Ren
		key   string
		value string
	}{
		{"HxTargetInherited", HxTargetInherited("#out"), "hx-target:inherited", "#out"},
		{"HxIncludeInherited", HxIncludeInherited("closest form"), "hx-include:inherited", "closest form"},
		{"HxSwapInherited", HxSwapInherited("outerHTML"), "hx-swap:inherited", "outerHTML"},
		{"HxBoostInherited", HxBoostInherited("true"), "hx-boost:inherited", "true"},
		{"HxConfirmInherited", HxConfirmInherited("Sure?"), "hx-confirm:inherited", "Sure?"},
		{"HxHeadersInherited", HxHeadersInherited(`{"X-Token":"abc"}`), "hx-headers:inherited", `{"X-Token":"abc"}`},
		{"HxIndicatorInherited", HxIndicatorInherited("#spinner"), "hx-indicator:inherited", "#spinner"},
		{"HxSyncInherited", HxSyncInherited("this:drop"), "hx-sync:inherited", "this:drop"},
		{"HxConfigInherited", HxConfigInherited(`{"timeout":5000}`), "hx-config:inherited", `{"timeout":5000}`},
		{"HxEncodingInherited", HxEncodingInherited("multipart/form-data"), "hx-encoding:inherited", "multipart/form-data"},
		{"HxValidateInherited", HxValidateInherited("true"), "hx-validate:inherited", "true"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ar, ok := tc.attr.(*AttributeR)
			assert.True(t, ok, "expected *AttributeR from %s", tc.name)
			assert.Equal(t, tc.key, ar.Name)
			assert.Equal(t, tc.value, ar.Value)
		})
	}
}
